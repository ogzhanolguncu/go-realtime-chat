package ui_manager

import (
	"context"
	"fmt"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	ui_manager "github.com/ogzhanolguncu/go-chat/client/ui_manager/components"
	"github.com/ogzhanolguncu/go-chat/client/utils"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

type ChannelName string

func HandleChatUI(client *internal.Client) (bool, ChannelName, error) {
	chatUI := ui_manager.NewChatUI(client.GetUsername())
	defer chatUI.Close()

	header, commandBox, chatBox, inputBox, userList, err := chatUI.InitUI()
	if err != nil {
		return false, "", fmt.Errorf("failed to initialize termui: %v", err)
	}
	chatUI.UpdateChatBox(fmt.Sprintf("[%s] [Welcome to the chat!](fg:cyan)", time.Now().Format("01-02 15:04")), chatBox)
	draw := chatUI.Draw(header, commandBox, chatBox, inputBox, userList)
	draw()

	uiEvents := ui.PollEvents()
	incomingChan := make(chan protocol.Payload)
	errorChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.FetchChatHistory()
	go client.FetchActiveUserList()
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
				return false, "", nil
			case "<Enter>":
				if chatUI.IsInputMode() && len(inputBox.Text) > 0 {
					if inputBox.Text == "/quit" {
						return false, "", nil
					}
					if inputBox.Text == "/clear" {
						chatUI.ClearChatBox(chatBox)
						inputBox.Text = ""
						draw()
						continue
					}
					message, err := client.HandleSend(inputBox.Text)
					if err != nil {
						return false, "", err
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
			// If its a channel message action and success status return true to switch to Channel UI
			if client.CheckIfSuccessfulChannel(payload) {
				return true, ChannelName(payload.ChannelPayload.ChannelName), nil
			}
			if payload.MessageType == protocol.MessageTypeWSP {
				notificationMsg := payload.Content
				// Make notification shorter to fit it into small notification window
				if len(payload.Content) >= 10 {
					notificationMsg = payload.Content[0:10] + "..."
				}
				go utils.NotifyUser(fmt.Sprintf("Whisper from %s", payload.Sender), notificationMsg, "/System/Library/Sounds/Purr.aiff")
			}
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
			return false, "", err
		}
		draw()
	}
}
