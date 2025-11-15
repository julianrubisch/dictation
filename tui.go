package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// dialogState represents the state of a dialog
type dialogState int

const (
	dialogHidden dialogState = iota
	dialogShowing
)

// dialogType represents the type of dialog
type dialogType int

const (
	dialogCorrect dialogType = iota
	dialogIncorrect
)

// appModel is the main TUI model for the dictation practice app
// It uses a viewport to maintain a steady window with title bar and content area
type appModel struct {
	viewport     viewport.Model
	ready        bool
	width        int
	height       int
	
	// Application state
	words        []string  // Queue of words to practice
	originalCount int      // Original word count for progress
	currentWord  string
	wordIndex    int       // Current word index in practice
	correctCount int
	totalAttempts int
	correctWords []string
	language     string
	localizer    *i18n.Localizer
	
	// Dialog state
	dialogState  dialogState
	dialogType   dialogType
	dialogMsg    string
	dialogDiff   string
	
	// Input state
	inputText    string
	showInput    bool
	inputError   string
	
	// State management
	waitingForAudio bool  // Waiting for TTS to finish
}

// Styles for the TUI
var (
	titleBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).  // White
			Background(lipgloss.Color("6")).   // Turquoise
			Bold(true).
			Padding(0, 1)
	
	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)
	
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).  // Turquoise
			Padding(1, 2).
			Margin(1, 0).
			Width(60)  // Set minimum width for dialog
	
	dialogTitleStyle = lipgloss.NewStyle().
			Bold(true).
			MarginBottom(1)
	
	correctDialogStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("10")).  // Green
			Foreground(lipgloss.Color("10"))
	
	incorrectDialogStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("9")).  // Red
			Foreground(lipgloss.Color("9"))
)

// initialAppModel creates a new app model
func initialAppModel(localizer *i18n.Localizer, language string, words []string) appModel {
	return appModel{
		localizer:      localizer,
		language:       language,
		words:          words,
		originalCount:  len(words),
		correctWords:   []string{},
		wordIndex:      0,
		showInput:      false,
		dialogState:    dialogHidden,
		waitingForAudio: false,
	}
}

// Init initializes the model and starts the first word
func (m appModel) Init() tea.Cmd {
	// Start with the first word
	model := m
	return model.startNextWord()
}

// Update handles messages and updates the model
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		if !m.ready {
			// Initialize viewport with space for title bar
			headerHeight := 1
			footerHeight := 0
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 1 // Reserve space for title bar
		}
		return m, nil
		
	case tuiRepeatAudioMsg:
		// Audio repetition completed
		return m, nil
		
	case speakWordMsg:
		// Word spoken, now show input
		// Ensure currentWord is still set (it should be, but double-check)
		if m.currentWord == "" && m.wordIndex < len(m.words) {
			m.currentWord = m.words[m.wordIndex]
		}
		m.waitingForAudio = false
		m.showInput = true
		m.updateViewportContent()
		return m, nil
		
	case validationCompleteMsg:
		// Validation complete, dialog is already shown
		// When dialog is closed, we'll move to next word
		return m, nil
		
	case tea.KeyMsg:
		// Handle dialog interactions
		if m.dialogState == dialogShowing {
			switch msg.String() {
			case "enter", " ":
				// Close dialog and continue to next word
				return m, m.handleDialogClose()
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		
		// Handle normal input
		if m.showInput {
			switch msg.String() {
			case "enter":
				// Submit input
				input := strings.TrimSpace(m.inputText)
				if input == "" {
					validationError, _ := m.localizer.Localize(&i18n.LocalizeConfig{
						MessageID: "ValidationError",
					})
					m.inputError = validationError
					m.updateViewportContent()
					return m, nil
				}
				// Validate and show feedback
				return m.validateInput(input)
			case "tab":
				// Repeat audio
				return m, m.repeatAudio()
			case "backspace":
				if len(m.inputText) > 0 {
					m.inputText = m.inputText[:len(m.inputText)-1]
					m.inputError = ""
					m.updateViewportContent()
				}
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				// Add character to input
				if len(msg.Runes) > 0 {
					m.inputText += string(msg.Runes)
					m.inputError = ""
					m.updateViewportContent()
				}
				return m, nil
			}
		}
		
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	
	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the TUI
func (m appModel) View() string {
	if !m.ready {
		return "Initializing..."
	}
	
	var s strings.Builder
	
	// Title bar
	titleBar := m.renderTitleBar()
	s.WriteString(titleBar)
	s.WriteString("\n")
	
	// Content area (viewport)
	content := m.viewport.View()
	s.WriteString(content)
	
	// Dialog overlay (if showing)
	if m.dialogState == dialogShowing {
		dialog := m.renderDialog()
		// Create overlay (centered)
		overlay := lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialog,
		)
		s.WriteString(overlay)
	}
	
	return s.String()
}

// renderTitleBar renders the title bar with progress information
func (m appModel) renderTitleBar() string {
	wordsList := strings.Join(m.correctWords, ", ")
	coloredWordsList := ""
	if wordsList != "" {
		coloredWordsList = turquoiseStyle.Render(wordsList)
	}
	
		progressMsg, _ := m.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "ProgressMessage",
		TemplateData: map[string]interface{}{
			"Current":   m.wordIndex + 1,
			"Completed": m.correctCount,
			"Total":     m.originalCount,
			"Words":     coloredWordsList,
		},
	})
	
	return titleBarStyle.Width(m.width).Render("🔊 " + progressMsg)
}

