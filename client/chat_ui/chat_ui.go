package chat_ui

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type ChatUI struct {
	inputMode            bool
	messages             []string
	userListScrollOffset int
}

func (cu *ChatUI) InitUI() (*widgets.Paragraph, *widgets.Paragraph, *widgets.List, *widgets.Paragraph, *widgets.List, error) {
	if err := ui.Init(); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to initialize termui: %v", err)
	}

	p1, p2, l1, p3, l2 := cu.prepareUIItems()
	return p1, p2, l1, p3, l2, nil
}

func (cu *ChatUI) Draw(header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.List, inputBox *widgets.Paragraph, userList *widgets.List) func() {
	return func() {
		ui.Render(header, commandBox, chatBox, inputBox, userList)
	}
}

func (cu *ChatUI) prepareUIItems() (header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.List, inputBox *widgets.Paragraph, userList *widgets.List) {
	termWidth, termHeight := ui.TerminalDimensions()

	// Header
	header = widgets.NewParagraph()
	header.Text = "WELCOME TO CHATROOM"
	header.SetRect(0, 0, termWidth, 3)
	header.Border = true
	header.TextStyle.Fg = ui.ColorYellow
	header.BorderStyle.Fg = ui.ColorCyan

	// Command Box
	commandBox = widgets.NewParagraph()
	commandBox.Title = "Available Commands"
	commandBox.Text = "/whisper <recipient> <message> - Send a private message\n" +
		"/reply <message>              - Reply to the last whisper\n" +
		"/clear                        - Clear the screen\n" +
		"/users                        - Show active users\n" +
		"/help                         - Show commands\n" +
		"/quit                         - Exit the chat\n\n" +
		"To send a public message, just type and press Enter"
	commandBox.SetRect(0, 3, termWidth*3/4, 13)
	commandBox.Border = true
	commandBox.TitleStyle.Fg = ui.ColorGreen
	commandBox.BorderStyle.Fg = ui.ColorWhite

	// Chat Box
	chatBox = widgets.NewList()
	chatBox.Title = "Chat Messages"
	chatBox.SetRect(0, 13, termWidth*3/4, termHeight-3)
	chatBox.BorderStyle.Fg = ui.ColorCyan
	chatBox.TitleStyle.Fg = ui.ColorYellow
	chatBox.WrapText = true

	// Input Box
	inputBox = widgets.NewParagraph()
	inputBox.Title = "Type your message"
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)
	inputBox.TextStyle.Fg = ui.ColorGreen
	inputBox.BorderStyle.Fg = ui.ColorCyan
	inputBox.TitleStyle.Fg = ui.ColorYellow

	// User List
	userList = widgets.NewList()
	userList.Title = "Active Users"
	//TODO: add users to here
	userList.Rows = nil
	userList.TextStyle = ui.NewStyle(ui.ColorGreen)
	userList.WrapText = false
	userList.SetRect(termWidth*3/4, 3, termWidth, termHeight-3)
	userList.BorderStyle.Fg = ui.ColorCyan
	userList.TitleStyle.Fg = ui.ColorYellow

	return header, commandBox, chatBox, inputBox, userList
}
