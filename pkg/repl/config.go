package repl

// Config holds REPL shell configuration.
type Config struct {
	Title                string
	Prompt               string
	Placeholder          string
	Width                int
	FocusToggleKey       string
	StartMultiline       bool
	EnableExternalEditor bool
	EnableHistory        bool
	MaxHistorySize       int
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Title:                "REPL",
		Prompt:               "> ",
		Placeholder:          "Enter code or /command",
		Width:                80,
		FocusToggleKey:       "",
		StartMultiline:       false,
		EnableExternalEditor: false,
		EnableHistory:        true,
		MaxHistorySize:       1000,
	}
}
