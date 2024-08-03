package protocol

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	errInvalidFormat      = "invalid %s format: %s"
	errMissingTimestamp   = "missing timestamp separator"
	errMissingSender      = "missing sender separator"
	errMissingName        = "missing name separator"
	errMissingRecipient   = "missing recipient separator"
	errMissingMessage     = "missing message separator"
	errMissingRequester   = "missing requester separator"
	errMissingActiveUsers = "missing rawActiveUsers separator"
	errMissingContent     = "missing content separator"
	errInvalidTimestamp   = "invalid timestamp format: %v"
	errUnsupportedMsgType = "unsupported message type %s"
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
	messageType, parts, found := strings.Cut(sanitizedMessage, "|")
	if !found {
		return Payload{}, fmt.Errorf("message has missing parts")
	}

	switch MessageType(messageType) {
	case MessageTypeMSG:
		timestamp, sender, content, err := parseMSG(parts)
		if err != nil {
			return Payload{}, err
		}
		return Payload{
			Content:     content,
			Timestamp:   timestamp,
			Sender:      sender,
			MessageType: MessageTypeMSG,
		}, nil
	case MessageTypeWSP:
		timestamp, sender, recipient, content, err := parseWSP(parts)
		if err != nil {
			return Payload{}, err
		}
		return Payload{MessageType: MessageTypeWSP, Timestamp: timestamp, Content: content, Sender: sender, Recipient: recipient}, nil
	case MessageTypeSYS:
		timestamp, content, status, err := parseSYS(parts)
		if err != nil {
			return Payload{}, err
		}
		return Payload{Content: content, Timestamp: timestamp, MessageType: MessageTypeSYS, Status: status}, nil

	case MessageTypeUSR:
		timestamp, name, status, err := parseUSR(parts)
		if err != nil {
			return Payload{}, err
		}

		return Payload{MessageType: MessageTypeUSR, Timestamp: timestamp, Username: name, Status: status}, nil
	case MessageTypeACT_USRS:
		timestamp, activeUsers, status, err := parseACT_USRS(parts)
		if err != nil {
			return Payload{}, err
		}

		return Payload{MessageType: MessageTypeACT_USRS, Timestamp: timestamp, ActiveUsers: activeUsers, Status: status}, nil
	case MessageTypeHSTRY:
		timestamp, requester, status, parsedChatHistory, err := parseHSTRY(parts, encoding)
		if err != nil {
			return Payload{}, err
		}
		return Payload{MessageType: MessageTypeHSTRY, Sender: requester, Timestamp: timestamp, DecodedChatHistory: parsedChatHistory, Status: status}, nil

	case MessageTypeBLCK_USR:
		timestamp, sender, recipient, content, err := parseBLCK_USR(parts)
		if err != nil {
			return Payload{}, err
		}
		return Payload{MessageType: MessageTypeBLCK_USR, Timestamp: timestamp, Content: content, Sender: sender, Recipient: recipient}, nil
	default:
		return Payload{}, fmt.Errorf("unsupported message type %s", messageType)
	}
}

func parseMSG(msg string) (timestamp int64, sender, content string, err error) {
	timestampStr, rest, found := strings.Cut(msg, "|")
	if !found {
		return 0, "", "", fmt.Errorf(errInvalidFormat, "MSG", errMissingTimestamp)
	}

	timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, "", "", fmt.Errorf(errInvalidTimestamp, err)
	}

	sender, content, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", fmt.Errorf(errInvalidFormat, "MSG", errMissingSender)
	}

	return timestamp, sender, content, nil
}
func parseWSP(msg string) (timestamp int64, sender, recipient, content string, err error) {
	timestampStr, rest, found := strings.Cut(msg, "|")
	if !found {
		return 0, "", "", "", fmt.Errorf(errInvalidFormat, "WSP", errMissingTimestamp)
	}
	timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, "", "", "", fmt.Errorf(errInvalidTimestamp, err)
	}
	sender, rest, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", "", fmt.Errorf(errInvalidFormat, "WSP", errMissingSender)
	}
	recipient, content, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", "", fmt.Errorf(errInvalidFormat, "WSP", errMissingRecipient)
	}

	return timestamp, sender, recipient, content, nil
}

