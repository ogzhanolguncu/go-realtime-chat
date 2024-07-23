package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
)

type Client struct {
	conn                       net.Conn
	config                     Config
	name                       string
	lastWhispererFromGroupChat string

	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
	groupChatKey string
}

func NewClient(config Config) (*Client, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	publicKey := &privateKey.PublicKey
	return &Client{
		config:     config,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%s", c.config.Port))
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) UpdateLastWhispererFromGroupChat(lastWhisperer string) {
	c.lastWhispererFromGroupChat = lastWhisperer
}
