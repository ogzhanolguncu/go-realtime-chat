package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeGeneralMessage(t *testing.T) {
	t.Run("should encode general message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			sender   string
			expected string
		}{
			{"Hello", "John", "MSG|John|5|Hello\r\n"},
			{"World", "Oz", "MSG|Oz|5|World\r\n"},
			{"HeyHey", "Frey", "MSG|Frey|6|HeyHey\r\n"},
			{"", "John", "MSG|John|0|\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{ContentType: MessageTypeMSG, Content: test.content, Sender: test.sender, Status: ""}))
		}
	})

	t.Run("should fail to encode when length and given text mismatch", func(t *testing.T) {
		assert.NotEqual(t, "MSG|John|5|HeyHey\r\n", EncodeMessage(Payload{ContentType: MessageTypeMSG, Content: "HeyHey", Sender: "John", Status: ""}))
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
			{"Hello", "Oz", "John", "WSP|Oz|John|5|Hello\r\n"},
			{"World", "John", "Oz", "WSP|John|Oz|5|World\r\n"},
			{"HeyHey", "Frey", "Oz", "WSP|Frey|Oz|6|HeyHey\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, EncodeMessage(Payload{ContentType: MessageTypeWSP, Content: test.content, Sender: test.sender, Recipient: test.recipient, Status: ""}))
		}
	})

	t.Run("should fail to encode when length and given text mismatch", func(t *testing.T) {
		assert.NotEqual(t, "WSP|5|HeyHey\r\n", EncodeMessage(Payload{ContentType: MessageTypeWSP, Content: "HeyHey", Sender: "John", Recipient: "Oz", Status: ""}))
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
			assert.Equal(t, test.expected, EncodeMessage(Payload{ContentType: MessageTypeSYS, Content: test.content, Status: test.status}))
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
			assert.Equal(t, test.expected, EncodeMessage(Payload{ContentType: MessageTypeERR, Content: "Errr!"}))
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
			assert.Equal(t, test.expected, EncodeMessage(Payload{ContentType: MessageTypeUSR, Username: "Oz", Status: "success"}))
		}
	})

}
