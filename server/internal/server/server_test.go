package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/ogzhanolguncu/go-chat/server/internal/auth"
	"github.com/ogzhanolguncu/go-chat/server/internal/block_user"
	"github.com/ogzhanolguncu/go-chat/server/internal/chat_history"
	"github.com/stretchr/testify/assert"
)

// ==================== TestClient Section ====================

// TestClient represents a simplified version of the client for testing purposes
type TestClient struct {
	conn     net.Conn
	username string
	encodeFn func(payload protocol.Payload) string
	decodeFn func(message string) (protocol.Payload, error)
}

// NewTestClient creates a new TestClient instance
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

// Close closes the client connection
func (c *TestClient) Close() error {
	return c.conn.Close()
}

// SendMessage sends a message to the server
func (c *TestClient) SendMessage(payload protocol.Payload) error {
	msg := c.encodeFn(payload)
	_, err := c.conn.Write([]byte(msg))
	return err
}

// ReadMessage reads a message from the server
func (c *TestClient) ReadMessage() (protocol.Payload, error) {
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return protocol.Payload{}, err
	}
	return c.decodeFn(msg)
}

// Authenticate performs client authentication
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

// SendPublicMessage sends a public message
func (c *TestClient) SendPublicMessage(content string) error {
	return c.SendMessage(protocol.Payload{
		MessageType: protocol.MessageTypeMSG,
		Sender:      c.username,
		Content:     content,
	})
}

// SendWhisper sends a private message to a specific user
func (c *TestClient) SendWhisper(recipient, content string) error {
	return c.SendMessage(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Sender:      c.username,
		Recipient:   recipient,
		Content:     content,
	})
}

// ==================== Test Setup Section ====================

// UserBehavior defines the behavior of a test user
type UserBehavior struct {
	PublicMessages   []string
	Whispers         map[string]string
	ConnectionDelay  time.Duration
	RequestHistory   bool
	ExpectedMessages []string
}

const dbPath = "./test_db.db"

