package chat_history

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func (ch *ChatHistory) SaveToDisk() error {
	dir := filepath.Join(rootDir(), "chat_history")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file := filepath.Join(dir, "history.json")
	data, err := json.Marshal(ch.messages)
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0644)
}

func (ch *ChatHistory) DeleteFromDisk() error {
	dir := filepath.Join(rootDir(), "chat_history")
	return os.RemoveAll(dir)
}

func rootDir() string {
	// This is a placeholder. In a real application, you'd return the actual root directory.
	return "."
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
