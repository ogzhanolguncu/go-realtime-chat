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
			{"Hello", "John", fmt.Sprintf("MSG|%d|John|5|Hello\r\n", time.Now().Unix())},
			{"World", "Oz", fmt.Sprintf("MSG|%d|Oz|5|World\r\n", time.Now().Unix())},
			{"HeyHey", "Frey", fmt.Sprintf("MSG|%d|Frey|6|HeyHey\r\n", time.Now().Unix())},
			{"", "John", fmt.Sprintf("MSG|%d|John|0|\r\n", time.Now().Unix())},
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
			{"Hello", "Oz", "John", fmt.Sprintf("WSP|%d|Oz|John|5|Hello\r\n", time.Now().Unix())},
			{"World", "John", "Oz", fmt.Sprintf("WSP|%d|John|Oz|5|World\r\n", time.Now().Unix())},
			{"HeyHey", "Frey", "Oz", fmt.Sprintf("WSP|%d|Frey|Oz|6|HeyHey\r\n", time.Now().Unix())},
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
			{"John has left the chat!", "left", "SYS|23|John has left the chat!|left\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeSYS, Content: test.content, Status: test.status}))
		}
	})

}

func TestEncodeErrMessage(t *testing.T) {
	t.Run("should encode error message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			expected string
		}{
			{"Errr!", "ERR|5|Errr!\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeERR, Content: "Errr!"}))
		}
	})

}

func TestEncodeUsrMessage(t *testing.T) {
	t.Run("should encode username message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			expected string
		}{
			{"Oz", "USR|2|Oz|success\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeUSR, Username: "Oz", Status: "success"}))
		}
	})

}

func TestEncodeActiveUsrMessage(t *testing.T) {
	t.Run("should encode active users message successfully", func(t *testing.T) {
		tests := []struct {
			content  []string
			expected string
		}{
			{[]string{"hey", "there"}, "ACT_USRS|2|hey,there\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{MessageType: MessageTypeACT_USRS, ActiveUsers: []string{"hey", "there"}}))
		}
	})

}
