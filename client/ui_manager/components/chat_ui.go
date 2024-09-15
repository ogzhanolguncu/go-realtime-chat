package ui_manager

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
	rawChatMessages      []string // Required for input history
	currentUserName      string
	cursorVisible        bool
	inputText            string
}

func NewChatUI(username string) *ChatUI {
	return &ChatUI{
		userListScrollOffset: 0,
		chatScrollOffset:     0,
		inputMode:            true,
		chatMessages:         []string{},
		rawChatMessages:      []string{},
		currentUserName:      username,
		cursorVisible:        true,
		inputText:            "",
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
	commandBox.Text = "General:\n" +
		"  /clear - Clear chat                                  |  /quit - Exit app\n\n" +
		"User Interactions:\n" +
		"  /whisper <username> <message> - Send PM              |  /reply <message> - Reply to last PM\n" +
		"  /mute <username> - Hide messages                     |  /unmute <username> - Show messages\n" +
		"  /block <username> - Block user                       |  /unblock <username> - Unblock user\n\n" +
		"Channel Commands:\n" +
		"  /ch create <name> <password> <max_users> <public|private> - Create channel\n" +
		"  /ch join <name> <password> - Join channel            |  /ch leave - Leave current channel\n" +
		"  /ch users - List users in current channel            |  /ch list - Show active channels\n" +
		"Channel Owner Commands:\n" +
		"  /ch kick <username> - Kick user from channel         |  /ch ban <username> - Ban user from channel"
	commandBox.SetRect(0, 3, termWidth*3/4, 19)
	commandBox.Border = true
	commandBox.TitleStyle.Fg = ui.ColorYellow
	commandBox.BorderStyle.Fg = ui.ColorCyan
	commandBox.TextStyle.Fg = ui.ColorWhite
	commandBox.WrapText = true

	// Chat Box
	chatBox = widgets.NewParagraph()
	chatBox.Title = "Chat Messages"
	chatBox.SetRect(0, 19, termWidth*3/4, termHeight-3)
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
func (cu *ChatUI) UpdateRawChatBox(input string) {
	if len(cu.rawChatMessages) > 4 {
		cu.rawChatMessages = []string{input}
		return
	}
	cu.rawChatMessages = append(cu.rawChatMessages, input)
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
	commandBox.SetRect(0, 3, termWidth*3/4, 14)
	chatBox.SetRect(0, 14, termWidth*3/4, termHeight-3)
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

func (cu *ChatUI) UpdateInputText(text string) {
	cu.inputText = text
}

func (cu *ChatUI) GetInputText() string {
	return cu.inputText
}

func (cu *ChatUI) HandleKeyPress(key string) {
	switch key {
	case "<Backspace>":
		if len(cu.inputText) > 0 {
			cu.inputText = cu.inputText[:len(cu.inputText)-1]
		}
	case "<Space>":
		cu.inputText += " "
	default:
		if len(key) == 1 {
			cu.inputText += key
		}
	}
}

func (cu *ChatUI) ToggleCursor() {
	cu.cursorVisible = !cu.cursorVisible
}

func (cu *ChatUI) RenderInput(inputBox *widgets.Paragraph) {
	if cu.cursorVisible {
		inputBox.Text = cu.inputText + "|"
	} else {
		inputBox.Text = cu.inputText
	}
}

// This method keeps rotating raw chat messages when pressed <Up> to type faster.
func (cu *ChatUI) GetMessageFromHistory() func(reset bool) {
	position := len(cu.rawChatMessages) // Start from the end
	return func(reset bool) {
		if reset {
			position = len(cu.rawChatMessages) // Start from the end
		}
		if len(cu.rawChatMessages) == 0 {
			return // No messages to retrieve
		}

		// Move up in history (older messages)
		if position > 0 {
			position--
		} else {
			// If we've reached the oldest message, wrap around to the newest
			position = len(cu.rawChatMessages) - 1
		}

		cu.UpdateInputText(cu.rawChatMessages[position])
	}
}