func parseSYS(msg string) (timestamp int64, content, status string, err error) {
	timestampStr, rest, found := strings.Cut(msg, "|")
	if !found {
		return 0, "", "", fmt.Errorf(errInvalidFormat, "SYS", errMissingTimestamp)
	}
	timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, "", "", fmt.Errorf(errInvalidTimestamp, err)
	}
	content, status, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", fmt.Errorf(errInvalidFormat, "SYS", errMissingContent)
	}

	return timestamp, content, status, nil
}

func parseUSR(msg string) (timestamp int64, name, status string, err error) {
	timestampStr, rest, found := strings.Cut(msg, "|")
	if !found {
		return 0, "", "", fmt.Errorf(errInvalidFormat, "USR", errMissingTimestamp)
	}
	timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, "", "", fmt.Errorf(errInvalidTimestamp, err)
	}
	name, status, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", fmt.Errorf(errInvalidFormat, "USR", errMissingName)
	}

	return timestamp, name, status, nil
}

func parseACT_USRS(msg string) (timestamp int64, activeUsers []string, status string, err error) {
	timestampStr, rest, found := strings.Cut(msg, "|")
	if !found {
		return 0, nil, "", fmt.Errorf(errInvalidFormat, "ACT_USRS", errMissingTimestamp)
	}
	timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, nil, "", fmt.Errorf(errInvalidTimestamp, err)
	}

	rawActiveUsers, status, found := strings.Cut(rest, "|")
	if !found {
		return 0, nil, "", fmt.Errorf(errInvalidFormat, "ACT_USRS", errMissingActiveUsers)
	}

	if rawActiveUsers != "" {
		activeUsers = strings.Split(rawActiveUsers, ",")
	}

	return timestamp, activeUsers, status, nil
}

func parseHSTRY(msg string, encoding bool) (timestamp int64, requester, status string, parsedChatHistory []Payload, err error) {
	timestampStr, rest, found := strings.Cut(msg, "|")
	if !found {
		return 0, "", "", nil, fmt.Errorf(errInvalidFormat, "HSTRY", errMissingTimestamp)
	}
	timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, "", "", nil, fmt.Errorf(errInvalidTimestamp, err)
	}

	requester, rest, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", nil, fmt.Errorf(errInvalidFormat, "HSTRY", errMissingRequester)
	}

	// Picks only the status part, last part from string
	status = rest[strings.LastIndex(rest, "|")+1:]
	// Uses  "|" + last part to cutsuffix to end up only with messages
	messages, found := strings.CutSuffix(rest, Separator+status)
	if !found {
		return 0, "", "", nil, fmt.Errorf(errInvalidFormat, "HSTRY", messages)
	}

	for _, v := range strings.Split(messages, ",") {
		msg, err := decodeProtocol(encoding, v)
		if err != nil {
			continue
		}
		parsedChatHistory = append(parsedChatHistory, msg)
	}

	return timestamp, requester, status, parsedChatHistory, nil
}

func parseBLCK_USR(msg string) (timestamp int64, sender, recipient, content string, err error) {
	timestampStr, rest, found := strings.Cut(msg, "|")
	if !found {
		return 0, "", "", "", fmt.Errorf(errInvalidFormat, "BLCK_USR", errMissingTimestamp)
	}
	timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, "", "", "", fmt.Errorf(errInvalidTimestamp, err)
	}
	sender, rest, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", "", fmt.Errorf(errInvalidFormat, "BLCK_USR", errMissingSender)
	}
	recipient, content, found = strings.Cut(rest, "|")
	if !found {
		return 0, "", "", "", fmt.Errorf(errInvalidFormat, "BLCK_USR", recipient)
	}

	return timestamp, sender, recipient, content, nil
}
