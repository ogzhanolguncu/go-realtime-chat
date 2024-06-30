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
	Color      string
}

type TCPServer struct {
	ConnectionMap map[net.Conn]*ConnectionInfo
	ConnLock      sync.Mutex
	Colors        []string
}

func NewTCPServer() *TCPServer {
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

func (s *TCPServer) HandleConnection(c net.Conn) {
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

		messagePayload, err := RequestType.ParseMessage(rawMessage)
		if err != nil {
			// Write back to client that his message is malformed
			c.Write([]byte(err.Error()))
		}

		switch messagePayload.MessageType {
		//If messageType is group message call broadcast message
		case RequestType.GROUP_MESSAGE:
			s.BroadcastMessage(messagePayload, c)
		//If client requests a quit we break out of CONNECTION_LOOP
		case RequestType.QUIT:
			break CONNECTION_LOOP
		}

	}
	log.Printf("Connection closed for %s\n", c.RemoteAddr().String())
}

func (s *TCPServer) BroadcastMessage(msgPayload RequestType.Message, sender net.Conn) {
	s.ConnLock.Lock()
	defer s.ConnLock.Unlock()

	fmtMsg := fmt.Sprintf("%s: %s\n", msgPayload.Name, msgPayload.Message)
	senderInfo := s.ConnectionMap[sender]

	coloredMessage := fmt.Sprintf("%s%s\033[0m\n", senderInfo.Color, fmtMsg)

	for conn := range s.ConnectionMap {
		if conn != sender {
			_, err := conn.Write([]byte(coloredMessage))
			if err != nil {
				log.Println("Error broadcasting message:", err)
			}
		}
	}
}

func (s *TCPServer) GetColorForConnection() string {
	connCount := len(s.ConnectionMap)
	colorIndex := connCount % len(s.Colors)
	return s.Colors[colorIndex]
}

func main() {
	server := NewTCPServer()

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
		color := server.GetColorForConnection()
		server.ConnectionMap[conn] = &ConnectionInfo{Connection: conn, Color: color}
		server.ConnLock.Unlock()

		go server.HandleConnection(conn)
	}

}
