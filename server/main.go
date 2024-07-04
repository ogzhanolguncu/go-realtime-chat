package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const port = 7007

func main() {
	server := newServer()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Chat server started on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}

		name := handleUsernameSet(conn)
		log.Printf("Recently joined user's name: %s\n", name)
		server.ConnectionMap[conn] = &ConnectionInfo{Connection: conn, OwnerName: name}

		connectedUsers := len(server.ConnectionMap)
		log.Printf("Connection from %s\n", conn.RemoteAddr().String())
		log.Printf("Connected users: %d\n", connectedUsers)

		go server.sendSystemNotice(name, conn, "joined")
		go server.handleConnection(conn)
	}
}

func handleUsernameSet(conn net.Conn) string {
	conn.Write([]byte("USERNAME_REQUIRED\n"))
	connReader := bufio.NewReader(conn)
	var name string

	for {
		data, err := connReader.ReadString('\n')

		if err != nil {
			break
		}

		name = strings.TrimSuffix(data, "\n")
		if len(name) < 2 {
			conn.Write([]byte(encodeSystemMessage("Username cannot be empty or less than two characters", "fail")))
		} else {
			conn.Write([]byte(encodeSystemMessage(fmt.Sprintf("Username successfully set to %s", name), "success")))
			break
		}
	}
	return name
}
