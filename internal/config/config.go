package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Local     string    `json:"local"`
	Data      string    `json:"data"`
	Prefix    []string  `json:"prefix"`
	Transport Transport `json:"transport"`
	Default   Default   `json:"default"`
}

type Transport struct {
	HTTP  string `json:"http"`
	Sse   string `json:"sse"`
	Stdin string `json:"stdin"`
}

type Default struct {
	Transport string `json:"transport"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &cfg, nil
}

// ExpandPath handles tilde expansion and absolute path resolution for the data directory
func (c *Config) GetDataDir() string {
	path := c.Data
	if path == "" {
		return "./data" // Default value
	}

	// Note: In a real app we might use a library for tilde expansion,
	// but here we'll just return the value or resolve relative to project root.
	// Since config.json has "./data", it works fine with filepath.Abs.
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}
