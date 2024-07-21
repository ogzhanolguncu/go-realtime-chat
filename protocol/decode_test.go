package protocol

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeGeneralMessage(t *testing.T) {
	t.Run("should decode server message into payload successfully", func(t *testing.T) {
		timestamp := time.Now().Unix()
		payload, _ := DecodeMessage(fmt.Sprintf("MSG|%d|Frey|HeyHey\r\n", timestamp))
		assert.Equal(t, Payload{Content: "HeyHey", Timestamp: timestamp, MessageType: "MSG", Sender: "Frey"}, payload)
	})

	t.Run("should check for at least 4 parts of message MSG", func(t *testing.T) {
		_, err := DecodeMessage("MSG|Frey\r\n")
		assert.EqualError(t, err, "insufficient parts in MSG message")
	})

}

func TestDecodeWhisperMessage(t *testing.T) {
	t.Run("should decode whisper message into payload successfully", func(t *testing.T) {
		timestamp := time.Now().Unix()
		payload, _ := DecodeMessage(fmt.Sprintf("WSP|%d|Oz|John|HeyHey\r\n", timestamp))
		assert.Equal(t, Payload{MessageType: MessageTypeWSP, Timestamp: timestamp, Content: "HeyHey", Sender: "Oz", Recipient: "John", Status: ""}, payload)
	})
	t.Run("should check for at least 4 parts of message WSP", func(t *testing.T) {
		_, err := DecodeMessage("WSP|John|HeyHey\r\n")
		assert.EqualError(t, err, "insufficient parts in WSP message")
	})
}

func TestDecodeSystemMessage(t *testing.T) {
	timestamp := time.Now().Unix()
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage(fmt.Sprintf("SYS|%d|Oops|fail\r\n", timestamp))
		assert.Equal(t, Payload{MessageType: MessageTypeSYS, Timestamp: timestamp, Content: "Oops", Status: "fail"}, payload)
	})
	t.Run("should check for at least 4 parts of message SYS", func(t *testing.T) {
		_, err := DecodeMessage(fmt.Sprintf("SYS|%d|fail\r\n", timestamp))
		assert.EqualError(t, err, "insufficient parts in SYS message")
	})
}

func TestDecodeUsernameMessage(t *testing.T) {
	timestamp := time.Now().Unix()
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		payload, _ := DecodeMessage(fmt.Sprintf("USR|%d|Oz|success\r\n", timestamp))
		assert.Equal(t, Payload{MessageType: MessageTypeUSR, Timestamp: timestamp, Username: "Oz", Status: "success"}, payload)
	})

	t.Run("should check for at least 4 parts of message USR", func(t *testing.T) {
		_, err := DecodeMessage("USR|fail\r\n")
		assert.EqualError(t, err, "insufficient parts in USR message")
	})
}

func TestDecodeActiveUsrMessage(t *testing.T) {
	timestamp := time.Now().Unix()
	payload, _ := DecodeMessage(fmt.Sprintf("ACT_USRS|%d|hey,there|res\r\n", timestamp))
	assert.Equal(t, Payload{MessageType: MessageTypeACT_USRS, Timestamp: timestamp, ActiveUsers: []string{"hey", "there"}, Status: "res"}, payload)

}
func TestDecodeChatHistory(t *testing.T) {
	timestamp := time.Now().Unix()
	payload, _ := DecodeMessage(fmt.Sprintf("HSTRY|%d|Oz|MSG|1721160403|Oz|aaa|res\r\n", timestamp))
	require.Equal(t, Payload{MessageType: MessageTypeHSTRY, Sender: "Oz", Timestamp: timestamp, DecodedChatHistory: []Payload{{MessageType: MessageTypeMSG, Timestamp: 1721160403, Sender: "Oz", Content: "aaa"}}, Status: "res"}, payload)
}
