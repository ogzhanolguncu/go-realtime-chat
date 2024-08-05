package internal

import "github.com/ogzhanolguncu/go-chat/protocol"

func (c *Client) FetchActiveUserList() {
	message := c.prepareActiveUserListPayload(c.name)
	c.conn.Write([]byte(message))
}

func (c *Client) prepareActiveUserListPayload(requester string) string {
	return c.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeACT_USRS,
		Sender:      requester,
		Status:      "req",
	})
}
