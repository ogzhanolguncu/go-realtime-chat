// color/color.go
package color

import (
	"fmt"
	"time"
)

// ANSI escape codes for colors and styles
const (
	Reset        = "\033[0m"
	Bold         = "\033[1m"
	Blink        = "\033[5m"
	Red          = "\033[31m"
	Yellow       = "\033[33m"
	White        = "\033[37m"
	BoldRed      = "\033[31;1m"
	BlinkRed     = "\033[31;1;5m"
	Cyan         = "\033[36m"
	Purple       = "\033[35m"
	Green        = "\033[32m"
	Magenta      = "\033[35m"
	Blue         = "\033[34m"
	BoldBlue     = "\033[34;1m"
	BlinkBlue    = "\033[34;1;5m"
	BrightYellow = "\033[33;1m"

	ClearScreen = "\033[2J"
	MoveCursor  = "\033[%d;%dH"
)

// Colorify applies the given color to the text
func Colorify(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

// ColorifyWithTimestamp adds a timestamp and applies the given color to the text
func ColorifyWithTimestamp(text, color string) string {
	timestamp := time.Now().Format("[15:04]")
	return fmt.Sprintf("\r%s %s%s%s", timestamp, color, text, Reset)
}
