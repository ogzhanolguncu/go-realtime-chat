package protocol

import (
	"fmt"
	"strings"
	"time"
)

func EncodeMessage(payload Payload) string {
	length := len(payload.Content)

	switch payload.MessageType {
	// MSG|sender|message_length|message_content\r\n
	case MessageTypeMSG:
		return fmt.Sprintf("%s|%d|%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, length, payload.Content)
	// WSP|sender|recipient|message_length|message_content\r\n
	case MessageTypeWSP:
		return fmt.Sprintf("%s|%d|%s|%s|%d|%s\r\n", payload.MessageType, time.Now().Unix(), payload.Sender, payload.Recipient, length, payload.Content)
	// SYS|message_length|message_content|status \r\n status = "fail" | "success"
	case MessageTypeSYS:
		if payload.Status == "" {
			return fmt.Sprintf("%s|%d|%d|%s\r\n", payload.MessageType, time.Now().Unix(), length, payload.Content)
		}
		return fmt.Sprintf("%s|%d|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), length, payload.Content, payload.Status)
	// ERR|message_length|error_message\r\n
	case MessageTypeERR:
		return fmt.Sprintf("%s|%d|%d|%s\r\n", payload.MessageType, time.Now().Unix(), length, payload.Content)
	// USR|name_length|name_content|status\r\n status = "fail | "success"
	case MessageTypeUSR:
		nameLength := len(payload.Username)
		return fmt.Sprintf("%s|%d|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), nameLength, payload.Username, payload.Status)
	// ACT_USRS|active_user_length|active_user_array|status\r\n status = "req" | "res"
	case MessageTypeACT_USRS:
		activeUserLen := len(payload.ActiveUsers)
		return fmt.Sprintf("%s|%d|%d|%s|%s\r\n", payload.MessageType, time.Now().Unix(), activeUserLen, strings.Join(payload.ActiveUsers, ","), payload.Status)
	// ERR|message_length|error_message\r\n
	default:
		return fmt.Sprintf("ERR|%d|Invalid message type\r\n", len("Invalid message type"))
	}
}
