package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

const (
	cmdCreate  = "create"
	cmdJoin    = "join"
	cmdMessage = "message"
	cmdLeave   = "leave"
	cmdUsers   = "users"
	cmdKick    = "kick"
	cmdBan     = "ban"
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
	case cmdList:
		return handleGetChannelList(c)
	case cmdKick:
		return handleKickUser(c, channelName, args)
	default:
		return fmt.Sprintf("[%s] [Unknown action: %s](fg:red)", time.Now().Format("01-02 15:04"), action), nil
	}
}

func handleCreateChannel(c *Client, channelName string, args []string) (string, error) {
	var password string
	var size int = 2
	visibility := protocol.VisibilityPublic

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
	if len(args) > 2 {
		selectedVisibility := args[2]
		if selectedVisibility == "private" {
			visibility = protocol.VisibilityPrivate
		}
	}

	payload, err := buildChannelPayload(c, protocol.CreateChannel, channelName, password, size, visibility)
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

	payload, err := buildChannelPayload(c, protocol.JoinChannel, channelName, password, 0, "")
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
		message = strings.Join(args[1:], " ")
	}
	if len(args) == 1 {
		message = strings.Join(args, " ")
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

func handleGetChannelList(c *Client) (string, error) {
	payload, err := protocol.NewChannelPayloadBuilder().
		SetRequester(c.name).
		SetChannelAction(protocol.GetChannels).
		Build()
	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [Requested channel list](fg:magenta)", time.Now().Format("01-02 15:04")), nil
}

func handleLeaveChannel(c *Client, channelName string) (string, error) {
	payload, err := buildChannelPayload(c, protocol.LeaveChannel, channelName, "", 0, "")
	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [Requested leave channel: '%s'](fg:magenta)", time.Now().Format("01-02 15:04"), channelName), nil
}

func buildChannelPayload(c *Client, action protocol.ChannelActionType, channelName, password string, size int, visibility protocol.Visibility) (*protocol.Payload, error) {
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
		builder.AddOptionalArg("visibility", visibility)
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

func handleKickUser(c *Client, channelName string, args []string) (string, error) {
	var password, target_user string
	if len(args) > 1 {
		password = args[0]
		target_user = args[1]
	}
	if len(args) == 1 {
		target_user = args[0]
	}

	payload, err := protocol.NewChannelPayloadBuilder().
		SetRequester(c.name).
		SetChannelAction(protocol.KickUser).
		SetChannelName(channelName).
		SetChannelPassword(password).
		AddOptionalArg("target_user", target_user).
		Build()
	if err != nil {
		return "", err
	}

	if err := sendPayload(c, payload); err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s] [User '%s' has requested to kick '%s' from the channel](fg:magenta)",
		time.Now().Format("01-02 15:04"), c.name, target_user), nil
}

func (c *Client) HandleChReceive(payload protocol.Payload) (msg string, shouldExit bool) {
	//If received message is not a channel payload skip the rest
	if payload.ChannelPayload == nil {
		return "", false
	}
	var message string
	unixTimeUTC := time.Unix(payload.Timestamp, 0)
	switch payload.MessageType {
	case protocol.MessageTypeCH:
		switch {
		case payload.ChannelPayload.OptionalChannelArgs.Status == protocol.StatusFail:
			// Any failed message will be caught here
			message = fmt.Sprintf("[%s] [%s](fg:red)",
				unixTimeUTC.Format("01-02 15:04"),
				payload.ChannelPayload.OptionalChannelArgs.Reason)

		case payload.ChannelPayload.ChannelAction == protocol.LeaveChannel &&
			payload.ChannelPayload.OptionalChannelArgs.Status == protocol.StatusSuccess:
			return "", true

		case payload.ChannelPayload.ChannelAction == protocol.KickUser &&
			payload.ChannelPayload.OptionalChannelArgs.Status == protocol.StatusSuccess:
			return fmt.Sprintf("[%s] [You have been kicked by '%s'](fg:magenta)",
				unixTimeUTC.Format("01-02 15:04"),
				payload.ChannelPayload.Requester), true

		case payload.ChannelPayload.ChannelAction == protocol.MessageChannel:
			//Message Channel
			message = fmt.Sprintf("[%s] [%s: %s](fg:green)",
				unixTimeUTC.Format("01-02 15:04"),
				payload.ChannelPayload.Requester,
				strings.Trim(payload.ChannelPayload.OptionalChannelArgs.Message, "\r\n"))

		case payload.ChannelPayload.ChannelAction == protocol.NoticeChannel &&
			payload.ChannelPayload.OptionalChannelArgs.Notice != "":
			message = fmt.Sprintf("[%s] [%s](fg:magenta)", unixTimeUTC.Format("01-02 15:04"), payload.ChannelPayload.OptionalChannelArgs.Notice)

		case payload.ChannelPayload.ChannelAction == protocol.GetUsers &&
			payload.ChannelPayload.OptionalChannelArgs != nil &&
			payload.ChannelPayload.OptionalChannelArgs.Users != nil &&
			payload.ChannelPayload.OptionalChannelArgs.Status == protocol.StatusSuccess:
			//Get Users
			message = fmt.Sprintf("[%s] [%s](fg:magenta)",
				unixTimeUTC.Format("01-02 15:04"),
				strings.Join(payload.ChannelPayload.OptionalChannelArgs.Users, fmt.Sprintf("%s ", protocol.OptionalUserAndChannelsSeparator)))

		default:
			// Handle any other cases for MessageTypeCH
			message = fmt.Sprintf("[%s] [Unhandled channel action](fg:red)",
				unixTimeUTC.Format("01-02 15:04"))
		}
	default:
		message = fmt.Sprintf("[%s] [Unknown message type](fg:red)", unixTimeUTC.Format("01-02 15:04"))
	}
	return message, false
}
