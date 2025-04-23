//go:build windows
// +build windows

package windows

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
)

// GetWindowTitle returns the title of the currently active window
func GetWindowTitle() (string, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "", fmt.Errorf("no active window")
	}

	// Get window title
	var title [256]uint16
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&title[0])), uintptr(len(title)))
	return syscall.UTF16ToString(title[:]), nil
}
