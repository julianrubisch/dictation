package main

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

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
				return errors.New(validationError)
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
