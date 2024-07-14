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

func (c *Client) sendMessages(outgoingChan chan<- string, done <-chan struct{}) {
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
			return
		}

		var message string
		if strings.HasPrefix(text, "/reply") {
			message, err = c.sendReply(text)
		} else if strings.HasPrefix(text, "/whisper") {
			message, err = c.sendWhisper(text)
		} else {
			message = c.sendPublicMessage(text)
		}

		if err != nil {
			log.Println("Error preparing message:", err)
			continue
		}

		select {
		case outgoingChan <- message:
			// Message sent successfully
		case <-done:
			return
		}
	}
}

func (c *Client) sendReply(text string) (message string, err error) {
	return c.sendWhisper(fmt.Sprintf("/whisper %s %s", c.lastWhispererFromGroupChat, strings.TrimSpace(strings.Split(text, "/reply")[1])))
}

func (c *Client) sendWhisper(text string) (message string, err error) {
	re := regexp.MustCompile(`^\/whisper\s+(\S+)\s+(.*)$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		recipient := matches[1]
		msg := matches[2]
		return protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeWSP, Recipient: recipient, Sender: c.name, Content: msg}), nil
	} else {
		fmt.Println("Invalid whisper command format")
		return "", nil
	}
}

func (c *Client) sendPublicMessage(rawInput string) (message string) {
	return protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeMSG, Sender: c.name, Content: rawInput})
}
