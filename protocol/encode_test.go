package protocol

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			result := encodeProtocol(true, Payload{MessageType: MessageTypeMSG, Content: test.content, Sender: test.sender})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			assert.Equal(t, test.expected, string(decoded))
		}
	})

	t.Run("should fail to encode when message type is invalid", func(t *testing.T) {
		result := encodeProtocol(true, Payload{MessageType: "INVALID", Content: "HeyHey", Sender: "John"})
		decoded, _ := base64.StdEncoding.DecodeString(result)

		expected := "ERR|Invalid message type\r\n"
		assert.Equal(t, expected, string(decoded))
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
			{"", "Alice", "Bob", fmt.Sprintf("WSP|%d|Alice|Bob|\r\n", time.Now().Unix())},
		}
		for _, test := range tests {
			result := encodeProtocol(true, Payload{
				MessageType: MessageTypeWSP,
				Content:     test.content,
				Sender:      test.sender,
				Recipient:   test.recipient,
			})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			assert.Equal(t, test.expected, string(decoded))
		}
	})

	t.Run("should encode whisper message with empty content", func(t *testing.T) {
		result := encodeProtocol(true, Payload{
			MessageType: MessageTypeWSP,
			Content:     "",
			Sender:      "John",
			Recipient:   "Oz",
		})
		decoded, err := base64.StdEncoding.DecodeString(result)
		assert.NoError(t, err)

		pattern := "WSP|\\d+|John|Oz|\r\n"
		matched, err := regexp.Match(pattern, decoded)

		assert.NoError(t, err)
		assert.True(t, matched, "Encoded message with empty content does not match expected pattern")
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
			result := encodeProtocol(true, Payload{MessageType: MessageTypeSYS, Content: test.content, Status: test.status})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			assert.Equal(t, test.expected, string(decoded))
		}
	})
}

func TestEncodeUsrMessage(t *testing.T) {
	t.Run("should encode username message successfully", func(t *testing.T) {
		tests := []struct {
			username string
			password string
			status   string
			expected string
		}{
			{"Oz", "123456", "success", fmt.Sprintf("USR|%d|Oz|123456|success\r\n", time.Now().Unix())},
		}
		for _, test := range tests {
			result := encodeProtocol(true, Payload{MessageType: MessageTypeUSR, Password: test.password, Username: test.username, Status: test.status})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			assert.Equal(t, test.expected, string(decoded))
		}
	})
}

func TestEncodeActiveUsrMessage(t *testing.T) {
	t.Run("should encode active users message successfully", func(t *testing.T) {
		tests := []struct {
			activeUsers []string
			status      string
			expected    string
		}{
			{[]string{"Oz", "John"}, "res", fmt.Sprintf("ACT_USRS|%d|Oz,John|res\r\n", time.Now().Unix())},
		}
		for _, test := range tests {
			result := encodeProtocol(true, Payload{MessageType: MessageTypeACT_USRS, ActiveUsers: test.activeUsers, Status: test.status})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			assert.Equal(t, test.expected, string(decoded))
		}
	})
}

func TestEncodeChatHistory(t *testing.T) {
	t.Run("should encode chat history successfully", func(t *testing.T) {
		tests := []struct {
			sender   string
			history  []string
			status   string
			expected string
		}{
			{
				"Oz",
				[]string{"MSG|1721160403|Oz|aaa", "MSG|1721160403|Oz|aaaa"},
				"res",
				fmt.Sprintf("HSTRY|%d|Oz|MSG|1721160403|Oz|aaa,MSG|1721160403|Oz|aaaa|res\r\n", time.Now().Unix()),
			},
		}
		for _, test := range tests {
			result := encodeProtocol(true, Payload{
				MessageType:        MessageTypeHSTRY,
				Sender:             test.sender,
				EncodedChatHistory: test.history,
				Status:             test.status,
			})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			assert.Equal(t, test.expected, string(decoded))
		}
	})
}

func TestEncodeBlockUserMessage(t *testing.T) {
	t.Run("should encode blocker user message successfully", func(t *testing.T) {
		tests := []struct {
			content   string
			sender    string
			recipient string
			expected  string
		}{
			{"Hello", "Oz", "John", fmt.Sprintf("BLCK_USR|%d|Oz|John|Hello\r\n", time.Now().Unix())},
			{"World", "John", "Oz", fmt.Sprintf("BLCK_USR|%d|John|Oz|World\r\n", time.Now().Unix())},
		}
		for _, test := range tests {
			result := encodeProtocol(true, Payload{
				MessageType: MessageTypeBLCK_USR,
				Content:     test.content,
				Sender:      test.sender,
				Recipient:   test.recipient,
			})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			assert.Equal(t, test.expected, string(decoded))
		}
	})

}

