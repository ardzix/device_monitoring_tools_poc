package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"employeemonitoring/monitoring"
	"employeemonitoring/transport"

	"github.com/joho/godotenv"
)

// getEnvBool returns true if the environment variable is set to "true" (case insensitive)
func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return strings.ToLower(val) == "true"
}

// getEnvDuration returns the duration from environment variable in seconds
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("Environment variable %s not set, using default value: %v", key, defaultValue)
		return defaultValue
	}

	seconds, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		log.Printf("Error parsing %s value '%s': %v, using default value: %v", key, val, err, defaultValue)
		return defaultValue
	}

	if seconds <= 0 {
		log.Printf("Invalid %s value '%s': must be positive, using default value: %v", key, val, defaultValue)
		return defaultValue
	}

	duration := time.Duration(seconds) * time.Second
	log.Printf("Environment variable %s set to %d seconds (%v)", key, seconds, duration)
	return duration
}

func main() {
	// Parse command line flags
	hostFlag := flag.String("host", "http://localhost:8000", "Host server URL")
	flag.Parse()

	// Load environment variables from the correct path
	envPath := filepath.Join("..", "..", ".env") // Go up two levels to project root
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Error loading .env file from %s: %v", envPath, err)
		// Try loading from current directory as fallback
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file from current directory: %v", err)
		}
	}

	// Get host URL from environment variable or fall back to command line flag
	host := os.Getenv("HOST_URL")
	if host == "" {
		host = *hostFlag
	}

	// Get intervals from environment variables with proper defaults
	appInterval := getEnvDuration("APP_MONITOR_INTERVAL", 1*time.Second)
	websiteInterval := getEnvDuration("WEBSITE_MONITOR_INTERVAL", 1*time.Second)
	fileInterval := getEnvDuration("FILE_MONITOR_INTERVAL", 5*time.Second)
	usbInterval := getEnvDuration("USB_MONITOR_INTERVAL", 5*time.Second)
	screenshotInterval := getEnvDuration("SCREENSHOT_INTERVAL", 5*time.Second)
	sendInterval := getEnvDuration("SEND_DATA_INTERVAL", 6*time.Second)

	// Log the actual values being used
	log.Printf("Environment variables loaded:")
	log.Printf("  HOST_URL: %s", host)
	log.Printf("  APP_MONITOR_INTERVAL: %v", appInterval)
	log.Printf("  WEBSITE_MONITOR_INTERVAL: %v", websiteInterval)
	log.Printf("  FILE_MONITOR_INTERVAL: %v", fileInterval)
	log.Printf("  USB_MONITOR_INTERVAL: %v", usbInterval)
	log.Printf("  SCREENSHOT_INTERVAL: %v", screenshotInterval)
	log.Printf("  SEND_DATA_INTERVAL: %v", sendInterval)

	// Initialize HTTP client
	client := transport.NewHTTPClient(host)

	// Create channels for different types of monitoring data
	appCh := make(chan map[string]interface{}, 100)
	websiteCh := make(chan map[string]interface{}, 100)
	fileCh := make(chan map[string]interface{}, 100)
	usbCh := make(chan map[string]interface{}, 100)
	screenshotCh := make(chan map[string]interface{}, 100)

	// Start monitoring goroutines with configured intervals
	go monitoring.NewAppMonitor().Monitor(appCh)
	go monitoring.NewWebsiteMonitor().Monitor(websiteCh)
	go monitoring.NewFileMonitor([]string{"/home/ardz/Documents"}).Monitor(fileCh)
	go monitoring.NewUSBMonitor().Monitor(usbCh)

	// Create and start screenshot monitor with the configured interval
	screenshotMonitor := monitoring.NewScreenshotMonitor(screenshotInterval)
	go screenshotMonitor.Monitor(screenshotCh)

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Monitor agent is now running. Press Ctrl+C to stop.")

	// Create a ticker for bulk sending using the configured interval
	sendTicker := time.NewTicker(sendInterval)
	defer sendTicker.Stop()

	// Process incoming data
	for {
		select {
		case data := <-appCh:
			log.Printf("Received app usage data: %v", data)
			client.AddData("app_usage", data)
		case data := <-websiteCh:
			log.Printf("Received website visit data: %v", data)
			client.AddData("website_visits", data)
		case data := <-fileCh:
			log.Printf("Received file access data: %v", data)
			client.AddData("file_access", data)
		case data := <-usbCh:
			log.Printf("Received USB device data: %v", data)
			client.AddData("usb_devices", data)
		case data := <-screenshotCh:
			log.Printf("Received screenshot data: %v", data)
			client.AddData("activity_logs", data)
		case <-sendTicker.C:
			log.Printf("Sending bulk data...")
			// Send all collected data in bulk
			if err := client.SendBulkData(); err != nil {
				log.Printf("Error sending bulk data: %v", err)
			} else {
				log.Printf("Bulk data sent successfully")
			}
		case sig := <-sigChan:
			log.Printf("Received signal %v, sending final data and shutting down...", sig)
			// Send any remaining data before shutting down
			if err := client.SendBulkData(); err != nil {
				log.Printf("Error sending final bulk data: %v", err)
			}
			return
		}
	}
}
