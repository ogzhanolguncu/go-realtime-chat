package internal

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/stretchr/testify/assert"
)

func TestClientSendMessages(t *testing.T) {
	client := &Client{name: "TestUser", lastWhispererFromGroupChat: "LastWhisperer"}
	outgoingChan := make(chan string, 10)
	done := make(chan struct{})

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal message", "Hello", "MSG|TestUser|5|Hello\r\n"},
		{"Whisper", "/whisper John Hey", "WSP|TestUser|John|3|Hey\r\n"},
		{"Reply", "/reply Hey", "WSP|TestUser|LastWhisperer|3|Hey\r\n"},
		{"Quit", "quit", ""},
	}

	r, w, _ := os.Pipe()
	os.Stdin = r
	go client.SendMessages(outgoingChan, done)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w.Write([]byte(tc.input + "\n"))
			if tc.input == "quit" {
				close(done)
				return
			}
			assert.Equal(t, tc.expected, <-outgoingChan)
		})
	}
}

func TestClientReadMessages(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		name:                       "TestUser",
		lastWhispererFromGroupChat: "LastWhisperer",
		conn:                       clientConn,
	}

	incomingChan := make(chan protocol.Payload, 10)
	errChan := make(chan error, 10)
	done := make(chan struct{})

	go client.ReadMessages(incomingChan, errChan, done)

	testCases := []struct {
		name     string
		input    string
		expected protocol.Payload
	}{
		{"Normal message", "MSG|John|5|Hello\r\n", protocol.Payload{MessageType: protocol.MessageTypeMSG, Sender: "John", Content: "Hello"}},
		{"Whisper", "WSP|Oz|TestUser|5|Hello\r\n", protocol.Payload{MessageType: protocol.MessageTypeWSP, Sender: "Oz", Content: "Hello", Recipient: "TestUser"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serverConn.Write([]byte(tc.input))
			select {
			case msg := <-incomingChan:
				assert.Equal(t, tc.expected, msg)
			case err := <-errChan:
				t.Fatalf("Unexpected error: %v", err)
			case <-time.After(time.Second):
				t.Fatal("Timeout waiting for message")
			}
		})
	}

	t.Run("Connection close", func(t *testing.T) {
		serverConn.Close()
		select {
		case err := <-errChan:
			assert.NotNil(t, err)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for connection close error")
		}
	})

	close(done)
}
