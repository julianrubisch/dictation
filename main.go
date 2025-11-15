package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
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

// promptWord prompts the user to type a word and validates it
// This uses the Huh library for beautiful terminal prompts
func promptWord(word string, attempt int) (string, error) {
	var input string  // Variable to store user input

	// Build prompt title based on attempt number
	title := fmt.Sprintf("Word %d: Type what you heard", attempt)
	if attempt > 1 {
		title = fmt.Sprintf("Word %d: Try again (attempt %d)", attempt, attempt)
	}

	// Huh provides a fluent API for building forms
	// NewInput() creates a text input field
	// Value(&input) binds the input to our variable (pointer needed)
	// Validate() adds custom validation logic
	err := huh.NewInput().
		Title(title).
		Placeholder("Type the word here...").
		Value(&input).  // & gets address of input variable
		Validate(func(s string) error {
			// Anonymous function for validation
			// Returns error if validation fails, nil if OK
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("please enter a word")
			}
			return nil
		}).
		Run()  // Run() blocks until user submits

	if err != nil {
		return "", err
	}

	// Trim whitespace and return
	return strings.TrimSpace(input), nil
}

func main() {
	// main() is the entry point of every Go program
	// os.Args contains command-line arguments
	// os.Args[0] is the program name, os.Args[1:] are arguments
	
	// Default config file path
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]  // Use first argument as config file
	}

	// Load configuration - handle errors with log.Fatalf
	// Fatalf prints error and exits program (os.Exit(1))
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Shuffle words for variety in practice sessions
	words := shuffleWords(config.Words)

	// Print welcome message
	fmt.Println("🎯 German Dictation Practice")
	fmt.Println("============================")
	fmt.Printf("You will practice %d word(s).\n\n", len(words))
	fmt.Println("Listen carefully to each word and type it correctly.")
	fmt.Println("Press Enter after typing each word.\n")

	// Track progress
	correctCount := 0
	totalAttempts := 0

	// Practice each word - range loop iterates over slice
	// i is index, word is value
	for i, word := range words {
		attempt := 1
		correct := false

		// Keep trying until user gets it right
		for !correct {
			totalAttempts++

			// Speak the word using TTS
			fmt.Printf("\n🔊 Speaking word %d of %d...\n", i+1, len(words))
			if err := speakWord(word); err != nil {
				// log.Printf doesn't exit, just logs warning
				log.Printf("Warning: Failed to speak word: %v", err)
			}

			// Small delay to let TTS finish speaking
			time.Sleep(500 * time.Millisecond)

			// Prompt user for input
			userInput, err := promptWord(word, attempt)
			if err != nil {
				log.Fatalf("Error getting input: %v", err)
			}

			// Check if correct (case-insensitive comparison)
			// strings.EqualFold compares ignoring case
			if strings.EqualFold(userInput, word) {
				fmt.Println("✅ Correct! Well done!")
				correct = true
				correctCount++
			} else {
				// Show feedback and increment attempt counter
				fmt.Printf("❌ Incorrect. You typed: '%s', but the correct word is: '%s'\n", userInput, word)
				attempt++
			}
		}
	}

	// Print summary statistics
	fmt.Println("\n" + strings.Repeat("=", 30))
	fmt.Println("🎉 Practice Complete!")
	fmt.Printf("Words practiced: %d\n", len(words))
	fmt.Printf("Total attempts: %d\n", totalAttempts)
	// Calculate accuracy percentage
	fmt.Printf("Accuracy: %.1f%%\n", float64(len(words))/float64(totalAttempts)*100)
	fmt.Println(strings.Repeat("=", 30))
}
