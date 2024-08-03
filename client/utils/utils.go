package utils

import "os/exec"

func NotifyUser(title, message, soundPath string) {
	// Requires terminal notifier, install it from brew.
	cmd := exec.Command("terminal-notifier",
		"-title", title,
		"-message", message,
		"-activate", "com.apple.Terminal",
		"-timeout", "3") // Set timeout to 2 seconds
	cmdPlaySound := exec.Command("afplay", soundPath)

	cmd.Run()
	cmdPlaySound.Run()
}
