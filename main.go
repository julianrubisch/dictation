package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration file structure
// In Go, structs define data structures with named fields
// The `yaml:"words"` tag tells the YAML parser which field to map to
type Config struct {
	Words []string `yaml:"words"`
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

	// Return a pointer to the config (&config) and nil error
	return &config, nil
}

// shuffleWords shuffles a slice of words using Fisher-Yates algorithm
// This function takes a slice (Go's dynamic array type) and returns
// a new shuffled slice without modifying the original.
func shuffleWords(words []string) []string {
	// make() creates a slice with the specified length
	// We copy the original to avoid mutating it
	shuffled := make([]string, len(words))
	copy(shuffled, words)

	// Create a new random number generator seeded with current time
	// This ensures different shuffles each run
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Fisher-Yates shuffle: iterate backwards, swap each element
	// with a random element from the unshuffled portion
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)  // Random index from 0 to i
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]  // Swap
	}

	return shuffled
}

// speakWord uses macOS's native 'say' command to speak a word in German
// This demonstrates executing external commands in Go using os/exec
func speakWord(word string) error {
	// exec.Command creates a command struct (doesn't run it yet)
	// -v "Anna" specifies the German voice
	// -r 180 sets speech rate (words per minute)
	cmd := exec.Command("say", "-v", "Anna", "-r", "180", word)
	
	// cmd.Run() executes the command and waits for completion
	if err := cmd.Run(); err != nil {
		// Fallback to default voice if German voice unavailable
		cmd := exec.Command("say", "-r", "180", word)
		return cmd.Run()
	}
	return nil
}
