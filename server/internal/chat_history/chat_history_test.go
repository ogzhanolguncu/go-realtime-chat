package chat_history

import (
	"os"
	"strings"
	"testing"

	"github.com/ogzhanolguncu/go-chat/server/internal/block_user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dbPath = "./chat_test.db"

func TestGetHistoryWithBlockedUser(t *testing.T) {
	ch, err := NewChatHistory(false, dbPath)
	require.NoError(t, err)

	bm, err := block_user.NewBlockUserManager(dbPath)
	require.NoError(t, err)

	require.NoError(t, bm.BlockUser("Oz", "Frey"))

	tests := []struct {
		testName      string
		inputMessages []string
		requester     string
		output        []string
	}{
		{
			inputMessages: []string{
				"MSG|1724188406|Oz|Hello there group chat\r\n",
				"MSG|1724188406|John|Hello there group chat\r\n",
				"MSG|1724188406|Frey|Hello there group chat\r\n",
				"WSP|1724188406|Oz|John|John this is a whisper\r\n",
				"SYS|1724188406|Sys messages|\r\n",
				"ACT_USRS|1724188406|test_value|status\r\n",
				"WSP|1724188406|John|Oz|I know\r\n"},
			requester: "Oz",
			testName:  "Requested messages by Oz",
			output: []string{
				"MSG|1724188406|Oz|Hello there group chat",
				"MSG|1724188406|John|Hello there group chat",
				"WSP|1724188406|Oz|John|John this is a whisper",
				"WSP|1724188406|John|Oz|I know",
			},
		},
		{
			inputMessages: []string{
				"WSP|1724188406|John|Frey|I know\r\n"},
			requester: "John",
			testName:  "Requested messages by John",
			output: []string{
				"MSG|1724188406|Oz|Hello there group chat",
				"MSG|1724188406|John|Hello there group chat",
				"MSG|1724188406|Frey|Hello there group chat\r\n",
				"WSP|1724188406|Oz|John|John this is a whisper",
				"WSP|1724188406|John|Oz|I know",
				"WSP|1724188406|John|Frey|I know",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			for _, inputMessage := range tt.inputMessages {
				err := ch.AddMessage(inputMessage)
				require.NoError(t, err)
			}

			messages, err := ch.GetHistory(tt.requester, "MSG", "WSP")
			require.NoError(t, err)

			assert.Equal(t, len(tt.output), len(messages), "Number of messages doesn't match")

			for i, expectedMsg := range tt.output {
				actualMsg := messages[i]
				// Trim any trailing whitespace (including newlines) for comparison
				assert.Equal(t, strings.TrimSpace(expectedMsg), strings.TrimSpace(actualMsg), "Message mismatch at index %d", i)
			}
		})
	}
	assert.NoError(t, os.Remove(dbPath))
}
