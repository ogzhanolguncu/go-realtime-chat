package main

import (
	"fmt"
	"log"
	"strings"
)

type RequestType int32

const messageSeparator string = "#"

const (
	GROUP_MESSAGE RequestType = iota
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
	MessageType      RequestType
	MessageSender    string
	MessageContent   string
	MessageRecipient string
}

// ParseMessage parses a message string formatted as "GROUP_MESSAGE#Name#Content#[Recipient - OPTIONAL]" into a Message object.
// It sanitizes the name and message fields, determines the message type (GROUP_MESSAGE, WHISPER, etc.),
// and returns a Message struct containing these sanitized values.
// If the message format is invalid or the message type is unknown, an error is returned.
func parseMessage(msg string) (Message, error) {
	parsedMessage := strings.Split(msg, messageSeparator)
	log.Printf("Parsed message: %v", parsedMessage)

	if len(parsedMessage) < 3 {
		return Message{}, fmt.Errorf("invalid message format: expected at least 3 parts, got %d", len(parsedMessage))
	}

	// Trim any whitespace around the parsed elements
	messageTypeStr := strings.TrimSpace(parsedMessage[0])
	messageSender := strings.TrimSpace(parsedMessage[1])
	messageContent := strings.TrimSpace(parsedMessage[2])

	// Check if there is a recipient
	var recipient string
	if len(parsedMessage) > 3 {
		recipient = strings.TrimSpace(parsedMessage[3])
	}

	messageType, ok := requestTypeMap[messageTypeStr]
	if !ok {
		log.Printf("Unknown message type: '%s'", messageTypeStr)
		return Message{}, fmt.Errorf("unknown message type: '%s'", messageTypeStr)
	}

	log.Printf("Returning values - messageType: %d, name: '%s', message: '%s', recipient: '%s'", messageType, messageSender, messageContent, recipient)
	return Message{messageType, messageSender, messageContent, recipient}, nil
}
