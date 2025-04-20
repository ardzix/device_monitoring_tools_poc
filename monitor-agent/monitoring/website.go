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

type WebsiteMonitor struct {
	lastURL   string
	lastTitle string
	startTime time.Time
	interval  time.Duration
}

func NewWebsiteMonitor() *WebsiteMonitor {
	interval := 1 // default interval
	if envInterval := os.Getenv("WEBSITE_MONITOR_INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			interval = i
		}
	}
	return &WebsiteMonitor{
		startTime: time.Now(),
		interval:  time.Duration(interval) * time.Second,
	}
}

func (m *WebsiteMonitor) GetActiveWebsite() (string, string, error) {
	// Mock implementation for Linux
	// In a real implementation, this would use browser-specific methods
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting active window: %v", err)
	}

	title := strings.TrimSpace(string(output))
	// For now, we'll use a mock URL based on the window title
	// In a real implementation, this would get the actual URL from the browser
	url := fmt.Sprintf("https://mock-url.com/%s", strings.ReplaceAll(title, " ", "-"))

	return url, title, nil
}

func (m *WebsiteMonitor) Monitor(ch chan<- map[string]interface{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		url, title, err := m.GetActiveWebsite()
		if err != nil {
			log.Printf("Error getting active website: %v", err)
			continue
		}

		if url != m.lastURL || title != m.lastTitle {
			// If there was a previous website, send its visit data
			if m.lastURL != "" {
				duration := int(time.Since(m.startTime).Seconds())
				ch <- map[string]interface{}{
					"url":      m.lastURL,
					"title":    m.lastTitle,
					"duration": duration,
				}
			}
			// Update the current website info
			m.lastURL = url
			m.lastTitle = title
			m.startTime = time.Now()
		}
	}
}
