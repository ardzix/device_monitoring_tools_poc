package activity

// GetActiveWindowTitle returns a mock window title on Linux
func GetActiveWindowTitle() string {
	// This is a mock implementation for Linux
	// In the Windows version, this will use the actual Windows API
	return "Mock Window Title - Linux"
}
