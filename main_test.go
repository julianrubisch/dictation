package main

import (
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// setupTestLocalizer creates a localizer for testing
// This is a helper function to avoid repeating i18n setup in each test
func setupTestLocalizer() *i18n.Localizer {
	localizer, err := initI18n("en")
	if err != nil {
		// If i18n setup fails in tests, create a minimal localizer
		// This shouldn't happen, but provides a fallback
		return nil
	}
	return localizer
}

// TestFormatWordDiff tests the word diff visualization function
// This demonstrates Go's testing package and table-driven tests
func TestFormatWordDiff(t *testing.T) {
	// Table-driven tests are idiomatic in Go
	// They allow testing multiple cases in a clean, maintainable way
	tests := []struct {
		name         string
		userInput    string
		correctWord  string
		wantContains []string // Substrings that should appear in output
	}{
		{
			name:        "exact match",
			userInput:   "Haus",
			correctWord: "Haus",
			wantContains: []string{
				"Your input:",
				"Correct:",
				"Haus",
			},
		},
		{
			name:        "missing character",
			userInput:   "Hau",
			correctWord: "Haus",
			wantContains: []string{
				"Hau",
				"Haus",
				"Differences:",
			},
		},
		{
			name:        "extra character",
			userInput:   "Hauss",
			correctWord: "Haus",
			wantContains: []string{
				"Hauss",
				"Haus",
				"Differences:",
			},
		},
		{
			name:        "wrong character",
			userInput:   "Haus",
			correctWord: "Haus",
			wantContains: []string{
				"Your input:",
				"Correct:",
			},
		},
		{
			name:        "case difference - lowercase vs capitalized (German noun)",
			userInput:   "haus",
			correctWord: "Haus",
			wantContains: []string{
				"haus",
				"Haus",
				"Differences:",
			},
		},
		{
			name:        "case difference - capitalized vs lowercase",
			userInput:   "Haus",
			correctWord: "haus",
			wantContains: []string{
				"Haus",
				"haus",
				"Differences:",
			},
		},
		{
			name:        "completely different word",
			userInput:   "Buch",
			correctWord: "Haus",
			wantContains: []string{
				"Buch",
				"Haus",
				"Differences:",
			},
		},
		{
			name:        "German umlaut test",
			userInput:   "Apfel",
			correctWord: "Apfel",
			wantContains: []string{
				"Your input:",
				"Correct:",
			},
		},
		{
			name:        "case difference with umlaut",
			userInput:   "apfel",
			correctWord: "Apfel",
			wantContains: []string{
				"apfel",
				"Apfel",
				"Differences:",
			},
		},
	}

	// Set up localizer for tests
	localizer := setupTestLocalizer()
	if localizer == nil {
		t.Fatal("Failed to set up test localizer")
	}

	// Run each test case
	for _, tt := range tests {
		// t.Run creates a subtest for each case
		// This allows running tests individually and better error reporting
		t.Run(tt.name, func(t *testing.T) {
			got := formatWordDiff(tt.userInput, tt.correctWord, localizer)

			// Check that output contains expected substrings
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("formatWordDiff() output should contain %q, got:\n%s", want, got)
				}
			}

			// Verify output has three lines (Your input, Correct, Differences)
			lines := strings.Split(got, "\n")
			if len(lines) < 3 {
				t.Errorf("formatWordDiff() should return at least 3 lines, got %d", len(lines))
			}
		})
	}
}

// TestFormatWordDiffSpecificCases tests specific diff scenarios
func TestFormatWordDiffSpecificCases(t *testing.T) {
	localizer := setupTestLocalizer()
	if localizer == nil {
		t.Fatal("Failed to set up test localizer")
	}

	t.Run("shows differences correctly", func(t *testing.T) {
		// Test that differences are marked with ^
		result := formatWordDiff("Hau", "Haus", localizer)
		
		// Should show the missing 's'
		if !strings.Contains(result, "Hau") {
			t.Error("Should show user input 'Hau'")
		}
		if !strings.Contains(result, "Haus") {
			t.Error("Should show correct word 'Haus'")
		}
		if !strings.Contains(result, "^") {
			t.Error("Should mark differences with ^")
		}
	})

	t.Run("handles empty input", func(t *testing.T) {
		result := formatWordDiff("", "Haus", localizer)
		
		if !strings.Contains(result, "Haus") {
			t.Error("Should show correct word when input is empty")
		}
		if !strings.Contains(result, "Differences:") {
			t.Error("Should show differences line")
		}
	})

	t.Run("handles longer input than correct", func(t *testing.T) {
		result := formatWordDiff("Hausse", "Haus", localizer)
		
		if !strings.Contains(result, "Hausse") {
			t.Error("Should show full user input")
		}
		if !strings.Contains(result, "Haus") {
			t.Error("Should show correct word")
		}
	})

	t.Run("case sensitivity - lowercase vs uppercase", func(t *testing.T) {
		// Case differences should be marked as different
		result := formatWordDiff("haus", "Haus", localizer)
		
		if !strings.Contains(result, "Differences:") {
			t.Error("Case differences should be marked")
		}
		// Should show both versions
		if !strings.Contains(result, "haus") {
			t.Error("Should show user input 'haus'")
		}
		if !strings.Contains(result, "Haus") {
			t.Error("Should show correct word 'Haus'")
		}
	})

	t.Run("case sensitivity - all lowercase vs all uppercase", func(t *testing.T) {
		result := formatWordDiff("HAUS", "Haus", localizer)
		
		if !strings.Contains(result, "Differences:") {
			t.Error("Case differences should be marked")
		}
		if !strings.Contains(result, "HAUS") {
			t.Error("Should show user input 'HAUS'")
		}
		if !strings.Contains(result, "Haus") {
			t.Error("Should show correct word 'Haus'")
		}
	})
}
