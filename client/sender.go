package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) sendMessages(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	for {
		askForInput()
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		text = strings.TrimSpace(text)
		if text == "quit" {
			break
		}

		if strings.HasPrefix(text, "/whisper ") {
			c.sendWhisper(conn, text)
		} else {
			c.sendPublicMessage(conn, text)
		}
	}
}

func (c *Client) sendWhisper(conn net.Conn, rawInput string) {
	re := regexp.MustCompile(`^\/whisper\s+(\S+)\s+(.*)$`)
	matches := re.FindStringSubmatch(rawInput)
	if len(matches) == 3 {
		recipient := matches[1]
		msg := matches[2]
		message := protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeWSP, Recipient: recipient, Sender: c.name, Content: msg})

		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Fatal("Error sending whisper message:", err)
		}
	} else {
		fmt.Println("Invalid whisper command format")
	}
}

func (c *Client) sendPublicMessage(conn net.Conn, rawInput string) {
	message := protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeMSG, Sender: c.name, Content: rawInput})
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Fatal("Error sending group message:", err)
	}
}
