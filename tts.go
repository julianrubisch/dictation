package main

import (
	"os/exec"
)

// getVoiceForLanguage returns the macOS TTS voice name for a language code
// Maps language codes to appropriate voices for better pronunciation
func getVoiceForLanguage(langCode string) string {
	voices := map[string]string{
		"de": "Anna",    // German voice
		"en": "Alex",    // English voice (US)
		"fr": "Thomas",  // French voice (for future use)
	}

	if voice, ok := voices[langCode]; ok {
		return voice
	}
	// Fallback to default system voice
	return ""
}

// speakWord uses macOS's native 'say' command to speak a word
// Uses the appropriate voice for the specified language
func speakWord(word string, langCode string) error {
	voice := getVoiceForLanguage(langCode)
	
	var cmd *exec.Cmd
	if voice != "" {
		// Use language-specific voice
		// -v specifies the voice, -r sets speech rate (words per minute)
		cmd = exec.Command("say", "-v", voice, "-r", "180", word)
	} else {
		// Fallback to default system voice
		cmd = exec.Command("say", "-r", "180", word)
	}
	
	// cmd.Run() executes the command and waits for completion
	if err := cmd.Run(); err != nil {
		// If voice-specific command fails, try default voice
		cmd := exec.Command("say", "-r", "180", word)
		return cmd.Run()
	}
	return nil
}
