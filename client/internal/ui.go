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
	switch payload.ContentType {
	case protocol.MessageTypeSYS:
		fmtedMsg := fmt.Sprintf("System: %s\n", payload.Content)
		if payload.Status == "fail" {
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Red)) // Red fail messages
		} else {
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Cyan)) // Cyan for system messages
		}
	case protocol.MessageTypeWSP:
		fmtedMsg := fmt.Sprintf("Whisper from %s: %s\n", payload.Sender, payload.Content)
		fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Purple)) // Purple for whisper messages
	case protocol.MessageTypeUSR:
		var fmtedMsg string
		if payload.Status == "success" {
			fmtedMsg = fmt.Sprintf("Username successfully set to %s\n", payload.Username)
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Yellow)) // Purple for username messages
		} else {
			fmtedMsg = fmt.Sprintf("%s\n", payload.Username)
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Red)) // Purple for username messages
		}
	default:
		fmtedMsg := fmt.Sprintf("%s: %s\n", payload.Sender, payload.Content)
		fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Blue)) // Blue for group messages
	}
}
