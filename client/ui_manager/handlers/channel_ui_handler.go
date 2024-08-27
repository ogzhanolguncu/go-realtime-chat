package ui_manager

import (
	"context"
	"fmt"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	ui_manager "github.com/ogzhanolguncu/go-chat/client/ui_manager/components"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func HandleChannelUI(client *internal.Client, channelName string) error {
	channelUi := ui_manager.NewChannelUI(client.GetUsername(), channelName)
	defer channelUi.Close()

	header, chatBox, inputBox, err := channelUi.InitUI()
	if err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	channelUi.UpdateChatBox(fmt.Sprintf("[%s] [Welcome to the %s!](fg:cyan)", time.Now().Format("01-02 15:04"), channelName), chatBox)
	draw := channelUi.Draw(header, chatBox, inputBox)
	draw()

	uiEvents := ui.PollEvents()
	incomingChan := make(chan protocol.Payload)
	errorChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.ReadMessages(ctx, incomingChan, errorChan)
	cursorTicker := time.NewTicker(500 * time.Millisecond)
	defer cursorTicker.Stop()

	for {
		select {
		case <-cursorTicker.C:
			channelUi.ToggleCursor()
			channelUi.RenderInput(inputBox)
			draw()
		case e := <-uiEvents:
			switch e.ID {
			case "<MouseWheelUp>":
				channelUi.ScrollChatBox(chatBox, -1)
			case "<MouseWheelDown>":
				channelUi.ScrollChatBox(chatBox, 1)
			case "<C-c>":
				return nil
			case "<Enter>":
				if len(channelUi.GetInputText()) > 0 {
					inputText := channelUi.GetInputText()
					if inputText == "/quit" {
						return nil
					}
					if inputText == "/clear" {
						channelUi.ClearChatBox(chatBox)
						channelUi.UpdateInputText("")
					} else {
						formattedMessage := fmt.Sprintf("[%s] [You: %s](fg:cyan)", time.Now().Format("01-02 15:04"), inputText)
						channelUi.UpdateChatBox(formattedMessage, chatBox)
						channelUi.UpdateInputText("")
					}
				}
			case "<Space>":
				channelUi.HandleKeyPress("<Space>")
			default:
				channelUi.HandleKeyPress(e.ID)
			}
			channelUi.RenderInput(inputBox)
			draw()
		case payload := <-incomingChan:
			channelUi.UpdateChatBox(client.HandleReceive(payload), chatBox)
		case err := <-errorChan:
			return err
		}
	}
}
