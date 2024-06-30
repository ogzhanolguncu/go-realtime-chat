package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	RequestType "github.com/ogzhanolguncu/go-chat/server/requestType"
)

const port = 7007

// connectionInfo holds connection-related information and color per client.
type ConnectionInfo struct {
	Connection net.Conn
	OwnerName  string
	Color      string
}

type TCPServer struct {
	ConnectionMap map[net.Conn]*ConnectionInfo
	ConnLock      sync.Mutex
	Colors        []string
}

func newTCPServer() *TCPServer {
	return &TCPServer{
		ConnectionMap: make(map[net.Conn]*ConnectionInfo),
		ConnLock:      sync.Mutex{},
		Colors: []string{
			"\033[31m", // Red
			"\033[32m", // Green
			"\033[33m", // Yellow
			"\033[34m", // Blue
			"\033[35m", // Purple
		},
	}
}

func (s *TCPServer) handleConnection(c net.Conn) {
	defer func() {
		s.ConnLock.Lock()
		delete(s.ConnectionMap, c)
		s.ConnLock.Unlock()
		c.Close()
	}()

	connReader := bufio.NewReader(c)

	log.Printf("Connection from %s\n", c.RemoteAddr().String())

CONNECTION_LOOP:
	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from %s: %v\n", c.RemoteAddr().String(), err)
			break
		}

		//TODO: if messageType is whisper only send message to receiver
		//TODO: if message type is history simple echo back history to client

		rawMessage := strings.TrimSpace(string(data))
		log.Printf("Message from %s: %s\n", c.RemoteAddr().String(), rawMessage)

		msgPayload, err := RequestType.ParseMessage(rawMessage)

		// Handle case where c is not found in ConnectionMap
		updateConnectionOwner(s, c, msgPayload)

		if err != nil {
			// Write back to client that his message is malformed
			c.Write([]byte(err.Error()))
		}

		switch msgPayload.MessageType {
		case RequestType.GROUP_MESSAGE:
			s.broadcastMessage(msgPayload, c)
		case RequestType.WHISPER:
			s.sendWhisper(msgPayload, c)

		case RequestType.QUIT:
			break CONNECTION_LOOP // Exit the loop if QUIT message received
		default:
			log.Printf("Unknown message type received from %s\n", c.RemoteAddr().String())
		}
	}
	log.Printf("Connection closed for %s\n", c.RemoteAddr().String())
}

func (s *TCPServer) sendWhisper(msgPayload RequestType.Message, sender net.Conn) {
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

func (s *TCPServer) broadcastMessage(msgPayload RequestType.Message, sender net.Conn) {
	s.ConnLock.Lock()
	defer s.ConnLock.Unlock()

	fmtedMsg := fmt.Sprintf("%s: %s\n", msgPayload.MessageSender, msgPayload.MessageContent)
	senderInfo := s.ConnectionMap[sender]

	msg := fmt.Sprintf("%s%s\033[0m\n", senderInfo.Color, fmtedMsg)

	for conn := range s.ConnectionMap {
		if conn != sender {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Println("Error broadcasting message:", err)
			}
		}
	}
}

func (s *TCPServer) getColorForConnection() string {
	connCount := len(s.ConnectionMap)
	colorIndex := connCount % len(s.Colors)
	return s.Colors[colorIndex]
}

func main() {
	server := newTCPServer()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server started on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}

		server.ConnLock.Lock()
		color := server.getColorForConnection()
		server.ConnectionMap[conn] = &ConnectionInfo{Connection: conn, Color: color, OwnerName: "Unknown"}
		server.ConnLock.Unlock()

		go server.handleConnection(conn)
	}
}

func (s *TCPServer) findConnectionByOwnerName(ownerName string) (net.Conn, bool) {
	s.ConnLock.Lock()
	defer s.ConnLock.Unlock()

	for conn, info := range s.ConnectionMap {
		if info.OwnerName == ownerName {
			return conn, true
		}
	}

	return nil, false
}

func updateConnectionOwner(s *TCPServer, c net.Conn, msgPayload RequestType.Message) {
	s.ConnLock.Lock()
	if info, ok := s.ConnectionMap[c]; ok {
		info.OwnerName = msgPayload.MessageSender
	} else {
		log.Printf("Connection not found for %s\n", c.RemoteAddr().String())
	}
	s.ConnLock.Unlock()
}
