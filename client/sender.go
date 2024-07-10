package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) sendMessages() {
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

		if strings.HasPrefix(text, "/reply") {
			c.sendReply(text)
		}
		if strings.HasPrefix(text, "/whisper") {
			c.sendWhisper(text)
		} else {
			c.sendPublicMessage(text)
		}
	}
}

func (c *Client) sendReply(text string) {
	message := strings.TrimSpace(strings.Split(text, "/reply")[1])
	c.sendWhisper(fmt.Sprintf("/whisper %s %s", c.lastWhispererFromGroupChat, message))
}

func (c *Client) sendWhisper(text string) {
	re := regexp.MustCompile(`^\/whisper\s+(\S+)\s+(.*)$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		recipient := matches[1]
		msg := matches[2]
		message := protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeWSP, Recipient: recipient, Sender: c.name, Content: msg})

		_, err := c.conn.Write([]byte(message))
		if err != nil {
			log.Fatal("Error sending whisper message:", err)
		}
	} else {
		fmt.Println("Invalid whisper command format")
	}
}

func (c *Client) sendPublicMessage(rawInput string) {
	message := protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeMSG, Sender: c.name, Content: rawInput})
	_, err := c.conn.Write([]byte(message))
	if err != nil {
		log.Fatal("Error sending group message:", err)
	}
}
