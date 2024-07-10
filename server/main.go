package main

import (
	"fmt"
	"log"
	"net"
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
