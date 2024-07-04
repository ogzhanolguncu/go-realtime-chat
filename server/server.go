package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// ConnectionInfo holds connection-related information.
type ConnectionInfo struct {
	Connection net.Conn
	OwnerName  string
}

type TCPServer struct {
	ConnectionMap map[net.Conn]*ConnectionInfo
	ConnLock      sync.Mutex
}

func newServer() *TCPServer {
	return &TCPServer{
		ConnectionMap: make(map[net.Conn]*ConnectionInfo),
		ConnLock:      sync.Mutex{},
	}
}

func (s *TCPServer) handleConnection(c net.Conn) {
	defer func() {
		s.ConnLock.Lock()
		ownerName := s.ConnectionMap[c].OwnerName
		delete(s.ConnectionMap, c)
		defer s.ConnLock.Unlock()
		s.sendSystemNotice(ownerName, nil, "left")
		c.Close()
	}()

	connReader := bufio.NewReader(c)

CONNECTION_LOOP:
	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from %s: %v\n", c.RemoteAddr().String(), err)
			break
		}

		rawMessage := strings.TrimSpace(data)
		log.Printf("Message from %s: %s\n", c.RemoteAddr().String(), rawMessage)

		msgPayload, err := parseMessage(rawMessage)
		if err != nil {
			// Write back to client that their message is malformed
			c.Write([]byte(err.Error()))
		}

		switch msgPayload.MessageType {
		case GROUP_MESSAGE:
			s.broadcastMessage(msgPayload, c)
		case WHISPER:
			s.sendWhisper(msgPayload, c)
		case QUIT:
			break CONNECTION_LOOP // Exit the loop if QUIT message received
		default:
			log.Printf("Unknown message type received from %s\n", c.RemoteAddr().String())
		}
	}
	log.Printf("Connection closed for %s\n", c.RemoteAddr().String())
}

// sendWhisper looks up the recipient's connection in the connectionList. If found, it sends a whisper message to the recipient.
// If not found, it sends a system failure message back to the sender.
func (s *TCPServer) sendWhisper(msgPayload Message, sender net.Conn) {
	// Look up the recipient's connection by their name in the connectionList
	recipientConn, found := s.findConnectionByOwnerName(msgPayload.MessageRecipient)

	// If the recipient is not found or their connection is lost, send a system failure message to the sender
	if !found || recipientConn == nil {
		// Encode and send a system message indicating the recipient was not found or the connection was lost
		_, err := sender.Write([]byte(encodeSystemMessage("Recipient not found or connection lost", "fail")))
		if err != nil {
			// Log any error that occurs while sending the system message
			log.Println("Error sending recipient not found message:", err)
		}
		return
	}

	// If the recipient's connection is found, send the whisper message to the recipient
	_, err := recipientConn.Write([]byte(encodeWhisperMessage(msgPayload.MessageContent, msgPayload.MessageSender)))
	if err != nil {
		log.Println("Error sending whisper:", err)
	}
}

// broadcastMessage sends a message to all connections except the sender
func (s *TCPServer) broadcastMessage(msgPayload Message, sender net.Conn) {
	// Iterate over all connections in the connection map
	for conn := range s.ConnectionMap {
		// Skip the sender's connection
		if conn != sender {
			// Send the message to the current connection
			_, err := conn.Write([]byte(encodeGeneralMessage(msgPayload.MessageContent, msgPayload.MessageSender)))
			if err != nil {
				log.Println("Error broadcasting message:", err)
			}
		}
	}
}

// sendSystemNotice sends a system notice to all connections except the sender.
// The notice informs about an action performed by the sender (e.g., joining or leaving the chat).
func (s *TCPServer) sendSystemNotice(senderName string, sender net.Conn, action string) {
	// Iterate over all connections in the connection map
	for conn := range s.ConnectionMap {
		// Skip the sender's connection
		if conn != sender {
			_, err := conn.Write([]byte(encodeSystemMessage(fmt.Sprintf("%s has %s the chat.", senderName, action), "success")))
			if err != nil {
				log.Println("Error sending system notice:", err)
			}
		}
	}
}

// findConnectionByOwnerName searches for a connection by the owner's name in the connection map.
// It returns the connection and true if found, otherwise returns nil and false.
func (s *TCPServer) findConnectionByOwnerName(ownerName string) (net.Conn, bool) {
	// Iterate over all connections in the connection map
	for conn, info := range s.ConnectionMap {
		// Check if the current connection's owner name matches the specified owner name
		if info.OwnerName == ownerName {
			// Return the matching connection and true
			return conn, true
		}
	}
	return nil, false
}
