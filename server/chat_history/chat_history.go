package chat_history

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type ChatHistory struct {
	messages []string
}

func NewChatHistory() *ChatHistory {
	return &ChatHistory{
		messages: []string{},
	}
}

func (ch *ChatHistory) AddMessage(messages ...string) {
	ch.messages = append(ch.messages, messages...)
}

// Get messages from memory if they are from requester user and contain allowed messageTypes
func (ch *ChatHistory) GetHistory(user string, messageTypes ...string) []string {
	var filteredMessages []string
	for _, msg := range ch.messages {
		parts := strings.Split(msg, "|")
		if len(parts) < 3 {
			continue
		}

		msgType := parts[0]
		msgUser := parts[2]

		if contains(messageTypes, msgType) && (user == "" || msgUser == user || (msgType == "WSP" && len(parts) > 3 && parts[3] == user)) {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return filteredMessages
}

// Save to disk. And add a timestamp to first line so we can check and delete if its older than a day.
func (ch *ChatHistory) SaveToDisk() error {
	file := filepath.Join(rootDir(), "chat_history.txt")
	return os.WriteFile(file, []byte(strings.Join(ch.messages, "\n")), 0644)
}

// Remove file from disk if it exists.
func (ch *ChatHistory) DeleteFromDisk() error {
	filePath := filepath.Join(rootDir(), "chat_history.txt")
	return os.Remove(filePath)
}

// Read chat_history.txt from disk to in-memory.
func (ch *ChatHistory) ReadFromDiskToInMemory() error {
	filePath := filepath.Join(rootDir(), "chat_history.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	// Split the data by newline character
	ch.messages = strings.Split(string(data), "\n")

	// Remove empty strings that may result from splitting
	ch.messages = removeEmpty(ch.messages)

	return nil
}

// Helper function to remove empty strings from a slice
func removeEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func prepend[T any](slice []T, elems ...T) []T {
	return append(elems, slice...)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
