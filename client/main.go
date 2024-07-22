package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func main() {
	client, err := internal.NewClient(internal.NewConfig())
	if err != nil {
		fmt.Printf("failed to create client: %v", err)
	}
	defer client.Close()

	client.Connect()
	internal.PrintHeader(true)

	if err := client.SetUsername(); err != nil {
		fmt.Printf("failed to set username: %v", err)
	}

	users, err := client.FetchActiveUsersAfterUsername()
	if err != nil {
		fmt.Printf("failed to fetch active users: %v", err)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	header, commandBox, chatBox, inputBox, userList := prepareUIItems(users)

	draw := func() {
		ui.Render(header, commandBox, chatBox, inputBox, userList)
	}

	messages := []string{}
	updateChatBox := func() {
		chatBox.Text = strings.Join(messages, "\n")
	}

	inputMode := true
	draw()

	uiEvents := ui.PollEvents()
	incomingChan := make(chan protocol.Payload)
	outgoingChan := make(chan string)
	errChan := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		client.ReadMessages(incomingChan, errChan, done)
	}()

	userListScrollOffset := 0

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<Up>":
				userListScrollOffset--
				if userListScrollOffset < 0 {
					userListScrollOffset = len(userList.Rows) - 1
				}
				userList.ScrollTop()
				for i := 0; i < userListScrollOffset; i++ {
					userList.ScrollDown()
				}
			case "<Down>":
				userListScrollOffset++
				if userListScrollOffset >= len(userList.Rows) {
					userListScrollOffset = 0
				}
				userList.ScrollTop()
				for i := 0; i < userListScrollOffset; i++ {
					userList.ScrollDown()
				}
			case "q", "<C-c>":
				close(done)
				return
			case "<Enter>":
				if inputMode && len(inputBox.Text) > 0 {
					if strings.HasPrefix(inputBox.Text, "/") {
						handleCommand(inputBox.Text, &messages, client, userList)
					} else {
						messages = append(messages, fmt.Sprintf("[%s] You: %s", time.Now().Format("15:04:05"), inputBox.Text))
						client.SendMessages(inputBox.Text)
					}
					updateChatBox()
					inputBox.Text = ""
				}
			default:
				if len(e.ID) == 1 {
					inputBox.Text += e.ID
				}
			}
		case msg := <-outgoingChan:
			messages = append(messages, msg)
			updateChatBox()
		case msg := <-incomingChan:
			messages = append(messages, msg.Content)
			updateChatBox()
		}
		draw()
	}
}

func prepareUIItems(users []string) (*widgets.Paragraph, *widgets.Paragraph, *widgets.Paragraph, *widgets.Paragraph, *widgets.List) {
	header := widgets.NewParagraph()
	header.Text = "WELCOME TO CHATROOM"
	header.SetRect(0, 0, 100, 3)
	header.Border = true

	commandBox := widgets.NewParagraph()
	commandBox.Title = "Available commands:"
	commandBox.Text = "/whisper <recipient> <message> - Send a private message\n" +
		"/reply <message>              - Reply to the last whisper\n" +
		"/clear                        - Clear the screen\n" +
		"/users                        - Show active users\n" +
		"/help                         - Show commands\n" +
		"/quit                         - Exit the chat\n\n" +
		"To send a public message, just type and press Enter"
	commandBox.SetRect(0, 3, 75, 13)
	commandBox.Border = true

	chatBox := widgets.NewParagraph()
	chatBox.Title = "Chat Messages"
	chatBox.SetRect(0, 13, 75, 27)
	chatBox.TextStyle.Fg = ui.ColorBlue
	chatBox.BorderStyle.Fg = ui.ColorCyan

	inputBox := widgets.NewParagraph()
	inputBox.Title = "Type your message"
	inputBox.SetRect(0, 27, 100, 30)
	inputBox.TextStyle.Fg = ui.ColorYellow
	inputBox.BorderStyle.Fg = ui.ColorWhite

	userList := widgets.NewList()
	userList.Title = "Active Users"
	userList.Rows = users
	userList.TextStyle = ui.NewStyle(ui.ColorYellow)
	userList.WrapText = false
	userList.SetRect(75, 3, 100, 27)

	return header, commandBox, chatBox, inputBox, userList
}

func handleCommand(cmd string, messages *[]string, client *internal.Client, userList *widgets.List) {
	parts := strings.Fields(cmd)
	switch parts[0] {
	case "/whisper":
		if len(parts) < 3 {
			*messages = append(*messages, "Usage: /whisper <recipient> <message>")
		} else {
			recipient := parts[1]
			message := strings.Join(parts[2:], " ")
			// Implement whisper functionality here
			*messages = append(*messages, fmt.Sprintf("Whispered to %s: %s", recipient, message))
		}
	case "/reply":
		if len(parts) < 2 {
			*messages = append(*messages, "Usage: /reply <message>")
		} else {
			message := strings.Join(parts[1:], " ")
			// Implement reply functionality here
			*messages = append(*messages, fmt.Sprintf("Replied: %s", message))
		}
	case "/clear":
		*messages = []string{}
	case "/users":
		*messages = append(*messages, "Active users: "+strings.Join(userList.Rows, ", "))
	case "/help":
		*messages = append(*messages, "Available commands: /whisper, /reply, /clear, /users, /help, /quit")
	case "/quit":
		// This will be handled in the main loop
	default:
		*messages = append(*messages, "Unknown command. Type /help for available commands.")
	}
}
