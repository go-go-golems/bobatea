package commandpalette

import "github.com/charmbracelet/lipgloss"

// Styles defines the styling for the command palette
type Styles struct {
	Palette            lipgloss.Style
	Header             lipgloss.Style
	Query              lipgloss.Style
	Command            lipgloss.Style
	SelectedCommand    lipgloss.Style
	CommandName        lipgloss.Style
	CommandDescription lipgloss.Style
	Help               lipgloss.Style
}

// DefaultStyles returns the default styles for the command palette
func DefaultStyles() Styles {
	return Styles{
		Palette: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Background(lipgloss.Color("235")).
			Padding(1).
			Margin(2, 4),

		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			Margin(0, 0, 1, 0),

		Query: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("240")).
			Padding(0, 1).
			Margin(0, 0, 1, 0),

		Command: lipgloss.NewStyle().
			Padding(0, 1),

		SelectedCommand: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1),

		CommandName: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),

		CommandDescription: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
	}
}
