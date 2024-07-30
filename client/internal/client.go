package internal

import (
	"flag"
	"fmt"
	"log"
	"net"
	"slices"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

type Client struct {
	conn                       net.Conn
	config                     Config
	name                       string
	lastWhispererFromGroupChat string

	encodeFn func(payload protocol.Payload) string
	decodeFn func(message string) (protocol.Payload, error)

	mutedUsers []string
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

func (c *Client) GetMutedList() []string {
	return c.mutedUsers
}

func (c *Client) CheckIfUserMuted(user string) bool {
	return slices.ContainsFunc(c.mutedUsers, func(u string) bool {
		return u == user
	})
}

func (c *Client) AddUserToMutedList(user string) bool {
	if c.CheckIfUserMuted(user) {
		return false
	}
	c.mutedUsers = append(c.mutedUsers, user)
	return true
}

func (c *Client) RemoveUserFromMutedList(user string) bool {
	initialLength := len(c.mutedUsers)
	c.mutedUsers = slices.DeleteFunc(c.mutedUsers, func(u string) bool {
		return u == user
	})
	return len(c.mutedUsers) < initialLength
}
