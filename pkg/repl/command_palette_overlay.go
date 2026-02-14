package repl

import "github.com/charmbracelet/lipgloss"

type commandPaletteOverlayLayout struct {
	PanelX int
	PanelY int
	View   string
}

func (m *Model) computeCommandPaletteOverlayLayout() (commandPaletteOverlayLayout, bool) {
	if !m.palette.enabled || !m.palette.ui.IsVisible() || m.width <= 0 || m.height <= 0 {
		return commandPaletteOverlayLayout{}, false
	}

	paletteView := m.palette.ui.View()
	if paletteView == "" {
		return commandPaletteOverlayLayout{}, false
	}

	panelWidth := lipgloss.Width(paletteView)
	panelHeight := lipgloss.Height(paletteView)
	if panelWidth <= 0 || panelHeight <= 0 {
		return commandPaletteOverlayLayout{}, false
	}

	panelX := (m.width - panelWidth) / 2
	panelY := (m.height - panelHeight) / 2
	panelX = clampInt(panelX, 0, max(0, m.width-panelWidth))
	panelY = clampInt(panelY, 0, max(0, m.height-panelHeight))

	return commandPaletteOverlayLayout{
		PanelX: panelX,
		PanelY: panelY,
		View:   paletteView,
	}, true
}
