package repl

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandPaletteOverlayLayoutUsesBoundedPanel(t *testing.T) {
	m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, nil)
	m.width = 120
	m.height = 40
	m.palette.ui.SetSize(m.width, m.height)
	m.openCommandPalette()

	layout, ok := m.computeCommandPaletteOverlayLayout()
	require.True(t, ok)
	require.NotEmpty(t, layout.View)

	panelWidth := lipgloss.Width(layout.View)
	panelHeight := lipgloss.Height(layout.View)
	assert.Greater(t, panelWidth, 0)
	assert.Greater(t, panelHeight, 0)
	assert.Less(t, panelHeight, m.height, "palette overlay should not consume full terminal height")
	assert.GreaterOrEqual(t, layout.PanelX, 0)
	assert.GreaterOrEqual(t, layout.PanelY, 0)
	assert.Greater(t, layout.PanelY, 0, "bounded overlay should be positioned within the canvas, not drawn as a full-screen layer")
}

func TestCommandPaletteOverlayLayoutHiddenWhenPaletteClosed(t *testing.T) {
	m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, nil)
	m.width = 120
	m.height = 40
	m.palette.ui.SetSize(m.width, m.height)
	m.palette.ui.Hide()

	layout, ok := m.computeCommandPaletteOverlayLayout()
	assert.False(t, ok)
	assert.Empty(t, layout.View)
}
