package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go/v4"
	ui "github.com/gizak/termui/v3"
	"github.com/joho/godotenv"
	"github.com/ogzhanolguncu/go-chat/client/chat_ui"
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
		return fmt.Errorf("failed to create client: %v", err)
	}
	client.Close()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect server: %v", err)
	}

	if err := client.SetUsername(); err != nil {
		return fmt.Errorf("failed to set username: %v", err)
	}

	chatUI := chat_ui.NewChatUI(client.GetUsername())
	header, commandBox, chatBox, inputBox, userList, err := chatUI.InitUI()
	chatBox.Text += fmt.Sprintf("[%s] [System: Welcome %s to the chat!](fg:cyan)", time.Now().Format("15:04"), client.GetUsername())

	if err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
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
				return nil
			case "<Enter>":
				if chatUI.IsInputMode() && len(inputBox.Text) > 0 {
					message, err := client.HandleSend(inputBox.Text)
					if err != nil {
						return err
					}
					chatUI.UpdateChatBox(message, chatBox)
					inputBox.Text = ""
				}
			case "<Backspace>":
				if len(inputBox.Text) > 0 {
					inputBox.Text = inputBox.Text[:len(inputBox.Text)-1]
				}
			case "<Resize>":
				chatUI.ResizeUI(header, commandBox, chatBox, inputBox, userList)
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
