package internal

import (
	"bufio"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) FetchActiveUsersAfterUsername() error {
	serverReader := bufio.NewReader(c.conn)
	message, err := handleActiveUsers("", "", "")
	if err != nil {
		return err
	}

	_, err = c.conn.Write([]byte(message))
	if err != nil {
		return err
	}

	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	serverResp, err := serverReader.ReadString('\n')
	c.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return err
	}
	decodedMsg, err := protocol.DecodeMessage(serverResp)
	if err != nil {
		return err
	}

	printActiveUsers(decodedMsg.ActiveUsers)
	return nil
}