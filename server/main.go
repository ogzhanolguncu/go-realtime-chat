package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

const port = 7007

func main() {
	s := newServer()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Chat server started on port %d\n", port)

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go s.handleNewConnection(c)
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
			conn.Write([]byte(protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeSYS, Content: "Username cannot be empty or less than two characters", SysStatus: "fail"})))
		} else {
			conn.Write([]byte(protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeSYS, Content: fmt.Sprintf("Username successfully set to => %s", name), SysStatus: "success"})))
			break
		}
	}
	return name
}
