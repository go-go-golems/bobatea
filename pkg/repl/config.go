package repl

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

	// Optional helper content rendered above the input when toggled
	HelperMarkdown string

	// Keybinding to toggle focus between input and timeline (default: "tab")
	FocusToggleKey string
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
		FocusToggleKey:       "tab",
	}
}
