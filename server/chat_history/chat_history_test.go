package chat_history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddMessage(t *testing.T) {
	chatHistory := NewChatHistory()
	chatHistory.AddMessage(
		"MSG|1721160403|Oz|3|aaa\r\n",
		"MSG|1721160403|Oz|4|aaaa\r\n",
		"MSG|1721160403|Oz|5|aaaaa\r\n")
	assert.Equal(t, 3, len(chatHistory.messages))
}

func TestFilterMessages(t *testing.T) {
	tests := []struct {
		name             string
		messages         []string
		user             string
		messageTypes     []string
		expectedMessages []string
	}{
		{
			name: "filter messages by user and message types",
			messages: []string{
				"MSG|1721160403|Oz|3|aaa\r\n",
				"MSG|1721160403|Oz|4|aaaa\r\n",
				"MSG|1721160403|Oz|5|aaaaa\r\n",
				"WSP|1721160403|Oz|John|9|Hello Oz.\r\n",
				"SYS|1721160403|3|Aaa|success",
				"WSP|1721160403|John|Oz|11|Hello John.\r\n",
				"WSP|1721160403|John|Frey|11|Hello Frey.\r\n",
			},
			user:         "Oz",
			messageTypes: []string{"MSG", "WSP"},
			expectedMessages: []string{
				"MSG|1721160403|Oz|3|aaa\r\n",
				"MSG|1721160403|Oz|4|aaaa\r\n",
				"MSG|1721160403|Oz|5|aaaaa\r\n",
				"WSP|1721160403|Oz|John|9|Hello Oz.\r\n",
				"WSP|1721160403|John|Oz|11|Hello John.\r\n",
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatHistory := NewChatHistory()
			chatHistory.AddMessage(tt.messages...)
			filteredHistory := chatHistory.GetHistory(tt.user, tt.messageTypes...)
			assert.Equal(t, tt.expectedMessages, filteredHistory)
		})
	}
}

func TestSaveToDisk(t *testing.T) {
	chatHistory := NewChatHistory()
	chatHistory.AddMessage(
		"MSG|1721160403|Oz|4|3333",
		"MSG|1721160403|Oz|4|4444",
		"MSG|1721160403|Oz|4|5555",
		"MSG|1721160403|Oz|4|6666",
		"MSG|1721160403|Oz|3|aaa",
		"MSG|1721160403|Oz|4|aaaa",
		"MSG|1721160403|Oz|5|aaaaa")
	err := chatHistory.SaveToDisk(4)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(rootDir(), "chat_history.txt"))

	// Read the file to verify its contents
	content, err := os.ReadFile(filepath.Join(rootDir(), "chat_history.txt"))
	require.NoError(t, err)
	lines := strings.Split(strings.ReplaceAll(string(content), "\r", ""), "\n")
	require.GreaterOrEqual(t, len(lines), 3) // Timestamp + 3 messages

	assert.Equal(t, "MSG|1721160403|Oz|3|aaa", lines[0])
	assert.Equal(t, "MSG|1721160403|Oz|4|aaaa", lines[1])
	assert.Equal(t, "MSG|1721160403|Oz|5|aaaaa", lines[2])
}

func TestDeleteFromDisk(t *testing.T) {
	chatHistory := NewChatHistory()
	chatHistory.AddMessage(
		"MSG|1721160403|Oz|3|aaa\r\n",
		"MSG|1721160403|Oz|4|aaaa\r\n",
		"MSG|1721160403|Oz|5|aaaaa\r\n")
	err := chatHistory.SaveToDisk(200)
	require.NoError(t, err)

	err = chatHistory.DeleteFromDisk()
	require.NoError(t, err)
	require.NoFileExists(t, filepath.Join(rootDir(), "chat_history.txt"))
}

func TestReadFromDiskToInMemory(t *testing.T) {
	// Setup
	chatHistory := NewChatHistory()
	testMessages := []string{
		"1234567890", // Timestamp
		"MSG|1721160403|Oz|3|aaa\r",
		"MSG|1721160403|Oz|4|aaaa\r",
		"MSG|1721160403|Oz|5|aaaaa\r",
	}

	testFile := filepath.Join(rootDir(), "chat_history.txt")
	err := os.WriteFile(testFile, []byte(strings.Join(testMessages, "\n")), 0644)
	require.NoError(t, err)

	defer os.Remove(testFile)

	err = chatHistory.ReadFromDiskToInMemory()
	require.NoError(t, err)

	require.Equal(t, len(testMessages), len(chatHistory.messages))
	for i, msg := range testMessages {
		assert.Equal(t, msg, chatHistory.messages[i])
	}

	os.Remove(testFile)
	err = chatHistory.ReadFromDiskToInMemory()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not read file")
}
