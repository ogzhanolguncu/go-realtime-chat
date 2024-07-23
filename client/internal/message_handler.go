package internal

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/ogzhanolguncu/go-chat/client/terminal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) HandleCommand(cmd string, messages *[]string) {
	if !strings.HasPrefix(cmd, "/") {
		*messages = append(*messages, fmt.Sprintf("[%s You: %s](fg:cyan)", time.Now().Format("15:04"), cmd))
		c.conn.Write([]byte(preparePublicMessagePayload(cmd, c.name)))
		return
	}
	parts := strings.Fields(cmd)
	switch parts[0] {
	case "/whisper":
		if len(parts) < 3 {
			*messages = append(*messages, "Usage: /whisper <recipient> <message>")
		} else {
			recipient := parts[1]
			message := strings.Join(parts[2:], " ")
			// Implement whisper functionality here
			*messages = append(*messages, fmt.Sprintf("Whispered to %s: %s", recipient, message))
			message = prepareWhisperPayload(message, c.name, recipient)
			c.conn.Write([]byte(message))
		}
	case "/reply":
		if len(parts) < 2 {
			*messages = append(*messages, "Usage: /reply <message>")
		} else {
			message := strings.Join(parts[1:], " ")
			// Implement reply functionality here
			*messages = append(*messages, fmt.Sprintf("Replied: %s", message))
			message = prepareReplyPayload(message, c.name, c.lastWhispererFromGroupChat)
			c.conn.Write([]byte(message))

		}
	case "/clear":
		*messages = []string{}
	case "/quit":
		// This will be handled in the main loop
	default:
		*messages = append(*messages, "Unknown command. Type /help for available commands.")
	}
}

func prepareReplyPayload(input, sender, recipient string) string {
	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     input,
		Sender:      sender,
	})
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

func prepareWhisperPayload(input, sender, recipient string) string {
	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     input,
		Sender:      sender,
	})
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
