# German Dictation Practice

A CLI application built with Go and [Huh](https://github.com/charmbracelet/huh) that helps students practice German spelling through dictation exercises.

## Features

- 📝 Reads German words from a YAML configuration file
- 🔀 Shuffles words for varied practice sessions
- 🔊 Uses macOS native Text-to-Speech to pronounce words
- ⌨️ Interactive prompts for typing practice
- ✅ Validates spelling and provides feedback
- 📊 Shows progress and accuracy statistics

## Requirements

- macOS (for native TTS support)
- Go 1.17 or later

## Installation

1. Clone or download this repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Build the application:
   ```bash
   go build -o dictation
   ```

## Usage

1. Edit `config.yaml` to add your German words:
   ```yaml
   words:
     - Haus
     - Buch
     - Schule
     - Freund
   ```

2. Run the application:
   ```bash
   ./dictation
   ```

   Or specify a custom config file:
   ```bash
   ./dictation my-words.yaml
   ```

3. The application will:
   - Shuffle the words
   - Speak each word using macOS TTS
   - Prompt you to type what you heard
   - Continue until you spell each word correctly
   - Show a summary with your accuracy

## Configuration

The `config.yaml` file should contain a language code and a list of words:

```yaml
language: de  # Language code: 'en' for English, 'de' for German
words:
  - Haus
  - Buch
  - Schule
  - Freund
  - Wasser
  - Apfel
  - Auto
  - Garten
```

### Language Configuration

The `language` field specifies the interface language and TTS voice:
- `en` - English (uses Alex voice)
- `de` - German (uses Anna voice)
- Defaults to `en` if not specified

The language setting affects:
- All user-facing text (prompts, messages, labels)
- Text-to-Speech voice selection
- Error messages and feedback

### Adding New Languages

To add support for a new language:

1. **Create a translation file:**
   Create `active.XX.toml` where `XX` is the language code (e.g., `active.fr.toml` for French)

2. **Copy the structure from an existing file:**
   ```bash
   cp active.en.toml active.fr.toml
   ```

3. **Translate all messages:**
   Edit `active.fr.toml` and translate all the `other = "..."` values:
   ```toml
   [Title]
   other = "Exercice de Dictée"  # Translated title
   
   [WordPrompt]
   other = "Mot {{.Number}}: Écrivez ce que vous avez entendu"
   # ... translate all other messages
   ```

4. **Update `i18n.go`:**
   Add a line to load the new translation file:
   ```go
   _, err = bundle.LoadMessageFile("active.fr.toml")
   if err != nil {
       return nil, fmt.Errorf("failed to load French translations: %w", err)
   }
   ```

5. **Add TTS voice mapping:**
   In `tts.go`, add the voice to `getVoiceForLanguage()`:
   ```go
   voices := map[string]string{
       "de": "Anna",
       "en": "Alex",
       "fr": "Thomas",  // Add your new language
   }
   ```

6. **Test:**
   Update `config.yaml` to use the new language code and test the application.

**Note:** Template variables like `{{.Number}}`, `{{.Count}}`, etc. should remain unchanged - only translate the surrounding text.

## How It Works

1. **Load Configuration**: Reads words from the YAML file
2. **Shuffle**: Randomly shuffles the word order for each session
3. **Practice Loop**: For each word:
   - Uses macOS `say` command with German voice (Anna) to pronounce the word
   - Prompts you to type the word using Huh's interactive input
   - Validates your spelling (case-insensitive)
   - Repeats if incorrect until you get it right
4. **Summary**: Displays statistics about your practice session

## Text-to-Speech

The application uses macOS's built-in `say` command with language-specific voices:
- **German (de)**: Uses "Anna" voice
- **English (en)**: Uses "Alex" voice
- **Other languages**: Falls back to default system voice

The speech rate is set to 180 words per minute for clarity.

To see available voices on your system:
```bash
# List all voices
say -v '?'

# List German voices
say -v '?' | grep -i german

# List English voices
say -v '?' | grep -i english
```

If a language-specific voice is not available, the application falls back to the default system voice.

## Best Practices

- Start with familiar words and gradually add more challenging ones
- Practice regularly for best results
- Focus on accuracy rather than speed
- Review incorrect words after each session

## Dependencies

- [Huh](https://github.com/charmbracelet/huh) - Beautiful, accessible terminal forms
- [go-i18n](https://github.com/nicksnyder/go-i18n) - Internationalization library
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing
- [go-toml](https://github.com/pelletier/go-toml) - TOML parsing for translations

## License

This project is open source and available for educational use.
