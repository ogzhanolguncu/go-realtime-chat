package main

import (
	"bufio"
	"fmt"

	"github.com/ogzhanolguncu/go-chat/client/color"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) readMessages(incomingChan chan protocol.Payload, errChan chan error) {
	for {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			errChan <- err
			return
		}
		payload, err := protocol.DecodeMessage(message)
		incomingChan <- payload
		if err != nil {
			fmt.Print(color.ColorifyWithTimestamp(err.Error(), color.Red))
		}

	}
}
