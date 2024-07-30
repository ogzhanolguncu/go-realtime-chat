package main

import (
	"context"
	"fmt"
	"io"
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
			if err.Error() == io.EOF.Error() {
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
	defer client.Close()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect server: %v", err)
	}

	if err := client.SetUsername(); err != nil {
		return fmt.Errorf("failed to set username: %v", err)
	}

	chatUI := chat_ui.NewChatUI(client.GetUsername())
	defer chatUI.Close()

	header, commandBox, chatBox, inputBox, userList, err := chatUI.InitUI()
	if err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	chatUI.UpdateChatBox(fmt.Sprintf("[%s] [Welcome to the chat!](fg:cyan)", time.Now().Format("15:04")), chatBox)

	draw := chatUI.Draw(header, commandBox, chatBox, inputBox, userList)
	draw()

	uiEvents := ui.PollEvents()
	incomingChan := make(chan protocol.Payload)
	errorChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This will signal all operations to stop

	go client.FetchChatHistory()
	go client.ReadMessages(ctx, incomingChan, errorChan)

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<Up>":
				userList.ScrollUp()
			case "<Down>":
				userList.ScrollDown()
			case "<MouseWheelUp>":
				chatUI.ScrollChatBox(chatBox, -1)
			case "<MouseWheelDown>":
				chatUI.ScrollChatBox(chatBox, 1)
			case "<C-c>":
				return nil
			case "<Enter>":
				if chatUI.IsInputMode() && len(inputBox.Text) > 0 {
					if inputBox.Text == "/quit" {
						return nil
					}
					if inputBox.Text == "/clear" {
						chatUI.ClearChatBox(chatBox)
						inputBox.Text = ""
						draw()   // Redraw the UI
						continue // Skip the rest of the loop iteration
					}
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
			if client.CheckIfUserMuted(payload.Sender) {
				continue
			}
			if payload.MessageType == protocol.MessageTypeHSTRY {
				if len(payload.DecodedChatHistory) != 0 {
					chatUI.UpdateChatBox("---- CHAT HISTORY ----", chatBox)
				}
				for _, v := range payload.DecodedChatHistory {
					chatUI.UpdateChatBox(client.HandleReceive(v), chatBox)
				}
				if len(payload.DecodedChatHistory) != 0 {
					chatUI.UpdateChatBox("---- CHAT HISTORY ----", chatBox)
				}
				draw()
				continue
			}
			if payload.MessageType == protocol.MessageTypeACT_USRS {
				// If we recieve MessageTypeACT_USRS it means either someone joined or left. We have to update userList UI.
				fakeNames := []string{
					"Alice", "Bob", "Charlie", "David", "Eve",
					"Frank", "Grace", "Henry", "Ivy", "Jack",
					"Kate", "Liam", "Mia", "Noah", "Olivia",
					"Peter", "Quinn", "Rachel", "Sam", "Tina",
					"Ursula", "Victor", "Wendy", "Xander", "Yara",
				}
				payload.ActiveUsers = append(payload.ActiveUsers, fakeNames...)
				chatUI.UpdateUserList(userList, payload.ActiveUsers)
				draw()
				continue
			}
			chatUI.UpdateChatBox(client.HandleReceive(payload), chatBox)
		case err := <-errorChan:
			return err
		}
		draw()
	}
}
