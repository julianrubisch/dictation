package main

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// inputModel is a Bubble Tea model for text input with TAB key support
// This allows us to listen for TAB to repeat audio while typing
type inputModel struct {
	textInput   textinput.Model
	title       string
	placeholder string
	word        string        // The word being practiced (for repeating audio)
	language    string        // Language code for TTS
	localizer   *i18n.Localizer
	done        bool          // Whether user has submitted
	err         error         // Any error that occurred
}

// repeatAudioMsg is a message to trigger audio repetition
type repeatAudioMsg struct{}

// initialModel creates a new input model
func initialModel(word, language string, title, placeholder string, localizer *i18n.Localizer) inputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return inputModel{
		textInput:   ti,
		title:       title,
		placeholder: placeholder,
		word:        word,
		language:    language,
		localizer:   localizer,
		done:        false,
	}
}

// Init is required by the Bubble Tea Model interface
func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// User cancelled
			m.err = errors.New("cancelled")
			m.done = true
			return m, tea.Quit

		case "enter":
			// User submitted
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				// Empty input - show validation error but don't quit
				validationError, _ := m.localizer.Localize(&i18n.LocalizeConfig{
					MessageID: "ValidationError",
				})
				m.err = errors.New(validationError)
				return m, nil
			}
			m.done = true
			return m, tea.Quit

		case "tab":
			// TAB pressed - repeat audio
			// Use tea.ExecProcess to run TTS asynchronously without blocking UI
			voice := getVoiceForLanguage(m.language)
			var cmd *exec.Cmd
			if voice != "" {
				cmd = exec.Command("say", "-v", voice, "-r", "180", m.word)
			} else {
				cmd = exec.Command("say", "-r", "180", m.word)
			}
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				// If TTS fails, try fallback with default voice
				if err != nil && voice != "" {
					fallbackCmd := exec.Command("say", "-r", "180", m.word)
					_ = fallbackCmd.Run() // Ignore errors in fallback
				}
				return repeatAudioMsg{}
			})

		default:
			// Handle normal text input
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

	case repeatAudioMsg:
		// Audio repetition completed, continue normally
		return m, nil

	default:
		// Handle other messages (like window resize)
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// View renders the UI
func (m inputModel) View() string {
	// Get hint text from translations
	tabHint, _ := m.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "TabHint",
	})

	var s strings.Builder
	s.WriteString(m.title)
	s.WriteString("\n\n")
	s.WriteString(m.textInput.View())
	s.WriteString("\n\n")
	if m.err != nil {
		s.WriteString("❌ " + m.err.Error() + "\n")
	}
	s.WriteString(tabHint)
	s.WriteString("\n")
	return s.String()
}

// promptWord prompts the user to type a word and validates it
// This uses Bubble Tea for custom keyboard handling (TAB to repeat audio)
// Uses go-i18n localizer for translations
func promptWord(word string, attempt int, language string, localizer *i18n.Localizer) (string, error) {
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

	// Create and run the Bubble Tea program
	model := initialModel(word, language, title, placeholder, localizer)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	// Type assert to get our model back
	m := finalModel.(inputModel)
	if m.err != nil {
		return "", m.err
	}

	// Return the trimmed input
	return strings.TrimSpace(m.textInput.Value()), nil
}
