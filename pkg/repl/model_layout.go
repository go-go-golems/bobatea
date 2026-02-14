package repl

import "github.com/charmbracelet/lipgloss"

// applyLayoutFromState syncs dependent widget sizes and returns whether the
// timeline viewport dimensions changed.
func (m *Model) applyLayoutFromState() bool {
	if m.width <= 0 {
		return false
	}

	m.help.Width = max(0, m.width)
	m.textInput.Width = max(10, m.width-10)
	m.palette.ui.SetSize(m.width, max(0, m.height))

	if m.height <= 0 {
		return false
	}

	helpHeight := lipgloss.Height(m.renderHelp())
	helpBarHeight := 0
	if helpBarView := m.renderHelpBar(); helpBarView != "" {
		helpBarHeight = lipgloss.Height(helpBarView)
	}
	timelineHeight := max(0, m.height-helpHeight-helpBarHeight-4)
	changed := m.timelineWidth != m.width || m.timelineHeight != timelineHeight
	if !changed {
		return false
	}

	m.sh.SetSize(m.width, timelineHeight)
	m.timelineWidth = m.width
	m.timelineHeight = timelineHeight
	return true
}

func (m *Model) applyLayoutAndRefresh() {
	if m.applyLayoutFromState() {
		m.sh.RefreshView(false)
	}
}
