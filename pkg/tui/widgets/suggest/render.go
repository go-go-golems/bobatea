package suggest

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (w *Widget) PopupStyle(baseStyle lipgloss.Style) lipgloss.Style {
	if !w.noBorder {
		return baseStyle
	}
	return baseStyle.
		Border(lipgloss.HiddenBorder(), false, false, false, false).
		Padding(0, 0)
}

func (w *Widget) RenderPopup(styles Styles, layout OverlayLayout) string {
	if layout.VisibleRows <= 0 || layout.ContentWidth <= 0 {
		return ""
	}
	suggestions := w.lastResult.Suggestions
	if len(suggestions) == 0 {
		return ""
	}

	start := clampInt(w.scrollTop, 0, maxInt(0, len(suggestions)-1))
	end := minInt(len(suggestions), start+layout.VisibleRows)
	lines := make([]string, 0, layout.VisibleRows)
	for i := start; i < end; i++ {
		itemText := "  " + suggestions[i].DisplayText
		itemStyle := styles.Item
		if i == w.selection {
			itemStyle = styles.Selected
			itemText = "› " + suggestions[i].DisplayText
		}
		itemText = runewidth.Truncate(itemText, layout.ContentWidth, "")
		if delta := layout.ContentWidth - runewidth.StringWidth(itemText); delta > 0 {
			itemText += strings.Repeat(" ", delta)
		}
		lines = append(lines, itemStyle.Render(itemText))
	}
	return w.PopupStyle(styles.Popup).Width(layout.PopupWidth).Render(strings.Join(lines, "\n"))
}
