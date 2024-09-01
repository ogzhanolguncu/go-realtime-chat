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

func HandleChannelUI(client *internal.Client) error {
	channelUi := ui_manager.NewChannelUI(client.GetUsername(), client.GetChannelInfo().ChName)
	defer channelUi.Close()

	header, chatBox, inputBox, err := channelUi.InitUI()
	if err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	channelUi.UpdateChatBox(fmt.Sprintf("[%s] [Welcome to the %s!](fg:cyan)", time.Now().Format("01-02 15:04"), client.GetChannelInfo().ChName), chatBox)
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
				chMsgPayload := fmt.Sprintf("/ch leave %s", client.GetChannelInfo().ChName)
				message, err := client.HandleSend(chMsgPayload)
				if err != nil {
					message = err.Error()
				}
				channelUi.UpdateChatBox(message, chatBox)
				channelUi.UpdateInputText("")
				return nil
			case "<Enter>":
				if len(channelUi.GetInputText()) > 0 {
					inputText := channelUi.GetInputText()
					if inputText == "/clear" {
						channelUi.ClearChatBox(chatBox)
						channelUi.UpdateInputText("")
					}
					if inputText == "/quit" {
						chMsgPayload := fmt.Sprintf("/ch leave %s", client.GetChannelInfo().ChName)
						message, err := client.HandleSend(chMsgPayload)
						if err != nil {
							message = err.Error()
						}
						channelUi.UpdateChatBox(message, chatBox)
						channelUi.UpdateInputText("")
						return nil
					}
					if inputText == "/users" {
						chMsgPayload := fmt.Sprintf("/ch users %s %s", client.GetChannelInfo().ChName, client.GetChannelInfo().ChPassword)
						message, err := client.HandleSend(chMsgPayload)
						if err != nil {
							message = err.Error()
						}
						channelUi.UpdateChatBox(message, chatBox)
						channelUi.UpdateInputText("")
					} else {
						chMsgPayload := fmt.Sprintf("/ch message %s %s %s", client.GetChannelInfo().ChName, client.GetChannelInfo().ChPassword, inputText)
						message, err := client.HandleSend(chMsgPayload)
						if err != nil {
							message = err.Error()
						}
						channelUi.UpdateChatBox(message, chatBox)
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
			channelUi.UpdateChatBox(client.HandleChReceive(payload), chatBox)
		case err := <-errorChan:
			return err
		}
	}
}
