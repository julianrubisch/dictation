package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Version is set at build time using ldflags
// Example: go build -ldflags "-X main.Version=v1.0.0"
var Version = "dev"

func main() {
	// main() is the entry point of every Go program
	// os.Args contains command-line arguments
	// os.Args[0] is the program name, os.Args[1:] are arguments
	
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Printf("dictation version %s\n", Version)
		os.Exit(0)
	}
	
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
		progressMsg, _ := localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "ProgressMessage",
			TemplateData: map[string]interface{}{
				"Current":   i + 1,
				"Completed": correctCount,
				"Total":     originalWordCount,
			},
		})
		fmt.Printf("\n🔊 %s\n", progressMsg)
		if err := speakWord(word, config.Language); err != nil {
			// log.Printf doesn't exit, just logs warning
			log.Printf("Warning: Failed to speak word: %v", err)
		}

		// Small delay to let TTS finish speaking
		time.Sleep(500 * time.Millisecond)

		// Prompt user for input with i18n localizer
		// Note: attempt number is always 1 since we don't retry immediately
		// Pass language code for TTS when TAB is pressed
		userInput, err := promptWord(word, 1, config.Language, localizer)
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
