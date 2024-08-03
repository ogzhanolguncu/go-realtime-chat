package chat_history

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/elliotchance/pie/v2"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

const fileName = "chat_history.txt"

type ChatHistory struct {
	messages []string
	encoding bool
}

func NewChatHistory(encoding bool) *ChatHistory {
	return &ChatHistory{
		messages: []string{},
		encoding: encoding,
	}
}

func (ch *ChatHistory) AddMessage(messages ...string) {
	ch.messages = append(ch.messages, messages...)
}

// Get messages from memory if they are from requester user and contain allowed messageTypes
func (ch *ChatHistory) GetHistory(user string, messageTypes ...string) []string {
	blockMap := make(map[string]map[string]bool)

	log.Printf("Calling GetHistory")
	if len(ch.messages) == 0 {
		ch.ReadFromDiskToInMemory()
		log.Printf("Loaded %d messages from disk to memory", len(ch.messages))
	}

	for _, v := range ch.messages {
		decodedMsg, err := protocol.InitDecodeProtocol(ch.encoding)(v)
		if err != nil {
			continue // Skip undecodable messages
		}
		if decodedMsg.MessageType == protocol.MessageTypeBLCK_USR {
			if blockMap[decodedMsg.Sender] == nil {
				blockMap[decodedMsg.Sender] = make(map[string]bool)
			}
			blockMap[decodedMsg.Sender][decodedMsg.Recipient] = true
		}

	}

	msgs := pie.Filter(ch.messages, func(msg string) bool {
		decodedMsg, err := protocol.InitDecodeProtocol(ch.encoding)(msg)
		if err != nil {
			return false // Skip undecodable messages
		}

		if blockMap[decodedMsg.Sender][user] || blockMap[user][decodedMsg.Sender] {
			return false // Skip message if requesting user was blocked by sender or sender is blocked by requesting user
		}

		msgType := string(decodedMsg.MessageType)
		if !slices.Contains(messageTypes, msgType) {
			return false // Skip messages with unallowed types
		}
		//If message is not WSP return it
		if decodedMsg.MessageType != "WSP" {
			return true
		}
		//If message is WSP make sure recipient or sender is user
		return decodedMsg.Recipient == user || decodedMsg.Sender == user
	})
	log.Printf("Returning %d messages from GetHistory", len(msgs))
	return msgs
}

func (ch *ChatHistory) SaveToDisk(msgLimit int) error {
	filePath := filepath.Join(rootDir(), fileName)

	if checkIfFileExists(filePath) {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		lineCount, err := lineCounter(file)

		if err != nil {
			return fmt.Errorf("line count failed: %w", err)
		}

		if lineCount > msgLimit {
			ch.messages = pie.DropTop(ch.messages, msgLimit)
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove old file: %w", err)
			}
		}
	}

	// // Write new content
	return os.WriteFile(filePath, []byte(strings.Join(ch.messages, "\n")), 0644)
}

// Remove file from disk if it exists.
func (ch *ChatHistory) DeleteFromDisk() error {
	filePath := filepath.Join(rootDir(), fileName)
	return os.Remove(filePath)
}

// Read chat_history.txt from disk to in-memory.
func (ch *ChatHistory) ReadFromDiskToInMemory() error {
	filePath := filepath.Join(rootDir(), fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	// Split the data by newline character
	ch.messages = strings.Split(string(data), "\n")

	// Remove empty strings that may result from splitting
	ch.messages = removeEmpty(ch.messages)
	log.Printf("Reading messages from disk to memory. Count is: %d", len(ch.messages))
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

// https://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func checkIfFileExists(name string) bool {
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
