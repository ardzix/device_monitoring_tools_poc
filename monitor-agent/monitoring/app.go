package monitoring

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type AppMonitor struct {
	lastApp   string
	lastTitle string
	startTime time.Time
	isActive  bool
	interval  time.Duration
}

func NewAppMonitor() *AppMonitor {
	interval := 1 // default interval
	if envInterval := os.Getenv("APP_MONITOR_INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			interval = i
		}
	}
	return &AppMonitor{
		startTime: time.Now(),
		interval:  time.Duration(interval) * time.Second,
	}
}

func (m *AppMonitor) GetActiveApp() (string, string, error) {
	// Mock implementation for Linux
	// In a real implementation, this would use platform-specific methods
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting active window: %v", err)
	}

	title := strings.TrimSpace(string(output))
	// For demo purposes, we'll use the window title as both app name and title
	return title, title, nil
}

func (m *AppMonitor) Monitor(ch chan<- map[string]interface{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		appName, windowTitle, err := m.GetActiveApp()
		if err != nil {
			log.Printf("Error getting active app: %v", err)
			continue
		}

		if appName != m.lastApp || windowTitle != m.lastTitle {
			// If there was a previous app running, send its usage data
			if m.lastApp != "" {
				duration := int(time.Since(m.startTime).Seconds())
				ch <- map[string]interface{}{
					"app_name":     m.lastApp,
					"window_title": m.lastTitle,
					"duration":     duration,
					"is_active":    m.isActive,
				}
			}
			// Update the current app info
			m.lastApp = appName
			m.lastTitle = windowTitle
			m.startTime = time.Now()
		}
	}
}
