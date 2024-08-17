package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/stretchr/testify/assert"
)

// TestClient is a simplified version of the client for testing purposes
type TestClient struct {
	conn     net.Conn
	username string
	encodeFn func(payload protocol.Payload) string
	decodeFn func(message string) (protocol.Payload, error)
}

func NewTestClient(address string) (*TestClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TestClient{
		conn:     conn,
		encodeFn: protocol.InitEncodeProtocol(false),
		decodeFn: protocol.InitDecodeProtocol(false),
	}, nil
}

func (c *TestClient) Close() error {
	return c.conn.Close()
}

func (c *TestClient) SendMessage(payload protocol.Payload) error {
	msg := c.encodeFn(payload)
	_, err := c.conn.Write([]byte(msg))
	return err
}

func (c *TestClient) ReadMessage() (protocol.Payload, error) {
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return protocol.Payload{}, err
	}
	return c.decodeFn(msg)
}

func (c *TestClient) Authenticate(username, password string) error {
	err := c.SendMessage(protocol.Payload{
		MessageType: protocol.MessageTypeUSR,
		Username:    username,
		Password:    password,
	})
	if err != nil {
		return err
	}

	resp, err := c.ReadMessage()
	if err != nil {
		return err
	}
	if resp.MessageType != protocol.MessageTypeUSR || resp.Status != "success" {
		return fmt.Errorf("authentication failed: %+v", resp)
	}
	c.username = username
	return nil
}

func (c *TestClient) SendPublicMessage(content string) error {
	return c.SendMessage(protocol.Payload{
		MessageType: protocol.MessageTypeMSG,
		Sender:      c.username,
		Content:     content,
	})
}

func (c *TestClient) SendWhisper(recipient, content string) error {
	return c.SendMessage(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Sender:      c.username,
		Recipient:   recipient,
		Content:     content,
	})
}

func TestTwoClientsChat(t *testing.T) {
	// Start the server
	s, err := NewServer(0, ":memory:", false)
	assert.NoError(t, err)

	// Add test users
	err = s.authManager.AddUser("user1", "Password1!")
	assert.NoError(t, err)
	err = s.authManager.AddUser("user2", "Password2!")
	assert.NoError(t, err)

	go s.Start()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Get the address the server is listening on
	address := s.listener.Addr().String()

	// Create a context with timeout for the entire test
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	clientErr := make(chan error, 2)

	// Run client 1
	go func() {
		defer wg.Done()
		err := runClient(address, "user1", "Password1!")
		if err != nil {
			clientErr <- fmt.Errorf("client1 error: %v", err)
		}
	}()

	// Give client 1 a head start
	time.Sleep(100 * time.Millisecond)

	// Run client 2
	go func() {
		defer wg.Done()
		err := runClient(address, "user2", "Password2!")
		if err != nil {
			clientErr <- fmt.Errorf("client2 error: %v", err)
		}
	}()

	// Wait for both clients to finish or context to timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		t.Fatalf("Test timed out")
	case err := <-clientErr:
		t.Fatalf("Client error: %v", err)
	case <-done:
		// Test completed successfully
	}

	// Close all resources
	assert.NoError(t, s.Close())
}

func runClient(address, username, password string) error {
	log.Printf("Starting client for %s", username)
	client, err := NewTestClient(address)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Authenticate(username, password); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}
	log.Printf("%s authenticated successfully", username)

	if username == "user1" {
		// Wait for ACT_USRS message
		msg, err := client.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read ACT_USRS message: %v", err)
		}
		if msg.MessageType == protocol.MessageTypeACT_USRS {
			log.Printf("user1 received ACT_USRS message: %+v", msg)
		} else {
			return fmt.Errorf("expected ACT_USRS message, got: %+v", msg)
		}
		// Wait for SYS Notice
		msg, err = client.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read SYS_NOTICE message: %v", err)
		}
		if msg.MessageType == protocol.MessageTypeSYS {
			log.Printf("user1 received SYS_NOTICE message: %+v", msg)
		} else {
			return fmt.Errorf("expected SYS_NOTICE message, got: %+v", msg)
		}

		// Send a public message
		if err := client.SendPublicMessage("Hello, user2!"); err != nil {
			return fmt.Errorf("failed to send public message: %v", err)
		}
		log.Printf("user1 sent public message")

		// Wait for whisper from user2
		msg, err = client.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read whisper: %v", err)
		}
		if msg.MessageType != protocol.MessageTypeWSP || msg.Sender != "user2" {
			return fmt.Errorf("unexpected message received: %+v", msg)
		}
		log.Printf("user1 received whisper: %s", msg.Content)
	} else if username == "user2" {
		// Wait for ACT_USRS message
		msg, err := client.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read ACT_USRS message: %v", err)
		}
		if msg.MessageType == protocol.MessageTypeACT_USRS {
			log.Printf("user2 received ACT_USRS message: %+v", msg)
		} else {
			return fmt.Errorf("expected ACT_USRS message, got: %+v", msg)
		}

		// Wait for message from user1
		msg, err = client.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read message: %v", err)
		}
		if msg.MessageType != protocol.MessageTypeMSG || msg.Sender != "user1" {
			return fmt.Errorf("unexpected message received: %+v", msg)
		}
		log.Printf("user2 received message: %s", msg.Content)

		// Send whisper to user1
		if err := client.SendWhisper("user1", "This is a private message."); err != nil {
			return fmt.Errorf("failed to send whisper: %v", err)
		}
		log.Printf("user2 sent whisper")
	}

	return nil
}