// renderDialog renders the feedback dialog
func (m appModel) renderDialog() string {
	var dialog strings.Builder
	
	// Dialog title and style
	var title string
	var style lipgloss.Style
	
	if m.dialogType == dialogCorrect {
		title, _ = m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "Correct"})
		style = dialogBoxStyle.Copy().Inherit(correctDialogStyle)
	} else {
		title, _ = m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "IncorrectSpelling"})
		style = dialogBoxStyle.Copy().Inherit(incorrectDialogStyle)
	}
	
	// Title only (no duplicate message)
	dialog.WriteString(dialogTitleStyle.Render(title))
	dialog.WriteString("\n\n")
	
	// Show diff if available (for incorrect answers)
	if m.dialogDiff != "" {
		// The diff already contains newlines, so we don't need to add extra spacing
		dialog.WriteString(m.dialogDiff)
	}
	
	// Instructions
	dialog.WriteString("\n")
	dialog.WriteString("(Press Enter to continue)")
	
	return style.Render(dialog.String())
}

// updateViewportContent updates the viewport content
func (m *appModel) updateViewportContent() {
	var content strings.Builder
	
	if m.showInput {
		// Show input prompt
		var title string
		title, _ = m.localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "WordPrompt",
			TemplateData: map[string]interface{}{
				"Number": m.wordIndex + 1,
			},
		})
		
		placeholder, _ := m.localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "Placeholder",
		})
		
		tabHint, _ := m.localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "TabHint",
		})
		
		content.WriteString(title)
		content.WriteString("\n\n")
		
		// Input field
		if m.inputText == "" {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(placeholder))
		} else {
			content.WriteString(m.inputText)
		}
		content.WriteString("█") // Cursor
		content.WriteString("\n\n")
		
		if m.inputError != "" {
			content.WriteString(errorStyle.Render("❌ " + m.inputError))
			content.WriteString("\n")
		}
		
		content.WriteString(tabHint)
	} else {
		// Show welcome or waiting message
		content.WriteString("Waiting for next word...")
	}
	
	m.viewport.SetContent(content.String())
}

