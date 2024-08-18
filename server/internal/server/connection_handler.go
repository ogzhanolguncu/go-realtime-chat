package server

import (
	"bufio"
	"log"
	"net"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/ogzhanolguncu/go-chat/server/internal/auth"
	"github.com/ogzhanolguncu/go-chat/server/internal/connection"
)

type ConnectionHandler struct {
	conn           net.Conn
	server         *TCPServer
	reader         *bufio.Reader
	encodeFn       func(payload protocol.Payload) string
	decodeFn       func(message string) (protocol.Payload, error)
	connectionInfo *connection.ConnectionInfo
}

func NewConnectionHandler(conn net.Conn, server *TCPServer) *ConnectionHandler {
	return &ConnectionHandler{
		conn:     conn,
		server:   server,
		reader:   bufio.NewReader(conn),
		encodeFn: server.encodeFn,
		decodeFn: server.decodeFn,
	}
}

// Main Connection Handling
// -----------------------------

// Handle manages the lifecycle of a client connection
func (ch *ConnectionHandler) Handle() {
	defer ch.conn.Close()

	if !ch.authenticate() {
		return
	}

	ch.server.OnClientJoin(ch.connectionInfo)
	defer ch.server.OnClientLeave(ch.connectionInfo)

	ch.handleMessages()
}

// handleMessages continuously reads and processes incoming messages
func (ch *ConnectionHandler) handleMessages() {
	for {
		message, err := ch.reader.ReadString('\n')
		if err != nil {
			log.Printf("Client left the chat '%s': %v\n", ch.connectionInfo.OwnerName, err)
			break
		}

		ch.server.OnMessageReceived(ch.connectionInfo, message)
	}
}

// Authentication
// -----------------------------

func (ch *ConnectionHandler) authenticate() bool {
	for {
		data, err := ch.reader.ReadString('\n')
		if err != nil {
			log.Printf("User closed connection during auth: %v", err)
			return false
		}

		payload, err := ch.decodeFn(data)
		if err != nil {
			log.Printf("Failed to decode auth data: %s. Error: %v", data, err)
			ch.sendAuthResponse("Invalid data format", "fail")
			continue
		}

		authenticated, err := ch.server.authManager.AuthenticateUser(payload.Username, payload.Password)
		if err != nil {
			ch.handleAuthError(err)
			continue
		}

		if authenticated {
			ch.connectionInfo = &connection.ConnectionInfo{
				Connection: ch.conn,
				OwnerName:  payload.Username,
			}
			ch.sendAuthResponse(payload.Username, "success")
			return true
		}

		ch.sendAuthResponse("Invalid username or password", "fail")
	}
}

func (ch *ConnectionHandler) sendAuthResponse(message, status string) {
	msg := ch.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeUSR,
		Username:    message,
		Status:      status,
	})
	ch.conn.Write([]byte(msg))
}

// handleAuthError processes authentication errors and sends appropriate mapped responses
func (ch *ConnectionHandler) handleAuthError(err error) {
	var message string
	switch err {
	case auth.ErrWeakPassword:
		message = "Password does not meet strength requirements"
	case auth.ErrInvalidUsername:
		message = "Username must be at least 2 characters long"
	default:
		message = "Authentication failed"
	}
	ch.sendAuthResponse(message, "fail")
}
