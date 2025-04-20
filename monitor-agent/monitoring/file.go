package monitoring

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileMonitor struct {
	watchedPaths []string
	interval     time.Duration
}

func NewFileMonitor(paths []string) *FileMonitor {
	interval := 5 // default interval
	if envInterval := os.Getenv("FILE_MONITOR_INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			interval = i
		}
	}
	return &FileMonitor{
		watchedPaths: paths,
		interval:     time.Duration(interval) * time.Second,
	}
}

func (m *FileMonitor) isWatchedPath(path string) bool {
	for _, watchedPath := range m.watchedPaths {
		if strings.HasPrefix(path, watchedPath) {
			return true
		}
	}
	return false
}

func (m *FileMonitor) Monitor(ch chan<- map[string]interface{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		for _, watchedPath := range m.watchedPaths {
			err := filepath.Walk(watchedPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}

				// Skip directories
				if info.IsDir() {
					return nil
				}

				// Check if file was accessed within the last interval
				if time.Since(info.ModTime()) < m.interval {
					log.Printf("File access detected: %s", path)
					ch <- map[string]interface{}{
						"file_path":    path,
						"operation":    "read",    // Mock operation type
						"process_name": "unknown", // Mock process name
					}
				}

				return nil
			})

			if err != nil {
				log.Printf("Error walking path %s: %v", watchedPath, err)
			}
		}
	}
}
