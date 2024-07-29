package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) prepareReplyPayload(message, sender, recipient string) string {

	return c.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	})
}

func (c *Client) prepareWhisperPayload(message, sender, recipient string) string {
	return c.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeWSP,
		Recipient:   recipient,
		Content:     message,
		Sender:      sender,
	})
}

func (c *Client) preparePublicMessagePayload(message, sender string) string {
	return c.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeMSG, Sender: sender, Content: message})
}

//RECEIVER

func (c *Client) ReadMessages(ctx context.Context, incomingChan chan<- protocol.Payload, errorChan chan error) {
	reader := bufio.NewReader(c.conn)
	for {
		select {
		case <-ctx.Done():
			// Context was canceled, time to exit
			close(incomingChan) // Close channel to signal the end of incoming messages
			close(errorChan)    // Close channel to signal the end of error reporting
			return
		default:
			// Set a deadline for the read operation
			err := c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			if err != nil {
				errorChan <- fmt.Errorf("failed to set read deadline: %w", err)
				continue
			}

			message, err := reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is a timeout, just continue the loop
					continue
				}
				if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
					// Connection was closed, signal the end
					errorChan <- io.EOF
					return
				}
				errorChan <- io.EOF
				continue
			}

			payload, err := c.decodeFn(message)
			if err != nil {
				// Client keep reading it payload is broken, its safe
				continue
			}

			incomingChan <- payload
		}
	}
}

func (c *Client) HandleSend(userInput string) (string, error) {
	if !strings.HasPrefix(userInput, "/") {
		if _, err := c.conn.Write([]byte(c.preparePublicMessagePayload(userInput, c.name))); err != nil {
			return "", fmt.Errorf("error sending message: %v", err)
		}
		return fmt.Sprintf("[%s] [You: %s](fg:cyan)", time.Now().Format("01-02 15:04"), userInput), nil

	}
	parts := strings.Fields(userInput)
	switch parts[0] {
	case "/whisper":
		if len(parts) < 3 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /whisper <recipient> <message>"), nil
		} else {
			recipient := parts[1]
			message := strings.Join(parts[2:], " ")

			if _, err := c.conn.Write([]byte(c.prepareWhisperPayload(message, c.name, recipient))); err != nil {
				return "", fmt.Errorf("error sending whisper: %v", err)
			}
			return fmt.Sprintf("[%s] [Whispered to %s: %s](fg:magenta)", time.Now().Format("01-02 15:04"), recipient, message), nil
		}
	case "/reply":
		if len(parts) < 2 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /reply <message>"), nil
		} else {
			message := strings.Join(parts[1:], " ")
			if c.lastWhispererFromGroupChat == "" {
				return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "No one to reply to"), nil
			}

			if _, err := c.conn.Write([]byte(c.prepareReplyPayload(message, c.name, c.lastWhispererFromGroupChat))); err != nil {
				return "", fmt.Errorf("error sending whisper: %v", err)
			}

			return fmt.Sprintf("[%s] [Replied: %s](fg:magenta)", time.Now().Format("01-02 15:04"), message), nil
		}
	default:
		return fmt.Sprintf("[%s] [Unknown '%s' command](fg:red)", time.Now().Format("01-02 15:04"), parts[0]), nil
	}
}

func (c *Client) HandleReceive(payload protocol.Payload) string {
	var message string

	unixTimeUTC := time.Unix(payload.Timestamp, 0)
	switch payload.MessageType {
	case protocol.MessageTypeMSG:
		return fmt.Sprintf("[%s] [%s: %s](fg:green)", unixTimeUTC.Format("01-02 15:04"), payload.Sender, payload.Content)
	case protocol.MessageTypeWSP:
		message = fmt.Sprintf("[%s] [Whisper from %s: %s](fg:magenta)", unixTimeUTC.Format("01-02 15:04"), payload.Sender, payload.Content)
	case protocol.MessageTypeSYS:
		if payload.Status == "fail" {
			message = fmt.Sprintf("[%s] [System: %s](fg:red)", unixTimeUTC.Format("01-02 15:04"), payload.Content)
		} else {
			message = fmt.Sprintf("[%s] [System: %s](fg:cyan)", unixTimeUTC.Format("01-02 15:04"), payload.Content)
		}
	default:
		return fmt.Sprintf("[%s] [Unknown message](fg:red)", unixTimeUTC.Format("01-02 15:04"))
	}

	return message
}
