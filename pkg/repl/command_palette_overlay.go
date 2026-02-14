package repl

import "github.com/charmbracelet/lipgloss"

func (m *Model) renderCommandPaletteOverlay() string {
	if !m.palette.enabled || !m.palette.ui.IsVisible() || m.width <= 0 || m.height <= 0 {
		return ""
	}

	paletteView := m.palette.ui.View()
	if paletteView == "" {
		return ""
	}

	return lipgloss.Place(
		max(1, m.width),
		max(1, m.height),
		lipgloss.Center,
		lipgloss.Center,
		paletteView,
		lipgloss.WithWhitespaceChars(" "),
	)
}
