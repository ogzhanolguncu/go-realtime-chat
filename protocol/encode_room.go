package protocol

import (
	"fmt"
	"strings"
	"time"
)

const (
	// If args are empty it will return empty string
	optionalArgsSeparator         = ";"
	optionalUserAndRoomsSeparator = ","
	emptyRoomField                = "-"
)

type RoomPayloadBuilder struct {
	payload *Payload
}

func NewRoomPayloadBuilder() *RoomPayloadBuilder {
	return &RoomPayloadBuilder{
		payload: &Payload{
			MessageType: MessageTypeROOM,
			Timestamp:   time.Now().Unix(),
			RoomPayload: &RoomPayload{},
		},
	}
}

func (b *RoomPayloadBuilder) SetRoomAction(action RoomActionType) *RoomPayloadBuilder {
	b.payload.RoomPayload.RoomAction = action
	return b
}

func (b *RoomPayloadBuilder) SetRequester(requester string) *RoomPayloadBuilder {
	b.payload.RoomPayload.Requester = requester
	return b
}

func (b *RoomPayloadBuilder) SetRoomName(name string) *RoomPayloadBuilder {
	b.payload.RoomPayload.RoomName = name
	return b
}

func (b *RoomPayloadBuilder) SetRoomPassword(password string) *RoomPayloadBuilder {
	b.payload.RoomPayload.RoomPassword = password
	return b
}

func (b *RoomPayloadBuilder) SetRoomSize(size int) *RoomPayloadBuilder {
	b.payload.RoomPayload.RoomSize = size
	return b
}

func (b *RoomPayloadBuilder) AddOptionalArg(key string, value interface{}) *RoomPayloadBuilder {
	if b.payload.RoomPayload.OptionalRoomArgs == nil {
		b.payload.RoomPayload.OptionalRoomArgs = &OptionalRoomArgs{}
	}

	switch key {
	case "status":
		b.payload.RoomPayload.OptionalRoomArgs.Status = value.(Status)
	case "visibility":
		b.payload.RoomPayload.OptionalRoomArgs.Visibility = value.(Visibility)
	case "message":
		b.payload.RoomPayload.OptionalRoomArgs.Message = value.(string)
	case "reason":
		b.payload.RoomPayload.OptionalRoomArgs.Reason = value.(string)
	case "rooms":
		b.payload.RoomPayload.OptionalRoomArgs.Rooms = value.([]string)
	case "users":
		b.payload.RoomPayload.OptionalRoomArgs.Users = value.([]string)
	case "target_user":
		b.payload.RoomPayload.OptionalRoomArgs.TargetUser = value.(string)
	}
	return b
}

func (b *RoomPayloadBuilder) Build() (*Payload, error) {
	if b.payload.RoomPayload.Requester == "" {
		return nil, fmt.Errorf("requester is required")
	}
	if b.payload.RoomPayload.RoomAction == 0 {
		return nil, fmt.Errorf("room action is required")
	}
	return b.payload, nil
}

//Example usage:
// builder := NewRoomPayloadBuilder().SetRoomAction(CreateRoom).
//	SetRequester("user123").
//	SetRoomName("My Room").
//	SetRoomPassword("secret").
//	SetRoomSize(10).
//	AddOptionalArg("visibility", VisibilityPublic)

func encodeROOM(payload *Payload) string {
	var sb strings.Builder
	rp := payload.RoomPayload

	timestamp := payload.Timestamp
	if timestamp == 0 {
		payload.Timestamp = time.Now().Unix() // Fallback if timestamp is somehow 0
	}

	sb.WriteString(fmt.Sprintf("%s|%d|%s|%s|", MessageTypeROOM, payload.Timestamp, rp.RoomAction, rp.Requester))

	if rp.RoomName != "" {
		sb.WriteString(rp.RoomName)
	} else {
		sb.WriteString(emptyRoomField)
	}
	sb.WriteString("|")

	if rp.RoomPassword != "" {
		sb.WriteString(rp.RoomPassword)
	} else {
		sb.WriteString(emptyRoomField)
	}
	sb.WriteString("|")

	if rp.RoomSize != 0 {
		sb.WriteString(fmt.Sprintf("%d", rp.RoomSize))
	} else {
		sb.WriteString(emptyRoomField)
	}

	if rp.OptionalRoomArgs != nil {
		optionalArgs := serializeRoomOptionalArgs(rp.OptionalRoomArgs)
		if optionalArgs != "" {
			sb.WriteString("|")
			sb.WriteString(optionalArgs)
		}
	}

	return sb.String()
}

func serializeRoomOptionalArgs(args *OptionalRoomArgs) string {
	var optsParts []string

	if args.Status != "" {
		optsParts = append(optsParts, "status="+string(args.Status))
	}

	if args.Visibility != "" {
		optsParts = append(optsParts, "visibility="+string(args.Visibility))
	}

	if args.Message != "" {
		optsParts = append(optsParts, "message="+string(args.Message))
	}

	if args.Reason != "" {
		optsParts = append(optsParts, "reason="+string(args.Reason))
	}

	if args.Rooms != nil {
		optsParts = append(optsParts, "rooms="+strings.Join(args.Rooms, optionalUserAndRoomsSeparator))
	}

	if args.Users != nil {
		optsParts = append(optsParts, "users="+strings.Join(args.Users, optionalUserAndRoomsSeparator))
	}

	if args.TargetUser != "" {
		optsParts = append(optsParts, "target_user="+args.TargetUser)
	}

	return strings.Join(optsParts, optionalArgsSeparator)
}
