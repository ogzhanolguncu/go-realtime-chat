package main

import (
	"log"

	ui "github.com/gizak/termui/v3"
	"github.com/joho/godotenv"

	"github.com/ogzhanolguncu/go-chat/client/chat_ui"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := internal.NewClient(internal.NewConfig())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("failed to connect server: %v", err)
	}

	if err := client.SetUsername(); err != nil {
		log.Fatalf("failed to set username: %v", err)
	}
	chatUI := chat_ui.NewChatUI()

	header, commandBox, chatBox, inputBox, userList, err := chatUI.InitUI()
	if err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer chatUI.Close()
	draw := chatUI.Draw(header, commandBox, chatBox, inputBox, userList)

	draw()
	uiEvents := ui.PollEvents()
	incomingChan := make(chan protocol.Payload)
	// outgoingChan := make(chan string)
	errChan := make(chan error, 1) // Buffered channel to prevent goroutine leak
	done := make(chan struct{})

	go client.ReadMessages(incomingChan, errChan, done)

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<C-c>":
				return
			case "<Enter>":
				if chatUI.IsInputMode() && len(inputBox.Text) > 0 {
					chatUI.UpdateChatBox(inputBox.Text, chatBox)
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
		// case msg := <-incomingChan:
		// 	if msg.MessageType == protocol.MessageTypeWSP {
		// 		client.UpdateLastWhispererFromGroupChat(msg.Sender)
		// 	}
		// 	coloredMessage := internal.ColorifyAndFormatContent(msg)
		// 	if coloredMessage != "" {
		// 		messages = append(messages, coloredMessage)
		// 		updateChatBox()
		// 	}

		// case msg := <-outgoingChan:
		// 	coloredMessage := fmt.Sprintf("[%s %s](fg:green)", time.Now().Format("15:04"), msg)
		// 	messages = append(messages, coloredMessage)
		// 	// updateChatBox()
		case err, ok := <-errChan:
			if !ok {
				return // Channel closed, exit loop
			}
			log.Fatal(err)
		}
		draw()
	}

}
