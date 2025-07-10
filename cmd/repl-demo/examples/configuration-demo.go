package main

import (
	"log"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// ConfigurationDemo demonstrates various configuration options
func main() {
	// Create evaluator
	evaluator := repl.NewExampleEvaluator()

	// Create highly customized configuration
	config := repl.Config{
		Title:                "Configuration Demo REPL",
		Prompt:               "config> ",
		Placeholder:          "Enter commands or type /help for assistance",
		Width:                100,
		StartMultiline:       false,
		EnableExternalEditor: true,
		EnableHistory:        true,
		MaxHistorySize:       50,
	}

	// Create model
	model := repl.NewModel(evaluator, config)

	// Use dark theme
	model.SetTheme(repl.BuiltinThemes["dark"])

	// Add configuration-related commands
	model.AddCustomCommand("config", func(args []string) tea.Cmd {
		return func() tea.Msg {
			configInfo := `Current Configuration:
- Title: Configuration Demo REPL
- Prompt: config> 
- Width: 100
- Multiline: Enabled (Ctrl+J to add lines)
- External Editor: Enabled (Ctrl+E to open)
- History: Enabled (Up/Down to navigate, max 50 entries)
- Theme: Dark
- Placeholder: Enter commands or type /help for assistance`

			return repl.EvaluationCompleteMsg{
				Input:  "/config",
				Output: configInfo,
				Error:  nil,
			}
		}
	})

	model.AddCustomCommand("features", func(args []string) tea.Cmd {
		return func() tea.Msg {
			features := `REPL Features Demonstrated:

🎨 Theming:
- Built-in themes: default, dark, light
- Custom theme support
- Runtime theme switching

📝 Input Modes:
- Single-line input (default)
- Multiline input (Ctrl+J)
- External editor (Ctrl+E)

📚 History:
- Command history persistence
- Up/Down arrow navigation
- Configurable history size

⚙️ Configuration:
- Custom prompts and titles
- Adjustable width
- Feature toggles
- Placeholder text

🔧 Extensibility:
- Custom evaluators
- Custom slash commands
- Pluggable architecture`

			return repl.EvaluationCompleteMsg{
				Input:  "/features",
				Output: features,
				Error:  nil,
			}
		}
	})

	model.AddCustomCommand("demo", func(args []string) tea.Cmd {
		return func() tea.Msg {
			demo := `Try these demo commands:

Basic Usage:
- echo Hello, World!
- 5 + 3
- Type any text to see it echoed

Features:
- /config - Show current configuration
- /features - List all features
- /help - Show help information
- /multiline - Toggle multiline mode
- /edit - Open external editor (if $EDITOR is set)
- /clear - Clear screen history
- /quit - Exit the REPL

Navigation:
- Up/Down arrows - Navigate command history
- Ctrl+J - Add line in multiline mode
- Ctrl+E - Open external editor
- Ctrl+C - Exit REPL`

			return repl.EvaluationCompleteMsg{
				Input:  "/demo",
				Output: demo,
				Error:  nil,
			}
		}
	})

	// Start the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
