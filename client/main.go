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
	"github.com/ogzhanolguncu/go-chat/client/login_ui"
	"github.com/ogzhanolguncu/go-chat/client/terminal"
	"github.com/ogzhanolguncu/go-chat/client/utils"
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

	if err := handleLoginUI(client); err != nil {
		return err
	}
	return handleChatUI(client)
}

func handleChatUI(client *internal.Client) error {
	chatUI := chat_ui.NewChatUI(client.GetUsername())
	defer chatUI.Close()

	header, commandBox, chatBox, inputBox, userList, err := chatUI.InitUI()
	if err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
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
				return nil
			case "<Enter>":
				if chatUI.IsInputMode() && len(inputBox.Text) > 0 {
					if inputBox.Text == "/quit" {
						return nil
					}
					if inputBox.Text == "/clear" {
						chatUI.ClearChatBox(chatBox)
						inputBox.Text = ""
						draw()
						continue
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
			if payload.MessageType == protocol.MessageTypeWSP {
				notificationMsg := payload.Content
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
			return err
		}
		draw()
	}
}

func handleLoginUI(client *internal.Client) error {
	lu := login_ui.NewLoginUI()
	defer lu.Close()

	container, description, usernameBox, passwordBox, errorBox, err := lu.InitUI()
	if err != nil {
		log.Fatalf("Failed to initialize UI: %v", err)
	}

	draw := lu.Draw(container, description, usernameBox, errorBox, passwordBox)

	responseChan := make(chan protocol.Payload)
	errorChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.ReadMessages(ctx, responseChan, errorChan)
	uiEvents := ui.PollEvents()

	var loginAttemptInProgress bool
	var showLoader bool
	loaderFrame := 0

	loaderTicker := time.NewTicker(100 * time.Millisecond)
	defer loaderTicker.Stop()

	cursorTicker := time.NewTicker(500 * time.Millisecond)
	defer cursorTicker.Stop()

	for {
		draw()
		select {
		case <-cursorTicker.C:
			lu.ToggleCursor()
			lu.UpdateUsernameBox(usernameBox)
			lu.UpdatePasswordBox(passwordBox)
		case <-loaderTicker.C:
			if showLoader {
				loaderFrame++
				lu.UpdateLoader(errorBox, loaderFrame)
			}
		case e := <-uiEvents:
			switch e.ID {
			case "<C-c>":
				return nil
			case "<Tab>":
				lu.SwitchField()
				lu.ResetErrorBox(errorBox)

			case "<Enter>":
				if !loginAttemptInProgress {
					username, password := lu.GetCredentials()
					loginAttemptInProgress = true
					showLoader = true
					client.SendUsernameReq(username, password)
				}
			case "<Backspace>":
				lu.DeleteLastChar()
				lu.ResetErrorBox(errorBox)

			case "<Space>":
				lu.UpdateCurrentField(" ")
				lu.ResetErrorBox(errorBox)

			default:
				if len(e.ID) == 1 {
					lu.UpdateCurrentField(string(e.ID[0]))
					lu.ResetErrorBox(errorBox)
				}
			}
		case payload := <-responseChan:
			loginAttemptInProgress = false
			showLoader = false

			switch payload.Status {
			case "success":
				client.SetUsername(payload.Username)
				return nil // Login successful
			case "fail":
				errorBox.Text = payload.Username
				errorBox.TextStyle.Fg = ui.ColorRed
			default:
				lu.ResetErrorBox(errorBox)
			}
		case err := <-errorChan:
			loginAttemptInProgress = false
			showLoader = false

			errorBox.Text = err.Error()
			errorBox.TextStyle.Fg = ui.ColorYellow
		}

		lu.UpdateUsernameBox(usernameBox)
		lu.UpdatePasswordBox(passwordBox)
	}
}
