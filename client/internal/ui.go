package internal

import (
	"fmt"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func ColorifyAndFormatContent(payload protocol.Payload) string {
	colorMap := map[protocol.MessageType]string{
		protocol.MessageTypeSYS:      "yellow",
		protocol.MessageTypeWSP:      "magenta",
		protocol.MessageTypeUSR:      "cyan",
		protocol.MessageTypeMSG:      "blue",
		protocol.MessageTypeACT_USRS: "green",
	}

	var message string
	color := colorMap[payload.MessageType]
	timestamp := time.Now().Format("15:04")

	switch payload.MessageType {
	case protocol.MessageTypeSYS:
		message = fmt.Sprintf("System: %s", payload.Content)
		if payload.Status == "fail" {
			color = "red"
		}
	case protocol.MessageTypeWSP:
		message = fmt.Sprintf("Whisper from %s: %s", payload.Sender, payload.Content)
	case protocol.MessageTypeUSR:
		if payload.Status == "success" {
			return ""
		} else {
			message = payload.Username
			color = "red"
		}
	default:
		message = fmt.Sprintf("%s: %s", payload.Sender, payload.Content)
	}

	if message != "" {
		return fmt.Sprintf("[%s %s](fg:%s)", timestamp, message, color)
	}

	return ""
}
