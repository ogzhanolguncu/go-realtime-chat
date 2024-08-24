package protocol

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

func InitEncodeProtocol(isBase64 bool) func(payload Payload) string {
	return func(payload Payload) string {
		return encodeProtocol(isBase64, payload)
	}
}

// TODO: start adding room payload here, then tests, CREATE, JOIN, MESSAGE

func encodeProtocol(isBase64 bool, payload Payload) string {
	var sb strings.Builder

	writeCommonPrefix := func(messageType MessageType) {
		timestamp := payload.Timestamp
		if timestamp == 0 {
			timestamp = time.Now().Unix()
		}
		sb.WriteString(fmt.Sprintf("%s|%d|", messageType, timestamp))
	}

	messageFormatters := map[MessageType]func(){
		MessageTypeMSG: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(fmt.Sprintf("%s|%s", payload.Sender, payload.Content))
		},
		MessageTypeWSP: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(fmt.Sprintf("%s|%s|%s", payload.Sender, payload.Recipient, payload.Content))
		},
		MessageTypeBLCK_USR: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(fmt.Sprintf("%s|%s|%s", payload.Sender, payload.Recipient, payload.Content))
		},
		MessageTypeSYS: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(payload.Content)
			if payload.Status != "" {
				sb.WriteString(fmt.Sprintf("|%s", payload.Status))
			}
		},
		MessageTypeUSR: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(fmt.Sprintf("%s|%s|%s", payload.Username, payload.Password, payload.Status))
		},
		MessageTypeACT_USRS: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(fmt.Sprintf("%s|%s", strings.Join(payload.ActiveUsers, ","), payload.Status))
		},
		MessageTypeHSTRY: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(fmt.Sprintf("%s|%s|%s", payload.Sender, strings.Join(payload.EncodedChatHistory, ","), payload.Status))
		},
		MessageTypeENC: func() {
			writeCommonPrefix(payload.MessageType)
			sb.WriteString(payload.EncryptedKey)
		},

		MessageTypeROOM: func() {
			// There four mandatory fields those are: MessageType, Timestamp, RoomAction and Requester rest of them are interchangeable
			writeCommonPrefix(payload.MessageType)
			var parts []string

			parts = append(parts, payload.RoomPayload.RoomAction.String())
			parts = append(parts, payload.RoomPayload.Requester)

			if payload.RoomPayload.RoomName != nil && *payload.RoomPayload.RoomName != "" {
				parts = append(parts, *payload.RoomPayload.RoomName)
			}

			if payload.RoomPayload.RoomPassword != nil && *payload.RoomPayload.RoomPassword != "" {
				parts = append(parts, *payload.RoomPayload.RoomPassword)
			}

			if payload.RoomPayload.RoomSize != nil && *payload.RoomPayload.RoomSize != 0 {
				parts = append(parts, fmt.Sprintf("%d", *payload.RoomPayload.RoomSize))
			}

			if payload.RoomPayload.OptionalRoomArgs != nil {
				optionalArgs := serializeRoomOptionalArgs(payload.RoomPayload.OptionalRoomArgs)
				if optionalArgs != "" {
					parts = append(parts, optionalArgs)
				}

			}

			sb.WriteString(strings.Join(parts, "|"))
		},
	}

	if formatter, ok := messageFormatters[payload.MessageType]; ok {
		formatter()
	} else {
		sb.WriteString("ERR|Invalid message type")
	}

	sb.WriteString("\r\n")
	if isBase64 {
		return base64.StdEncoding.EncodeToString([]byte(sb.String())) + "\r\n"
	}
	return sb.String()

}

// If args are empty it will return empty string
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
		optsParts = append(optsParts, "rooms="+strings.Join(args.Rooms, ","))
	}

	if args.Users != nil {
		optsParts = append(optsParts, "users="+strings.Join(args.Users, ","))
	}

	if args.TargetUser != "" {
		optsParts = append(optsParts, "target_user="+args.TargetUser)
	}

	return strings.Join(optsParts, ";")
}
