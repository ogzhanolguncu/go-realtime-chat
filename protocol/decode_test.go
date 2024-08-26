package protocol

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeGeneralMessage(t *testing.T) {
	t.Run("should decode server message into payload successfully", func(t *testing.T) {
		timestamp := time.Now().Unix()
		encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("MSG|%d|Frey|HeyHey\r\n", timestamp)))
		payload, _ := decodeProtocol(true, encodedString)
		require.Equal(t, Payload{Content: "HeyHey", Timestamp: timestamp, MessageType: "MSG", Sender: "Frey"}, payload)
	})

	t.Run("should check for at least 4 parts of message MSG", func(t *testing.T) {
		encodedString := base64.StdEncoding.EncodeToString([]byte("MSG|Frey\r\n"))
		_, err := decodeProtocol(true, encodedString)
		require.EqualError(t, err, "invalid MSG format: missing timestamp separator")
	})

}

func TestDecodeWhisperMessage(t *testing.T) {
	t.Run("should decode whisper message into payload successfully", func(t *testing.T) {
		timestamp := time.Now().Unix()
		encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("WSP|%d|Oz|John|HeyHey\r\n", timestamp)))

		payload, _ := decodeProtocol(true, encodedString)
		assert.Equal(t, Payload{MessageType: MessageTypeWSP, Timestamp: timestamp, Content: "HeyHey", Sender: "Oz", Recipient: "John", Status: ""}, payload)
	})
	t.Run("should check for at least 4 parts of message WSP", func(t *testing.T) {
		encodedString := base64.StdEncoding.EncodeToString([]byte("WSP|John|HeyHey\r\n"))
		_, err := decodeProtocol(true, encodedString)
		assert.EqualError(t, err, "invalid timestamp format: strconv.ParseInt: parsing \"John\": invalid syntax")
	})
}

func TestDecodeSystemMessage(t *testing.T) {
	timestamp := time.Now().Unix()
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("SYS|%d|Oops|fail\r\n", timestamp)))
		payload, _ := decodeProtocol(true, encodedString)
		assert.Equal(t, Payload{MessageType: MessageTypeSYS, Timestamp: timestamp, Content: "Oops", Status: "fail"}, payload)
	})
	t.Run("should check for at least 4 parts of message SYS", func(t *testing.T) {
		encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("SYS|%d|fail\r\n", timestamp)))

		_, err := decodeProtocol(true, encodedString)
		assert.EqualError(t, err, "invalid SYS format: missing content separator")
	})
}

func TestDecodeUsernameMessage(t *testing.T) {
	timestamp := time.Now().Unix()
	t.Run("should decode system message into payload successfully", func(t *testing.T) {
		encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("USR|%d|Oz|123456|success\r\n", timestamp)))

		payload, _ := decodeProtocol(true, encodedString)
		assert.Equal(t, Payload{MessageType: MessageTypeUSR, Timestamp: timestamp, Username: "Oz", Password: "123456", Status: "success"}, payload)
	})

	t.Run("should check for at least 4 parts of message USR", func(t *testing.T) {
		encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("USR|%d|fail\r\n", timestamp)))
		_, err := decodeProtocol(true, encodedString)
		assert.EqualError(t, err, "invalid USR format: missing name separator")
	})
}

func TestDecodeActiveUsrMessage(t *testing.T) {
	timestamp := time.Now().Unix()
	encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("ACT_USRS|%d|hey,there|res\r\n", timestamp)))

	payload, _ := decodeProtocol(true, encodedString)
	assert.Equal(t, Payload{MessageType: MessageTypeACT_USRS, Timestamp: timestamp, ActiveUsers: []string{"hey", "there"}, Status: "res"}, payload)

}
func TestDecodeChatHistory(t *testing.T) {
	timestamp := time.Now().Unix()
	encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("HSTRY|%d|Oz|TVNHfDE3MjExNjA0MDN8T3p8YWFh|res\r\n", timestamp)))

	payload, _ := decodeProtocol(true, encodedString)
	require.Equal(
		t,
		Payload{
			MessageType: MessageTypeHSTRY,
			Sender:      "Oz",
			Timestamp:   timestamp,
			DecodedChatHistory: []Payload{{
				MessageType: MessageTypeMSG,
				Timestamp:   1721160403,
				Sender:      "Oz", Content: "aaa"}},
			Status: "res"},
		payload)
}

func TestDecodeBlockUserMessage(t *testing.T) {
	t.Run("should decode block user message into payload successfully", func(t *testing.T) {
		timestamp := time.Now().Unix()
		encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("BLCK_USR|%d|Oz|John|block\r\n", timestamp)))

		payload, _ := decodeProtocol(true, encodedString)
		assert.Equal(t, Payload{MessageType: MessageTypeBLCK_USR, Timestamp: timestamp, Content: "block", Sender: "Oz", Recipient: "John", Status: ""}, payload)
	})
	t.Run("should check for at least 4 parts of message WSP", func(t *testing.T) {
		encodedString := base64.StdEncoding.EncodeToString([]byte("BLCK_USR|John|HeyHey\r\n"))
		_, err := decodeProtocol(true, encodedString)
		assert.EqualError(t, err, "invalid timestamp format: strconv.ParseInt: parsing \"John\": invalid syntax")
	})
}

