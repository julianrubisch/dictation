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
	correctWords []string
	language     string
	localizer    *i18n.Localizer
	
	// Dialog state
	dialogState  dialogState
	dialogType   dialogType
	dialogDiff   string
	
	// Input state
	inputText    string
	showInput    bool
	inputError   string
}

// Styles for the TUI
var (
	titleBarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true).
			BorderForeground(lipgloss.Color("6")).  // Turquoise border
			Foreground(lipgloss.Color("15")).       // White text
			Bold(true).
			Padding(0, 1)
	
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
	}
}

// Init initializes the model and starts the first word
func (m appModel) Init() tea.Cmd {
	return m.startNextWord()
}

// Update handles messages and updates the model
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		headerHeight := 3 // Title bar with borders
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight
		}
		return m, nil
		
	case tuiRepeatAudioMsg:
		// Audio repetition completed - no action needed
		return m, nil
		
	case speakWordMsg:
		// Word spoken, show input prompt
		m.showInput = true
		m.updateViewportContent()
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
		
		// Handle input when showing input prompt
		if m.showInput {
			switch msg.String() {
			case "enter":
				input := strings.TrimSpace(m.inputText)
				if input == "" {
					validationError, _ := m.localizer.Localize(&i18n.LocalizeConfig{
						MessageID: "ValidationError",
					})
					m.inputError = validationError
					m.updateViewportContent()
					return m, nil
				}
				return m.validateInput(input)
			case "tab":
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
				if len(msg.Runes) > 0 {
					m.inputText += string(msg.Runes)
					m.inputError = ""
					m.updateViewportContent()
				}
				return m, nil
			}
		}
		
		// Global quit handler
		if msg.String() == "q" || msg.String() == "ctrl+c" {
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
	titleBar := m.renderTitleBar()
	s.WriteString(titleBar)
	
	if m.dialogState == dialogShowing {
		// Show dialog centered below title bar
		titleBarHeight := strings.Count(titleBar, "\n") + 1
		remainingHeight := m.height - titleBarHeight
		if remainingHeight < 0 {
			remainingHeight = m.height
		}
		
		dialog := m.renderDialog()
		centeredDialog := lipgloss.Place(
			m.width, remainingHeight,
			lipgloss.Center, lipgloss.Center,
			dialog,
		)
		s.WriteString(centeredDialog)
	} else {
		// Show viewport content
		s.WriteString(m.viewport.View())
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
	
	// Width minus 2 for border characters (left + right)
	contentWidth := m.width - 2
	if contentWidth < 0 {
		contentWidth = m.width
	}
	return titleBarStyle.Width(contentWidth).Render("🔊 " + progressMsg)
}

// renderDialog renders the feedback dialog
func (m appModel) renderDialog() string {
	var title string
	var style lipgloss.Style
	
	if m.dialogType == dialogCorrect {
		title, _ = m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "Correct"})
		style = dialogBoxStyle.Copy().Inherit(correctDialogStyle)
	} else {
		title, _ = m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "IncorrectSpelling"})
		style = dialogBoxStyle.Copy().Inherit(incorrectDialogStyle)
	}
	
	var dialog strings.Builder
	dialog.WriteString(dialogTitleStyle.Render(title))
	dialog.WriteString("\n\n")
	
	if m.dialogDiff != "" {
		dialog.WriteString(m.dialogDiff)
	}
	
	pressEnterMsg, _ := m.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "PressEnterToContinue",
	})
	dialog.WriteString("\n(" + pressEnterMsg + ")")
	
	return style.Render(dialog.String())
}

// updateViewportContent updates the viewport content
func (m *appModel) updateViewportContent() {
	if !m.showInput {
		m.viewport.SetContent("Waiting for next word...")
		return
	}
	
	var content strings.Builder
	
	title, _ := m.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "WordPrompt",
		TemplateData: map[string]interface{}{"Number": m.wordIndex + 1},
	})
	placeholder, _ := m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "Placeholder"})
	tabHint, _ := m.localizer.Localize(&i18n.LocalizeConfig{MessageID: "TabHint"})
	
	content.WriteString(title)
	content.WriteString("\n\n")
	
	if m.inputText == "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(placeholder))
	} else {
		content.WriteString(m.inputText)
	}
	content.WriteString("█\n\n")
	
	if m.inputError != "" {
		content.WriteString(errorStyle.Render("❌ " + m.inputError))
		content.WriteString("\n")
	}
	
	content.WriteString(tabHint)
	m.viewport.SetContent(content.String())
}

// validateInput validates the user input and shows feedback
func (m *appModel) validateInput(input string) (tea.Model, tea.Cmd) {
	if m.currentWord == "" {
		// Fallback: try to get word from array (shouldn't be needed)
		if m.wordIndex < len(m.words) {
			m.currentWord = m.words[m.wordIndex]
		} else {
			return m, nil // Can't validate without a word
		}
	}
	
	if input == m.currentWord {
		m.correctCount++
		m.correctWords = append(m.correctWords, m.currentWord)
		m.dialogType = dialogCorrect
		m.dialogDiff = ""
	} else {
		m.dialogType = dialogIncorrect
		m.dialogDiff = formatWordDiff(input, m.currentWord, m.localizer)
	}
	
	m.dialogState = dialogShowing
	m.inputText = ""
	m.inputError = ""
	m.showInput = false
	
	return m, nil
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
		return tea.Quit
	}
	
	word := m.words[m.wordIndex]
	if word == "" {
		return tea.Quit
	}
	
	m.currentWord = word
	m.inputText = ""
	m.inputError = ""
	m.showInput = false
	m.dialogState = dialogHidden
	m.updateViewportContent()
	
	// Speak the word
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
	// If word was incorrect, add it back to the end of the queue
	if m.dialogType == dialogIncorrect && m.currentWord != "" {
		m.words = append(m.words, m.currentWord)
	}
	
	m.dialogState = dialogHidden
	m.dialogDiff = ""
	m.wordIndex++
	
	return m.startNextWord()
}
