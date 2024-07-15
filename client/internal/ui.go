package internal

import (
	"fmt"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
	"golang.org/x/term"
)

const (
	headerWidth   = 70
	userBoxWidth  = 24
	separatorLine = "------------------------------------"
)

func printBox(width int, title string, content []string, align string) {
	box := []string{
		"+" + strings.Repeat("-", width-2) + "+",
		fmt.Sprintf("| %-*s |", width-4, title),
		"+" + strings.Repeat("-", width-2) + "+",
	}

	for _, line := range content {
		box = append(box, fmt.Sprintf("| %-*s |", width-4, line))
	}

	box = append(box, "+"+strings.Repeat("-", width-2)+"+")

	termWidth, _, _ := term.GetSize(0) // Implement this function to get terminal width
	fmt.Print(color.Cyan)
	for _, line := range box {
		switch align {
		case "right":
			fmt.Printf("%*s\n", termWidth, line)
		case "center":
			padding := (termWidth - len(line)) / 2
			fmt.Printf("%*s%s\n", padding, "", line)
		default:
			fmt.Println(line)
		}
	}
	fmt.Print(color.Reset)
}

func PrintHeader() {
	fmt.Print(color.ClearScreen)
	content := []string{
		"Available commands:",
		"/whisper <recipient> <message> - Send a private message",
		"/reply <message>               - Reply to the last whisper",
		"/clear                         - Clear the screen",
		"/users                         - Show active users",
		"/help                          - Show commands",
		"/quit                          - Exit the chat",
		"",
		"To send a public message, just type and press Enter",
	}
	printBox(headerWidth, "WELCOME TO CHATROOM", content, "left")
}

func printActiveUsers(users []string) {
	content := append([]string{}, users...)
	printBox(userBoxWidth, fmt.Sprintf("ACTIVE USERS (%d)", len(users)), content, "left")
}

func askForInput() {
	fmt.Println(separatorLine)
	coloredPrompt := color.ColorifyWithTimestamp("You:", color.Yellow)
	fmt.Printf("%s ", coloredPrompt)
	fmt.Print("\033[s") // Save cursor position
	fmt.Println()       // Move to next line
	fmt.Println(separatorLine)
	fmt.Print("\033[u")    // Restore cursor position
	fmt.Print("\033[?25h") // Show the cursor
}

func colorifyAndFormatContent(payload protocol.Payload) {
	colorMap := map[protocol.MessageType]string{
		protocol.MessageTypeSYS:      color.Cyan,
		protocol.MessageTypeWSP:      color.Purple,
		protocol.MessageTypeUSR:      color.Yellow,
		protocol.MessageTypeACT_USRS: color.Blue,
	}

	var message string
	colorCode := colorMap[payload.MessageType]

	switch payload.MessageType {
	case protocol.MessageTypeSYS:
		message = fmt.Sprintf("System: %s\n", payload.Content)
		if payload.Status == "fail" {
			colorCode = color.Red
		}
	case protocol.MessageTypeWSP:
		message = fmt.Sprintf("Whisper from %s: %s\n", payload.Sender, payload.Content)
	case protocol.MessageTypeUSR:
		if payload.Status == "success" {
			// message = fmt.Sprintf("Username successfully set to %s\n", payload.Username)
		} else {
			message = fmt.Sprintf("%s\n", payload.Username)
			colorCode = color.Red
		}
	case protocol.MessageTypeACT_USRS:
		if payload.Status == "res" {
			message = fmt.Sprintf("Active users: %s\n", strings.Join(payload.ActiveUsers, ", "))
		}
	default:
		message = fmt.Sprintf("%s: %s\n", payload.Sender, payload.Content)
	}

	if message != "" {
		fmt.Print(color.ColorifyWithTimestamp(message, colorCode))
	}
}