func TestDecodeRoomMessage(t *testing.T) {
	timeNow := time.Now().Unix()
	tests := []struct {
		expected      Payload
		expectedError bool
		testName      string
		input         string
	}{
		{
			expected: Payload{
				MessageType: MessageTypeCH,
				Timestamp:   timeNow,
				ChannelPayload: &ChannelPayload{
					ChannelAction:   CreateChannel,
					Requester:       "Oz",
					ChannelName:     "testRoom",
					ChannelPassword: "testPassword",
					ChannelSize:     2,
					OptionalChannelArgs: &OptionalChannelArgs{
						Visibility: VisibilityPublic,
					},
				},
			},
			input:    fmt.Sprintf("CH|%d|CreateChannel|Oz|testRoom|testPassword|2|visibility=public", timeNow),
			testName: "Create Room Request",
		},
		{
			expected: Payload{
				MessageType: MessageTypeCH,
				Timestamp:   timeNow,
				ChannelPayload: &ChannelPayload{
					ChannelAction:   CreateChannel,
					Requester:       "Oz",
					ChannelName:     "testRoom",
					ChannelPassword: "testPassword",
					ChannelSize:     2,
					OptionalChannelArgs: &OptionalChannelArgs{
						Status: StatusSuccess,
					},
				},
			},
			input:    fmt.Sprintf("CH|%d|CreateChannel|Oz|testRoom|testPassword|2|status=success", timeNow),
			testName: "Create Room Success",
		},
		{
			expected: Payload{
				MessageType: MessageTypeCH,
				Timestamp:   timeNow,
				ChannelPayload: &ChannelPayload{
					ChannelAction:   JoinChannel,
					Requester:       "John",
					ChannelName:     "testRoom",
					ChannelPassword: "testPassword",
				},
			},
			input:    fmt.Sprintf("CH|%d|JoinChannel|John|testRoom|testPassword|-", timeNow),
			testName: "Join Room Request",
		},
		{
			expected: Payload{
				MessageType: MessageTypeCH,
				Timestamp:   timeNow,
				ChannelPayload: &ChannelPayload{
					ChannelAction: GetChannels,
					Requester:     "John",
				},
			},
			input:    fmt.Sprintf("CH|%d|GetChannels|John|-|-|-", timeNow),
			testName: "Get Rooms Request",
		},
		{
			expected: Payload{
				MessageType: MessageTypeCH,
				Timestamp:   timeNow,
				ChannelPayload: &ChannelPayload{
					ChannelAction: GetChannels,
					Requester:     "John",
					OptionalChannelArgs: &OptionalChannelArgs{
						Status:   StatusSuccess,
						Channels: []string{"golang", "nodejs", "test"},
					},
				},
			},
			input:    fmt.Sprintf("CH|%d|GetChannels|John|-|-|-|status=success;rooms=golang,nodejs,test", timeNow),
			testName: "Get Rooms Success",
		},
		{
			expected: Payload{
				MessageType: MessageTypeCH,
				Timestamp:   timeNow,
				ChannelPayload: &ChannelPayload{
					ChannelAction: GetChannels,
					Requester:     "John",
					OptionalChannelArgs: &OptionalChannelArgs{
						Status: StatusFail,
						Reason: "no_active_rooms",
					},
				},
			},
			input:    fmt.Sprintf("CH|%d|GetChannels|John|-|-|-|status=fail;reason=no_active_rooms", timeNow),
			testName: "Get Rooms Fail",
		},
		{
			expected: Payload{
				MessageType: MessageTypeCH,
				Timestamp:   timeNow,
				ChannelPayload: &ChannelPayload{
					ChannelAction:   LeaveChannel,
					Requester:       "Alice",
					ChannelName:     "specialRoom!@#$",
					ChannelPassword: "",
					OptionalChannelArgs: &OptionalChannelArgs{
						Status:  StatusSuccess,
						Message: "Left the room successfully",
						Users:   []string{"Bob", "Charlie"},
					},
				},
			},
			input:    fmt.Sprintf("CH|%d|LeaveChannel|Alice|specialRoom!@#$|-|-|status=success;message=Left the room successfully;users=Bob,Charlie", timeNow),
			testName: "Leave Room with Special Characters and Multiple Optional Args",
		},
		{
			expectedError: true,
			input:         fmt.Sprintf("CH|%d|InvalidAction|John|-|-|-", timeNow),
			testName:      "Invalid Room Action",
		},
		{
			expectedError: true,
			input:         "CH|1234567890",
			testName:      "Missing Required Fields",
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			payload, err := decodeProtocol(false, test.input)
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, payload)
			}
		})
	}
}
