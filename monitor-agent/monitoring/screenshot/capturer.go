//go:build windows
// +build windows

package screenshot

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")

	getDesktopWindow       = user32.NewProc("GetDesktopWindow")
	getWindowDC            = user32.NewProc("GetWindowDC")
	createCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	selectObject           = gdi32.NewProc("SelectObject")
	bitBlt                 = gdi32.NewProc("BitBlt")
	deleteDC               = gdi32.NewProc("DeleteDC")
	releaseRC              = user32.NewProc("ReleaseDC")
	deleteObject           = gdi32.NewProc("DeleteObject")
	getSystemMetrics       = user32.NewProc("GetSystemMetrics")
)

const (
	SRCCOPY     = 0x00CC0020
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
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
	// Create directory if it doesn't exist
	if err := os.MkdirAll(c.screenshotDir, 0755); err != nil {
		log.Printf("Failed to create screenshots directory %s: %v", c.screenshotDir, err)
		return "", fmt.Errorf("failed to create screenshots directory: %v", err)
	}
	log.Printf("Using screenshot directory: %s", c.screenshotDir)

	// Generate filename with timestamp
	filename := fmt.Sprintf("screenshot_%s.png", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(c.screenshotDir, filename)
	log.Printf("Saving screenshot to: %s", filepath)

	// Get screen dimensions
	width, _, _ := getSystemMetrics.Call(uintptr(SM_CXSCREEN))
	height, _, _ := getSystemMetrics.Call(uintptr(SM_CYSCREEN))
	log.Printf("Screen dimensions: %dx%d", width, height)

	// Get the desktop window
	hwnd, _, _ := getDesktopWindow.Call()
	if hwnd == 0 {
		return "", fmt.Errorf("failed to get desktop window")
	}

	// Get the window DC
	hwndDC, _, _ := getWindowDC.Call(hwnd)
	if hwndDC == 0 {
		return "", fmt.Errorf("failed to get window DC")
	}
	defer releaseRC.Call(hwnd, hwndDC)

	// Create a compatible DC
	memDC, _, _ := createCompatibleDC.Call(hwndDC)
	if memDC == 0 {
		return "", fmt.Errorf("failed to create compatible DC")
	}
	defer deleteDC.Call(memDC)

	// Create a compatible bitmap
	hBitmap, _, _ := createCompatibleBitmap.Call(hwndDC, width, height)
	if hBitmap == 0 {
		return "", fmt.Errorf("failed to create compatible bitmap")
	}
	defer deleteObject.Call(hBitmap)

	// Select the bitmap into the compatible DC
	oldBitmap, _, _ := selectObject.Call(memDC, hBitmap)
	if oldBitmap == 0 {
		return "", fmt.Errorf("failed to select object")
	}
	defer selectObject.Call(memDC, oldBitmap)
	defer deleteObject.Call(oldBitmap)

	// Copy the window content
	ret, _, _ := bitBlt.Call(
		memDC, 0, 0, width, height,
		hwndDC, 0, 0,
		uintptr(SRCCOPY))
	if ret == 0 {
		return "", fmt.Errorf("failed to copy screen content")
	}

	// Create image from bitmap
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	// Get the bitmap bits
	if err := getHBitmapBits(hBitmap, width, height, img.Pix); err != nil {
		return "", fmt.Errorf("failed to get bitmap bits: %v", err)
	}

	// Save to file
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

func getHBitmapBits(hBitmap uintptr, width, height uintptr, pBits []uint8) error {
	// Get the bitmap bits
	ret, _, _ := gdi32.NewProc("GetBitmapBits").Call(
		hBitmap,
		uintptr(len(pBits)),
		uintptr(unsafe.Pointer(&pBits[0])))

	if ret == 0 {
		return fmt.Errorf("GetBitmapBits failed")
	}

	return nil
}
