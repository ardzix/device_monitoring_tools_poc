package monitoring

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"employeemonitoring/monitor-agent/monitoring/windows"
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
	if runtime.GOOS != "windows" {
		return "", "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	title, err := windows.GetWindowTitle()
	if err != nil {
		return "", "", err
	}

	// Check if the window title contains browser names
	lowerTitle := strings.ToLower(title)
	isBrowser := strings.Contains(lowerTitle, "chrome") ||
		strings.Contains(lowerTitle, "firefox") ||
		strings.Contains(lowerTitle, "edge") ||
		strings.Contains(lowerTitle, "opera") ||
		strings.Contains(lowerTitle, "safari")

	if !isBrowser {
		return "", "", nil
	}

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

		// If no browser window is active or couldn't get URL
		if url == "" {
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
