package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

type Client struct {
	conn                       net.Conn
	config                     Config
	name                       string
	lastWhispererFromGroupChat string

	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
	groupChatKey string

	encodeFn func(payload protocol.Payload) string
	decodeFn func(message string) (protocol.Payload, error)
}

func NewClient(config Config) (*Client, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	publicKey := &privateKey.PublicKey

	encoding := flag.Bool("encoding", false, "enable encoding")
	flag.Parse()

	var encodingType string
	if *encoding {
		encodingType = "BASE64"
	} else {
		encodingType = "PLAIN-TEXT"
	}

	log.Printf("------ ENCODING SET TO %s ------", encodingType)

	return &Client{
		config:     config,
		privateKey: privateKey,
		publicKey:  publicKey,
		decodeFn:   protocol.InitDecodeProtocol(*encoding),
		encodeFn:   protocol.InitEncodeProtocol(*encoding),
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

func (c *Client) GetUsername() string {
	return c.name
}
