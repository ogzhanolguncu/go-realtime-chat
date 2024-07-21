package protocol

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodeGeneralMessage(t *testing.T) {
	t.Run("should encode general message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			sender   string
			expected string
		}{
			{"Hello", "John", fmt.Sprintf("MSG|%d|John|Hello\r\n", time.Now().Unix())},
			{"World", "Oz", fmt.Sprintf("MSG|%d|Oz|World\r\n", time.Now().Unix())},
			{"HeyHey", "Frey", fmt.Sprintf("MSG|%d|Frey|HeyHey\r\n", time.Now().Unix())},
			{"", "John", fmt.Sprintf("MSG|%d|John|\r\n", time.Now().Unix())},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeMSG, Content: test.content, Sender: test.sender, Status: ""}))
		}
	})

	t.Run("should fail to encode when length and given text mismatch", func(t *testing.T) {
		assert.NotEqual(t, "MSG|John|5|HeyHey\r\n", EncodeMessage(Payload{MessageType: MessageTypeMSG, Content: "HeyHey", Sender: "John", Status: ""}))
	})
}

func TestEncodeWhisperMessage(t *testing.T) {
	t.Run("should encode whisper message successfully", func(t *testing.T) {
		tests := []struct {
			content   string
			sender    string
			recipient string
			expected  string
		}{
			{"Hello", "Oz", "John", fmt.Sprintf("WSP|%d|Oz|John|Hello\r\n", time.Now().Unix())},
			{"World", "John", "Oz", fmt.Sprintf("WSP|%d|John|Oz|World\r\n", time.Now().Unix())},
			{"HeyHey", "Frey", "Oz", fmt.Sprintf("WSP|%d|Frey|Oz|HeyHey\r\n", time.Now().Unix())},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeWSP, Content: test.content, Sender: test.sender, Recipient: test.recipient, Status: ""}))
		}
	})

	t.Run("should fail to encode when length and given text mismatch", func(t *testing.T) {
		assert.NotEqual(t, "WSP|5|HeyHey\r\n", EncodeMessage(Payload{MessageType: MessageTypeWSP, Content: "HeyHey", Sender: "John", Recipient: "Oz", Status: ""}))
	})
}

func TestEncodeSystemMessage(t *testing.T) {
	t.Run("should encode system message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			status   string
			expected string
		}{
			{"John has left the chat!", "left", fmt.Sprintf("SYS|%d|John has left the chat!|left\r\n", time.Now().Unix())},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeSYS, Timestamp: time.Now().Unix(), Content: test.content, Status: test.status}))
		}
	})

}

func TestEncodeErrMessage(t *testing.T) {
	t.Run("should encode error message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			expected string
		}{
			{"Errr!", fmt.Sprintf("ERR|%d|Errr!\r\n", time.Now().Unix())},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeERR, Timestamp: time.Now().Unix(), Content: "Errr!"}))
		}
	})

}

func TestEncodeUsrMessage(t *testing.T) {
	t.Run("should encode username message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			expected string
		}{
			{"Oz", fmt.Sprintf("USR|%d|Oz|success\r\n", time.Now().Unix())},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeUSR, Timestamp: time.Now().Unix(), Username: "Oz", Status: "success"}))
		}
	})

}

func TestEncodeActiveUsrMessage(t *testing.T) {
	t.Run("should encode active users message successfully", func(t *testing.T) {
		tests := []struct {
			content  []string
			expected string
		}{
			{[]string{"Oz", "John"}, fmt.Sprintf("ACT_USRS|%d|hey,there|res\r\n", time.Now().Unix())},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeACT_USRS, Timestamp: time.Now().Unix(), ActiveUsers: []string{"hey", "there"}, Status: "res"}))
		}
	})

}

func TestEncodeChatHistory(t *testing.T) {
	t.Run("should encode chat history successfully", func(t *testing.T) {
		tests := []struct {
			content  []string
			expected string
		}{
			{[]string{
				"MSG|1721160403|Oz|aaa",
				"MSG|1721160403|Oz|aaaa",
			}, fmt.Sprintf("HSTRY|%d|Oz|MSG|1721160403|Oz|aaa,MSG|1721160403|Oz|aaaa|res\r\n", time.Now().Unix())},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeHSTRY, Sender: "Oz", Timestamp: time.Now().Unix(), EncodedChatHistory: []string{"MSG|1721160403|Oz|aaa",
				"MSG|1721160403|Oz|aaaa"}, Status: "res"}))
		}
	})

}
