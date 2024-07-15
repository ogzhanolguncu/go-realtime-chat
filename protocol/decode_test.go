package protocol

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodeGeneralMessage(t *testing.T) {
	t.Run("should decode server message into payload successfully", func(t *testing.T) {
		timestamp := time.Now().Unix()
		payload, _ := DecodeMessage(fmt.Sprintf("MSG|%d|Frey|6|HeyHey\r\n", timestamp))
		assert.Equal(t, Payload{Content: "HeyHey", Timestamp: timestamp, MessageType: "MSG", Sender: "Frey"}, payload)
	})
	t.Run("should check against content length", func(t *testing.T) {
		_, err := DecodeMessage("MSG|123123|Frey|5|HeyHey\r\n")
		assert.EqualError(t, err, "message content length does not match expected length in MSG message")
	})
	t.Run("should check for at least 4 parts of message MSG", func(t *testing.T) {
		_, err := DecodeMessage("MSG|Frey|5\r\n")
		assert.EqualError(t, err, "insufficient parts in MSG message")
	})

}

func TestDecodeWhisperMessage(t *testing.T) {
	t.Run("should decode whisper message into payload successfully", func(t *testing.T) {
		timestamp := time.Now().Unix()
		payload, _ := DecodeMessage(fmt.Sprintf("WSP|%d|Oz|John|6|HeyHey\r\n", timestamp))
		assert.Equal(t, Payload{MessageType: MessageTypeWSP, Timestamp: timestamp, Content: "HeyHey", Sender: "Oz", Recipient: "John", Status: ""}, payload)
	})
	t.Run("should check for at least 4 parts of message WSP", func(t *testing.T) {
		_, err := DecodeMessage("WSP|John|6|HeyHey\r\n")
		assert.EqualError(t, err, "insufficient parts in WSP message")
	})
}

func TestDecodeSystemMessage(t *testing.T) {
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("SYS|4|Oops|fail\r\n")
		assert.Equal(t, Payload{MessageType: MessageTypeSYS, Content: "Oops", Status: "fail"}, payload)
	})
	t.Run("should check for at least 4 parts of message SYS", func(t *testing.T) {
		_, err := DecodeMessage("SYS|4|fail\r\n")
		assert.EqualError(t, err, "insufficient parts in SYS message")
	})
}

func TestDecodeUsernameMessage(t *testing.T) {
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("USR|2|Oz|success\r\n")
		assert.Equal(t, Payload{MessageType: MessageTypeUSR, Username: "Oz", Status: "success"}, payload)
	})

	t.Run("should check for at least 4 parts of message USR", func(t *testing.T) {
		_, err := DecodeMessage("USR|4|fail\r\n")
		assert.EqualError(t, err, "insufficient parts in USR message")
	})
}

func TestDecodeActiveUsrMessage(t *testing.T) {
	t.Run("should decode active users message successfully", func(t *testing.T) {
		payload, _ := DecodeMessage("ACT_USRS|2|hey,there\r\n")
		assert.Equal(t, Payload{MessageType: MessageTypeACT_USRS, ActiveUsers: []string{"hey", "there"}}, payload)
	})

}
