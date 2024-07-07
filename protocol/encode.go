package protocol

import "fmt"

func EncodeMessage(payload Payload) string {
	length := len(payload.Content)

	switch payload.ContentType {
	// MSG|sender|message_length|message_content\r\n
	case MessageTypeMSG:
		return fmt.Sprintf("%s|%s|%d|%s\r\n", payload.ContentType, payload.Sender, length, payload.Content)
	// WSP|sender|recipient|message_length|message_content\r\n
	case MessageTypeWSP:
		return fmt.Sprintf("%s|%s|%s|%d|%s\r\n", payload.ContentType, payload.Sender, payload.Recipient, length, payload.Content)
	// SYS|message_length|message_content|status \r\n status = "fail" | "success"
	case MessageTypeSYS:
		if payload.Status == "" {
			return fmt.Sprintf("%s|%d|%s\r\n", payload.ContentType, length, payload.Content)
		}
		return fmt.Sprintf("%s|%d|%s|%s\r\n", payload.ContentType, length, payload.Content, payload.Status)
	// ERR|message_length|error_message\r\n
	case MessageTypeERR:
		return fmt.Sprintf("%s|%d|%s\r\n", payload.ContentType, length, payload.Content)
	// USR|name_length|name_content|status\r\n status = "fail | "success"
	case MessageTypeUSR:
		nameLength := len(payload.Username)
		return fmt.Sprintf("%s|%d|%s|%s\r\n", payload.ContentType, nameLength, payload.Username, payload.Status)
	// ERR|message_length|error_message\r\n
	default:
		return fmt.Sprintf("ERR|%d|Invalid message type\r\n", len("Invalid message type"))
	}
}
