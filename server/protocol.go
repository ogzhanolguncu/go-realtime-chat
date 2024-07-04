package main

import "fmt"

// Group/General Message (MSG): MSG|sender|message_length|message_content\r\n
// Whisper/DM Message (WSP): WSP|recipient|message_length|message_content\r\n
// System Notice (SYS): SYS|message_length|message_content|status\r\n => status = fail | success
// Error Message (ERR): ERR|message_length|error_message\r\n

func encodeGeneralMessage(content, sender string) string {
	length := len(content)
	return fmt.Sprintf("MSG|%s|%d|%s\r\n", sender, length, content)
}

func encodeWhisperMessage(content, sender string) string {
	length := len(content)
	return fmt.Sprintf("WSP|%s|%d|%s\r\n", sender, length, content)
}

func encodeSystemMessage(content, status string) string {
	length := len(content)
	if status == "" {
		return fmt.Sprintf("SYS|%d|%s\r\n", length, content)
	}
	return fmt.Sprintf("SYS|%d|%s|%s\r\n", length, content, status)
}

func encodeErrorMessage(content string) string {
	length := len(content)
	return fmt.Sprintf("ERR|%d|%s\r\n", length, content)
}
