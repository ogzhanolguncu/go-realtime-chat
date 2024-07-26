package main

import (
	"context"
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
	client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("failed to connect server: %v", err)
	}

	if err := client.SetUsername(); err != nil {
		log.Fatalf("failed to set username: %v", err)
	}

	chatUI := chat_ui.NewChatUI(client.GetUsername())
	header, commandBox, chatBox, inputBox, userList, err := chatUI.InitUI()
	if err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer chatUI.Close()

	draw := chatUI.Draw(header, commandBox, chatBox, inputBox, userList)
	draw()

	uiEvents := ui.PollEvents()
	incomingChan := make(chan protocol.Payload)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This will signal all operations to stop

	go client.ReadMessages(ctx, incomingChan)

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<MouseWheelUp>":
				chatUI.ScrollChatBox(chatBox, -1)
			case "<MouseWheelDown>":
				chatUI.ScrollChatBox(chatBox, 1)
			case "<C-c>":
				return
			case "<Enter>":
				if chatUI.IsInputMode() && len(inputBox.Text) > 0 {
					message := client.HandleSend(inputBox.Text)
					chatUI.UpdateChatBox(message, chatBox)
					inputBox.Text = ""
				}
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
		case payload := <-incomingChan:
			//TODO format incoming messages
			chatUI.UpdateChatBox(client.HandleReceive(payload), chatBox)
		}
		draw()
	}
}