// NewTestServer creates a new test server instance
func NewTestServer(t *testing.T) (*TCPServer, error) {
	authManager, err := auth.NewAuthManager(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	blockUserManager, err := block_user.NewBlockUserManager(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize block user manager: %w", err)
	}

	historyManager, err := chat_history.NewChatHistory(false, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat history manager: %w", err)
	}

	s, err := NewServer(0, dbPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	s.authManager = authManager
	s.blockUserManager = blockUserManager
	s.historyManager = historyManager

	return s, nil
}

// ==================== Main Test Function ====================

// TestMultipleClientsChat is the main test function
func TestMultipleClientsChat(t *testing.T) {
	s, err := NewTestServer(t)
	assert.NoError(t, err)

	userBehaviors := map[string]UserBehavior{
		"user1": {
			PublicMessages: []string{"Hello, everyone!", "How are you doing?"},
			Whispers: map[string]string{
				"user2": "Hey user2, this is a secret message.",
			},
		},
		"user2": {
			PublicMessages: []string{"Hi there!", "I'm doing great!"},
			Whispers: map[string]string{
				"user1": "Hello user1, here's a private reply.",
			},
		},
		"user3": {
			PublicMessages: []string{"Greetings!", "Nice to meet you all."},
			Whispers: map[string]string{
				"user1": "User1, can you help me with something?",
				"user2": "User2, let's plan a surprise for user1!",
			},
			RequestHistory:   true,
			ExpectedMessages: []string{"MSG: Hello", "MSG: Greetings!", "MSG: Hi there!", "MSG: Nice to meet you all.", "MSG: I'm doing great!", "MSG: How are you doing?", "WSP: User1", "WSP: User2"},
		},
	}

	// Add test users
	for username := range userBehaviors {
		err = s.authManager.AddUser(username, "Password123!")
		assert.NoError(t, err)
	}

	go s.Start()
	time.Sleep(100 * time.Millisecond)

	address := s.listener.Addr().String()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	clientErr := make(chan error, len(userBehaviors))

	clients := make(map[string]*TestClient)
	messageReceivedChan := make(chan struct{})
	var connectWg sync.WaitGroup
	connectWg.Add(len(userBehaviors))

	// Start client goroutines
	for username, behavior := range userBehaviors {
		wg.Add(1)
		go func(username string, behavior UserBehavior) {
			defer wg.Done()
			client, err := connectClient(address, username, "Password123!")
			if err != nil {
				clientErr <- fmt.Errorf("%s connection error: %v", username, err)
				return
			}
			clients[username] = client
			connectWg.Done()
			connectWg.Wait()

			err = runClientBehavior(ctx, t, client, behavior, messageReceivedChan)
			if err != nil {
				clientErr <- fmt.Errorf("%s error: %v", username, err)
			}
		}(username, behavior)
		time.Sleep(100 * time.Millisecond)
	}

	go func() {
		wg.Wait()
		close(messageReceivedChan)
	}()

	// Monitor test progress
	monitorTestProgress(t, ctx, clientErr, messageReceivedChan, clients, s)
}

// ==================== Helper Functions Section ====================

// connectClient creates and authenticates a new TestClient
func connectClient(address, username, password string) (*TestClient, error) {
	client, err := NewTestClient(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	if err := client.Authenticate(username, password); err != nil {
		client.Close()
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	return client, nil
}

// runClientBehavior executes the defined behavior for a client
func runClientBehavior(ctx context.Context, t *testing.T, client *TestClient, behavior UserBehavior, messageReceivedChan chan<- struct{}) error {
	done := make(chan struct{})
	go handleMessagesContiniuously(client, behavior.ExpectedMessages, done, t, messageReceivedChan)

	// Send public messages
	for _, msg := range behavior.PublicMessages {
		if err := client.SendPublicMessage(msg); err != nil {
			return fmt.Errorf("failed to send public message: %v", err)
		}
		t.Logf("%s sent public message: %s", client.username, msg)
		time.Sleep(50 * time.Millisecond)
	}

	// Send whispers
	for recipient, msg := range behavior.Whispers {
		if err := client.SendWhisper(recipient, msg); err != nil {
			return fmt.Errorf("failed to send whisper: %v", err)
		}
		t.Logf("%s sent whisper to %s: %s", client.username, recipient, msg)
		time.Sleep(50 * time.Millisecond)
	}

	// Request chat history if needed
	if behavior.RequestHistory {
		t.Logf("%s is about to request chat history", client.username)
		if err := client.SendMessage(protocol.Payload{
			MessageType: protocol.MessageTypeHSTRY,
			Sender:      client.username,
			Status:      "req",
		}); err != nil {
			return fmt.Errorf("failed to request chat history: %v", err)
		}
		t.Logf("%s requested chat history", client.username)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// handleMessagesContiniuously processes incoming messages for a client
func handleMessagesContiniuously(client *TestClient, expectedMessages []string, done chan<- struct{}, t *testing.T, messageReceivedChan chan<- struct{}) {
	defer close(done)
	var receivedHistory []string
	for {
		msg, err := client.ReadMessage()
		if err != nil {
			t.Logf("Error reading message for %s: %v", client.username, err)
			return
		}
		processMessage(t, client, msg, &receivedHistory, expectedMessages, messageReceivedChan)
	}
}

// processMessage handles different types of incoming messages
func processMessage(t *testing.T, client *TestClient, msg protocol.Payload, receivedHistory *[]string, expectedMessages []string, messageReceivedChan chan<- struct{}) {
	switch msg.MessageType {
	case protocol.MessageTypeACT_USRS:
		t.Logf("%s received ACT_USRS message: %+v", client.username, msg)
	case protocol.MessageTypeSYS:
		t.Logf("%s received SYS message: %+v", client.username, msg)
	case protocol.MessageTypeMSG:
		t.Logf("%s received public message: %s", client.username, msg.Content)
		messageReceivedChan <- struct{}{}
	case protocol.MessageTypeWSP:
		t.Logf("%s received whisper from %s: %s", client.username, msg.Sender, msg.Content)
		messageReceivedChan <- struct{}{}
	case protocol.MessageTypeHSTRY:
		processHistoryMessage(t, client, msg, receivedHistory, expectedMessages)
	default:
		t.Logf("%s received unknown message type: %+v", client.username, msg)
	}
}

// processHistoryMessage handles the chat history message
func processHistoryMessage(t *testing.T, client *TestClient, msg protocol.Payload, receivedHistory *[]string, expectedMessages []string) {
	for _, historyMsg := range msg.DecodedChatHistory {
		*receivedHistory = append(*receivedHistory, fmt.Sprintf("%s: %s", historyMsg.MessageType, historyMsg.Content))
	}
	if len(expectedMessages) > 0 {
		assert.Subset(t, *receivedHistory, expectedMessages, "Chat history doesn't contain all expected messages for %s", client.username)
	}
}

// monitorTestProgress watches for test completion or errors
func monitorTestProgress(t *testing.T, ctx context.Context, clientErr chan error, messageReceivedChan chan struct{}, clients map[string]*TestClient, s *TCPServer) {
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Test timed out")
		case err := <-clientErr:
			t.Fatalf("Client error: %v", err)
		case <-messageReceivedChan:
			// A message was received, reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(2 * time.Second)
		case <-timer.C:
			// No messages received for 2 seconds, consider the test complete
			t.Log("No new messages received for 2 seconds, test complete")
			cleanupTest(t, clients, s)
			return
		}
	}
}

// cleanupTest closes all client connections and cleans up the test environment
func cleanupTest(t *testing.T, clients map[string]*TestClient, s *TCPServer) {
	for _, client := range clients {
		client.Close()
	}

	assert.NoError(t, s.Close())
	assert.NoError(t, os.Remove(dbPath))
}
