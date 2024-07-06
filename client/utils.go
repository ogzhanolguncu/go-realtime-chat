package main

import (
	"fmt"

	"github.com/ogzhanolguncu/go-chat/client/color"
)

func colorifyAndFormatContent(payload Payload) {
	switch payload.contentType {
	case MessageTypeSYS:
		fmtedMsg := fmt.Sprintf("System: %s\n", payload.content)
		if payload.sysStatus == "fail" {
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Red)) // Red fail messages
		} else {
			fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Cyan)) // Cyan for system messages
		}
	case MessageTypeWSP:
		fmtedMsg := fmt.Sprintf("Whisper from %s: %s\n", payload.sender, payload.content)
		fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Purple)) // Purple for whisper messages
	default:
		fmtedMsg := fmt.Sprintf("%s: %s\n", payload.sender, payload.content)
		fmt.Print(color.ColorifyWithTimestamp(fmtedMsg, color.Blue)) // Blue for group messages
	}
}

func askForInput() {
	coloredPrompt := color.ColorifyWithTimestamp("You:", color.Yellow)
	fmt.Printf("%s ", coloredPrompt)
}
