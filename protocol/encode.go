package protocol

import "fmt"

func EncodeMessage(payload Payload) string {
	length := len(payload.Content)

	switch payload.ContentType {
	case MessageTypeMSG:
		return fmt.Sprintf("%s|%s|%d|%s\r\n", payload.ContentType, payload.Sender, length, payload.Content)
	// WSP|sender|recipient|message_length|message_content\r\n
	case MessageTypeWSP:
		return fmt.Sprintf("%s|%s|%s|%d|%s\r\n", payload.ContentType, payload.Sender, payload.Recipient, length, payload.Content)
	case MessageTypeSYS:
		if payload.SysStatus == "" {
			return fmt.Sprintf("%s|%d|%s\r\n", payload.ContentType, length, payload.Content)
		}
		return fmt.Sprintf("%s|%d|%s|%s\r\n", payload.ContentType, length, payload.Content, payload.SysStatus)
	case MessageTypeERR:
		return fmt.Sprintf("%s|%d|%s\r\n", payload.ContentType, length, payload.Content)
	default:
		return fmt.Sprintf("ERR|%d|Invalid message type\r\n", len("Invalid message type"))
	}
}
