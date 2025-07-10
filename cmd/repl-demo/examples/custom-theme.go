package main

import (
	"log"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// CustomThemeExample demonstrates creating and using custom themes
func main() {
	// Create evaluator
	evaluator := repl.NewExampleEvaluator()

	// Create custom theme
	customTheme := repl.Theme{
		Name: "Cyberpunk",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("201")). // Bright magenta
				Background(lipgloss.Color("57")).  // Dark purple
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")). // Bright green
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("51")), // Bright cyan

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")). // Bright red
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")). // Bright yellow
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")). // Dark gray
				Italic(true),
		},
	}

	// Create another custom theme - minimalist
	minimalistTheme := repl.Theme{
		Name: "Minimalist",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")). // White
				Background(lipgloss.Color("0")).  // Black
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")). // Light gray
				Bold(false),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")), // White

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")). // Light red
				Bold(false),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")). // Dark gray
				Italic(false),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")). // Dark gray
				Italic(false),
		},
	}

	// Create configuration
	config := repl.DefaultConfig()
	config.Title = "Custom Theme Example"

	// Create model
	model := repl.NewModel(evaluator, config)

	// Start with cyberpunk theme
	model.SetTheme(customTheme)

	// Add custom commands to switch themes
	model.AddCustomCommand("theme", func(args []string) tea.Cmd {
		if len(args) == 0 {
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/theme",
					Output: "Available themes: cyberpunk, minimalist, default, dark, light",
					Error:  nil,
				}
			}
		}

		themeName := args[0]
		switch themeName {
		case "cyberpunk":
			model.SetTheme(customTheme)
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/theme cyberpunk",
					Output: "Switched to Cyberpunk theme",
					Error:  nil,
				}
			}
		case "minimalist":
			model.SetTheme(minimalistTheme)
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/theme minimalist",
					Output: "Switched to Minimalist theme",
					Error:  nil,
				}
			}
		default:
			// Check built-in themes
			if theme, ok := repl.BuiltinThemes[themeName]; ok {
				model.SetTheme(theme)
				return func() tea.Msg {
					return repl.EvaluationCompleteMsg{
						Input:  "/theme " + themeName,
						Output: "Switched to " + theme.Name + " theme",
						Error:  nil,
					}
				}
			}
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/theme " + themeName,
					Output: "Unknown theme: " + themeName,
					Error:  nil,
				}
			}
		}
	})

	// Add info command
	model.AddCustomCommand("info", func(args []string) tea.Cmd {
		return func() tea.Msg {
			return repl.EvaluationCompleteMsg{
				Input:  "/info",
				Output: "Custom Theme Example\n\nThis example demonstrates:\n- Creating custom themes\n- Switching between themes at runtime\n- Combining built-in and custom themes\n\nTry:\n- /theme cyberpunk\n- /theme minimalist\n- /theme dark\n- /theme light",
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
