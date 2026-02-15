package repl

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/tui/widgets/contextpanel"
)

func (m *Model) computeHelpDrawerOverlayLayout(header, timelineView string) (helpDrawerOverlayLayout, bool) {
	headerHeight := lipgloss.Height(header)
	timelineHeight := lipgloss.Height(timelineView)
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil {
		return helpDrawerOverlayLayout{}, false
	}
	return m.helpDrawer.widget.ComputeOverlayLayout(m.width, m.height, headerHeight, timelineHeight)
}

func (m *Model) renderHelpDrawerPanel(layout helpDrawerOverlayLayout) string {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil || layout.ContentWidth <= 0 || layout.ContentHeight <= 0 {
		return ""
	}
	toggleKey := bindingPrimaryKey(m.keyMap.HelpDrawerToggle, "alt+h")
	refreshKey := bindingPrimaryKey(m.keyMap.HelpDrawerRefresh, "ctrl+r")
	pinKey := bindingPrimaryKey(m.keyMap.HelpDrawerPin, "ctrl+g")
	return m.helpDrawer.widget.RenderPanel(layout, contextpanel.RenderOptions{
		ToggleBinding:  toggleKey,
		RefreshBinding: refreshKey,
		PinBinding:     pinKey,
		FooterRenderer: func(s string) string {
			return m.styles.HelpText.Render(s)
		},
	})
}

func bindingPrimaryKey(b key.Binding, fallback string) string {
	if !b.Enabled() {
		return fallback
	}
	keyName := strings.TrimSpace(b.Help().Key)
	if keyName == "" {
		return fallback
	}
	return keyName
}
