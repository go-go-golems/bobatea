package repl

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles defines the visual styling for the REPL
type Styles struct {
	Title    lipgloss.Style
	Prompt   lipgloss.Style
	Result   lipgloss.Style
	Error    lipgloss.Style
	Info     lipgloss.Style
	HelpText lipgloss.Style
}

// DefaultStyles returns the default styling configuration
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("32")).
			Background(lipgloss.Color("240")).
			Padding(0, 1),

		Prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true),

		Result: lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true),

		HelpText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
	}
}

// Theme represents a color theme for the REPL
type Theme struct {
	Name   string
	Styles Styles
}

// BuiltinThemes provides predefined themes
var BuiltinThemes = map[string]Theme{
	"default": {
		Name:   "Default",
		Styles: DefaultStyles(),
	},
	"dark": {
		Name: "Dark",
		Styles: Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("236")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true),
		},
	},
	"light": {
		Name: "Light",
		Styles: Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("252")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("4")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("130")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Italic(true),
		},
	},
}
