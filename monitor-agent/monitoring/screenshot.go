package monitoring

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"employeemonitoring/monitor-agent/analysis"
)

type ScreenshotMonitor struct {
	lastCapture time.Time
	interval    time.Duration
	analyzer    *analysis.FakeAnalyzer
}

func NewScreenshotMonitor(interval time.Duration) *ScreenshotMonitor {
	log.Printf("Initializing screenshot monitor with interval: %v", interval)
	return &ScreenshotMonitor{
		interval: interval,
		analyzer: analysis.NewFakeAnalyzer(),
	}
}

func (m *ScreenshotMonitor) CaptureScreenshot() (string, error) {
	// Create screenshots directory if it doesn't exist
	dir := "screenshots"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create screenshots directory: %v", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(dir, fmt.Sprintf("screenshot_%s.png", timestamp))

	log.Printf("Attempting to capture screenshot to: %s", filename)

	// Use scrot to capture screenshot (Linux)
	cmd := exec.Command("scrot", filename)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to capture screenshot: %v", err)
	}

	log.Printf("Successfully captured screenshot: %s", filename)
	return filename, nil
}

func (m *ScreenshotMonitor) GetActiveWindow() (string, error) {
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting active window: %v", err)
	}
	return string(output), nil
}

func (m *ScreenshotMonitor) Monitor(dataChan chan<- map[string]interface{}) {
	log.Printf("Starting screenshot monitor with interval: %v", m.interval)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("Screenshot monitor tick")
		filename, err := m.CaptureScreenshot()
		if err != nil {
			log.Printf("Error capturing screenshot: %v", err)
			continue
		}

		// Get active window title
		windowTitle, err := m.GetActiveWindow()
		if err != nil {
			log.Printf("Error getting active window: %v", err)
			windowTitle = "Unknown"
		}

		// Analyze screenshot
		analysisResult := m.analyzer.AnalyzeScreenshot(filename)

		// Send screenshot data to channel
		data := map[string]interface{}{
			"timestamp":    time.Now().Format(time.RFC3339),
			"window_title": windowTitle,
			"clipboard":    "", // We'll leave this empty for now
			"screenshot":   filename,
			"analysis":     analysisResult.Description,
			"is_flagged":   analysisResult.IsFlagged,
			"keywords":     analysisResult.Keywords,
			"confidence":   analysisResult.Confidence,
		}
		log.Printf("Sending screenshot data to channel: %+v", data)
		dataChan <- data
	}
}
