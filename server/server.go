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

func (s *TCPServer) sendWhisper(msgPayload Message, sender net.Conn) {
	fmtedMsg := fmt.Sprintf("Whisper from %s: %s\n", msgPayload.MessageSender, msgPayload.MessageContent)
	recipientConn, found := s.findConnectionByOwnerName(msgPayload.MessageRecipient)

	if !found || recipientConn == nil {
		_, err := sender.Write([]byte("Recipient not found or connection lost\n"))
		if err != nil {
			log.Println("Error sending recipient not found message:", err)
		}
		return
	}

	_, err := recipientConn.Write([]byte(fmtedMsg))
	if err != nil {
		log.Println("Error sending whisper:", err)
	}
}

func (s *TCPServer) broadcastMessage(msgPayload Message, sender net.Conn) {
	s.ConnLock.Lock()
	defer s.ConnLock.Unlock()

	fmtedMsg := fmt.Sprintf("%s: %s\n", msgPayload.MessageSender, msgPayload.MessageContent)

	for conn := range s.ConnectionMap {
		if conn != sender {
			_, err := conn.Write([]byte(fmtedMsg))
			if err != nil {
				log.Println("Error broadcasting message:", err)
			}
		}
	}
}

func (s *TCPServer) sendSystemNotice(senderName string, sender net.Conn, action string) {
	msg := fmt.Sprintf("System: %s has %s the chat.\n", senderName, action)

	for conn := range s.ConnectionMap {
		if conn != sender {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Println("Error sending system notice:", err)
			}
		}
	}
}

func (s *TCPServer) findConnectionByOwnerName(ownerName string) (net.Conn, bool) {
	for conn, info := range s.ConnectionMap {
		if info.OwnerName == ownerName {
			return conn, true
		}
	}

	return nil, false
}
