package protocol

import "fmt"

func EncodeMessage(messageType, content, sender, status string) string {
	length := len(content)

	switch messageType {
	case MessageTypeMSG, MessageTypeWSP:
		return fmt.Sprintf("%s|%s|%d|%s\r\n", messageType, sender, length, content)
	case MessageTypeSYS:
		if status == "" {
			return fmt.Sprintf("%s|%d|%s\r\n", messageType, length, content)
		}
		return fmt.Sprintf("%s|%d|%s|%s\r\n", messageType, length, content, status)
	case MessageTypeERR:
		return fmt.Sprintf("%s|%d|%s\r\n", messageType, length, content)
	default:
		return fmt.Sprintf("ERR|%d|Invalid message type\r\n", len("Invalid message type"))
	}
}
