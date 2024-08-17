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
func encodeProtocol(isBase64 bool, payload Payload) string {
	var sb strings.Builder
	timestamp := time.Now().Unix()

	writeCommonPrefix := func(messageType MessageType) {
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
