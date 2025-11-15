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
	"github.com/charmbracelet/lipgloss"
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

// Define color styles for the diff output
// These are package-level variables that can be reused
var (
	// Error style for incorrect input
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).  // Red
			Bold(true)
	
	// Success style for correct parts
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))  // Green
	
	// Label style for section headers
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).  // Cyan
			Bold(true)
	
	// Diff marker style for difference indicators
	diffMarkerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).  // Yellow
			Bold(true)
	
	// Correct character style (when characters match)
	correctCharStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))  // Green
	
	// Wrong character style (when characters differ)
	wrongCharStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).  // Red
			Bold(true)
)

// formatWordDiff creates a visual comparison between user input and correct word
// It shows both words side by side with color-coded indicators for matches and differences
// This helps students see exactly where they made mistakes
func formatWordDiff(userInput, correctWord string) string {
	// Convert to rune slices to handle Unicode characters properly
	// Runes are Go's representation of Unicode code points
	userRunes := []rune(userInput)
	correctRunes := []rune(correctWord)
	
	// Find the maximum length for alignment
	maxLen := len(userRunes)
	if len(correctRunes) > maxLen {
		maxLen = len(correctRunes)
	}
	
	// Build the comparison strings with color coding
	// We'll show matching characters in green, differences in red
	var userLine strings.Builder
	var correctLine strings.Builder
	var diffLine strings.Builder
	
	// Iterate through each position up to the maximum length
	for i := 0; i < maxLen; i++ {
		var userChar, correctChar rune
		userExists := i < len(userRunes)
		correctExists := i < len(correctRunes)
		
		if userExists {
			userChar = userRunes[i]
		} else {
			userChar = ' '  // Padding for missing characters
		}
		
		if correctExists {
			correctChar = correctRunes[i]
		} else {
			correctChar = ' '  // Padding for missing characters
		}
		
		// Compare characters (case-sensitive)
		// This allows the diff to show case differences (e.g., "haus" vs "Haus")
		// Note: The main validation is still case-insensitive, but the diff
		// visualization highlights case differences to help students learn
		isMatch := userChar == correctChar && userExists && correctExists
		
		// Add characters to lines with appropriate styling
		if isMatch {
			// Both characters match - show in green
			userLine.WriteString(correctCharStyle.Render(string(userChar)))
			correctLine.WriteString(correctCharStyle.Render(string(correctChar)))
		} else {
			// Characters differ - show in red
			userLine.WriteString(wrongCharStyle.Render(string(userChar)))
			correctLine.WriteString(wrongCharStyle.Render(string(correctChar)))
		}
		
		// Mark differences with colored indicators
		if !isMatch {
			diffLine.WriteString(diffMarkerStyle.Render("^"))  // Mark difference in yellow
		} else {
			diffLine.WriteString(" ")  // Match - no marker
		}
	}
	
	// Format the output with colored labels
	// Use fixed-width labels (14 chars) to ensure proper alignment
	// This accounts for ANSI escape codes in colored text
	labelWidth := 14
	yourInputLabel := labelStyle.Width(labelWidth).Render("Your input:")
	correctLabel := labelStyle.Width(labelWidth).Render("Correct:")
	diffLabel := labelStyle.Width(labelWidth).Render("Differences:")
	
	return fmt.Sprintf(
		"%s  %s\n"+
			"%s  %s\n"+
			"%s  %s",
		yourInputLabel,
		userLine.String(),
		correctLabel,
		correctLine.String(),
		diffLabel,
		diffLine.String(),
	)
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
	originalWordCount := len(words)  // Store original count for progress display

	// Print welcome message
	fmt.Println("🎯 German Dictation Practice")
	fmt.Println("============================")
	fmt.Printf("You will practice %d word(s).\n\n", originalWordCount)
	fmt.Println("Listen carefully to each word and type it correctly.")
	fmt.Println("Press Enter after typing each word.")

	// Track progress
	correctCount := 0
	totalAttempts := 0

	// Practice words using a queue approach
	// When a word is incorrect, it's added back to the end of the queue
	// This gives students a break and lets them practice other words first
	for i := 0; i < len(words); i++ {
		word := words[i]
		totalAttempts++

		// Speak the word using TTS
		// Show progress: how many words completed correctly out of original total
		fmt.Printf("\n🔊 Word %d: %d of %d completed correctly\n", i+1, correctCount, originalWordCount)
		if err := speakWord(word); err != nil {
			// log.Printf doesn't exit, just logs warning
			log.Printf("Warning: Failed to speak word: %v", err)
		}

		// Small delay to let TTS finish speaking
		time.Sleep(500 * time.Millisecond)

		// Prompt user for input
		// Note: attempt number is always 1 since we don't retry immediately
		userInput, err := promptWord(word, 1)
		if err != nil {
			log.Fatalf("Error getting input: %v", err)
		}

		// Check if correct (case-sensitive comparison)
		// German requires proper capitalization (nouns are capitalized)
		// Direct string comparison ensures exact match including case
		if userInput == word {
			fmt.Println("✅ Correct! Well done!")
			correctCount++
		} else {
			// Show colorful feedback with visual diff to help learning
			fmt.Println(errorStyle.Render("❌ Incorrect spelling!"))
			fmt.Println(formatWordDiff(userInput, word))
			fmt.Print("\n")  // Empty line for readability
			
			// Add the word back to the end of the queue
			// This allows the student to practice other words first
			// and come back to this one later
			words = append(words, word)
		}
	}

	// Print summary statistics
	fmt.Println("\n" + strings.Repeat("=", 30))
	fmt.Println("🎉 Practice Complete!")
	fmt.Printf("Words practiced: %d\n", correctCount)
	fmt.Printf("Total attempts: %d\n", totalAttempts)
	// Calculate accuracy percentage
	fmt.Printf("Accuracy: %.1f%%\n", float64(correctCount)/float64(totalAttempts)*100)
	fmt.Println(strings.Repeat("=", 30))
}
