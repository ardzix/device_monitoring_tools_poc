package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ActivityLog struct {
	Timestamp   time.Time `json:"timestamp"`
	WindowTitle string    `json:"window_title"`
	Clipboard   string    `json:"clipboard,omitempty"`
	Screenshot  string    `json:"screenshot,omitempty"`
	Analysis    string    `json:"analysis,omitempty"`
	IsFlagged   bool      `json:"is_flagged"`
	Keywords    []string  `json:"keywords,omitempty"`
	Confidence  float64   `json:"confidence,omitempty"`
}

type Logger struct {
	logDir string
}

func NewLogger(logDir string) *Logger {
	return &Logger{
		logDir: logDir,
	}
}

func (l *Logger) SaveActivity(log ActivityLog) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("activity_%s.json", log.Timestamp.Format("2006-01-02"))
	filepath := fmt.Sprintf("%s/%s", l.logDir, filename)

	// Read existing logs if file exists
	var logs []ActivityLog
	if _, err := os.Stat(filepath); err == nil {
		file, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("failed to read existing log file: %v", err)
		}
		if err := json.Unmarshal(file, &logs); err != nil {
			return fmt.Errorf("failed to unmarshal existing logs: %v", err)
		}
	}

	// Append new log
	logs = append(logs, log)

	// Write updated logs
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %v", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write log file: %v", err)
	}

	return nil
}
