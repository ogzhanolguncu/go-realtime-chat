package internal

import (
	"fmt"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func PrintHeader() {
	fmt.Printf("\n\n")
	fmt.Println("---------WELCOME TO CHATROOM--------")
	fmt.Println("------------------------------------")
	fmt.Println("Available commands:")
	fmt.Println("/whisper <recipient> <message> - Send a private message")
	fmt.Println("/reply <message>               - Reply to the last whisper")
	fmt.Println("/clear                         - Clear the screen")
	fmt.Println("/users                         - Show active users")
	fmt.Println("/quit                          - Exit the chat")
	fmt.Println("")
	fmt.Println("To send a public message, just type and press Enter")
	fmt.Println("------------------------------------")
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

	switch payload.MessageType {
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
	case protocol.MessageTypeACT_USRS:
		if payload.Status == "res" {
			message = fmt.Sprintf("Active users: %s\n", strings.Join(payload.ActiveUsers, ", "))
			colorCode = color.Blue
		}
	default:
		message = fmt.Sprintf("%s: %s\n", payload.Sender, payload.Content)
		colorCode = color.Blue
	}
	fmt.Print(color.ColorifyWithTimestamp(message, colorCode))
}
