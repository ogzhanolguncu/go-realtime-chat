package chat_ui

import (
	"fmt"

	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type ChatUI struct {
	inputMode            bool
	userListScrollOffset int
	chatScrollOffset     int
	chatMessages         []string
	currentUserName      string
}

func NewChatUI(username string) *ChatUI {
	return &ChatUI{
		userListScrollOffset: 0,
		chatScrollOffset:     0,
		inputMode:            true,
		chatMessages:         []string{},
		currentUserName:      username,
	}
}

func (cu *ChatUI) InitUI() (header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph, userList *widgets.List, err error) {
	if err := ui.Init(); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to initialize termui: %v", err)
	}
	p1, p2, l1, p3, l2 := cu.prepareUIItems()
	return p1, p2, l1, p3, l2, nil
}

func (cu *ChatUI) Draw(header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph, userList *widgets.List) func() {
	return func() {
		ui.Render(header, commandBox, chatBox, inputBox, userList)
	}
}

func (cu *ChatUI) prepareUIItems() (header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph, userList *widgets.List) {
	termWidth, termHeight := ui.TerminalDimensions()

	// Header
	header = widgets.NewParagraph()
	header.Text = fmt.Sprintf("Welcome to chatroom, %s", cu.currentUserName)
	header.SetRect(0, 0, termWidth, 3)
	header.Border = true
	header.TextStyle.Fg = ui.ColorYellow
	header.BorderStyle.Fg = ui.ColorCyan

	// Command Box
	commandBox = widgets.NewParagraph()
	commandBox.Title = "Commands"
	commandBox.Text = "/whisper, /reply, /clear, /quit, /mute, /unmute"
	commandBox.SetRect(0, 3, termWidth*3/4, 6)
	commandBox.Border = true
	commandBox.TitleStyle.Fg = ui.ColorYellow
	commandBox.BorderStyle.Fg = ui.ColorCyan
	commandBox.TextStyle.Fg = ui.ColorWhite
	commandBox.WrapText = true

	// Chat Box
	chatBox = widgets.NewParagraph()
	chatBox.Title = "Chat Messages"
	chatBox.SetRect(0, 6, termWidth*3/4, termHeight-3)
	chatBox.BorderStyle.Fg = ui.ColorCyan
	chatBox.TitleStyle.Fg = ui.ColorYellow
	chatBox.WrapText = true

	// Input Box
	inputBox = widgets.NewParagraph()
	inputBox.Title = "Type your message"
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)
	inputBox.TextStyle.Fg = ui.ColorWhite
	inputBox.BorderStyle.Fg = ui.ColorCyan
	inputBox.TitleStyle.Fg = ui.ColorYellow

	// User List
	userList = widgets.NewList()
	userList.Title = "Active Users"
	userList.Rows = nil
	userList.TextStyle = ui.NewStyle(ui.ColorWhite)
	userList.WrapText = false
	userList.SetRect(termWidth*3/4, 3, termWidth, termHeight-3)
	userList.BorderStyle.Fg = ui.ColorCyan
	userList.TitleStyle.Fg = ui.ColorYellow
	userList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)

	return header, commandBox, chatBox, inputBox, userList
}

func (cu *ChatUI) UpdateChatBox(input string, chatBox *widgets.Paragraph) {
	cu.chatMessages = append(cu.chatMessages, input)
	cu.chatScrollOffset = len(cu.chatMessages) - (chatBox.Inner.Dy() - 1)
	if cu.chatScrollOffset < 0 {
		cu.chatScrollOffset = 0
	}
	cu.refreshChatBox(chatBox)
}

func (cu *ChatUI) refreshChatBox(chatBox *widgets.Paragraph) {
	visibleLines := chatBox.Inner.Dy() - 1
	if cu.chatScrollOffset+visibleLines > len(cu.chatMessages) {
		visibleLines = len(cu.chatMessages) - cu.chatScrollOffset
	}
	chatBox.Text = strings.Join(cu.chatMessages[cu.chatScrollOffset:cu.chatScrollOffset+visibleLines], "\n")
}

func (cu *ChatUI) ResizeUI(header *widgets.Paragraph, commandBox *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph, userList *widgets.List) {
	termWidth, termHeight := ui.TerminalDimensions()

	header.SetRect(0, 0, termWidth, 3)
	commandBox.SetRect(0, 3, termWidth*3/4, 13)
	chatBox.SetRect(0, 13, termWidth*3/4, termHeight-3)
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)
	userList.SetRect(termWidth*3/4, 3, termWidth, termHeight-3)

	ui.Clear()
	cu.Draw(header, commandBox, chatBox, inputBox, userList)
}

func (cu *ChatUI) ScrollChatBox(chatBox *widgets.Paragraph, direction int) {
	cu.chatScrollOffset += direction
	if cu.chatScrollOffset < 0 {
		cu.chatScrollOffset = 0
	}

	visibleLines := chatBox.Inner.Dy() - 1
	totalMessages := len(cu.chatMessages)

	if totalMessages <= visibleLines {
		cu.chatScrollOffset = 0
	} else if cu.chatScrollOffset > totalMessages-visibleLines {
		cu.chatScrollOffset = totalMessages - visibleLines
	}

	cu.refreshChatBox(chatBox)
}

func (cu *ChatUI) UpdateUserList(userList *widgets.List, users []string) {
	userList.Rows = users
}

func (cu *ChatUI) ClearChatBox(chatBox *widgets.Paragraph) {
	cu.chatMessages = []string{}
	cu.chatScrollOffset = 0
	cu.refreshChatBox(chatBox)
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