// validateInput validates the user input and shows feedback
func (m *appModel) validateInput(input string) (tea.Model, tea.Cmd) {
	// Store current word before any state changes
	// Try multiple sources to ensure we have the word
	currentWord := m.currentWord
	
	// If currentWord is empty, try to get it from the words array
	if currentWord == "" {
		if m.wordIndex < len(m.words) {
			currentWord = m.words[m.wordIndex]
			// Restore it to m.currentWord for consistency
			m.currentWord = currentWord
		}
	}
	
	// Final check - if still empty, we can't validate
	if currentWord == "" {
		// This is an error state - show error but don't crash
		m.dialogType = dialogIncorrect
		m.dialogMsg = "Error: No word available for comparison"
		m.dialogDiff = "Unable to compare input. Please restart the application."
		m.dialogState = dialogShowing
		return m, nil
	}
	
	if input == currentWord {
		// Correct!
		m.correctCount++
		m.correctWords = append(m.correctWords, currentWord)
		
		m.dialogType = dialogCorrect
		m.dialogMsg = ""  // Title will be shown, no need for separate message
		m.dialogDiff = ""
		m.dialogState = dialogShowing
	} else {
		// Incorrect - show diff
		m.dialogType = dialogIncorrect
		m.dialogMsg = ""  // Title will be shown, no need for separate message
		// Format diff with user input and correct word
		// Use the stored currentWord to ensure we have the value
		m.dialogDiff = formatWordDiff(input, currentWord, m.localizer)
		m.dialogState = dialogShowing
	}
	
	// Clear input (but keep currentWord - don't clear it!)
	m.inputText = ""
	m.inputError = ""
	m.showInput = false
	// NOTE: We intentionally do NOT clear m.currentWord here
	// It must remain available until we move to the next word
	
	// Return a message to notify that validation is complete
	return m, func() tea.Msg {
		return validationCompleteMsg{correct: input == m.currentWord}
	}
}

// validationCompleteMsg is sent when input validation completes
type validationCompleteMsg struct {
	correct bool
}

// repeatAudio repeats the audio for the current word
func (m *appModel) repeatAudio() tea.Cmd {
	return func() tea.Msg {
		if err := speakWord(m.currentWord, m.language); err != nil {
			// Silently fail
		}
		return tuiRepeatAudioMsg{}
	}
}

// tuiRepeatAudioMsg is sent when audio repetition completes in TUI
type tuiRepeatAudioMsg struct{}

// startNextWord starts the next word in the queue
func (m *appModel) startNextWord() tea.Cmd {
	if m.wordIndex >= len(m.words) {
		// All words completed
		return tea.Quit
	}
	
	// Ensure we have a word to practice
	if m.wordIndex >= len(m.words) || m.words[m.wordIndex] == "" {
		return tea.Quit
	}
	
	// Set current word BEFORE any other state changes
	word := m.words[m.wordIndex]
	m.currentWord = word
	m.totalAttempts++
	m.inputText = ""
	m.inputError = ""
	m.showInput = false
	m.waitingForAudio = true
	m.dialogState = dialogHidden
	m.updateViewportContent()
	
	// Speak the word (use local variable to ensure we speak the right word)
	return func() tea.Msg {
		if err := speakWord(word, m.language); err != nil {
			// Continue even if TTS fails
		}
		return speakWordMsg{}
	}
}

// speakWordMsg is sent when word has been spoken
type speakWordMsg struct{}

// handleDialogClose handles closing the dialog and moving to next word
func (m *appModel) handleDialogClose() tea.Cmd {
	// Store current word before clearing state (for queue management)
	currentWordForQueue := m.currentWord
	
	m.dialogState = dialogHidden
	m.dialogMsg = ""
	m.dialogDiff = ""
	
	// Check if word was incorrect - add to end of queue
	// Use stored word to ensure we have the value even if currentWord gets cleared
	if m.dialogType == dialogIncorrect && currentWordForQueue != "" {
		m.words = append(m.words, currentWordForQueue)
	}
	
	// Move to next word
	m.wordIndex++
	
	// Start next word (this will set a new currentWord)
	return m.startNextWord()
}
