package internal

import (
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) FetchChatHistory() {
	message := c.prepareChatHistoryPayload(c.name)
	c.conn.Write([]byte(message))
}

func (c *Client) prepareChatHistoryPayload(requester string) string {
	return c.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeHSTRY,
		Sender:      requester,
		Status:      "req",
	})
}
