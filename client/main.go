package main

import (
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/joho/godotenv"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	"github.com/ogzhanolguncu/go-chat/client/terminal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = retry.Do(
		func() error {
			return runClient()
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			if err.Error() == "EOF" {
				err = fmt.Errorf("server is not responding")
			}
			fmt.Println(terminal.ColorifyWithTimestamp(fmt.Sprintf("Trying to reconnect, but %v", err), terminal.Red, 0))
		}),
	)

	if err != nil {
		log.Fatalf(terminal.ColorifyWithTimestamp(fmt.Sprintf("Failed after max retries: %v", err), terminal.Red, 0))
	}
}

func runClient() error {
	client, err := internal.NewClient(internal.NewConfig())
	if err != nil {
		fmt.Printf("failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		return err
	}

	if err := client.SetUsername(); err != nil {
		return fmt.Errorf("failed to set username: %v", err)
	}

	users, err := client.FetchActiveUsersAfterUsername()
	if err != nil {
		return fmt.Errorf("failed to fetch active users: %v", err)
	}

	if err := client.FetchGroupChatKey(); err != nil {
		return fmt.Errorf("failed to fetch chat history: %v", err)
	}

	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	header, commandBox, chatBox, inputBox, userList := prepareUIItems(users)

	draw := func() {
		ui.Render(header, commandBox, chatBox, inputBox, userList)
	}

	messages := []string{}
	updateChatBox := func() {
		chatBox.Rows = messages
	}

	inputMode := true
	draw()

	uiEvents := ui.PollEvents()
	incomingChan := make(chan protocol.Payload)
	outgoingChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		client.ReadMessages(incomingChan, errChan)
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
			case "<C-c>":
				return nil
			case "<Enter>":
				if inputMode && len(inputBox.Text) > 0 {
					client.HandleCommand(inputBox.Text, &messages)
					updateChatBox()
					inputBox.Text = ""
				}
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				header.SetRect(0, 0, payload.Width, 3)
				commandBox.SetRect(0, 3, payload.Width*3/4, 13)
				chatBox.SetRect(0, 13, payload.Width*3/4, payload.Height-3)
				inputBox.SetRect(0, payload.Height-3, payload.Width, payload.Height)
				userList.SetRect(payload.Width*3/4, 3, payload.Width, payload.Height-3)
				ui.Clear()
				draw()
			case "<Backspace>":
				if len(inputBox.Text) > 0 {
					inputBox.Text = inputBox.Text[:len(inputBox.Text)-1]
				}
			case "<Space>":
				inputBox.Text += " "
			default:
				if len(e.ID) == 1 {
					inputBox.Text += e.ID
				}
			}
		case msg := <-incomingChan:
			if msg.MessageType == protocol.MessageTypeWSP {
				client.UpdateLastWhispererFromGroupChat(msg.Sender)
			}
			coloredMessage := internal.ColorifyAndFormatContent(msg)
			if coloredMessage != "" {
				messages = append(messages, coloredMessage)
				updateChatBox()
			}

		case msg := <-outgoingChan:
			coloredMessage := fmt.Sprintf("[%s %s](fg:green)", time.Now().Format("15:04"), msg)
			messages = append(messages, coloredMessage)
			updateChatBox()
		case err, ok := <-errChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			return err
		}
		draw()
	}
}

func prepareUIItems(users []string) (*widgets.Paragraph, *widgets.Paragraph, *widgets.List, *widgets.Paragraph, *widgets.List) {
	termWidth, termHeight := ui.TerminalDimensions()

	// Header
	header := widgets.NewParagraph()
	header.Text = "WELCOME TO CHATROOM"
	header.SetRect(0, 0, termWidth, 3)
	header.Border = true
	header.TextStyle.Fg = ui.ColorYellow
	header.BorderStyle.Fg = ui.ColorCyan

	// Command Box
	commandBox := widgets.NewParagraph()
	commandBox.Title = "Available Commands"
	commandBox.Text = "/whisper <recipient> <message> - Send a private message\n" +
		"/reply <message>              - Reply to the last whisper\n" +
		"/clear                        - Clear the screen\n" +
		"/users                        - Show active users\n" +
		"/help                         - Show commands\n" +
		"/quit                         - Exit the chat\n\n" +
		"To send a public message, just type and press Enter"
	commandBox.SetRect(0, 3, termWidth*3/4, 13)
	commandBox.Border = true
	commandBox.TitleStyle.Fg = ui.ColorGreen
	commandBox.BorderStyle.Fg = ui.ColorWhite

	// Chat Box
	chatBox := widgets.NewList()
	chatBox.Title = "Chat Messages"
	chatBox.SetRect(0, 13, termWidth*3/4, termHeight-3)
	chatBox.BorderStyle.Fg = ui.ColorCyan
	chatBox.TitleStyle.Fg = ui.ColorYellow
	chatBox.WrapText = true

	// Input Box
	inputBox := widgets.NewParagraph()
	inputBox.Title = "Type your message"
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)
	inputBox.TextStyle.Fg = ui.ColorGreen
	inputBox.BorderStyle.Fg = ui.ColorCyan
	inputBox.TitleStyle.Fg = ui.ColorYellow

	// User List
	userList := widgets.NewList()
	userList.Title = "Active Users"
	userList.Rows = users
	userList.TextStyle = ui.NewStyle(ui.ColorGreen)
	userList.WrapText = false
	userList.SetRect(termWidth*3/4, 3, termWidth, termHeight-3)
	userList.BorderStyle.Fg = ui.ColorCyan
	userList.TitleStyle.Fg = ui.ColorYellow

	return header, commandBox, chatBox, inputBox, userList
}
