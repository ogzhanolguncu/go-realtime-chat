package protocol

import (
	"fmt"
	"strings"
	"time"
)

func EncodeMessage(payload Payload) string {
	length := len(payload.Content)

	switch payload.MessageType {
	case MessageTypeMSG:
		return fmt.Sprintf("%s|%d|%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, length, payload.Content)
	case MessageTypeWSP:
		return fmt.Sprintf("%s|%d|%s|%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, payload.Recipient, length, payload.Content)
	case MessageTypeSYS:
		if payload.Status == "" {
			return fmt.Sprintf("%s|%d|%d|%s\r\n", payload.MessageType, time.Now().Unix(), length, payload.Content)
		}
		return fmt.Sprintf("%s|%d|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), length, payload.Content, payload.Status)
	case MessageTypeERR:
		return fmt.Sprintf("%s|%d|%d|%s\r\n", payload.MessageType, time.Now().Unix(), length, payload.Content)
	case MessageTypeUSR:
		nameLength := len(payload.Username)
		return fmt.Sprintf("%s|%d|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), nameLength, payload.Username, payload.Status)
	case MessageTypeACT_USRS:
		activeUserLen := len(payload.ActiveUsers)
		return fmt.Sprintf("%s|%d|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), activeUserLen, strings.Join(payload.ActiveUsers, ","), payload.Status)
	case MessageTypeHSTRY:
		return fmt.Sprintf("%s|%d|%s|%s|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, strings.Join(payload.EncodedChatHistory, ","), payload.Status)
	case MessageTypeENC:
		return fmt.Sprintf("%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.EncryptedKey)
	default:
		return fmt.Sprintf("ERR|%d|Invalid message type\r\n", len("Invalid message type"))
	}
}
