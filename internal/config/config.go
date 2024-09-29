package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var (
	noRouteDefinedError = errors.New("no route is defined")
)

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

	err = cfg.validateAll()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validateAll() error {
	if len(c.Routes) == 0 {
		return noRouteDefinedError
	}

	for _, r := range c.Routes {
		err := r.validateRoute()
		if err != nil {
			return err
		}

		err = r.Target.validateTarget()
		if err != nil {
			return err
		}
	}

	return nil
}
