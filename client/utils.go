package main

import (
	"fmt"

	"github.com/ogzhanolguncu/go-chat/client/color"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

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
	default:
		fmtedMsg := fmt.Sprintf("%s: %s\n", payload.Sender, payload.Content)
		fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Blue)) // Blue for group messages
	}
}

func askForInput() {
	coloredPrompt := color.ColorifyWithTimestamp("You:", color.Yellow)
	fmt.Printf("%s ", coloredPrompt)
}
