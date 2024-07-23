package internal

import (
	"fmt"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/terminal"
	"github.com/ogzhanolguncu/go-chat/protocol"
	"golang.org/x/term"
)

const (
	headerWidth   = 70
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
	fmt.Print(terminal.Cyan)
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
	fmt.Print(terminal.Reset)
}

func PrintHeader(shouldClear bool) {
	if shouldClear {
		fmt.Print(terminal.ClearScreen)
		fmt.Print("\033[H") // Moves cursor to top left after clear
	}
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

func ColorifyAndFormatContent(payload protocol.Payload) {
	colorMap := map[protocol.MessageType]string{
		protocol.MessageTypeSYS:      terminal.Cyan,
		protocol.MessageTypeWSP:      terminal.Purple,
		protocol.MessageTypeUSR:      terminal.Yellow,
		protocol.MessageTypeMSG:      terminal.Green,
		protocol.MessageTypeACT_USRS: terminal.Blue,
	}

	var message string
	colorCode := colorMap[payload.MessageType]

	switch payload.MessageType {
	case protocol.MessageTypeSYS:
		message = fmt.Sprintf("System: %s\n", payload.Content)
		if payload.Status == "fail" {
			colorCode = terminal.Red
		}
		// Don't let it print with day and month. Use default time format.
		payload.Timestamp = 0
	case protocol.MessageTypeWSP:
		message = fmt.Sprintf("Whisper from %s: %s\n", payload.Sender, payload.Content)
		// Don't let it print with day and month. Use default time format.
		payload.Timestamp = 0
	case protocol.MessageTypeUSR:
		if payload.Status == "success" {
			fmt.Print("\033[F") // Move cursor up one line
			fmt.Print("\033[K") // Clear the entire line
			// Save cursor position
			fmt.Print("\033[s")
			// Move to the line right after ACTIVE USERS
			fmt.Print("\033[18;1H") // Adjust this number if needed
			// Restore cursor position
			fmt.Print("\033[u")
			// Don't let it print with day and month. Use default time format.
			payload.Timestamp = 0
		} else {
			message = fmt.Sprintf("%s\n", payload.Username)
			colorCode = terminal.Red
		}
	default:
		message = fmt.Sprintf("%s: %s\n", payload.Sender, payload.Content)
		// Don't let it print with day and month. Use default time format.
		payload.Timestamp = 0
	}

	if message != "" {
		fmt.Print(terminal.ColorifyWithTimestamp(message, colorCode, payload.Timestamp))
	}
}
