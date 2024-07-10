package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

type Config struct {
	Port int
}

type Client struct {
	conn                       net.Conn
	name                       string
	reader                     *bufio.Reader
	config                     Config
	lastWhispererFromGroupChat string
}

func newClient(config Config) (*Client, error) {
	return &Client{
		reader: bufio.NewReader(os.Stdin),
		config: config,
	}, nil
}

func (c *Client) connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", c.config.Port))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) printHeader() {
	fmt.Printf("\n\n")
	fmt.Println("---------CHATROOM--------")
	fmt.Println("-------------------------")
}

func (c *Client) setUsername() error {
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
