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
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration file structure
// In Go, structs define data structures with named fields
// The `yaml:"words"` tag tells the YAML parser which field to map to
type Config struct {
	Language string   `yaml:"language"` // Language code (e.g., "en", "de", "fr")
	Words    []string `yaml:"words"`
}

// initI18n initializes the i18n bundle and loads translation files
// This is the idiomatic Go approach using go-i18n library
func initI18n(langCode string) (*i18n.Localizer, error) {
	// Create bundle with English as default language
	// The bundle manages all translation files
	bundle := i18n.NewBundle(language.English)
	
	// Register TOML unmarshal function
	// This allows go-i18n to parse TOML translation files
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	
	// Load translation files
	// These files contain all user-facing strings for each language
	// LoadMessageFile returns (*MessageFile, error)
	_, err := bundle.LoadMessageFile("active.en.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load English translations: %w", err)
	}
	_, err = bundle.LoadMessageFile("active.de.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load German translations: %w", err)
	}
	
	// Create localizer for the requested language
	// The localizer provides methods to get translated strings
	localizer := i18n.NewLocalizer(bundle, langCode)
	
	return localizer, nil
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


// getVoiceForLanguage returns the macOS TTS voice name for a language code
// Maps language codes to appropriate voices for better pronunciation
func getVoiceForLanguage(langCode string) string {
	voices := map[string]string{
		"de": "Anna",    // German voice
		"en": "Alex",    // English voice (US)
		"fr": "Thomas",  // French voice (for future use)
	}

	if voice, ok := voices[langCode]; ok {
		return voice
	}
	// Fallback to default system voice
	return ""
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

// speakWord uses macOS's native 'say' command to speak a word
// Uses the appropriate voice for the specified language
func speakWord(word string, langCode string) error {
	voice := getVoiceForLanguage(langCode)
	
	var cmd *exec.Cmd
	if voice != "" {
		// Use language-specific voice
		// -v specifies the voice, -r sets speech rate (words per minute)
		cmd = exec.Command("say", "-v", voice, "-r", "180", word)
	} else {
		// Fallback to default system voice
		cmd = exec.Command("say", "-r", "180", word)
	}
	
	// cmd.Run() executes the command and waits for completion
	if err := cmd.Run(); err != nil {
		// If voice-specific command fails, try default voice
		cmd := exec.Command("say", "-r", "180", word)
		return cmd.Run()
	}
	return nil
}

// promptWord prompts the user to type a word and validates it
// This uses the Huh library for beautiful terminal prompts
// Uses go-i18n localizer for translations
func promptWord(word string, attempt int, localizer *i18n.Localizer) (string, error) {
	var input string  // Variable to store user input

	// Build prompt title using i18n localizer
	// go-i18n supports template variables like {{.Number}}
	var title string
	if attempt > 1 {
		title, _ = localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "WordPromptRetry",
			TemplateData: map[string]interface{}{
				"Number":  attempt,
				"Attempt": attempt,
			},
		})
	} else {
		title, _ = localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "WordPrompt",
			TemplateData: map[string]interface{}{
				"Number": attempt,
			},
		})
	}

	// Get placeholder text from translations
	placeholder, _ := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "Placeholder",
	})

	// Get validation error message
	validationError, _ := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "ValidationError",
	})

	// Huh provides a fluent API for building forms
	// NewInput() creates a text input field
	// Value(&input) binds the input to our variable (pointer needed)
	// Validate() adds custom validation logic
	err := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&input).  // & gets address of input variable
		Validate(func(s string) error {
			// Anonymous function for validation
			// Returns error if validation fails, nil if OK
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf(validationError)
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
// Uses go-i18n localizer for translations
func formatWordDiff(userInput, correctWord string, localizer *i18n.Localizer) string {
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
	// Get labels from i18n localizer
	yourInputText, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "YourInput"})
	correctText, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "CorrectLabel"})
	diffText, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Differences"})
	
	labelWidth := 14
	yourInputLabel := labelStyle.Width(labelWidth).Render(yourInputText)
	correctLabel := labelStyle.Width(labelWidth).Render(correctText)
	diffLabel := labelStyle.Width(labelWidth).Render(diffText)
	
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

	// Initialize i18n with go-i18n library
	// This loads translation files and creates a localizer
	localizer, err := initI18n(config.Language)
	if err != nil {
		log.Fatalf("Error initializing i18n: %v", err)
	}

	// Shuffle words for variety in practice sessions
	words := shuffleWords(config.Words)
	originalWordCount := len(words)  // Store original count for progress display

	// Print welcome message using i18n localizer
	title, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Title"})
	subtitle, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Subtitle"})
	practiceInstructions, _ := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "PracticeInstructions",
		TemplateData: map[string]interface{}{"Count": originalWordCount},
	})
	pressEnter, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "PressEnter"})
	
	fmt.Printf("🎯 %s\n", title)
	fmt.Println(subtitle)
	fmt.Printf("%s\n\n", practiceInstructions)
	fmt.Println(pressEnter)

	// Track progress
	correctCount := 0
	totalAttempts := 0

	// Practice words using a queue approach
	// When a word is incorrect, it's added back to the end of the queue
	// This gives students a break and lets them practice other words first
	for i := 0; i < len(words); i++ {
		word := words[i]
		totalAttempts++

		// Speak the word using TTS with language-specific voice
		// Show progress: how many words completed correctly out of original total
		fmt.Printf("\n🔊 Word %d: %d of %d completed correctly\n", i+1, correctCount, originalWordCount)
		if err := speakWord(word, config.Language); err != nil {
			// log.Printf doesn't exit, just logs warning
			log.Printf("Warning: Failed to speak word: %v", err)
		}

		// Small delay to let TTS finish speaking
		time.Sleep(500 * time.Millisecond)

		// Prompt user for input with i18n localizer
		// Note: attempt number is always 1 since we don't retry immediately
		userInput, err := promptWord(word, 1, localizer)
		if err != nil {
			log.Fatalf("Error getting input: %v", err)
		}

		// Check if correct (case-sensitive comparison)
		// German requires proper capitalization (nouns are capitalized)
		// Direct string comparison ensures exact match including case
		correctMsg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Correct"})
		incorrectMsg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "IncorrectSpelling"})
		
		if userInput == word {
			fmt.Println(correctMsg)
			correctCount++
		} else {
			// Show colorful feedback with visual diff to help learning
			fmt.Println(errorStyle.Render(incorrectMsg))
			fmt.Println(formatWordDiff(userInput, word, localizer))
			fmt.Print("\n")  // Empty line for readability
			
			// Add the word back to the end of the queue
			// This allows the student to practice other words first
			// and come back to this one later
			words = append(words, word)
		}
	}

	// Print summary statistics using i18n localizer
	completeMsg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "PracticeComplete"})
	wordsPracticedMsg, _ := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "WordsPracticed",
		TemplateData: map[string]interface{}{"Count": correctCount},
	})
	totalAttemptsMsg, _ := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "TotalAttempts",
		TemplateData: map[string]interface{}{"Count": totalAttempts},
	})
	accuracyMsg, _ := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "Accuracy",
		TemplateData: map[string]interface{}{
			"Percent": fmt.Sprintf("%.1f", float64(correctCount)/float64(totalAttempts)*100),
		},
	})
	
	fmt.Println("\n" + strings.Repeat("=", 30))
	fmt.Println(completeMsg)
	fmt.Println(wordsPracticedMsg)
	fmt.Println(totalAttemptsMsg)
	fmt.Println(accuracyMsg)
	fmt.Println(strings.Repeat("=", 30))
}
