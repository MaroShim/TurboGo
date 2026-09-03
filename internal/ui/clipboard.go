package ui

import (
	"bytes"
	"os/exec"
	"runtime"
)

var internalClipboard string

// SetClipboard copies text to internal clipboard and system clipboard (pbcopy on macOS)
func SetClipboard(text string) {
	internalClipboard = text
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = bytes.NewBufferString(text)
		_ = cmd.Run()
	}
}

// GetClipboard retrieves text from system clipboard (pbpaste on macOS) or fallback to internal
func GetClipboard() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("pbpaste")
		out, err := cmd.Output()
		if err == nil {
			return string(out)
		}
	}
	return internalClipboard
}
