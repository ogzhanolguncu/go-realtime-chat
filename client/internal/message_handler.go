package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

type Command struct {
	name    string
	handler func(message, sender, recipient string) (string, error)
}

var commands = []Command{
	{name: "/whisper", handler: handleWhisper},
	{name: "/reply", handler: handleReply},
	{name: "/users", handler: handleActiveUsers},
}

func (c *Client) SendMessages(outgoingChan chan<- string, done <-chan struct{}) {
	reader := bufio.NewReader(os.Stdin)

	for {
		askForInput()
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		text = strings.TrimSpace(text)
		if text == "/quit" {
			os.Exit(0)
		}
		if text == "/clear" {
			fmt.Print("\033[H\033[2J")
			askForInput()
		}

		message, err := processInput(text, c.name, c.lastWhispererFromGroupChat)
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

func processInput(input, sender, recipient string) (string, error) {
	for _, cmd := range commands {
		if strings.HasPrefix(input, cmd.name) {
			return cmd.handler(input, sender, recipient)
		}

	}
	return handlePublicMessage(input, sender), nil
}

func handleReply(input, sender, recipient string) (string, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid reply format. Use: /reply <message>")
	}
	message := parts[1]
	if recipient == "" {
		return "", fmt.Errorf("no one to reply to")
	}
	return protocol.EncodeMessage(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	}), nil
}

func handleActiveUsers(input, _, _ string) (string, error) {
	return protocol.EncodeMessage(protocol.Payload{
		MessageType: protocol.MessageTypeACT_USRS, Status: "req",
	}), nil
}

func handleWhisper(input, sender, _ string) (string, error) {
	parts := strings.SplitN(input, " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid whisper format. Use: /whisper <recipient> <message>")
	}
	recipient := parts[1]
	message := parts[2]

	return protocol.EncodeMessage(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	}), nil
}

func handlePublicMessage(input, sender string) (message string) {
	return protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeMSG, Sender: sender, Content: input})
}

//RECEIVER

func (c *Client) ReadMessages(incomingChan chan<- protocol.Payload, errChan chan<- error, done <-chan struct{}) {
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

//MESSAGE LOOP

func (c *Client) MessageLoop(incomingChan <-chan protocol.Payload, outgoingChan <-chan string, errChan <-chan error, done chan struct{}) error {
	for {
		select {
		case incMessage, ok := <-incomingChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			if incMessage.MessageType == protocol.MessageTypeWSP {
				c.lastWhispererFromGroupChat = incMessage.Sender
			}
			colorifyAndFormatContent(incMessage)
			askForInput()
		case outMessage, ok := <-outgoingChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			_, err := c.conn.Write([]byte(outMessage))
			if err != nil {
				return fmt.Errorf("error sending message: %v", err)
			}
		case err, ok := <-errChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			return err
		case <-done:
			return nil // Done signal received, exit loop
		}
	}
}
