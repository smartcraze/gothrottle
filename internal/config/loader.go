package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// Load reads and parses the YAML configuration file
func Load(filePath string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into Config struct
	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	setDefaults(&config)

	return &config, nil
}

// validate checks if the configuration is valid
func validate(config *Config) error {
	if len(config.Routes) == 0 {
		return fmt.Errorf("at least one route must be configured")
	}

	for i, route := range config.Routes {
		if route.Path == "" {
			return fmt.Errorf("route[%d]: path cannot be empty", i)
		}
		if route.Target == "" {
			return fmt.Errorf("route[%d]: target cannot be empty", i)
		}
	}

	if config.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("requests_per_second must be greater than 0")
	}

	if config.RateLimit.Burst <= 0 {
		return fmt.Errorf("burst must be greater than 0")
	}

	return nil
}


func setDefaults(config *Config) {
	if config.Server.Port == 0 {
		config.Server.Port = 8080 // Default server port
	}
}