package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"employeemonitoring/monitor-agent/monitoring"
	"employeemonitoring/monitor-agent/transport"
)

//go:embed env.txt
var envFile string

func init() {
	// Read the embedded .env file and set environment variables
	scanner := bufio.NewScanner(strings.NewReader(envFile))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set environment variable if it's not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
			log.Printf("Set environment variable: %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading embedded .env file: %v", err)
	}
}

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

// getMACAddress returns the MAC address of the first non-loopback interface
func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// Skip loopback and interfaces without MAC
		if iface.Flags&net.FlagLoopback != 0 || iface.HardwareAddr == nil {
			continue
		}

		// Skip interfaces that are down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		mac := iface.HardwareAddr.String()
		if mac != "" {
			// Replace colons with dashes for better compatibility
			mac = strings.ReplaceAll(mac, ":", "-")
			return mac, nil
		}
	}

	return "", fmt.Errorf("no suitable network interface found")
}

func main() {
	// Parse command line flags
	hostFlag := flag.String("host", "http://localhost:8000", "Host server URL")
	flag.Parse()

	// Get host URL from environment variable or fall back to command line flag
	host := os.Getenv("HOST_URL")
	if host == "" {
		host = *hostFlag
		log.Printf("Warning: HOST_URL not found in environment, using default: %s", host)
	} else {
		log.Printf("Using HOST_URL from environment: %s", host)
	}

	// Get device MAC address for identifier
	deviceIdentifier, err := getMACAddress()
	if err != nil {
		log.Printf("Warning: Failed to get MAC address: %v", err)
		// Fallback to a timestamp-based identifier
		deviceIdentifier = fmt.Sprintf("device_%d", time.Now().Unix())
	}
	log.Printf("Using device identifier: %s", deviceIdentifier)

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

	// Initialize HTTP client with device identifier
	client := transport.NewHTTPClient(host, os.Getenv("API_KEY"), deviceIdentifier)

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

	// Start data collection goroutine
	go client.CollectAndSendData(appCh, websiteCh, fileCh, usbCh, screenshotCh, sendInterval)

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("Monitor agent is now running. Press Ctrl+C to stop.")

	<-sigChan

	log.Println("Received signal interrupt, sending final data and shutting down...")
	client.SendFinalData()
}
