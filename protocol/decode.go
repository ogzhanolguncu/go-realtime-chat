package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

func DecodeMessage(message string) (Payload, error) {
	sanitizedMessage := strings.TrimSpace(message) // Messages from server comes with \r\n, so we have to trim it

	parts := strings.Split(sanitizedMessage, Separator)
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

		return Payload{Content: messageContent, Sender: sender, ContentType: messageType}, nil

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

		return Payload{Content: messageContent, Sender: recipient, ContentType: messageType}, nil

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
		return Payload{Content: messageContent, ContentType: messageType, SysStatus: status}, nil

	default:
		return Payload{}, fmt.Errorf("unsupported message type %s", messageType)
	}
}
