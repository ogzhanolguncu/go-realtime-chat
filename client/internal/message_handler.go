package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/terminal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

type Command struct {
	name    string
	handler func(message, sender, recipient string) (string, error)
}

var commands = []Command{
	{name: "/whisper", handler: prepareWhisperPayload},
	{name: "/reply", handler: prepareReplyPayload},
	{name: "/users", handler: prepareActiveUserPayload},
}

func (c *Client) SendMessages(text string) {
	text = strings.TrimSpace(text)
	if text == "/quit" {
		os.Exit(0)
	}
	if text == "/clear" {
		fmt.Print("\033[H\033[2J") //Clears terminal
	}
	if text == "/help" {
		fmt.Println("")
		PrintHeader(false)
	}

	message, err := processInput(text, c.name, c.lastWhispererFromGroupChat)
	if err != nil {
		log.Println("Error preparing message:", err)
	}
	c.conn.Write([]byte(message))

}
func processInput(input, sender, recipient string) (string, error) {
	for _, cmd := range commands {
		if strings.HasPrefix(input, cmd.name) {
			return cmd.handler(input, sender, recipient)
		}
	}
	return preparePublicMessagePayload(input, sender), nil
}

func prepareReplyPayload(input, sender, recipient string) (string, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid reply format. Use: /reply <message>")
	}
	message := parts[1]
	if recipient == "" {
		return "", fmt.Errorf("no one to reply to")
	}
	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	}), nil
}

func prepareActiveUserPayload(_, _, _ string) (string, error) {
	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeACT_USRS, Status: "req",
	}), nil
}

func prepareChatHistoryPayload(requester string) (string, error) {
	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeHSTRY, Status: "req", Sender: requester,
	}), nil
}

func prepareWhisperPayload(input, sender, _ string) (string, error) {
	parts := strings.SplitN(input, " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid whisper format. Use: /whisper <recipient> <message>")
	}
	recipient := parts[1]
	message := parts[2]

	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	}), nil
}

func preparePublicMessagePayload(input, sender string) (message string) {
	return protocol.EncodeProtocol(protocol.Payload{MessageType: protocol.MessageTypeMSG, Sender: sender, Content: input})
}

//RECEIVER

func (c *Client) ReadMessages(incomingChan chan<- protocol.Payload, errChan chan<- error) {
	for {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			errChan <- err
		}
		payload, err := protocol.DecodeProtocol(message)
		if err != nil {
			fmt.Print(terminal.ColorifyWithTimestamp(err.Error(), terminal.Red, 0))
			continue
		}
		incomingChan <- payload
	}
}
