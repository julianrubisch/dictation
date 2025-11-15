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
			Margin(1, 0)
	
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
	
	// Dialog title and message
	var title string
	var style lipgloss.Style
	
	if m.dialogType == dialogCorrect {
		title, _ = m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "Correct"})
		style = dialogBoxStyle.Copy().Inherit(correctDialogStyle)
	} else {
		title, _ = m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "IncorrectSpelling"})
		style = dialogBoxStyle.Copy().Inherit(incorrectDialogStyle)
	}
	
	dialog.WriteString(dialogTitleStyle.Render(title))
	dialog.WriteString("\n")
	
	if m.dialogMsg != "" {
		dialog.WriteString(m.dialogMsg)
		dialog.WriteString("\n")
	}
	
	if m.dialogDiff != "" {
		dialog.WriteString("\n")
		dialog.WriteString(m.dialogDiff)
		dialog.WriteString("\n")
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
	if input == m.currentWord {
		// Correct!
		m.correctCount++
		m.correctWords = append(m.correctWords, m.currentWord)
		
		correctMsg, _ := m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "Correct"})
		m.dialogType = dialogCorrect
		m.dialogMsg = correctMsg
		m.dialogDiff = ""
		m.dialogState = dialogShowing
	} else {
		// Incorrect
		incorrectMsg, _ := m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "IncorrectSpelling"})
		m.dialogType = dialogIncorrect
		m.dialogMsg = incorrectMsg
		m.dialogDiff = formatWordDiff(input, m.currentWord, m.localizer)
		m.dialogState = dialogShowing
	}
	
	// Clear input
	m.inputText = ""
	m.inputError = ""
	m.showInput = false
	
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
	
	m.currentWord = m.words[m.wordIndex]
	m.totalAttempts++
	m.inputText = ""
	m.inputError = ""
	m.showInput = false
	m.waitingForAudio = true
	m.dialogState = dialogHidden
	m.updateViewportContent()
	
	// Speak the word
	return func() tea.Msg {
		if err := speakWord(m.currentWord, m.language); err != nil {
			// Continue even if TTS fails
		}
		return speakWordMsg{}
	}
}

// speakWordMsg is sent when word has been spoken
type speakWordMsg struct{}

// handleDialogClose handles closing the dialog and moving to next word
func (m *appModel) handleDialogClose() tea.Cmd {
	m.dialogState = dialogHidden
	m.dialogMsg = ""
	m.dialogDiff = ""
	
	// Check if word was incorrect - add to end of queue
	if m.dialogType == dialogIncorrect {
		m.words = append(m.words, m.currentWord)
	}
	
	// Move to next word
	m.wordIndex++
	
	// Start next word
	return m.startNextWord()
}
