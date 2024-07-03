package main

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodeGeneralMessage(t *testing.T) {
	t.Run("should decode server message into payload successfully", func(t *testing.T) {
		payload, _ := decodeMessage("MSG|Frey|6|HeyHey\r\n")
		assert.Equal(t, Payload{content: "HeyHey", contentType: "MSG", sender: "Frey"}, payload)
	})
	t.Run("should check against content length", func(t *testing.T) {
		_, err := decodeMessage("MSG|Frey|5|HeyHey\r\n")
		assert.EqualError(t, err, "message content length does not match expected length in MSG message")
	})
	t.Run("should check for at least 4 parts of message MSG", func(t *testing.T) {
		_, err := decodeMessage("MSG|Frey|5\r\n")
		assert.EqualError(t, err, "insufficient parts in MSG message")
	})
}

func TestColorifyAndFormatContent(t *testing.T) {
	timestamp := time.Now().Format("[15:04]")

	t.Run("should format system message with timestamp", func(t *testing.T) {
		payload := Payload{content: "System message", contentType: MessageTypeSYS}
		expectedOutput := fmt.Sprintf("\r\x1b[36m%s System: System message\n\x1b[0m", timestamp)

		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Equal(t, stdout, expectedOutput)
	})

	t.Run("should format whisper message with timestamp", func(t *testing.T) {
		payload := Payload{content: "Hello!", contentType: MessageTypeWSP, sender: "Alice"}

		expectedOutput := fmt.Sprintf("\r\x1b[35m%s Whisper from Alice: Hello!\n", timestamp)

		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Contains(t, stdout, expectedOutput)
	})

	t.Run("should format group message with timestamp", func(t *testing.T) {
		payload := Payload{content: "Hey everyone!", contentType: MessageTypeMSG, sender: "Bob"}
		expectedOutput := fmt.Sprintf("\r\x1b[34m%s Bob: Hey everyone!\n", timestamp)

		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Contains(t, stdout, expectedOutput)
	})
}

// Helper function to capture stdout for testing purposes
func captureStdout(f func()) string {
	old := *os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = &old

	return string(out)
}
