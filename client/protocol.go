package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
)

// Group/General Message (MSG): MSG|sender|message_length|message_content\r\n
// Whisper/DM Message (WSP): 	WSP|recipient|message_length|message_content\r\n
// System Notice (SYS): 		SYS|message_length|message_content|status \r\n status = "fail" | "success"
// Error Message (ERR): 		ERR|message_length|error_message\r\n
const separator = "|"

const (
	MessageTypeMSG = "MSG"
	MessageTypeWSP = "WSP"
	MessageTypeSYS = "SYS"
	MessageTypeERR = "ERR"
)

type Payload struct {
	content     string
	contentType string
	sender      string
	sysStatus   string
}

func decodeMessage(message string) (Payload, error) {
	sanitizedMessage := strings.TrimSpace(message) // Messages from server comes with \r\n, so we have to trim it

	parts := strings.Split(sanitizedMessage, separator)
	messageType := parts[0]

	switch messageType {
	case MessageTypeMSG:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in MSG message")
		}

		sender := parts[1]
		lengthStr := parts[2]
		messageContent := parts[3]
		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in MSG message: %v", err)
		}
		if len(messageContent) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in MSG message")
		}

		return Payload{content: messageContent, sender: sender, contentType: messageType}, nil

	case MessageTypeWSP:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in WSP message")
		}

		recipient := parts[1]
		lengthStr := parts[2]
		messageContent := parts[3]

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in WSP message: %v", err)
		}
		if len(messageContent) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in WSP message")
		}

		return Payload{content: messageContent, sender: recipient, contentType: messageType}, nil

	case MessageTypeSYS:
		if len(parts) < 3 {
			return Payload{}, fmt.Errorf("insufficient parts in SYS message")
		}

		lengthStr := parts[1]
		messageContent := parts[2]
		status := parts[3]

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in SYS message: %v", err)
		}
		if len(messageContent) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in SYS message")
		}
		return Payload{content: messageContent, contentType: messageType, sysStatus: status}, nil

	default:
		return Payload{}, fmt.Errorf("unsupported message type %s", messageType)
	}
}

func colorifyAndFormatContent(payload Payload) {
	switch payload.contentType {
	case MessageTypeSYS:
		fmtedMsg := fmt.Sprintf("System: %s\n", payload.content)
		if payload.sysStatus == "fail" {
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Red)) // Red fail messages
		} else {
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Cyan)) // Cyan for system messages
		}
	case MessageTypeWSP:
		fmtedMsg := fmt.Sprintf("Whisper from %s: %s\n", payload.sender, payload.content)
		fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Purple)) // Purple for whisper messages
	default:
		fmtedMsg := fmt.Sprintf("%s: %s\n", payload.sender, payload.content)
		fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Blue)) // Blue for group messages
	}
}

func handleIncomingMessage(content string, cb func()) {
	payload, err := decodeMessage(content)
	if err != nil {
		fmt.Print(color.ColorifyWithTimestamp(err.Error(), color.Red))
	}
	colorifyAndFormatContent(payload)
	cb()
}
