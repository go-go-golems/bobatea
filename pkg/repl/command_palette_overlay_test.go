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

func TestCommandPaletteOverlayPlacementOptions(t *testing.T) {
	testCases := []struct {
		name      string
		placement CommandPaletteOverlayPlacement
		check     func(t *testing.T, m *Model, layout commandPaletteOverlayLayout, panelWidth, panelHeight int)
	}{
		{
			name:      "top",
			placement: CommandPaletteOverlayPlacementTop,
			check: func(t *testing.T, m *Model, layout commandPaletteOverlayLayout, _, panelHeight int) {
				want := clampInt(m.palette.overlayMargin, 0, max(0, m.height-panelHeight))
				assert.Equal(t, want, layout.PanelY)
			},
		},
		{
			name:      "bottom",
			placement: CommandPaletteOverlayPlacementBottom,
			check: func(t *testing.T, m *Model, layout commandPaletteOverlayLayout, _, panelHeight int) {
				want := clampInt(m.height-m.palette.overlayMargin-panelHeight, 0, max(0, m.height-panelHeight))
				assert.Equal(t, want, layout.PanelY)
			},
		},
		{
			name:      "left",
			placement: CommandPaletteOverlayPlacementLeft,
			check: func(t *testing.T, m *Model, layout commandPaletteOverlayLayout, panelWidth, _ int) {
				want := clampInt(m.palette.overlayMargin, 0, max(0, m.width-panelWidth))
				assert.Equal(t, want, layout.PanelX)
			},
		},
		{
			name:      "right",
			placement: CommandPaletteOverlayPlacementRight,
			check: func(t *testing.T, m *Model, layout commandPaletteOverlayLayout, panelWidth, _ int) {
				want := clampInt(m.width-m.palette.overlayMargin-panelWidth, 0, max(0, m.width-panelWidth))
				assert.Equal(t, want, layout.PanelX)
			},
		},
		{
			name:      "center",
			placement: CommandPaletteOverlayPlacementCenter,
			check: func(t *testing.T, m *Model, layout commandPaletteOverlayLayout, panelWidth, panelHeight int) {
				wantX := clampInt((m.width-panelWidth)/2, 0, max(0, m.width-panelWidth))
				wantY := clampInt((m.height-panelHeight)/2, 0, max(0, m.height-panelHeight))
				assert.Equal(t, wantX, layout.PanelX)
				assert.Equal(t, wantY, layout.PanelY)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, func(cfg *Config) {
				cfg.CommandPalette.OverlayPlacement = tc.placement
				cfg.CommandPalette.OverlayMargin = 2
			})
			m.width = 220
			m.height = 70
			m.palette.ui.SetSize(m.width, m.height)
			m.openCommandPalette()

			layout, ok := m.computeCommandPaletteOverlayLayout()
			require.True(t, ok)
			panelWidth := lipgloss.Width(layout.View)
			panelHeight := lipgloss.Height(layout.View)
			require.Greater(t, panelWidth, 0)
			require.Greater(t, panelHeight, 0)
			tc.check(t, m, layout, panelWidth, panelHeight)
		})
	}
}
