package screenshot

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/kbinani/screenshot"
)

type Capturer struct {
	screenshotDir string
}

func NewCapturer(screenshotDir string) *Capturer {
	return &Capturer{
		screenshotDir: screenshotDir,
	}
}

func (c *Capturer) Capture() (string, error) {
	// Get the primary display
	display := 0
	bounds := screenshot.GetDisplayBounds(display)

	// Capture the screen
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", fmt.Errorf("failed to capture screen: %v", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("screenshot_%s.png", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(c.screenshotDir, filename)

	// Save the image
	if err := os.MkdirAll(c.screenshotDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create screenshots directory: %v", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %v", err)
	}

	return filepath, nil
}
