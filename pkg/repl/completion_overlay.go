package repl

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/tui/widgets/suggest"
)

func (m *Model) computeCompletionOverlayLayout(header, timelineView string) (completionOverlayLayout, bool) {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return completionOverlayLayout{}, false
	}
	m.syncCompletionWidgetFromLegacy()
	layout, ok := m.completion.widget.ComputeOverlayLayout(
		m.width,
		m.height,
		lipgloss.Height(header),
		lipgloss.Height(timelineView),
		m.textInput.Prompt,
		m.textInput.Value(),
		m.textInput.Position(),
		m.completionPopupStyle(),
	)
	m.syncCompletionLegacyFromWidget()
	return layout, ok
}

func (m *Model) renderCompletionPopup(layout completionOverlayLayout) string {
	m.ensureCompletionWidget()
	if m.completion.widget == nil || layout.VisibleRows <= 0 || layout.ContentWidth <= 0 {
		return ""
	}
	m.syncCompletionWidgetFromLegacy()
	ret := m.completion.widget.RenderPopup(suggest.Styles{
		Item:     m.styles.CompletionItem,
		Selected: m.styles.CompletionSelected,
		Popup:    m.styles.CompletionPopup,
	}, layout)
	m.syncCompletionLegacyFromWidget()
	return ret
}

func (m *Model) completionPopupStyle() lipgloss.Style {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return m.styles.CompletionPopup
	}
	m.syncCompletionWidgetFromLegacy()
	ret := m.completion.widget.PopupStyle(m.styles.CompletionPopup)
	m.syncCompletionLegacyFromWidget()
	return ret
}
