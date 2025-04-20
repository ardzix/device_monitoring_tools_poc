package activity

import (
	"fmt"
	"os/exec"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) GetActiveWindowTitle() (string, error) {
	// Mock implementation for Linux
	// In a real Windows implementation, this would use Windows API calls
	cmd := exec.Command("xdotool", "getwindowfocus", "getwindowname")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get window title: %v", err)
	}
	return string(output), nil
}
