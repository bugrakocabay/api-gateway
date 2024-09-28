package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Route defines the structure for route configuration.
type Route struct {
	Path   string  `json:"path"`
	Target *Target `json:"target"`
	Method string  `json:"method"`
	Limit  int64   `json:"limit"`
}

// Target defines the structure for target endpoint configuration.
type Target struct {
	Host   string `json:"host"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

// Config holds the application's route configurations.
type Config struct {
	Routes []*Route `json:"routes"`
}

// LoadConfig reads and parses the configuration file.
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	for i, route := range cfg.Routes {
		if route.Path == "" || route.Target == nil || route.Method == "" {
			return nil, fmt.Errorf("route %d has incomplete configuration", i)
		}
	}

	return &cfg, nil
}
