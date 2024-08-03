package protocol

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

func InitDecodeProtocol(encoding bool) func(message string) (Payload, error) {
	return func(message string) (Payload, error) {
		return decodeProtocol(encoding, message)
	}
}

func decodeProtocol(encoding bool, message string) (Payload, error) {
	if encoding {
		decodedMsg, err := base64.StdEncoding.DecodeString(message)
		message = string(decodedMsg)
		if err != nil {
			return Payload{}, fmt.Errorf("something went wrong when decoding message: %v", err)
		}
	}

	sanitizedMessage := strings.TrimSpace(string(message)) // Messages from server comes with \r\n, so we have to trim it

	parts := strings.Split(sanitizedMessage, Separator)
	messageType := parts[0]

	switch MessageType(messageType) {
	case MessageTypeMSG:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in MSG message")
		}

		timestamp := parts[1]
		sender := parts[2]
		content := parts[3]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in MSG message: %v", err)
		}

		return Payload{Content: content, Timestamp: unixTimestamp, Sender: sender, MessageType: MessageTypeMSG}, nil

	case MessageTypeWSP:
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in WSP message")
		}

		timestamp := parts[1]
		sender := parts[2]
		recipient := parts[3]
		content := parts[4]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in WSP message: %v", err)
		}
		return Payload{MessageType: MessageTypeWSP, Timestamp: unixTimestamp, Content: content, Sender: sender, Recipient: recipient}, nil
	case MessageTypeBLCK_USR:
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in BLCK_USR message")
		}

		timestamp := parts[1]
		sender := parts[2]
		recipient := parts[3]
		content := parts[4]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in BLCK_USR message: %v", err)
		}
		return Payload{MessageType: MessageTypeBLCK_USR, Timestamp: unixTimestamp, Content: content, Sender: sender, Recipient: recipient}, nil
	case MessageTypeSYS:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in SYS message")
		}

		timestamp := parts[1]
		content := parts[2]
		status := parts[3]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in SYS message: %v", err)
		}

		return Payload{Content: content, Timestamp: unixTimestamp, MessageType: MessageTypeSYS, Status: status}, nil

	case MessageTypeUSR:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in USR message")
		}

		timestamp := parts[1]
		name := parts[2]
		status := parts[3]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in USR message: %v", err)
		}

		return Payload{MessageType: MessageTypeUSR, Timestamp: unixTimestamp, Username: name, Status: status}, nil
	case MessageTypeACT_USRS:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in ACT_USRS message")
		}

		timestamp := parts[1]
		var activeUsers []string
		if parts[2] != "" {
			activeUsers = strings.Split(parts[2], ",")
		}
		status := parts[3]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in ACT_USRS message: %v", err)
		}

		return Payload{MessageType: MessageTypeACT_USRS, Timestamp: unixTimestamp, ActiveUsers: activeUsers, Status: status}, nil
	case MessageTypeHSTRY:
		if len(parts) < 4 {
			return Payload{}, fmt.Errorf("insufficient parts in HSTRY message")
		}

		payloadDetails, messages := parseChatHistory(encoding, sanitizedMessage)
		parts = strings.Split(payloadDetails, Separator)

		timestamp := parts[1]
		requester := parts[2]
		status := parts[3]

		var parsedChatHistory []Payload
		for _, v := range messages {
			msg, err := decodeProtocol(encoding, v)
			if err != nil {
				continue
			}
			parsedChatHistory = append(parsedChatHistory, msg)
		}

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in HSTRY message: %v", err)
		}
		return Payload{MessageType: MessageTypeHSTRY, Sender: requester, Timestamp: unixTimestamp, DecodedChatHistory: parsedChatHistory, Status: status}, nil

	case MessageTypeENC:
		if len(parts) < 3 {
			return Payload{}, fmt.Errorf("insufficient parts in ENC message")
		}
		timestamp := parts[1]
		encryptedKey := parts[2]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in USR message: %v", err)
		}
		return Payload{MessageType: MessageTypeENC, EncryptedKey: encryptedKey, Timestamp: unixTimestamp}, nil
	default:
		return Payload{}, fmt.Errorf("unsupported message type %s", messageType)
	}
}

func parseChatHistory(encoding bool, input string) (string, []string) {
	parts := strings.Split(input, "|")
	if len(parts) < 5 {
		return input, nil // Return original input if it doesn't have enough parts
	}

	// Construct the first part (HSTRY metadata)
	part1 := fmt.Sprintf("%s|%s|%s|%s", parts[0], parts[1], parts[2], parts[len(parts)-1])

	// Get the messages
	messagesPart := strings.Join(parts[3:len(parts)-1], "|")
	messages := strings.Split(messagesPart, ",")

	// If not encoding (i.e., plain text), we're done
	if !encoding {
		return part1, messages
	}

	// For base64 encoding (not implemented here)
	// You would add your base64 decoding logic here
	return part1, messages
}
