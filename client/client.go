package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type Config struct {
	Port int
}

type Client struct {
	conn   net.Conn
	name   string
	reader *bufio.Reader
	config Config
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
	message, err := serverReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading from server: %w", err)
	}
	if strings.TrimSpace(message) != "USERNAME_REQUIRED" {
		return fmt.Errorf("expected USERNAME_REQUIRED message from server, got: %s", message)
	}

	for retries := 0; retries < 3; retries++ {
		fmt.Print("Enter your username: ")
		nameInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading username input: %w", err)
		}
		nameInput = strings.TrimSpace(nameInput)

		if _, err := c.conn.Write([]byte(nameInput + "\n")); err != nil {
			return fmt.Errorf("error sending username to server: %w", err)
		}

		message, err = serverReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading server response: %w", err)
		}

		decodedMessage, err := decodeMessage(message)
		if err != nil {
			return fmt.Errorf("error decoding server message: %w", err)
		}

		if strings.Contains(decodedMessage.sysStatus, "fail") {
			colorifyAndFormatContent(decodedMessage)
			continue
		}

		if strings.Contains(decodedMessage.sysStatus, "success") {
			colorifyAndFormatContent(decodedMessage)
			c.name = strings.Split(strings.TrimSpace(decodedMessage.content), "=>")[1]
			return nil
		}

		fmt.Println("Unexpected response from server. Please try again.")
	}

	return fmt.Errorf("max retries reached for setting username")
}
