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

type USBMonitor struct {
	lastDevices map[string]bool
	interval    time.Duration
}

func NewUSBMonitor() *USBMonitor {
	interval := 5 // default interval
	if envInterval := os.Getenv("USB_MONITOR_INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			interval = i
		}
	}
	return &USBMonitor{
		lastDevices: make(map[string]bool),
		interval:    time.Duration(interval) * time.Second,
	}
}

func (m *USBMonitor) GetUSBDevices() ([]map[string]string, error) {
	cmd := exec.Command("lsusb")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running lsusb: %v", err)
	}

	var devices []map[string]string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Parse lsusb output format: Bus 001 Device 002: ID 8087:0024 Intel Corp. Integrated Rate Matching Hub
		parts := strings.Split(line, " ")
		if len(parts) < 6 {
			continue
		}
		device := map[string]string{
			"bus":        parts[1],
			"device":     parts[3],
			"vendor_id":  strings.Split(parts[5], ":")[0],
			"product_id": strings.Split(parts[5], ":")[1],
			"name":       strings.Join(parts[6:], " "),
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (m *USBMonitor) Monitor(ch chan<- map[string]interface{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		devices, err := m.GetUSBDevices()
		if err != nil {
			log.Printf("Error getting USB devices: %v", err)
			continue
		}

		currentDevices := make(map[string]bool)

		for _, device := range devices {
			deviceID := fmt.Sprintf("%s:%s", device["vendor_id"], device["product_id"])
			currentDevices[deviceID] = true

			// Check for new devices
			if !m.lastDevices[deviceID] {
				// Convert map[string]string to map[string]interface{}
				deviceData := make(map[string]interface{})
				for k, v := range device {
					deviceData[k] = v
				}
				deviceData["action"] = "connect"
				ch <- deviceData
			}
		}

		// Check for disconnected devices
		for deviceID := range m.lastDevices {
			if !currentDevices[deviceID] {
				parts := strings.Split(deviceID, ":")
				ch <- map[string]interface{}{
					"vendor_id":     parts[0],
					"product_id":    parts[1],
					"device_name":   "Unknown Device",
					"serial_number": "",
					"action":        "disconnect",
				}
			}
		}

		m.lastDevices = currentDevices
	}
}
