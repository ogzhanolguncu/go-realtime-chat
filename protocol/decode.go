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

	switch MessageType(messageType) {
	case MessageTypeMSG:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in MSG message")
		}

		sender := parts[1]
		lengthStr := parts[2]
		content := parts[3]

		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in MSG message: %v", err)
		}
		if len(content) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in MSG message")
		}

		return Payload{Content: content, Sender: sender, MessageType: MessageTypeMSG}, nil

	case MessageTypeWSP:
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in WSP message")
		}

		sender := parts[1]
		recipient := parts[2]
		lengthStr := parts[3]
		content := parts[4]

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in WSP message: %v", err)
		}
		if len(content) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in WSP message")
		}

		return Payload{MessageType: MessageTypeWSP, Content: content, Sender: sender, Recipient: recipient}, nil
	// SYS|message_length|message_content|status \r\n status = "fail" | "success"
	case MessageTypeSYS:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in SYS message")
		}

		lengthStr := parts[1]
		content := parts[2]
		status := parts[3]

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in SYS message: %v", err)
		}
		if len(content) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in SYS message")
		}
		return Payload{Content: content, MessageType: MessageTypeSYS, Status: status}, nil

	// USR|name_length|name_content|status\r\n status = "fail | "success"
	case MessageTypeUSR:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in USR message")
		}

		lengthStr := parts[1]
		name := parts[2]
		status := parts[3]

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in USR message: %v", err)
		}
		if len(name) != expectedLength {
			return Payload{}, fmt.Errorf("name length does not match expected length in USR message")
		}
		return Payload{MessageType: MessageTypeUSR, Username: name, Status: status}, nil

	default:
		return Payload{}, fmt.Errorf("unsupported message type %s", messageType)
	}
}