func TestEncodeRoomMessage(t *testing.T) {
	tests := []struct {
		roomAction   ChannelActionType
		requester    string
		roomName     string
		roomPassword string
		roomSize     int
		optionalArgs *OptionalChannelArgs
		expected     string
		testName     string
	}{
		{
			roomAction:   CreateChannel,
			requester:    "Oz",
			roomName:     "testRoom",
			roomPassword: "testPassword",
			roomSize:     2,
			optionalArgs: &OptionalChannelArgs{
				Visibility: VisibilityPublic,
			},
			expected: fmt.Sprintf("CH|%d|CreateChannel|Oz|testRoom|testPassword|2|visibility=public\r\n", time.Now().Unix()),
			testName: "Create Room Request",
		},
		{
			roomAction:   CreateChannel,
			requester:    "Oz",
			roomName:     "testRoom",
			roomPassword: "testPassword",
			roomSize:     2,
			optionalArgs: &OptionalChannelArgs{
				Status: StatusSuccess,
			},
			expected: fmt.Sprintf("CH|%d|CreateChannel|Oz|testRoom|testPassword|2|status=success\r\n", time.Now().Unix()),
			testName: "Create Room Success",
		},
		{
			roomAction:   JoinChannel,
			requester:    "John",
			roomName:     "testRoom",
			roomPassword: "testPassword",
			expected:     fmt.Sprintf("CH|%d|JoinChannel|John|testRoom|testPassword|-\r\n", time.Now().Unix()),
			testName:     "Join Room Request",
		},
		{
			roomAction: GetChannels,
			requester:  "John",
			expected:   fmt.Sprintf("CH|%d|GetChannels|John|-|-|-\r\n", time.Now().Unix()),
			testName:   "Get Rooms Request",
		},
		{
			roomAction: GetChannels,
			requester:  "John",
			optionalArgs: &OptionalChannelArgs{
				Status:     StatusSuccess,
				Channels:   []string{"golang", "nodejs", "test"},
				TargetUser: "heheh",
			},
			expected: fmt.Sprintf("CH|%d|GetChannels|John|-|-|-|status=success;rooms=golang,nodejs,test\r\n", time.Now().Unix()),
			testName: "Get Rooms Success",
		},
		{
			roomAction: GetChannels,
			requester:  "John",
			optionalArgs: &OptionalChannelArgs{
				Status: StatusFail,
				Reason: "no_active_rooms",
			},
			expected: fmt.Sprintf("CH|%d|GetChannels|John|-|-|-|status=fail;reason=no_active_rooms\r\n", time.Now().Unix()),
			testName: "Get Rooms Fail",
		},
		{
			roomAction: LeaveChannel,
			requester:  "Alice",
			roomName:   "specialRoom!@#$",
			optionalArgs: &OptionalChannelArgs{
				Status:  StatusSuccess,
				Message: "Left the room successfully",
				Users:   []string{"Bob", "Charlie"},
			},
			expected: fmt.Sprintf("CH|%d|LeaveChannel|Alice|specialRoom!@#$|-|-|status=success;message=Left the room successfully;users=Bob,Charlie\r\n", time.Now().Unix()),
			testName: "Leave Room with Special Characters and Multiple Optional Args",
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			roomPayload := &ChannelPayload{
				ChannelAction:   test.roomAction,
				Requester:       test.requester,
				ChannelName:     test.roomName,
				ChannelPassword: test.roomPassword,
				ChannelSize:     test.roomSize,
			}
			if test.optionalArgs != nil {
				roomPayload.OptionalChannelArgs = &OptionalChannelArgs{
					Status:     test.optionalArgs.Status,
					Visibility: test.optionalArgs.Visibility,
					Channels:   test.optionalArgs.Channels,
					Reason:     test.optionalArgs.Reason,
					Message:    test.optionalArgs.Message,
					Users:      test.optionalArgs.Users,
				}
			}
			result := encodeProtocol(true, Payload{
				MessageType:    MessageTypeCH,
				ChannelPayload: roomPayload,
			})
			decoded, _ := base64.StdEncoding.DecodeString(result)
			require.Equal(t, test.expected, string(decoded))
		})

	}

}
