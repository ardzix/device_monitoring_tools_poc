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
		isActive:  true, // Initialize as active
	}
}

func (m *AppMonitor) GetActiveApp() (string, string, bool, error) {
	// Get active window
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	output, err := cmd.Output()
	if err != nil {
		return "", "", false, fmt.Errorf("error getting active window: %v", err)
	}

	title := strings.TrimSpace(string(output))

	// Check if window is focused
	cmd = exec.Command("xdotool", "getwindowfocus")
	focusOutput, err := cmd.Output()
	if err != nil {
		return title, title, false, nil
	}

	// Get window state
	cmd = exec.Command("xprop", "-id", strings.TrimSpace(string(focusOutput)), "_NET_WM_STATE")
	stateOutput, err := cmd.Output()
	if err != nil {
		return title, title, true, nil // Assume active if we can't determine state
	}

	// Check if window is not minimized or hidden
	isActive := !strings.Contains(string(stateOutput), "_NET_WM_STATE_HIDDEN") &&
		!strings.Contains(string(stateOutput), "_NET_WM_STATE_MINIMIZED")

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
