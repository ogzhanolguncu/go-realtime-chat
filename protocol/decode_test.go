package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeGeneralMessage(t *testing.T) {
	t.Run("should decode server message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("MSG|Frey|6|HeyHey\r\n")
		assert.Equal(t, Payload{Content: "HeyHey", ContentType: "MSG", Sender: "Frey"}, payload)
	})
	t.Run("should check against content length", func(t *testing.T) {
		_, err := DecodeMessage("MSG|Frey|5|HeyHey\r\n")
		assert.EqualError(t, err, "message content length does not match expected length in MSG message")
	})
	t.Run("should check for at least 4 parts of message MSG", func(t *testing.T) {
		_, err := DecodeMessage("MSG|Frey|5\r\n")
		assert.EqualError(t, err, "insufficient parts in MSG message")
	})

}

func TestDecodeWhisperMessage(t *testing.T) {
	t.Run("should decode server message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("WSP|Oz|John|6|HeyHey\r\n")
		assert.Equal(t, Payload{ContentType: MessageTypeWSP, Content: "HeyHey", Sender: "Oz", Recipient: "John", Status: ""}, payload)
	})

}
