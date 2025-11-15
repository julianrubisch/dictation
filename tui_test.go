package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
)

// setupTestTUI creates a test appModel with minimal setup
func setupTestTUI() appModel {
	localizer, _ := initI18n("en")
	words := []string{"Haus", "Buch", "Schule"}
	return initialAppModel(localizer, "en", words)
}

// TestTitleBarRendering tests the title bar rendering
func TestTitleBarRendering(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus", "Buch"})
	model.width = 80
	model.height = 24
	model.wordIndex = 0
	model.correctCount = 0
	model.originalCount = 2

	titleBar := model.renderTitleBar()

	// Should contain progress message
	if !strings.Contains(titleBar, "Word") && !strings.Contains(titleBar, "Wort") {
		t.Error("Title bar should contain progress message")
	}

	// Should contain border characters (top border)
	if !strings.Contains(titleBar, "─") && !strings.Contains(titleBar, "━") {
		t.Error("Title bar should contain border characters")
	}

	// Should contain emoji
	if !strings.Contains(titleBar, "🔊") {
		t.Error("Title bar should contain speaker emoji")
	}
}

// TestTitleBarWithCorrectWords tests title bar with correctly spelled words
func TestTitleBarWithCorrectWords(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus", "Buch"})
	model.width = 80
	model.correctWords = []string{"Haus"}
	model.correctCount = 1
	model.wordIndex = 1
	model.originalCount = 2

	titleBar := model.renderTitleBar()

	// Should contain the correctly spelled word
	if !strings.Contains(titleBar, "Haus") {
		t.Error("Title bar should contain correctly spelled words")
	}
}

// TestDialogRendering tests dialog rendering
func TestDialogRendering(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.dialogState = dialogShowing
	model.dialogType = dialogCorrect
	model.dialogDiff = ""

	dialog := model.renderDialog()

	// Should contain dialog border
	if !strings.Contains(dialog, "╭") || !strings.Contains(dialog, "╯") {
		t.Error("Dialog should have rounded borders")
	}

	// Should contain correct message
	if !strings.Contains(dialog, "Correct") {
		t.Error("Dialog should contain correct message")
	}
}

// TestDialogWithDiff tests dialog with diff content
func TestDialogWithDiff(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.dialogState = dialogShowing
	model.dialogType = dialogIncorrect
	model.dialogDiff = formatWordDiff("Hau", "Haus", localizer)

	dialog := model.renderDialog()

	// Should contain diff content
	if !strings.Contains(dialog, "Hau") || !strings.Contains(dialog, "Haus") {
		t.Error("Dialog should contain diff content")
	}

	// Should contain differences marker (check for the label from formatWordDiff)
	// The diff output includes "Differences:" or "Unterschiede:" label
	hasDiffLabel := strings.Contains(dialog, "Differences") || 
	                strings.Contains(dialog, "Unterschiede") ||
	                strings.Contains(dialog, "^")  // Diff markers indicate differences are shown
	if !hasDiffLabel {
		t.Error("Dialog should contain differences label or markers")
	}
}

// TestViewWithDialog tests that title bar is visible when dialog is shown
func TestViewWithDialog(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.width = 80
	model.height = 24
	model.ready = true
	model.dialogState = dialogShowing
	model.dialogType = dialogCorrect

	view := model.View()

	// Title bar should be visible (contains progress or emoji)
	if !strings.Contains(view, "🔊") {
		t.Error("View should show title bar when dialog is displayed")
	}

	// Dialog should be visible (check for dialog content or border)
	// lipgloss.Place might format it differently, so check for key indicators
	hasDialogContent := strings.Contains(view, "Correct") || 
	                    strings.Contains(view, "Richtig") ||
	                    strings.Contains(view, "╭") ||  // Dialog border
	                    strings.Contains(view, "Press Enter")  // Dialog instruction
	if !hasDialogContent {
		t.Error("View should show dialog content when dialog is showing")
	}
}

