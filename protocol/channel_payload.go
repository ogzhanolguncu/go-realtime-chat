package protocol

import "fmt"

// Chat Channel(CH): CH|timestamp|ch_action|requester|chName|chPassword|chSize|optional_args
type ChannelActionType int

const (
	CreateChannel ChannelActionType = iota + 1
	JoinChannel
	LeaveChannel
	KickUser
	BanUser
	GetUsers
	GetChannels
	MessageChannel
)

// Status represents the result of an action
type Status string

const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
)

// Visibility represents the visibility of a room
type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

// OptionalChannelArgs contains optional arguments for room operations
type OptionalChannelArgs struct {
	Status     Status
	Visibility Visibility
	Message    string
	Reason     string
	Channels   []string // For GetRooms
	Users      []string // For GetUsers
	TargetUser string   // For KICK and BAN actions
}

// ChannelPayload represents the payload for room-related operations
type ChannelPayload struct {
	ChannelAction       ChannelActionType
	Requester           string
	ChannelName         string
	ChannelPassword     string
	ChannelSize         int
	OptionalChannelArgs *OptionalChannelArgs
}

func (rat ChannelActionType) String() string {
	switch rat {
	case CreateChannel:
		return "CreateChannel"
	case JoinChannel:
		return "JoinChannel"
	case LeaveChannel:
		return "LeaveChannel"
	case KickUser:
		return "KickUser"
	case BanUser:
		return "BanUser"
	case GetUsers:
		return "GetUsers"
	case GetChannels:
		return "GetChannels"
	case MessageChannel:
		return "MessageChannel"
	default:
		return "Unknown"
	}
}

var ChannelActionMap = map[string]ChannelActionType{
	"CreateChannel":  CreateChannel,
	"JoinChannel":    JoinChannel,
	"LeaveChannel":   LeaveChannel,
	"KickUser":       KickUser,
	"BanUser":        BanUser,
	"GetUsers":       GetUsers,
	"GetChannels":    GetChannels,
	"MessageChannel": MessageChannel,
}

var ClientChannelActionMap = map[string]ChannelActionType{
	"create":  CreateChannel,
	"join":    JoinChannel,
	"leave":   LeaveChannel,
	"kick":    KickUser,
	"ban":     BanUser,
	"users":   GetUsers,
	"list":    GetChannels,
	"message": MessageChannel,
}

// ParseChannelAction converts a string to ChannelActionType
func parseChannelAction(s string) (ChannelActionType, error) {
	action, ok := ChannelActionMap[s]
	if !ok {
		return 0, fmt.Errorf("invalid channel action: %s", s)
	}
	return action, nil
}
