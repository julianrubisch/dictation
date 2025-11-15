package main

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

// initI18n initializes the i18n bundle and loads translation files
// This is the idiomatic Go approach using go-i18n library
func initI18n(langCode string) (*i18n.Localizer, error) {
	// Create bundle with English as default language
	// The bundle manages all translation files
	bundle := i18n.NewBundle(language.English)
	
	// Register TOML unmarshal function
	// This allows go-i18n to parse TOML translation files
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	
	// Load translation files
	// These files contain all user-facing strings for each language
	// LoadMessageFile returns (*MessageFile, error)
	_, err := bundle.LoadMessageFile("active.en.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load English translations: %w", err)
	}
	_, err = bundle.LoadMessageFile("active.de.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load German translations: %w", err)
	}
	
	// Create localizer for the requested language
	// The localizer provides methods to get translated strings
	localizer := i18n.NewLocalizer(bundle, langCode)
	
	return localizer, nil
}
