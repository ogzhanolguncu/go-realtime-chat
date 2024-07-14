package internal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) SetUsername() error {
	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(c.conn)

	if err := c.readAndValidateInitialMessage(serverReader); err != nil {
		return err
	}

	for retries := 0; retries < 3; retries++ {
		fmt.Print(color.ColorifyWithTimestamp("Enter your username: ", color.White))
		nameInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading username input: %w", err)
		}

		nameInput = strings.TrimSpace(nameInput)
		if nameInput == "" {
			fmt.Println(color.ColorifyWithTimestamp("Username cannot be empty. Please try again.", color.Red))
			continue
		}

		if err := c.sendUsernameAndHandleResponse(nameInput, serverReader); err != nil {
			if err == errRetry {
				continue
			}
			return err
		}

		return nil
	}

	return fmt.Errorf("max retries reached for setting username")
}

func (c *Client) readAndValidateInitialMessage(reader *bufio.Reader) error {
	message, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading from server: %w", err)
	}

	decodedMessage, err := protocol.DecodeMessage(message)
	if err != nil {
		return fmt.Errorf("error reading server response: %w", err)
	}

	if decodedMessage.ContentType != protocol.MessageTypeUSR || decodedMessage.Status != "required" {
		return fmt.Errorf("expected username required message from server, got: %s", message)
	}

	return nil
}

var errRetry = errors.New("retry username")

func (c *Client) sendUsernameAndHandleResponse(username string, reader *bufio.Reader) error {
	if _, err := c.conn.Write([]byte(protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeUSR, Username: username}))); err != nil {
		return fmt.Errorf("error sending username to server: %w", err)
	}

	message, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading server response: %w", err)
	}

	decodedMessage, err := protocol.DecodeMessage(message)
	if err != nil {
		return fmt.Errorf("error decoding server message: %w", err)
	}

	switch decodedMessage.Status {
	case "fail":
		colorifyAndFormatContent(decodedMessage)
		return errRetry
	case "success":
		colorifyAndFormatContent(decodedMessage)
		c.name = decodedMessage.Username
		return nil
	default:
		return fmt.Errorf("unexpected response from server: %s", decodedMessage.Status)
	}
}
