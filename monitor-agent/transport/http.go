package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type MonitoringData struct {
	AppUsage      []map[string]interface{} `json:"app_usage"`
	WebsiteVisits []map[string]interface{} `json:"website_visits"`
	FileAccess    []map[string]interface{} `json:"file_access"`
	USBDevices    []map[string]interface{} `json:"usb_devices"`
	ActivityLogs  []map[string]interface{} `json:"activity_logs"`
}

type HTTPClient struct {
	hostURL          string
	apiKey           string
	deviceIdentifier string
	data             MonitoringData
	mutex            sync.Mutex
}

func NewHTTPClient(hostURL, apiKey, deviceIdentifier string) *HTTPClient {
	return &HTTPClient{
		hostURL:          hostURL,
		apiKey:           apiKey,
		deviceIdentifier: deviceIdentifier,
		data: MonitoringData{
			AppUsage:      []map[string]interface{}{},
			WebsiteVisits: []map[string]interface{}{},
			FileAccess:    []map[string]interface{}{},
			USBDevices:    []map[string]interface{}{},
			ActivityLogs:  []map[string]interface{}{},
		},
	}
}

func (c *HTTPClient) AddData(dataType string, data map[string]interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Add device identifier and timestamp to all data
	data["device_identifier"] = c.deviceIdentifier
	data["timestamp"] = time.Now().Format(time.RFC3339)

	switch dataType {
	case "app_usage":
		c.data.AppUsage = append(c.data.AppUsage, data)
	case "website_visits":
		c.data.WebsiteVisits = append(c.data.WebsiteVisits, data)
	case "file_access":
		c.data.FileAccess = append(c.data.FileAccess, data)
	case "usb_devices":
		// For USB devices, we only want to keep the most recent state
		c.data.USBDevices = []map[string]interface{}{data}
	case "activity_logs":
		c.data.ActivityLogs = append(c.data.ActivityLogs, data)
	}
}

func (c *HTTPClient) CollectAndSendData(appCh, websiteCh, fileCh, usbCh, screenshotCh <-chan map[string]interface{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case data := <-appCh:
			log.Printf("Received app usage data: %v", data)
			c.AddData("app_usage", data)
		case data := <-websiteCh:
			log.Printf("Received website visit data: %v", data)
			c.AddData("website_visits", data)
		case data := <-fileCh:
			log.Printf("Received file access data: %v", data)
			c.AddData("file_access", data)
		case data := <-usbCh:
			log.Printf("Received USB device data: %v", data)
			c.AddData("usb_devices", data)
		case data := <-screenshotCh:
			log.Printf("Received screenshot data: %v", data)
			c.AddData("activity_logs", data)
		case <-ticker.C:
			log.Printf("Sending bulk data...")
			if err := c.SendBulkData(); err != nil {
				log.Printf("Error sending bulk data: %v", err)
			} else {
				log.Printf("Bulk data sent successfully")
			}
		}
	}
}

func (c *HTTPClient) SendBulkData() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.data.AppUsage) == 0 && len(c.data.WebsiteVisits) == 0 &&
		len(c.data.FileAccess) == 0 && len(c.data.USBDevices) == 0 &&
		len(c.data.ActivityLogs) == 0 {
		return nil
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add JSON data
	jsonData, err := json.Marshal(c.data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}

	// Add device identifier to JSON data
	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return fmt.Errorf("error unmarshaling data: %v", err)
	}
	dataMap["device_identifier"] = c.deviceIdentifier
	jsonData, err = json.Marshal(dataMap)
	if err != nil {
		return fmt.Errorf("error marshaling data with device identifier: %v", err)
	}

	jsonPart, err := writer.CreateFormField("data")
	if err != nil {
		return fmt.Errorf("error creating form field: %v", err)
	}
	jsonPart.Write(jsonData)

	// Add screenshots
	for _, activityLog := range c.data.ActivityLogs {
		if screenshot, ok := activityLog["screenshot"].(string); ok && screenshot != "" {
			// Use the filename as the form field name to match server expectations
			filename := filepath.Base(screenshot)
			log.Printf("Processing screenshot file: %s (full path: %s)", filename, screenshot)

			// Check if file exists
			if _, err := os.Stat(screenshot); os.IsNotExist(err) {
				log.Printf("Screenshot file does not exist: %s", screenshot)
				continue
			}

			file, err := os.Open(screenshot)
			if err != nil {
				log.Printf("Error opening screenshot file %s: %v", screenshot, err)
				continue
			}
			defer file.Close()

			// Get file size for logging
			fileInfo, err := file.Stat()
			if err == nil {
				log.Printf("Screenshot file size: %d bytes", fileInfo.Size())
			}

			// Read file content for verification
			fileContent, err := io.ReadAll(file)
			if err != nil {
				log.Printf("Error reading file content: %v", err)
				continue
			}
			log.Printf("File content length: %d bytes", len(fileContent))

			// Reset file pointer to beginning
			_, err = file.Seek(0, 0)
			if err != nil {
				log.Printf("Error resetting file pointer: %v", err)
				continue
			}

			// Create form file with "screenshot" as the field name
			part, err := writer.CreateFormFile("screenshot", filepath.Base(screenshot))
			if err != nil {
				log.Printf("Error creating form file for %s: %v", screenshot, err)
				continue
			}

			// Copy file content to form part
			if _, err := io.Copy(part, file); err != nil {
				log.Printf("Error copying file %s: %v", screenshot, err)
				continue
			}
			log.Printf("Successfully attached screenshot %s to form", filename)

			// Close the file explicitly
			file.Close()
		}
	}

	writer.Close()

	// Print form data content for debugging
	formData := body.String()
	log.Printf("Form Data Content Length: %d bytes", len(formData))
	log.Printf("Form Data Content Preview (first 100 bytes): %x", formData[:100])

	// Create request with debug logging
	endpoint := c.hostURL + "/api/bulk/"
	log.Printf("Sending request to endpoint: %s", endpoint)
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	contentType := writer.FormDataContentType()
	log.Printf("Content-Type: %s", contentType)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-API-Key", c.apiKey)

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and log response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf("Response Status: %d", resp.StatusCode)
	log.Printf("Response Headers: %v", resp.Header)
	log.Printf("Response Body Length: %d bytes", len(respBody))
	if len(respBody) > 0 {
		log.Printf("Response Body: %s", string(respBody))
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Clear data after successful send
	c.data = MonitoringData{
		AppUsage:      []map[string]interface{}{},
		WebsiteVisits: []map[string]interface{}{},
		FileAccess:    []map[string]interface{}{},
		USBDevices:    []map[string]interface{}{},
		ActivityLogs:  []map[string]interface{}{},
	}

	return nil
}

func (c *HTTPClient) SendFinalData() {
	if err := c.SendBulkData(); err != nil {
		log.Printf("Error sending final data: %v", err)
	}
}

func (c *HTTPClient) SendActivityLog(data map[string]interface{}) error {
	c.AddData("activity_logs", data)
	return nil
}
