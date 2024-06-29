package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
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
	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from %s: %v\n", c.RemoteAddr().String(), err)
			break
		}
		message := strings.TrimSpace(string(data))

		log.Printf("Message from %s: %s\n", c.RemoteAddr().String(), message)
		if message == "quit" {
			break
		}
		s.BroadcastMessage(message, c)
	}
	log.Printf("Connection closed for %s\n", c.RemoteAddr().String())
}

func (s *TCPServer) BroadcastMessage(message string, sender net.Conn) {
	s.ConnLock.Lock()
	defer s.ConnLock.Unlock()

	senderInfo := s.ConnectionMap[sender]
	coloredMessage := fmt.Sprintf("%s%s\033[0m\n", senderInfo.Color, message)

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
