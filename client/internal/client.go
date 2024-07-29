package internal

import (
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

	encodeFn func(payload protocol.Payload) string
	decodeFn func(message string) (protocol.Payload, error)
}

func NewClient(config Config) (*Client, error) {
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
		config:   config,
		decodeFn: protocol.InitDecodeProtocol(*encoding),
		encodeFn: protocol.InitEncodeProtocol(*encoding),
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
