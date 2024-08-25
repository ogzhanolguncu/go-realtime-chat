package ui_manager

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type ChannelUI struct {
	chatScrollOffset int
	chatMessages     []string
	currentUserName  string
	channelName      string
	showCursor       bool
	cursorChar       string
	inputText        string
}

func NewChannelUI(username, channelName string) *ChannelUI {
	return &ChannelUI{
		chatScrollOffset: 0,
		chatMessages:     []string{},
		currentUserName:  username,
		channelName:      channelName,
		showCursor:       true,
		cursorChar:       "|",
		inputText:        "",
	}
}

func (cu *ChannelUI) InitUI() (header *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph, err error) {
	if err := ui.Init(); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize termui: %v", err)
	}
	h, c, i := cu.prepareUIItems()
	return h, c, i, nil
}

func (cu *ChannelUI) Draw(header *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph) func() {
	return func() {
		ui.Render(header, chatBox, inputBox)
	}
}

func (cu *ChannelUI) prepareUIItems() (header *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph) {
	termWidth, termHeight := ui.TerminalDimensions()

	header = widgets.NewParagraph()
	header.Text = fmt.Sprintf("Channel: %s | User: %s", cu.channelName, cu.currentUserName)
	header.SetRect(0, 0, termWidth, 3)
	header.Border = true
	header.TextStyle.Fg = ui.ColorYellow
	header.BorderStyle.Fg = ui.ColorCyan

	chatBox = widgets.NewParagraph()
	chatBox.Title = "Messages"
	chatBox.SetRect(0, 3, termWidth, termHeight-3)
	chatBox.BorderStyle.Fg = ui.ColorCyan
	chatBox.TitleStyle.Fg = ui.ColorYellow
	chatBox.WrapText = true

	inputBox = widgets.NewParagraph()
	inputBox.Title = "Input"
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)
	inputBox.TextStyle.Fg = ui.ColorWhite
	inputBox.BorderStyle.Fg = ui.ColorCyan
	inputBox.TitleStyle.Fg = ui.ColorYellow

	return header, chatBox, inputBox
}

func (cu *ChannelUI) UpdateChatBox(input string, chatBox *widgets.Paragraph) {
	cu.chatMessages = append(cu.chatMessages, input)
	cu.chatScrollOffset = len(cu.chatMessages) - (chatBox.Inner.Dy() - 1)
	if cu.chatScrollOffset < 0 {
		cu.chatScrollOffset = 0
	}
	cu.refreshChatBox(chatBox)
}

func (cu *ChannelUI) refreshChatBox(chatBox *widgets.Paragraph) {
	visibleLines := chatBox.Inner.Dy() - 1
	if cu.chatScrollOffset+visibleLines > len(cu.chatMessages) {
		visibleLines = len(cu.chatMessages) - cu.chatScrollOffset
	}
	chatBox.Text = strings.Join(cu.chatMessages[cu.chatScrollOffset:cu.chatScrollOffset+visibleLines], "\n")
}

func (cu *ChannelUI) ResizeUI(header *widgets.Paragraph, chatBox *widgets.Paragraph, inputBox *widgets.Paragraph) {
	termWidth, termHeight := ui.TerminalDimensions()

	header.SetRect(0, 0, termWidth, 3)
	chatBox.SetRect(0, 3, termWidth, termHeight-3)
	inputBox.SetRect(0, termHeight-3, termWidth, termHeight)

	ui.Clear()
	cu.Draw(header, chatBox, inputBox)
}

func (cu *ChannelUI) ScrollChatBox(chatBox *widgets.Paragraph, direction int) {
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

func (cu *ChannelUI) ClearChatBox(chatBox *widgets.Paragraph) {
	cu.chatMessages = []string{}
	cu.chatScrollOffset = 0
	cu.refreshChatBox(chatBox)
}

func (cu *ChannelUI) Close() {
	ui.Close()
}

func (cu *ChannelUI) UpdateInputText(text string) {
	cu.inputText = text
}

func (cu *ChannelUI) GetInputText() string {
	return cu.inputText
}

func (cu *ChannelUI) ToggleCursor() {
	cu.showCursor = !cu.showCursor
}

func (cu *ChannelUI) RenderInput(inputBox *widgets.Paragraph) {
	if cu.showCursor {
		inputBox.Text = cu.inputText + cu.cursorChar
	} else {
		inputBox.Text = cu.inputText
	}
}

func (cu *ChannelUI) HandleKeyPress(key string) {
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
