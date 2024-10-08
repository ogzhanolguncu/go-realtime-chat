package internal

import (
	"fmt"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) SendUsernameReq(username, password string) error {
	if _, err := c.conn.Write([]byte(c.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Password: password, Username: username}))); err != nil {
		return fmt.Errorf("error sending username to server: %w", err)
	}
	return nil
}
