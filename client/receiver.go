package main

import (
	"bufio"
	"fmt"
	"log"

	"github.com/ogzhanolguncu/go-chat/client/color"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) readMessages() {
	for {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}
		payload, err := protocol.DecodeMessage(message)

		if err != nil {
			fmt.Print(color.ColorifyWithTimestamp(err.Error(), color.Red))
		}
		// This is required for /reply function to work.
		if payload.ContentType == protocol.MessageTypeWSP {
			c.lastWhispererFromGroupChat = payload.Sender
		}
		colorifyAndFormatContent(payload)
		// When message received from server we append You: right after it.
		askForInput()
	}
}
