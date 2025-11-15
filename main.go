package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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

	// Create and run the TUI
	model := initialAppModel(localizer, config.Language, words)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
