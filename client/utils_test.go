package main

import (
	"io"
	"os"
	"testing"

	"github.com/ogzhanolguncu/go-chat/client/color"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"

	"github.com/stretchr/testify/assert"
)

func TestColorifyAndFormatContent(t *testing.T) {

	t.Run("should format system message with timestamp", func(t *testing.T) {
		payload := protocol.Payload{Content: "System message", ContentType: protocol.MessageTypeSYS}
		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Equal(t, stdout, color.ColorifyWithTimestamp("System: System message\n", color.Cyan))
	})

	t.Run("should format whisper message with timestamp", func(t *testing.T) {
		payload := protocol.Payload{Content: "Hello!", ContentType: protocol.MessageTypeWSP, Sender: "Alice"}
		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Contains(t, stdout, color.ColorifyWithTimestamp("Whisper from Alice: Hello!\n", color.Purple))
	})

	t.Run("should format group message with timestamp", func(t *testing.T) {
		payload := protocol.Payload{Content: "Hey everyone!", ContentType: protocol.MessageTypeMSG, Sender: "Bob"}
		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Contains(t, stdout, color.ColorifyWithTimestamp("Bob: Hey everyone!\n", color.Blue))
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
