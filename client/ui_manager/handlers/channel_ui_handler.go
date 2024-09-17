package ui_manager

import (
	"context"
	"fmt"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
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
	channelUi.UpdateHeader(header)
	draw := channelUi.Draw(header, chatBox, inputBox)
	draw()

	uiEvents := ui.PollEvents()

	incomingChan := make(chan protocol.Payload)
	errorChan := make(chan error, 1)
	exitChan := make(chan struct{})

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
				handleExit(client, channelUi, chatBox, exitChan)
			case "<Enter>":
				if len(channelUi.GetInputText()) > 0 {
					handleEnterKey(client, channelUi, chatBox, exitChan)
				}
			case "<Space>":
				channelUi.HandleKeyPress("<Space>")
			default:
				go client.HandleSend(fmt.Sprintf("/ch typing %s %s", client.GetChannelInfo().ChName, client.GetChannelInfo().ChPassword))
				channelUi.HandleKeyPress(e.ID)
			}
			channelUi.RenderInput(inputBox)
			draw()
		case payload := <-incomingChan:
			msg, shouldExit := client.HandleChReceive(payload)
			if strings.HasPrefix(msg, "T-") {
				username := strings.TrimPrefix(msg, "T-")
				channelUi.SetUserTyping(username)
				channelUi.UpdateHeader(header)
				draw()
				continue
			}
			if msg != "" {
				channelUi.UpdateChatBox(msg, chatBox)
				draw()
			}
			//Required for displaying message first otherwise function just quits
			if shouldExit {
				go func() {
					time.Sleep(1 * time.Second)
					exitChan <- struct{}{}
				}()
			}
		case err := <-errorChan:
			return err
		case <-exitChan:
			return nil
		}
	}
}

func handleExit(client *internal.Client, channelUi *ui_manager.ChannelUI, chatBox *widgets.Paragraph, exitChan chan struct{}) {
	chMsgPayload := fmt.Sprintf("/ch leave %s", client.GetChannelInfo().ChName)
	message, err := client.HandleSend(chMsgPayload)
	if err != nil {
		message = err.Error()
	}
	channelUi.UpdateChatBox(message, chatBox)
	channelUi.UpdateInputText("")
	go func() {
		time.Sleep(1 * time.Second)
		exitChan <- struct{}{}
	}()
}

func handleEnterKey(client *internal.Client, channelUi *ui_manager.ChannelUI, chatBox *widgets.Paragraph, exitChan chan struct{}) {
	inputText := channelUi.GetInputText()
	var message string
	var err error

	switch {
	case inputText == "/clear":
		channelUi.ClearChatBox(chatBox)
	case inputText == "/quit":
		handleExit(client, channelUi, chatBox, exitChan)
	case strings.HasPrefix(inputText, "/kick "):
		parts := strings.Fields(inputText)
		chMsgPayload := fmt.Sprintf("/ch kick %s %s %s", client.GetChannelInfo().ChName, client.GetChannelInfo().ChPassword, strings.TrimSpace(parts[1]))
		message, err = client.HandleSend(chMsgPayload)
	case strings.HasPrefix(inputText, "/ban "):
		parts := strings.Fields(inputText)
		chMsgPayload := fmt.Sprintf("/ch ban %s %s %s", client.GetChannelInfo().ChName, client.GetChannelInfo().ChPassword, strings.TrimSpace(parts[1]))
		message, err = client.HandleSend(chMsgPayload)
	case inputText == "/users":
		chMsgPayload := fmt.Sprintf("/ch users %s %s", client.GetChannelInfo().ChName, client.GetChannelInfo().ChPassword)
		message, err = client.HandleSend(chMsgPayload)
	default:
		chMsgPayload := fmt.Sprintf("/ch message %s %s %s", client.GetChannelInfo().ChName, client.GetChannelInfo().ChPassword, inputText)
		message, err = client.HandleSend(chMsgPayload)
	}

	if err != nil {
		message = err.Error()
	}
	if message != "" {
		channelUi.UpdateChatBox(message, chatBox)
	}
	channelUi.UpdateInputText("")
}
