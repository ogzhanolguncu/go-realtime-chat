package protocol

import "fmt"

// Chat Room(ROOM): 				ROOM|timestamp|room_action|requester|roomName|roomPassword|roomSize|optional_args

type RoomActionType int

const (
	CreateRoom RoomActionType = iota + 1
	JoinRoom
	LeaveRoom
	KickUser
	BanUser
	GetUsers
	GetRooms
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

// OptionalRoomArgs contains optional arguments for room operations
type OptionalRoomArgs struct {
	Status     Status
	Visibility Visibility
	Message    string
	Reason     string
	Rooms      []string // For GetRooms
	Users      []string // For GetUsers
	TargetUser string   // For KICK and BAN actions
}

// RoomPayload represents the payload for room-related operations
type RoomPayload struct {
	RoomAction       RoomActionType
	Requester        string
	RoomName         string
	RoomPassword     string
	RoomSize         int
	OptionalRoomArgs *OptionalRoomArgs
}

func (rat RoomActionType) String() string {
	switch rat {
	case CreateRoom:
		return "CreateRoom"
	case JoinRoom:
		return "JoinRoom"
	case LeaveRoom:
		return "LeaveRoom"
	case KickUser:
		return "KickUser"
	case BanUser:
		return "BanUser"
	case GetUsers:
		return "GetUsers"
	case GetRooms:
		return "GetRooms"
	default:
		return "Unknown"
	}
}

var RoomActionMap = map[string]RoomActionType{
	"CreateRoom": CreateRoom,
	"JoinRoom":   JoinRoom,
	"LeaveRoom":  LeaveRoom,
	"KickUser":   KickUser,
	"BanUser":    BanUser,
	"GetUsers":   GetUsers,
	"GetRooms":   GetRooms,
}

// ParseRoomAction converts a string to RoomActionType
func parseRoomAction(s string) (RoomActionType, error) {
	action, ok := RoomActionMap[s]
	if !ok {
		return 0, fmt.Errorf("invalid room action: %s", s)
	}
	return action, nil
}
