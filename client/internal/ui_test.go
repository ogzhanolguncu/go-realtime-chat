package internal

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/ogzhanolguncu/go-chat/client/terminal"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"

	"github.com/stretchr/testify/assert"
)

func TestColorifyAndFormatContent(t *testing.T) {

	t.Run("should format system message with timestamp", func(t *testing.T) {
		payload := protocol.Payload{Content: "System message", Timestamp: time.Now().Unix(), MessageType: protocol.MessageTypeSYS}
		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Equal(t, stdout, terminal.ColorifyWithTimestamp("System: System message\n", terminal.Cyan))
	})

	t.Run("should format whisper message with timestamp", func(t *testing.T) {
		payload := protocol.Payload{Content: "Hello!", Timestamp: time.Now().Unix(), MessageType: protocol.MessageTypeWSP, Sender: "Alice"}
		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Contains(t, stdout, terminal.ColorifyWithTimestamp("Whisper from Alice: Hello!\n", terminal.Purple))
	})

	t.Run("should format group message with timestamp", func(t *testing.T) {
		payload := protocol.Payload{Content: "Hey everyone!", Timestamp: time.Now().Unix(), MessageType: protocol.MessageTypeMSG, Sender: "Bob"}
		stdout := captureStdout(func() {
			colorifyAndFormatContent(payload)
		})

		assert.Contains(t, stdout, terminal.ColorifyWithTimestamp("Bob: Hey everyone!\n", terminal.Green))
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
