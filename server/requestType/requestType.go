package RequestType

import (
	"fmt"
	"log"
	"strings"
)

type RequestType int32

const (
	messageSeparator string = "#"

	GROUP_MESSAGE RequestType = iota + 1
	WHISPER
	CHAT_HISTORY
	QUIT
)

var requestTypeMap = map[string]RequestType{
	"GROUP_MESSAGE": GROUP_MESSAGE,
	"WHISPER":       WHISPER,
	"CHAT_HISTORY":  CHAT_HISTORY,
}

type Message struct {
	MessageType RequestType
	Name        string
	Message     string
}

// ParseMessage parses a message string formatted as "GROUP_MESSAGE#Name#Content" into a Message object.
// It sanitizes the name and message fields, determines the message type (GROUP_MESSAGE, WHISPER, etc.),
// and returns a Message struct containing these sanitized values.
// If the message format is invalid or the message type is unknown, an error is returned.
func ParseMessage(msg string) (Message, error) {
	parsedMessage := strings.Split(msg, messageSeparator)
	log.Printf("Parsed message: %v", parsedMessage)

	if len(parsedMessage) < 3 {
		return Message{}, fmt.Errorf("invalid message format")
	}

	// Trim any whitespace around the parsed elements
	messageTypeStr := strings.TrimSpace(parsedMessage[0])
	name := strings.TrimSpace(parsedMessage[1])
	message := strings.TrimSpace(parsedMessage[2])

	messageType, ok := requestTypeMap[messageTypeStr]
	if !ok {
		log.Printf("Unknown message type: '%s'", messageTypeStr)
		return Message{}, fmt.Errorf("unknown message type: '%s'", messageTypeStr)
	}

	log.Printf("Returning values - messageType: %d, name: '%s', message: '%s'", messageType, name, message)
	return Message{messageType, name, message}, nil
}
