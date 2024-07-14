package internal

import (
	"fmt"
	"net"
)

type Client struct {
	conn                       net.Conn
	config                     Config
	name                       string
	lastWhispererFromGroupChat string
}

func NewClient(config Config) (*Client, error) {
	return &Client{
		config: config,
	}, nil
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%s", c.config.Port))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
