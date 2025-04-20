package config

import (
	"path/filepath"
)

type Config struct {
	LogDir        string
	ScreenshotDir string
	Interval      int // in seconds
}

func NewConfig() *Config {
	return &Config{
		LogDir:        "logs",
		ScreenshotDir: "screenshots",
		Interval:      30,
	}
}

func (c *Config) GetLogPath(filename string) string {
	return filepath.Join(c.LogDir, filename)
}

func (c *Config) GetScreenshotPath(filename string) string {
	return filepath.Join(c.ScreenshotDir, filename)
}
