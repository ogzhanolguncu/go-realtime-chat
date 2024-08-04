package login_ui

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type LoginUI struct {
	username     string
	password     string
	currentField int // 0 for username, 1 for password
}

func NewLoginUI() *LoginUI {
	return &LoginUI{
		username:     "",
		password:     "",
		currentField: 0,
	}
}

func (lu *LoginUI) InitUI() (container *widgets.Paragraph, description *widgets.List, usernameBox *widgets.Paragraph, passwordBox *widgets.Paragraph, errorText *widgets.Paragraph, err error) {
	if err := ui.Init(); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to initialize termui: %v", err)
	}
	c, d, u, p, e := lu.prepareUIItems()
	return c, d, u, p, e, nil
}

func (lu *LoginUI) Draw(container *widgets.Paragraph, description *widgets.List, usernameBox *widgets.Paragraph, passwordBox *widgets.Paragraph, errorText *widgets.Paragraph) func() {
	return func() {
		ui.Render(container, description, usernameBox, passwordBox, errorText)
	}
}

func (lu *LoginUI) prepareUIItems() (container *widgets.Paragraph, description *widgets.List, usernameBox *widgets.Paragraph, passwordBox *widgets.Paragraph, errorText *widgets.Paragraph) {
	termWidth, termHeight := ui.TerminalDimensions()

	containerWidth := termWidth * 2 / 3
	containerHeight := 14
	containerStartX := (termWidth - containerWidth) / 2
	containerStartY := (termHeight - containerHeight) / 2

	// Container
	container = widgets.NewParagraph()
	container.SetRect(containerStartX, containerStartY, containerStartX+containerWidth, containerStartY+containerHeight)
	container.Border = true
	container.BorderStyle.Fg = ui.ColorBlue
	container.Title = "🔐 Secure Login"
	container.TitleStyle.Fg = ui.ColorYellow
	container.TitleStyle.Modifier = ui.ModifierBold

	// Description
	description = widgets.NewList()
	description.Title = "Instructions"
	description.Rows = []string{
		"• Enter your username and password",
		"• Use Tab to switch between fields",
		"• Press Enter to submit",
	}
	description.SetRect(containerStartX+1, containerStartY+1, containerStartX+containerWidth-1, containerStartY+6)
	description.BorderStyle.Fg = ui.ColorCyan
	description.TextStyle.Fg = ui.ColorWhite
	description.WrapText = true

	// Username Box
	usernameBox = widgets.NewParagraph()
	usernameBox.Title = "👤 Username"
	usernameBox.SetRect(containerStartX+1, containerStartY+6, containerStartX+containerWidth-1, containerStartY+9)
	usernameBox.BorderStyle.Fg = ui.ColorCyan
	usernameBox.TextStyle.Fg = ui.ColorWhite

	// Password Box
	passwordBox = widgets.NewParagraph()
	passwordBox.Title = "🔑 Password"
	passwordBox.SetRect(containerStartX+1, containerStartY+9, containerStartX+containerWidth-1, containerStartY+12)
	passwordBox.BorderStyle.Fg = ui.ColorCyan
	passwordBox.TextStyle.Fg = ui.ColorWhite

	errorText = widgets.NewParagraph()
	errorText.SetRect(containerStartX+1, containerStartY+12, containerStartX+containerWidth-1, containerStartY+13)
	errorText.TextStyle.Fg = ui.ColorRed
	errorText.Border = false

	return container, description, usernameBox, passwordBox, errorText
}

func (lu *LoginUI) UpdateUsernameBox(usernameBox *widgets.Paragraph) {
	usernameBox.Text = lu.username
	if lu.currentField == 0 {
		usernameBox.BorderStyle.Fg = ui.ColorYellow
	} else {
		usernameBox.BorderStyle.Fg = ui.ColorCyan
	}
}

func (lu *LoginUI) UpdatePasswordBox(passwordBox *widgets.Paragraph) {
	passwordBox.Text = strings.Repeat("•", len(lu.password))
	if lu.currentField == 1 {
		passwordBox.BorderStyle.Fg = ui.ColorYellow
	} else {
		passwordBox.BorderStyle.Fg = ui.ColorCyan
	}
}

func (lu *LoginUI) SwitchField() {
	lu.currentField = 1 - lu.currentField // Toggle between 0 and 1
}

func (lu *LoginUI) GetCurrentField() int {
	return lu.currentField
}

func (lu *LoginUI) UpdateCurrentField(char string) {
	if lu.currentField == 0 {
		lu.username += string(char)
	} else {
		lu.password += string(char)
	}
}

func (lu *LoginUI) DeleteLastChar() {
	if lu.currentField == 0 && len(lu.username) > 0 {
		lu.username = lu.username[:len(lu.username)-1]
	} else if lu.currentField == 1 && len(lu.password) > 0 {
		lu.password = lu.password[:len(lu.password)-1]
	}
}

func (lu *LoginUI) GetCredentials() (string, string) {
	return lu.username, lu.password
}

func (lu *LoginUI) Close() {
	ui.Close()
}
