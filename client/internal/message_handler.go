package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/ogzhanolguncu/go-chat/client/terminal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func prepareReplyPayload(message, sender, recipient string) string {

	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	})
}

func prepareWhisperPayload(message, sender, recipient string) string {
	return protocol.EncodeProtocol(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	})
}

func preparePublicMessagePayload(message, sender string) string {
	return protocol.EncodeProtocol(protocol.Payload{MessageType: protocol.MessageTypeMSG, Sender: sender, Content: message})
}

//RECEIVER

func (c *Client) ReadMessages(ctx context.Context, incomingChan chan<- protocol.Payload) {
	reader := bufio.NewReader(c.conn)
	for {
		select {
		case <-ctx.Done():
			// Context was canceled, time to exit
			return
		default:
			// Set a deadline for the read operation
			err := c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			if err != nil {
				fmt.Print(terminal.ColorifyWithTimestamp("Failed to set read deadline: "+err.Error(), terminal.Red, 0))
				continue
			}

			message, err := reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is a timeout, just continue the loop
					continue
				}
				if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
					// Connection was closed, time to exit
					return
				}
				fmt.Print(terminal.ColorifyWithTimestamp("Read error: "+err.Error(), terminal.Red, 0))
				continue
			}

			payload, err := protocol.DecodeProtocol(message)
			if err != nil {
				fmt.Print(terminal.ColorifyWithTimestamp("Decode error: "+err.Error(), terminal.Red, 0))
				continue
			}

			incomingChan <- payload
		}
	}
}

func (c *Client) HandleSend(userInput string) (string, error) {
	if !strings.HasPrefix(userInput, "/") {
		if _, err := c.conn.Write([]byte(preparePublicMessagePayload(userInput, c.name))); err != nil {
			return "", fmt.Errorf("error sending message: %v", err)
		}
		return fmt.Sprintf("[%s] [You: %s](fg:cyan)", time.Now().Format("15:04"), userInput), nil

	}
	parts := strings.Fields(userInput)
	switch parts[0] {
	case "/whisper":
		if len(parts) < 3 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("15:04"), "Usage: /whisper <recipient> <message>"), nil
		} else {
			recipient := parts[1]
			message := strings.Join(parts[2:], " ")

			if _, err := c.conn.Write([]byte(prepareWhisperPayload(message, c.name, recipient))); err != nil {
				return "", fmt.Errorf("error sending whisper: %v", err)
			}
			return fmt.Sprintf("[%s] [Whispered to %s: %s](fg:magenta)", time.Now().Format("15:04"), recipient, message), nil
		}
	case "/reply":
		if len(parts) < 2 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("15:04"), "Usage: /reply <message>"), nil
		} else {
			message := strings.Join(parts[1:], " ")
			if c.lastWhispererFromGroupChat == "" {
				return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("15:04"), "No one to reply to"), nil
			}

			if _, err := c.conn.Write([]byte(prepareReplyPayload(message, c.name, c.lastWhispererFromGroupChat))); err != nil {
				return "", fmt.Errorf("error sending whisper: %v", err)
			}

			return fmt.Sprintf("[%s] [Replied: %s](fg:magenta)", time.Now().Format("15:04"), message), nil
		}
	default:
		return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("15:04"), "Unknown command"), nil
	}
}

func (c *Client) HandleReceive(payload protocol.Payload) string {
	var message string

	switch payload.MessageType {
	case protocol.MessageTypeMSG:
		return fmt.Sprintf("[%s] [%s: %s](fg:green)", time.Now().Format("15:04"), payload.Sender, payload.Content)
	case protocol.MessageTypeWSP:
		message = fmt.Sprintf("[%s] [Whisper from %s: %s](fg:purple)\n", time.Now().Format("15:04"), payload.Sender, payload.Content)
	case protocol.MessageTypeSYS:
		if payload.Status == "fail" {
			message = fmt.Sprintf("[%s] [System: %s](fg:red)", time.Now().Format("15:04"), payload.Content)
		} else {
			message = fmt.Sprintf("[%s] [System: %s](fg:cyan)", time.Now().Format("15:04"), payload.Content)
		}
	default:
		message = fmt.Sprintf("%s: %s\n", payload.Sender, payload.Content)
	}

	return message
}
