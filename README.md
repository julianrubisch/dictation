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

The `config.yaml` file should contain a list of German words:

```yaml
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

The application uses macOS's built-in `say` command with the German voice "Anna". If Anna is not available, it falls back to the default system voice. The speech rate is set to 180 words per minute for clarity.

To see available German voices on your system:
```bash
say -v '?' | grep -i german
```

## Best Practices

- Start with familiar words and gradually add more challenging ones
- Practice regularly for best results
- Focus on accuracy rather than speed
- Review incorrect words after each session

## Dependencies

- [Huh](https://github.com/charmbracelet/huh) - Beautiful, accessible terminal forms
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing

## License

This project is open source and available for educational use.
