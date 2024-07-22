package internal

import (
	"bufio"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) FetchActiveUsersAfterUsername() ([]string, error) {
	serverReader := bufio.NewReader(c.conn)
	message, err := prepareActiveUserPayload("", "", "")
	if err != nil {
		return nil, err
	}

	_, err = c.conn.Write([]byte(message))
	if err != nil {
		return nil, err
	}

	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	serverResp, err := serverReader.ReadString('\n')
	c.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return nil, err
	}
	decodedMsg, err := protocol.DecodeProtocol(serverResp)
	if err != nil {
		return nil, err
	}

	return decodedMsg.ActiveUsers, nil
}
