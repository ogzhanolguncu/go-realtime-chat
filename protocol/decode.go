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
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in MSG message")
		}

		timestamp := parts[1]
		sender := parts[2]
		lengthStr := parts[3]
		content := parts[4]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in MSG message: %v", err)
		}

		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in MSG message: %v", err)
		}
		if len(content) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in MSG message")
		}

		return Payload{Content: content, Timestamp: unixTimestamp, Sender: sender, MessageType: MessageTypeMSG}, nil

	case MessageTypeWSP:
		if len(parts) < 6 {
			return Payload{}, fmt.Errorf("insufficient parts in WSP message")
		}

		timestamp := parts[1]
		sender := parts[2]
		recipient := parts[3]
		lengthStr := parts[4]
		content := parts[5]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in MSG message: %v", err)
		}

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in WSP message: %v", err)
		}
		if len(content) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in WSP message")
		}

		return Payload{MessageType: MessageTypeWSP, Timestamp: unixTimestamp, Content: content, Sender: sender, Recipient: recipient}, nil
	// SYS|message_length|message_content|status \r\n status = "fail" | "success"
	case MessageTypeSYS:
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in SYS message")
		}

		timestamp := parts[1]
		lengthStr := parts[2]
		content := parts[3]
		status := parts[4]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in SYS message: %v", err)
		}

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in SYS message: %v", err)
		}
		if len(content) != expectedLength {
			return Payload{}, fmt.Errorf("message content length does not match expected length in SYS message")
		}
		return Payload{Content: content, Timestamp: unixTimestamp, MessageType: MessageTypeSYS, Status: status}, nil

	// USR|name_length|name_content|status\r\n status = "fail | "success"
	case MessageTypeUSR:
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in USR message")
		}

		timestamp := parts[1]
		lengthStr := parts[2]
		name := parts[3]
		status := parts[4]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in USR message: %v", err)
		}

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in USR message: %v", err)
		}
		if len(name) != expectedLength {
			return Payload{}, fmt.Errorf("name length does not match expected length in USR message")
		}
		return Payload{MessageType: MessageTypeUSR, Timestamp: unixTimestamp, Username: name, Status: status}, nil
	// ACT_USRS|active_user_length|active_user_array|status  status = "res" | "req"
	case MessageTypeACT_USRS:
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in ACT_USRS message")
		}

		timestamp := parts[1]
		lengthStr := parts[2]
		var activeUsers []string
		if parts[3] != "" {
			activeUsers = strings.Split(parts[3], ",")
		}
		status := parts[4]

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in ACT_USRS message: %v", err)
		}

		// Validate message length
		expectedLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid length format in ACT_USRS message: %v", err)
		}
		if len(activeUsers) != expectedLength {
			return Payload{}, fmt.Errorf("list length does not match expected length in ACT_USRS message")
		}
		return Payload{MessageType: MessageTypeACT_USRS, Timestamp: unixTimestamp, ActiveUsers: activeUsers, Status: status}, nil
	// HSTRY|timestamp|messages_array|status\r\n status = "res" | "req"
	case MessageTypeHSTRY:
		if len(parts) < 5 {
			return Payload{}, fmt.Errorf("insufficient parts in HSTRY message")
		}

		payloadDetails, messages := parseChatHistory(sanitizedMessage)
		parts := strings.Split(payloadDetails, Separator)
		timestamp := parts[1]
		status := parts[2]
		var parsedChatHistory []Payload
		for _, v := range messages {
			msg, err := DecodeMessage(v)
			if err != nil {
				continue
			}
			parsedChatHistory = append(parsedChatHistory, msg)
		}

		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return Payload{}, fmt.Errorf("invalid timestamp format in HSTRY message: %v", err)
		}
		return Payload{MessageType: MessageTypeHSTRY, Timestamp: unixTimestamp, DecodedChatHistory: parsedChatHistory, Status: status}, nil

	default:
		return Payload{}, fmt.Errorf("unsupported message type %s", messageType)
	}
}

func parseChatHistory(input string) (string, []string) {
	// Split the string by '|'
	parts := strings.Split(input, "|")

	// Construct the first part
	part1 := fmt.Sprintf("%s|%s|%s", parts[0], parts[1], parts[len(parts)-1])

	// Reconstruct the MSG parts with comma separated segments
	var msgParts []string
	for i := 2; i < len(parts)-1; i++ {
		if strings.Contains(parts[i], "MSG") {
			msgParts = append(msgParts, parts[i])
		} else {
			msgParts[len(msgParts)-1] = msgParts[len(msgParts)-1] + "|" + parts[i]
		}
	}
	part2 := strings.Split(strings.TrimSuffix(strings.Join(msgParts, "|"), ","), ",")

	return part1, part2
}