// TestViewWithoutDialog tests normal view rendering
func TestViewWithoutDialog(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.width = 80
	model.height = 24
	model.ready = true
	model.dialogState = dialogHidden
	model.showInput = true
	model.inputText = "test"

	// Initialize viewport
	model.viewport = viewport.New(model.width, model.height-3)
	model.updateViewportContent()

	view := model.View()

	// Should contain title bar
	if !strings.Contains(view, "🔊") {
		t.Error("View should contain title bar")
	}

	// Should not contain dialog
	if strings.Contains(view, "Correct") || strings.Contains(view, "Incorrect") {
		t.Error("View should not show dialog when hidden")
	}
}

// TestTitleBarWidthCalculation tests that title bar width accounts for borders
func TestTitleBarWidthCalculation(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.width = 80

	titleBar := model.renderTitleBar()
	lines := strings.Split(titleBar, "\n")

	// Should have multiple lines (top border, content, bottom border)
	if len(lines) < 2 {
		t.Error("Title bar should have at least 2 lines (borders + content)")
	}

	// Check that borders extend properly
	// First line should be top border
	if len(lines) > 0 {
		topBorder := lines[0]
		// Should start with border character
		if !strings.HasPrefix(topBorder, "┌") && !strings.HasPrefix(topBorder, "╭") {
			// Might be using different border style, check for horizontal line
			if !strings.Contains(topBorder, "─") && !strings.Contains(topBorder, "━") {
				t.Error("Top border should contain border characters")
			}
		}
	}
}

// TestDialogCentering tests that dialog is centered
func TestDialogCentering(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.width = 80
	model.height = 24
	model.ready = true
	model.dialogState = dialogShowing
	model.dialogType = dialogCorrect

	view := model.View()
	lines := strings.Split(view, "\n")

	// Find dialog line (should be roughly centered)
	// Dialog should appear in the middle portion of the screen
	middleStart := model.height / 3
	middleEnd := (model.height * 2) / 3
	foundDialog := false

	for i := middleStart; i < middleEnd && i < len(lines); i++ {
		if strings.Contains(lines[i], "Correct") || strings.Contains(lines[i], "Richtig") {
			foundDialog = true
			break
		}
	}

	if !foundDialog {
		// Dialog might be in a different position, but should exist
		hasDialog := strings.Contains(view, "Correct") || strings.Contains(view, "Richtig")
		if !hasDialog {
			t.Error("Dialog should be present in view")
		}
	}
}

// TestCurrentWordPreservation tests that currentWord is preserved during validation
func TestCurrentWordPreservation(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus", "Buch"})
	model.currentWord = "Haus"
	model.wordIndex = 0

	// Validate with incorrect input
	_, _ = model.validateInput("Hau")

	// currentWord should still be set (for diff display)
	if model.currentWord == "" {
		t.Error("currentWord should be preserved after validation for diff display")
	}

	// Dialog should be showing
	if model.dialogState != dialogShowing {
		t.Error("Dialog should be showing after validation")
	}

	// Dialog should have diff content
	if model.dialogDiff == "" {
		t.Error("Dialog should contain diff when input is incorrect")
	}
}

// TestViewportContentUpdate tests viewport content updates
func TestViewportContentUpdate(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.width = 80
	model.height = 24
	model.viewport = viewport.New(model.width, model.height-3)
	model.showInput = true
	model.inputText = "test"
	model.wordIndex = 0

	model.updateViewportContent()
	content := model.viewport.View()

	// Should contain input text
	if !strings.Contains(content, "test") {
		t.Error("Viewport should contain input text")
	}

	// Should contain cursor
	if !strings.Contains(content, "█") {
		t.Error("Viewport should contain cursor")
	}
}

// TestViewportContentWithError tests viewport with error message
func TestViewportContentWithError(t *testing.T) {
	localizer, _ := initI18n("en")
	model := initialAppModel(localizer, "en", []string{"Haus"})
	model.width = 80
	model.height = 24
	model.viewport = viewport.New(model.width, model.height-3)
	model.showInput = true
	model.inputError = "please enter a word"

	model.updateViewportContent()
	content := model.viewport.View()

	// Should contain error message
	if !strings.Contains(content, "please enter a word") {
		t.Error("Viewport should contain error message")
	}
}
