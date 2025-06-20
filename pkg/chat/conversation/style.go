package conversation

import "github.com/charmbracelet/lipgloss"

type Style struct {
	UnselectedMessage lipgloss.Style
	SelectedMessage   lipgloss.Style
	FocusedMessage    lipgloss.Style
	MetadataStyle     lipgloss.Style
	ErrorMessage      lipgloss.Style
	ErrorSelected     lipgloss.Style
}

type BorderColors struct {
	Unselected string
	Selected   string
	Focused    string
}

func DefaultStyles() *Style {
	lightModeColors := BorderColors{
		Unselected: "#CCCCCC",
		Selected:   "#FFB6C1", // Light pink
		Focused:    "#FFFF99", // Light yellow
	}

	darkModeColors := BorderColors{
		Unselected: "#444444",
		Selected:   "#DD7090", // Desaturated pink for dark mode
		Focused:    "#DDDD77", // Desaturated yellow for dark mode
	}

	errorColors := BorderColors{
		Unselected: "#FF6B6B", // Light red
		Selected:   "#FF4444", // Darker red for selected
		Focused:    "#FF8888", // Medium red for focused
	}

	return &Style{
		UnselectedMessage: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: lightModeColors.Unselected,
				Dark:  darkModeColors.Unselected,
			}),
		SelectedMessage: lipgloss.NewStyle().Border(lipgloss.ThickBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: lightModeColors.Selected,
				Dark:  darkModeColors.Selected,
			}),
		FocusedMessage: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: lightModeColors.Focused,
				Dark:  darkModeColors.Focused,
			}),
		MetadataStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Right),
		ErrorMessage: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.Color(errorColors.Unselected)).
			Foreground(lipgloss.Color(errorColors.Unselected)),
		ErrorSelected: lipgloss.NewStyle().Border(lipgloss.ThickBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.Color(errorColors.Selected)).
			Foreground(lipgloss.Color(errorColors.Selected)),
	}
}
