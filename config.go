package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration file structure
// In Go, structs define data structures with named fields
// The `yaml:"words"` tag tells the YAML parser which field to map to
type Config struct {
	Language string   `yaml:"language"` // Language code (e.g., "en", "de", "fr")
	Words    []string `yaml:"words"`
}

// loadConfig reads and parses the YAML configuration file
// Functions in Go can return multiple values - here we return a pointer
// to Config and an error. This is the idiomatic Go error handling pattern.
func loadConfig(filename string) (*Config, error) {
	// os.ReadFile reads the entire file into a byte slice
	data, err := os.ReadFile(filename)
	if err != nil {
		// fmt.Errorf creates a formatted error with context
		// The %w verb wraps the original error for error unwrapping
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create an empty Config struct
	var config Config
	
	// yaml.Unmarshal parses YAML bytes into our struct
	// The & operator gets the address (pointer) of config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate that we have at least one word
	if len(config.Words) == 0 {
		return nil, fmt.Errorf("no words found in config file")
	}

	// Set default language if not specified
	if config.Language == "" {
		config.Language = "en"  // Default to English
	}

	// Return a pointer to the config (&config) and nil error
	return &config, nil
}
