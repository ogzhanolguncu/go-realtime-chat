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
	t.Run("should decode whisper message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("WSP|Oz|John|6|HeyHey\r\n")
		assert.Equal(t, Payload{ContentType: MessageTypeWSP, Content: "HeyHey", Sender: "Oz", Recipient: "John", Status: ""}, payload)
	})
	t.Run("should check for at least 4 parts of message WSP", func(t *testing.T) {
		_, err := DecodeMessage("WSP|John|6|HeyHey\r\n")
		assert.EqualError(t, err, "insufficient parts in WSP message")
	})
}

func TestDecodeSystemMessage(t *testing.T) {
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("SYS|4|Oops|fail\r\n")
		assert.Equal(t, Payload{ContentType: MessageTypeSYS, Content: "Oops", Status: "fail"}, payload)
	})
	t.Run("should check for at least 4 parts of message SYS", func(t *testing.T) {
		_, err := DecodeMessage("SYS|4|fail\r\n")
		assert.EqualError(t, err, "insufficient parts in SYS message")
	})
}

func TestDecodeUsernameMessage(t *testing.T) {
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("USR|2|Oz|success\r\n")
		assert.Equal(t, Payload{ContentType: MessageTypeUSR, Username: "Oz", Status: "success"}, payload)
	})

	t.Run("should check for at least 4 parts of message USR", func(t *testing.T) {
		_, err := DecodeMessage("USR|4|fail\r\n")
		assert.EqualError(t, err, "insufficient parts in USR message")
	})
}
