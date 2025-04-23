package monitoring

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"employeemonitoring/monitor-agent/monitoring/windows"
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
		isActive:  true, // Initialize as active
	}
}

func (m *AppMonitor) GetActiveApp() (string, string, bool, error) {
	if runtime.GOOS != "windows" {
		return "", "", false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	title, err := windows.GetWindowTitle()
	if err != nil {
		return "", "", false, err
	}

	// On Windows, we'll consider the window active if we can get its title
	isActive := title != ""

	// Use the window title as both app name and title for now
	return title, title, isActive, nil
}

func (m *AppMonitor) Monitor(ch chan<- map[string]interface{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		appName, windowTitle, isActive, err := m.GetActiveApp()
		if err != nil {
			log.Printf("Error getting active app: %v", err)
			continue
		}

		// Send data if the app, title, or active state has changed
		if appName != m.lastApp || windowTitle != m.lastTitle || isActive != m.isActive {
			// If there was a previous app running, send its usage data
			if m.lastApp != "" {
				duration := int(time.Since(m.startTime).Seconds())
				log.Printf("Sending app usage data: app=%s, title=%s, duration=%d, active=%v",
					m.lastApp, m.lastTitle, duration, m.isActive)
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
			m.isActive = isActive
			m.startTime = time.Now()
		}
	}
}
