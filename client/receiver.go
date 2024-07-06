package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/ogzhanolguncu/go-chat/client/color"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) readMessages(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}
		payload, err := protocol.DecodeMessage(message)
		if err != nil {
			fmt.Print(color.ColorifyWithTimestamp(err.Error(), color.Red))
		}
		colorifyAndFormatContent(payload)
		// When message received from server we append You: right after it.
		askForInput()
	}
}
