package internal

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

const (
	cmdCreate  = "create"
	cmdJoin    = "join"
	cmdMessage = "message"
	cmdLeave   = "leave"
	cmdUsers   = "users"
	cmdList    = "list" // Useful for getting channel list on group chat
)

func chMessageHandler(parts []string, c *Client) (string, error) {
	if len(parts) < 3 {
		return fmt.Sprintf("[%s] [Usage: /ch <action> <channelName> [args...]](fg:red)", time.Now().Format("01-02 15:04")), nil
	}

	action := parts[1]
	channelName := parts[2]
	args := parts[3:]

	switch action {
	case cmdCreate:
		return handleCreateChannel(c, channelName, args)
	case cmdJoin:
		return handleJoinChannel(c, channelName, args)
	case cmdMessage:
		return handleMessageChannel(c, channelName, args)
	case cmdLeave:
		return handleLeaveChannel(c, channelName)
	case cmdUsers:
		return handleGetUsersOfChannel(c, channelName, args)
	default:
		return fmt.Sprintf("[%s] [Unknown action: %s](fg:red)", time.Now().Format("01-02 15:04"), action), nil
	}
}

func handleCreateChannel(c *Client, channelName string, args []string) (string, error) {
	var password string
	var size int = 2

	if len(args) > 0 {
		password = args[0]
	}
	if len(args) > 1 {
		var err error
		size, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Sprintf("[%s] [Invalid channel size: %s](fg:red)", time.Now().Format("01-02 15:04"), args[1]), nil
		}
	}

	payload, err := buildChannelPayload(c, protocol.CreateChannel, channelName, password, size)
	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [Channel create request sent: %s](fg:magenta)", time.Now().Format("01-02 15:04"), channelName), nil
}

func handleJoinChannel(c *Client, channelName string, args []string) (string, error) {
	var password string
	if len(args) > 0 {
		password = args[0]
	}

	payload, err := buildChannelPayload(c, protocol.JoinChannel, channelName, password, 0)
	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [Channel join request sent: %s](fg:magenta)", time.Now().Format("01-02 15:04"), channelName), nil
}

func handleMessageChannel(c *Client, channelName string, args []string) (string, error) {
	if len(args) == 0 {
		return fmt.Sprintf("[%s] [Message content is required](fg:red)", time.Now().Format("01-02 15:04")), nil
	}

	var password, message string
	if len(args) > 1 {
		password = args[0]
		message = args[1]
	}
	if len(args) == 1 {
		message = args[0]
	}

	payload, err := protocol.NewChannelPayloadBuilder().
		SetRequester(c.name).
		SetChannelAction(protocol.MessageChannel).
		SetChannelName(channelName).
		SetChannelPassword(password).
		AddOptionalArg("message", message).
		Build()

	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [You: %s](fg:cyan)", time.Now().Format("01-02 15:04"), message), nil
}

func handleGetUsersOfChannel(c *Client, channelName string, args []string) (string, error) {
	var password string
	if len(args) > 1 {
		password = args[0]
	}

	payload, err := protocol.NewChannelPayloadBuilder().
		SetRequester(c.name).
		SetChannelAction(protocol.GetUsers).
		SetChannelName(channelName).
		SetChannelPassword(password).
		Build()
	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [Requested channel '%s' users](fg:magenta)", time.Now().Format("01-02 15:04"), channelName), nil
}

func handleLeaveChannel(c *Client, channelName string) (string, error) {
	payload, err := buildChannelPayload(c, protocol.LeaveChannel, channelName, "", 0)
	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [Left channel: %s](fg:cyan)", time.Now().Format("01-02 15:04"), channelName), nil
}

func buildChannelPayload(c *Client, action protocol.ChannelActionType, channelName, password string, size int) (*protocol.Payload, error) {
	builder := protocol.NewChannelPayloadBuilder().
		SetRequester(c.name).
		SetChannelAction(action).
		SetChannelName(channelName)

	if password != "" {
		builder.SetChannelPassword(password)
	}

	if size > 0 {
		builder.SetChannelSize(size)
	}

	if action == protocol.CreateChannel {
		builder.AddOptionalArg("visibility", protocol.VisibilityPublic)
	}

	return builder.Build()
}

func sendPayload(c *Client, payload *protocol.Payload) error {
	_, err := c.conn.Write([]byte(c.encodeFn(*payload)))
	if err != nil {
		return fmt.Errorf("error sending payload: %v", err)
	}
	return nil
}
