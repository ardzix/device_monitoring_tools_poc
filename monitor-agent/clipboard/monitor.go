package clipboard

import (
	"os/exec"
)

type Monitor struct {
	lastContent string
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start(ch chan<- string) {
	// Mock implementation for Linux
	// In a real Windows implementation, this would use Windows API calls
	for {
		cmd := exec.Command("xclip", "-o")
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		content := string(output)
		if content != m.lastContent {
			m.lastContent = content
			ch <- content
		}
	}
}
