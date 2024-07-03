package main

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
			assert.Equal(t, test.expected, encodeGeneralMessage(test.content, test.sender))
		}
	})

	t.Run("should fail to encode when length and given text mismatch", func(t *testing.T) {
		assert.NotEqual(t, "MSG|John|5|HeyHey\r\n", encodeGeneralMessage("HeyHey", "John"))
	})
}

func TestEncodeWhisperMessage(t *testing.T) {
	t.Run("should encode whisper message successfully", func(t *testing.T) {
		tests := []struct {
			content   string
			recipient string
			expected  string
		}{
			{"Hello", "Oz", "WSP|Oz|5|Hello\r\n"},
			{"World", "John", "WSP|John|5|World\r\n"},
			{"HeyHey", "Frey", "WSP|Frey|6|HeyHey\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, encodeWhisperMessage(test.content, test.recipient))
		}
	})

	t.Run("should fail to encode when length and given text mismatch", func(t *testing.T) {
		assert.NotEqual(t, "WSP|5|HeyHey\r\n", encodeWhisperMessage("HeyHey", "John"))
	})
}

func TestEncodeSystemMessage(t *testing.T) {
	t.Run("should encode whisper message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			expected string
		}{
			{"John has left the chat!", "SYS|23|John has left the chat!\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, encodeSystemMessage(test.content))
		}
	})

}

func TestEncodeErrMessage(t *testing.T) {
	t.Run("should encode whisper message successfully", func(t *testing.T) {
		tests := []struct {
			content  string
			expected string
		}{
			{"Errr!", "ERR|5|Errr!\r\n"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, encodeErrorMessage(test.content))
		}
	})

}
