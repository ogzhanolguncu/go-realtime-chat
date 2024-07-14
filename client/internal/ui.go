package internal

import (
	"fmt"

	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) PrintHeader() {
	fmt.Printf("\n\n")
	fmt.Println("---------CHATROOM--------")
	fmt.Println("-------------------------")
}

func askForInput() {
	coloredPrompt := color.ColorifyWithTimestamp("You:", color.Yellow)
	fmt.Printf("%s ", coloredPrompt)
}

func colorifyAndFormatContent(payload protocol.Payload) {
	var (
		message   string
		colorCode string
	)

	switch payload.ContentType {
	case protocol.MessageTypeSYS:
		message = fmt.Sprintf("System: %s\n", payload.Content)
		if payload.Status == "fail" {
			colorCode = color.Red
		} else {
			colorCode = color.Cyan
		}
	case protocol.MessageTypeWSP:
		message = fmt.Sprintf("Whisper from %s: %s\n", payload.Sender, payload.Content)
		colorCode = color.Purple
	case protocol.MessageTypeUSR:
		if payload.Status == "success" {
			message = fmt.Sprintf("Username successfully set to %s\n", payload.Username)
			colorCode = color.Yellow
		} else {
			message = fmt.Sprintf("%s\n", payload.Username)
			colorCode = color.Red
		}
	default:
		message = fmt.Sprintf("%s: %s\n", payload.Sender, payload.Content)
		colorCode = color.Blue
	}
	fmt.Print(color.ColorifyWithTimestamp(message, colorCode))
}
