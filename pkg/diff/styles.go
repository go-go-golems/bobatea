package diff

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles contains style definitions used by the diff component.
type Styles struct {
	Title            lipgloss.Style
	ListBase         lipgloss.Style
	ListFocused      lipgloss.Style
	DetailBase       lipgloss.Style
	DetailFocused    lipgloss.Style
	CategoryHeader   lipgloss.Style
	Path             lipgloss.Style
	RemovedLine      lipgloss.Style
	AddedLine        lipgloss.Style
	UpdatedLine      lipgloss.Style
	SensitiveValue   lipgloss.Style
}

func defaultStyles() Styles {
	focusedBorderColor := lipgloss.AdaptiveColor{Light: "#3B82F6", Dark: "#60A5FA"}
	blurredBorderColor := lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#1F2937"}

	return Styles{
		Title: lipgloss.NewStyle().Bold(true).Padding(0, 1),
		ListBase: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blurredBorderColor).
			Padding(0, 1).
			MarginRight(1),

		ListFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(focusedBorderColor).
			Padding(0, 1).
			MarginRight(1),

		DetailBase: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blurredBorderColor).
			Padding(0, 1),

		DetailFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(focusedBorderColor).
			Padding(0, 1),

		CategoryHeader: lipgloss.NewStyle().Bold(true),
 		Path:           lipgloss.NewStyle().Faint(true),
 		RemovedLine:    lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")),
 		AddedLine:      lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")),
 		UpdatedLine:    lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")),
 		SensitiveValue: lipgloss.NewStyle().Faint(true),
 	}
}


