package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
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

func (c *Client) prepareBlockPayload(message, sender, recipient string) string {
	return c.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeBLCK_USR,
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
	case "/ch":
		return chMessageHandler(parts, c)
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

			return fmt.Sprintf("[%s] [Replied to %s: %s](fg:magenta)", time.Now().Format("01-02 15:04"), c.lastWhispererFromGroupChat, message), nil
		}
	case "/block":
		if len(parts) < 2 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /block <user>"), nil
		}
		user := parts[1]
		if _, err := c.conn.Write([]byte(c.prepareBlockPayload("block", c.name, user))); err != nil {
			return "", fmt.Errorf("error sending blocking: %v", err)
		}
		return fmt.Sprintf("[%s] [%s successfully blocked](fg:magenta)", time.Now().Format("01-02 15:04"), user), nil
	case "/unblock":
		if len(parts) < 2 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /unblock <user>"), nil
		}
		user := parts[1]
		if _, err := c.conn.Write([]byte(c.prepareBlockPayload("unblock", c.name, user))); err != nil {
			return "", fmt.Errorf("error sending unblocking: %v", err)
		}
		return fmt.Sprintf("[%s] [%s successfully unblocked](fg:magenta)", time.Now().Format("01-02 15:04"), user), nil
	case "/mute":
		if len(parts) < 2 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /mute <user>"), nil
		}
		user := parts[1]
		c.AddUserToMutedList(user)
		return fmt.Sprintf("[%s] [%s successfully muted](fg:magenta)", time.Now().Format("01-02 15:04"), user), nil
	case "/unmute":
		if len(parts) < 2 {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /unmute <user>"), nil
		}
		user := parts[1]
		c.RemoveUserFromMutedList(user)
		return fmt.Sprintf("[%s] [%s successfully unmuted](fg:magenta)", time.Now().Format("01-02 15:04"), user), nil
	default:
		return fmt.Sprintf("[%s] [Unknown '%s' command](fg:red)", time.Now().Format("01-02 15:04"), parts[0]), nil
	}
}

func chMessageHandler(parts []string, c *Client) (string, error) {
	partsWithoutCommand, found := strings.CutPrefix(strings.Join(parts, " "), "/ch")
	if !found {
		return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /ch <join|create> <channelName> [channelPassword] [channelSize]"), nil
	}
	roomParts := strings.Fields(partsWithoutCommand)
	if len(roomParts) < 2 {
		return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Usage: /ch <join|create> <channelName> [channelPassword] [channelSize]"), nil
	}

	roomAction := roomParts[0]
	roomName := roomParts[1]
	roomPassword := ""
	roomSizeInt := 10

	if len(roomParts) > 2 {
		roomPassword = roomParts[2]
	}
	if len(roomParts) > 3 {
		roomSize := roomParts[3]
		var err error
		roomSizeInt, err = strconv.Atoi(roomSize)
		if err != nil {
			return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), "Please enter a valid room size"), nil
		}
	}

	payload, err := protocol.NewRoomPayloadBuilder().
		SetRequester(c.name).
		SetChannelAction(protocol.ClientChannelActionMap[roomAction]).
		SetChannelName(roomName).
		SetChannelPassword(roomPassword).
		SetChannelSize(roomSizeInt).
		AddOptionalArg("visibility", protocol.VisibilityPublic).
		Build()
	if err != nil {
		return fmt.Sprintf("[%s] [%s](fg:red)", time.Now().Format("01-02 15:04"), err.Error()), nil
	}
	if _, err := c.conn.Write([]byte(c.encodeFn(*payload))); err != nil {
		return "", fmt.Errorf("error sending channel: %v", err)
	}
	return fmt.Sprintf("[%s] [Channel request sent: %s %s](fg:cyan)", time.Now().Format("01-02 15:04"), roomAction, roomName), nil
}

func (c *Client) HandleReceive(payload protocol.Payload) string {
	var message string

	unixTimeUTC := time.Unix(payload.Timestamp, 0)
	switch payload.MessageType {
	case protocol.MessageTypeCH:
		if payload.ChannelPayload.ChannelAction == protocol.MessageChannel {
			return fmt.Sprintf("[%s] [%s: %s](fg:green)", unixTimeUTC.Format("01-02 15:04"), payload.ChannelPayload.Requester, payload.ChannelPayload.OptionalChannelArgs.Message)
		}
	case protocol.MessageTypeMSG:
		// If the sender is the current user, display "You" instead of the username
		if payload.Sender == c.name {
			return fmt.Sprintf("[%s] [You: %s](fg:cyan)", unixTimeUTC.Format("01-02 15:04"), payload.Content)
		}
		// For messages from other users, display their username
		return fmt.Sprintf("[%s] [%s: %s](fg:green)", unixTimeUTC.Format("01-02 15:04"), payload.Sender, payload.Content)
	case protocol.MessageTypeWSP:
		c.lastWhispererFromGroupChat = payload.Sender
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
