package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

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
