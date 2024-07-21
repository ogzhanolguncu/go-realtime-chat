package protocol

import (
	"fmt"
	"strings"
	"time"
)

func EncodeMessage(payload Payload) string {
	switch payload.MessageType {
	case MessageTypeMSG:
		return fmt.Sprintf("%s|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, payload.Content)
	case MessageTypeWSP:
		return fmt.Sprintf("%s|%d|%s|%s|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, payload.Recipient, payload.Content)
	case MessageTypeSYS:
		if payload.Status == "" {
			return fmt.Sprintf("%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Content)
		}
		return fmt.Sprintf("%s|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Content, payload.Status)
	case MessageTypeERR:
		return fmt.Sprintf("%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Content)
	case MessageTypeUSR:
		return fmt.Sprintf("%s|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Username, payload.Status)
	case MessageTypeACT_USRS:
		return fmt.Sprintf("%s|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), strings.Join(payload.ActiveUsers, ","), payload.Status)
	case MessageTypeHSTRY:
		return fmt.Sprintf("%s|%d|%s|%s|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, strings.Join(payload.EncodedChatHistory, ","), payload.Status)
	case MessageTypeENC:
		return fmt.Sprintf("%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.EncryptedKey)
	default:
		return fmt.Sprintf("ERR|%d|Invalid message type\r\n", len("Invalid message type"))
	}
}
