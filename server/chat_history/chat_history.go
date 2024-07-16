package chat_history

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
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
	file := filepath.Join(rootDir(), "chat_history.txt")
	ch.messages = prepend(ch.messages, strconv.FormatInt(time.Now().Unix(), 10)+"\n")
	return os.WriteFile(file, []byte(strings.Join(ch.messages, "")), 0644)
}

func (ch *ChatHistory) DeleteFromDisk() error {
	dir := filepath.Join(rootDir(), "chat_history")
	return os.RemoveAll(dir)
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
