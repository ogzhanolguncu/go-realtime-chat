package internal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/terminal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) SetUsername() error {
	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(c.conn)

	if err := c.readAndValidateInitialMessage(serverReader); err != nil {
		return err
	}

	for retries := 0; retries < 3; retries++ {
		fmt.Print(terminal.ColorifyWithTimestamp("Enter your username: ", terminal.White, 0))
		nameInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading username input: %w", err)
		}

		nameInput = strings.TrimSpace(nameInput)
		if nameInput == "" {
			fmt.Println(terminal.ColorifyWithTimestamp("Username cannot be empty. Please try again.", terminal.Red, 0))
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

	decodedMessage, err := protocol.DecodeProtocol(message)
	if err != nil {
		return fmt.Errorf("error reading server response: %w", err)
	}

	if decodedMessage.MessageType != protocol.MessageTypeUSR || decodedMessage.Status != "required" {
		return fmt.Errorf("expected username required message from server, got: %s", message)
	}

	return nil
}

var errRetry = errors.New("retry username")

func (c *Client) sendUsernameAndHandleResponse(username string, reader *bufio.Reader) error {
	if _, err := c.conn.Write([]byte(protocol.EncodeProtocol(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: username}))); err != nil {
		return fmt.Errorf("error sending username to server: %w", err)
	}

	message, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading server response: %w", err)
	}

	decodedMessage, err := protocol.DecodeProtocol(message)
	if err != nil {
		return fmt.Errorf("error decoding server message: %w", err)
	}

	switch decodedMessage.Status {
	case "fail":
		ColorifyAndFormatContent(decodedMessage)
		return errRetry
	case "success":
		ColorifyAndFormatContent(decodedMessage)
		c.name = decodedMessage.Username
		return nil
	default:
		return fmt.Errorf("unexpected response from server: %s", decodedMessage.Status)
	}
}
