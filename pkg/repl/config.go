package repl

import "time"

// AutocompleteConfig controls autocomplete request and key-routing behavior.
type AutocompleteConfig struct {
	// Enabled toggles REPL autocomplete integration.
	Enabled bool
	// Debounce is the delay after input edits before a completion request is sent.
	Debounce time.Duration
	// RequestTimeout bounds a single completion request.
	RequestTimeout time.Duration
	// TriggerKeys are optional explicit shortcut keys (for example "tab").
	TriggerKeys []string
	// AcceptKeys apply the selected completion while the popup is visible.
	AcceptKeys []string
	// FocusToggleKey switches between input and timeline focus.
	FocusToggleKey string
	// MaxSuggestions limits the number of rendered popup items.
	MaxSuggestions int
}

// DefaultAutocompleteConfig returns default autocomplete settings.
func DefaultAutocompleteConfig() AutocompleteConfig {
	return AutocompleteConfig{
		Enabled:        true,
		Debounce:       120 * time.Millisecond,
		RequestTimeout: 400 * time.Millisecond,
		TriggerKeys:    []string{"tab"},
		AcceptKeys:     []string{"enter", "tab"},
		FocusToggleKey: "",
		MaxSuggestions: 8,
	}
}

// Config holds REPL shell configuration.
type Config struct {
	Title                string
	Prompt               string
	Placeholder          string
	Width                int
	StartMultiline       bool
	EnableExternalEditor bool
	EnableHistory        bool
	MaxHistorySize       int
	// Autocomplete controls completion scheduling, shortcuts, and popup behavior.
	Autocomplete AutocompleteConfig
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Title:                "REPL",
		Prompt:               "> ",
		Placeholder:          "Enter code or /command",
		Width:                80,
		StartMultiline:       false,
		EnableExternalEditor: false,
		EnableHistory:        true,
		MaxHistorySize:       1000,
		Autocomplete:         DefaultAutocompleteConfig(),
	}
}
