package monitoring

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"employeemonitoring/monitor-agent/analysis"
	"employeemonitoring/monitor-agent/monitoring/screenshot"
	"employeemonitoring/monitor-agent/monitoring/windows"
)

type ScreenshotMonitor struct {
	lastCapture time.Time
	interval    time.Duration
	analyzer    *analysis.FakeAnalyzer
	capturer    *screenshot.Capturer
}

func NewScreenshotMonitor(interval time.Duration) *ScreenshotMonitor {
	log.Printf("Initializing screenshot monitor with interval: %v", interval)

	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v", err)
		exePath = "."
	}
	screenshotDir := filepath.Join(filepath.Dir(exePath), "screenshots")
	log.Printf("Screenshot directory: %s", screenshotDir)

	return &ScreenshotMonitor{
		interval: interval,
		analyzer: analysis.NewFakeAnalyzer(),
		capturer: screenshot.NewCapturer(screenshotDir),
	}
}

func (m *ScreenshotMonitor) CaptureScreenshot() (string, error) {
	if runtime.GOOS == "windows" {
		return m.capturer.Capture()
	}
	return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
}

func (m *ScreenshotMonitor) GetActiveWindow() (string, error) {
	if runtime.GOOS == "windows" {
		return windows.GetWindowTitle()
	}
	return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
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
