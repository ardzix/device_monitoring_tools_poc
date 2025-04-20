package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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

func main() {
	// Parse command line flags
	host := flag.String("host", "http://localhost:8000", "Host server URL")
	interval := flag.Duration("interval", 30*time.Second, "Monitoring interval")
	flag.Parse()

	log.Printf("Starting monitor agent with host=%s, interval=%v", *host, *interval)

	// Load environment variables from root directory
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file from %s: %v", envPath, err)
	}

	// Initialize HTTP client
	client := transport.NewHTTPClient(os.Getenv("HOST_URL"))

	// Create channels for different types of monitoring data
	appCh := make(chan map[string]interface{}, 100)
	websiteCh := make(chan map[string]interface{}, 100)
	fileCh := make(chan map[string]interface{}, 100)
	usbCh := make(chan map[string]interface{}, 100)
	screenshotCh := make(chan map[string]interface{}, 100)

	// Start monitoring goroutines
	go monitoring.NewAppMonitor().Monitor(appCh)
	go monitoring.NewWebsiteMonitor().Monitor(websiteCh)
	go monitoring.NewFileMonitor([]string{"/home/ardz/Documents"}).Monitor(fileCh)
	go monitoring.NewUSBMonitor().Monitor(usbCh)
	go monitoring.NewScreenshotMonitor(30 * time.Second).Monitor(screenshotCh)

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Monitor agent is now running. Press Ctrl+C to stop.")

	// Create a ticker for bulk sending
	sendTicker := time.NewTicker(*interval)
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
