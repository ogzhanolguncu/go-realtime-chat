package ui_manager

import (
	"context"
	"log"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	ui_manager "github.com/ogzhanolguncu/go-chat/client/ui_manager/components"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func HandleLoginUI(client *internal.Client) (bool, error) {
	lu := ui_manager.NewLoginUI()
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
				return true, nil
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
				return false, nil // Login successful
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
