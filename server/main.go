package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const PORT = 7007

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)

	}

}

func handleConnection(c net.Conn) {
	defer c.Close()

	connReader := bufio.NewReader(c)

	log.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		message := strings.TrimSpace(string(data))

		log.Printf("Request from %s: %s\n", c.RemoteAddr().String(), message)
		if message == "quit" {
			break
		}
		c.Write([]byte(data))
	}
	log.Printf("Stop serving %s\n", c.RemoteAddr().String())
}
