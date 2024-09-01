package protocol

import (
	"fmt"
	"strings"
	"time"
)

const (
	// If args are empty it will return empty string
	optionalArgsSeparator            = ";"
	optionalUserAndChannelsSeparator = ","
	emptyChannelField                = "-"
)

type ChannelPayloadBuilder struct {
	payload *Payload
}

func NewChannelPayloadBuilder() *ChannelPayloadBuilder {
	return &ChannelPayloadBuilder{
		payload: &Payload{
			MessageType:    MessageTypeCH,
			Timestamp:      time.Now().Unix(),
			ChannelPayload: &ChannelPayload{},
		},
	}
}

func (b *ChannelPayloadBuilder) SetChannelAction(action ChannelActionType) *ChannelPayloadBuilder {
	b.payload.ChannelPayload.ChannelAction = action
	return b
}

func (b *ChannelPayloadBuilder) SetRequester(requester string) *ChannelPayloadBuilder {
	b.payload.ChannelPayload.Requester = requester
	return b
}

func (b *ChannelPayloadBuilder) SetChannelName(name string) *ChannelPayloadBuilder {
	b.payload.ChannelPayload.ChannelName = name
	return b
}

func (b *ChannelPayloadBuilder) SetChannelPassword(password string) *ChannelPayloadBuilder {
	b.payload.ChannelPayload.ChannelPassword = password
	return b
}

func (b *ChannelPayloadBuilder) SetChannelSize(size int) *ChannelPayloadBuilder {
	b.payload.ChannelPayload.ChannelSize = size
	return b
}

func (b *ChannelPayloadBuilder) AddOptionalArg(key string, value interface{}) *ChannelPayloadBuilder {
	if b.payload.ChannelPayload.OptionalChannelArgs == nil {
		b.payload.ChannelPayload.OptionalChannelArgs = &OptionalChannelArgs{}
	}

	switch key {
	case "status":
		b.payload.ChannelPayload.OptionalChannelArgs.Status = value.(Status)
	case "visibility":
		b.payload.ChannelPayload.OptionalChannelArgs.Visibility = value.(Visibility)
	case "message":
		b.payload.ChannelPayload.OptionalChannelArgs.Message = value.(string)
	case "reason":
		b.payload.ChannelPayload.OptionalChannelArgs.Reason = value.(string)
	case "channels":
		b.payload.ChannelPayload.OptionalChannelArgs.Channels = value.([]string)
	case "users":
		b.payload.ChannelPayload.OptionalChannelArgs.Users = value.([]string)
	case "target_user":
		b.payload.ChannelPayload.OptionalChannelArgs.TargetUser = value.(string)
	}
	return b
}

func (b *ChannelPayloadBuilder) Build() (*Payload, error) {
	if b.payload.ChannelPayload.Requester == "" {
		return nil, fmt.Errorf("requester is required")
	}
	if b.payload.ChannelPayload.ChannelAction == 0 {
		return nil, fmt.Errorf("channel action is required")
	}
	return b.payload, nil
}

func encodeCH(payload *Payload) string {
	var sb strings.Builder
	rp := payload.ChannelPayload

	timestamp := payload.Timestamp
	if timestamp == 0 {
		payload.Timestamp = time.Now().Unix() // Fallback if timestamp is somehow 0
	}

	sb.WriteString(fmt.Sprintf("%s|%d|%s|%s|", MessageTypeCH, payload.Timestamp, rp.ChannelAction, rp.Requester))

	if rp.ChannelName != "" {
		sb.WriteString(rp.ChannelName)
	} else {
		sb.WriteString(emptyChannelField)
	}
	sb.WriteString("|")

	if rp.ChannelPassword != "" {
		sb.WriteString(rp.ChannelPassword)
	} else {
		sb.WriteString(emptyChannelField)
	}
	sb.WriteString("|")

	if rp.ChannelSize != 0 {
		sb.WriteString(fmt.Sprintf("%d", rp.ChannelSize))
	} else {
		sb.WriteString(emptyChannelField)
	}

	if rp.OptionalChannelArgs != nil {
		optionalArgs := serializeChannelOptionalArgs(rp.OptionalChannelArgs)
		if optionalArgs != "" {
			sb.WriteString("|")
			sb.WriteString(optionalArgs)
		}
	}

	return sb.String()
}

func serializeChannelOptionalArgs(args *OptionalChannelArgs) string {
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

	if args.Channels != nil {
		optsParts = append(optsParts, "channels="+strings.Join(args.Channels, optionalUserAndChannelsSeparator))
	}

	if args.Users != nil {
		optsParts = append(optsParts, "users="+strings.Join(args.Users, optionalUserAndChannelsSeparator))
	}

	if args.TargetUser != "" {
		optsParts = append(optsParts, "target_user="+args.TargetUser)
	}

	return strings.Join(optsParts, optionalArgsSeparator)
}
