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
	baseURL    string
	httpClient *http.Client
	data       *MonitoringData
	mutex      sync.Mutex
}

func NewHTTPClient(baseURL string) *HTTPClient {
	transport := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,
	}

	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		data: &MonitoringData{
			AppUsage:      make([]map[string]interface{}, 0),
			WebsiteVisits: make([]map[string]interface{}, 0),
			FileAccess:    make([]map[string]interface{}, 0),
			USBDevices:    make([]map[string]interface{}, 0),
			ActivityLogs:  make([]map[string]interface{}, 0),
		},
	}
}

func (c *HTTPClient) AddData(dataType string, data map[string]interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	switch dataType {
	case "app_usage":
		c.data.AppUsage = append(c.data.AppUsage, data)
	case "website_visits":
		c.data.WebsiteVisits = append(c.data.WebsiteVisits, data)
	case "file_access":
		c.data.FileAccess = append(c.data.FileAccess, data)
	case "usb_devices":
		c.data.USBDevices = append(c.data.USBDevices, data)
	case "activity_logs":
		c.data.ActivityLogs = append(c.data.ActivityLogs, data)
	}
}

func (c *HTTPClient) SendBulkData() error {
	c.mutex.Lock()
	data := c.data
	// Reset the data store
	c.data = &MonitoringData{
		AppUsage:      make([]map[string]interface{}, 0),
		WebsiteVisits: make([]map[string]interface{}, 0),
		FileAccess:    make([]map[string]interface{}, 0),
		USBDevices:    make([]map[string]interface{}, 0),
		ActivityLogs:  make([]map[string]interface{}, 0),
	}
	c.mutex.Unlock()

	// Create a multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add screenshot files first
	var screenshotPath string
	for _, activityLog := range data.ActivityLogs {
		if path, ok := activityLog["screenshot"].(string); ok && path != "" {
			screenshotPath = path
			break // Only send one screenshot at a time
		}
	}

	// If we have a screenshot, add it to the form
	if screenshotPath != "" {
		file, err := os.Open(screenshotPath)
		if err != nil {
			log.Printf("Error opening screenshot file: %v", err)
		} else {
			defer file.Close()

			// Create form file with the same name as in the JSON data
			part, err := writer.CreateFormFile("screenshot", filepath.Base(screenshotPath))
			if err != nil {
				log.Printf("Error creating form file: %v", err)
			} else {
				if _, err := io.Copy(part, file); err != nil {
					log.Printf("Error copying file: %v", err)
				}
			}
		}
	}

	// Add JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}
	jsonPart, err := writer.CreateFormField("data")
	if err != nil {
		return fmt.Errorf("error creating form field: %v", err)
	}
	if _, err := jsonPart.Write(jsonData); err != nil {
		return fmt.Errorf("error writing JSON data: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/bulk/", &requestBody)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Connection", "keep-alive")
	req.Close = false

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}

	// Read the entire response body and close it
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *HTTPClient) SendActivityLog(data map[string]interface{}) error {
	c.AddData("activity_logs", data)
	return nil
}
