package main

import (
	"bufio"
	"fmt"

	"github.com/ogzhanolguncu/go-chat/client/color"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) readMessages(incomingChan chan<- protocol.Payload, errChan chan<- error, done <-chan struct{}) {
	for {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			select {
			case errChan <- err:
			case <-done:
			}
			return
		}
		payload, err := protocol.DecodeMessage(message)
		if err != nil {
			fmt.Print(color.ColorifyWithTimestamp(err.Error(), color.Red))
			continue
		}
		select {
		case incomingChan <- payload:
		case <-done:
			return
		}
	}
}
