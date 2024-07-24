package chat_ui

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type ChatUI struct {
	inputMode            bool
	userListScrollOffset int
	config               Config
}

type Config struct {
	HeaderColor   ui.Color
	CommandColor  ui.Color
	ChatColor     ui.Color
	InputColor    ui.Color
	UserListColor ui.Color
}

func NewChatUI() *ChatUI {
	return &ChatUI{
		config: Config{
			HeaderColor:   ui.ColorMagenta, // Neon Pink
			CommandColor:  ui.ColorCyan,    // Neon Blue
			ChatColor:     ui.ColorGreen,   // Neon Green
			InputColor:    ui.ColorYellow,  // Neon Yellow
			UserListColor: ui.ColorRed,     // Neon Red
		},
		userListScrollOffset: 0,
		inputMode:            true,
	}
}

func (cu *ChatUI) InitUI() (header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.List, inputBox *widgets.Paragraph, userList *widgets.List, err error) {
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
	header.TextStyle.Fg = cu.config.HeaderColor
	header.BorderStyle.Fg = ui.ColorMagenta

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
	commandBox.TitleStyle.Fg = cu.config.CommandColor
	commandBox.BorderStyle.Fg = ui.ColorMagenta
	commandBox.WrapText = true

	// Chat Box
	chatBox = widgets.NewList()
	chatBox.Title = "Chat Messages"
	chatBox.SetRect(0, 13, termWidth*3/4, termHeight-3)
	chatBox.BorderStyle.Fg = ui.ColorMagenta
	chatBox.TitleStyle.Fg = cu.config.ChatColor
	chatBox.WrapText = true

	// Input Box
	inputBox = widgets.NewParagraph()
	inputBox.Title = "Type your message"
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)
	inputBox.TextStyle.Fg = cu.config.InputColor
	inputBox.BorderStyle.Fg = ui.ColorMagenta
	inputBox.TitleStyle.Fg = cu.config.InputColor

	// User List
	userList = widgets.NewList()
	userList.Title = "Active Users"
	userList.Rows = nil
	userList.TextStyle = ui.NewStyle(cu.config.UserListColor)
	userList.WrapText = false
	userList.SetRect(termWidth*3/4, 3, termWidth, termHeight-3)
	userList.BorderStyle.Fg = ui.ColorMagenta
	userList.TitleStyle.Fg = cu.config.UserListColor

	return header, commandBox, chatBox, inputBox, userList
}

func (cu *ChatUI) UpdateChatBox(input string, chatBox *widgets.List) {
	chatBox.Rows = append(chatBox.Rows, input)
}

func (cu *ChatUI) ResizeUI(header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.List, inputBox *widgets.Paragraph, userList *widgets.List) {
	termWidth, termHeight := ui.TerminalDimensions()

	header.SetRect(0, 0, termWidth, 3)
	commandBox.SetRect(0, 3, termWidth*3/4, 13)
	chatBox.SetRect(0, 13, termWidth*3/4, termHeight-3)
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)
	userList.SetRect(termWidth*3/4, 3, termWidth, termHeight-3)

	ui.Clear()
	cu.Draw(header, commandBox, chatBox, inputBox, userList)
}

func (cu *ChatUI) ScrollChatBox(chatBox *widgets.List, direction int) {
	chatBox.ScrollAmount(direction)
}

func (cu *ChatUI) UpdateUserList(userList *widgets.List, users []string) {
	userList.Rows = users
}

func (cu *ChatUI) ClearChatBox(chatBox *widgets.List) {
	chatBox.Rows = []string{}
}

func (cu *ChatUI) SetInputMode(mode bool) {
	cu.inputMode = mode
}

func (cu *ChatUI) IsInputMode() bool {
	return cu.inputMode
}

func (cu *ChatUI) Close() {
	ui.Close()
}
