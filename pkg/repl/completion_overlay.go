package repl

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *Model) computeCompletionOverlayLayout(header, timelineView string) (completionOverlayLayout, bool) {
	if !m.completion.visible || m.width <= 0 || m.height <= 0 {
		return completionOverlayLayout{}, false
	}
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		return completionOverlayLayout{}, false
	}

	inputY := lipgloss.Height(header) + 1 + lipgloss.Height(timelineView)
	popupStyle := m.completionPopupStyle()
	frameWidth := popupStyle.GetHorizontalFrameSize()
	frameHeight := popupStyle.GetVerticalFrameSize()

	contentWidth := 1
	for _, suggestion := range suggestions {
		w := runewidth.StringWidth("  " + suggestion.DisplayText)
		if w > contentWidth {
			contentWidth = w
		}
	}

	popupWidth := contentWidth + frameWidth
	if m.completion.minWidth > 0 {
		popupWidth = max(popupWidth, m.completion.minWidth)
	}
	if m.completion.maxWidth > 0 {
		popupWidth = min(popupWidth, m.completion.maxWidth)
	}
	popupWidth = min(popupWidth, m.width)
	contentWidth = max(1, popupWidth-frameWidth)

	desiredRows := len(suggestions)
	if m.completion.maxVisible > 0 {
		desiredRows = min(desiredRows, m.completion.maxVisible)
	}
	maxHeight := m.completion.maxHeight
	if maxHeight <= 0 {
		maxHeight = m.height
	}
	maxHeight = min(maxHeight, m.height)
	maxRowsByConfig := max(1, maxHeight-frameHeight)
	desiredRows = min(desiredRows, maxRowsByConfig)
	if desiredRows <= 0 {
		return completionOverlayLayout{}, false
	}

	margin := max(0, m.completion.margin)
	availableBelow := max(0, m.height-(inputY+1+margin))
	availableAbove := max(0, inputY-margin)
	belowRows := max(0, min(availableBelow, maxHeight)-frameHeight)
	aboveRows := max(0, min(availableAbove, maxHeight)-frameHeight)

	bottomRows := max(0, min(maxHeight, m.height-margin)-frameHeight)
	if belowRows == 0 && aboveRows == 0 && bottomRows == 0 {
		return completionOverlayLayout{}, false
	}

	visibleRows := desiredRows
	popupY := inputY + 1 + margin
	switch m.completion.placement {
	case CompletionOverlayPlacementAbove:
		visibleRows = min(visibleRows, aboveRows)
		popupY = inputY - margin - (visibleRows + frameHeight)
	case CompletionOverlayPlacementBelow:
		visibleRows = min(visibleRows, belowRows)
		popupY = inputY + 1 + margin
	case CompletionOverlayPlacementBottom:
		visibleRows = min(visibleRows, bottomRows)
		popupY = m.height - margin - (visibleRows + frameHeight)
	case CompletionOverlayPlacementAuto:
		placeBelow := belowRows >= desiredRows || belowRows >= aboveRows
		if placeBelow {
			visibleRows = min(visibleRows, belowRows)
		} else {
			visibleRows = min(visibleRows, aboveRows)
			popupY = inputY - margin - (visibleRows + frameHeight)
		}
	default:
		visibleRows = min(visibleRows, aboveRows)
		popupY = inputY - margin - (visibleRows + frameHeight)
	}
	if visibleRows <= 0 {
		return completionOverlayLayout{}, false
	}

	anchorX := m.completionAnchorColumn()
	popupX := anchorX
	if m.completion.horizontal == CompletionOverlayHorizontalGrowLeft {
		popupX -= popupWidth
	}
	popupX += m.completion.offsetX
	popupY += m.completion.offsetY
	popupX = clampInt(popupX, 0, max(0, m.width-popupWidth))
	popupY = clampInt(popupY, 0, max(0, m.height-1))

	return completionOverlayLayout{
		PopupX:       popupX,
		PopupY:       popupY,
		PopupWidth:   popupWidth,
		VisibleRows:  visibleRows,
		ContentWidth: contentWidth,
	}, true
}

func (m *Model) renderCompletionPopup(layout completionOverlayLayout) string {
	if layout.VisibleRows <= 0 || layout.ContentWidth <= 0 {
		return ""
	}
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		return ""
	}

	start := clampInt(m.completion.scrollTop, 0, max(0, len(suggestions)-1))
	end := min(len(suggestions), start+layout.VisibleRows)
	lines := make([]string, 0, layout.VisibleRows)
	for i := start; i < end; i++ {
		itemText := "  " + suggestions[i].DisplayText
		itemStyle := m.styles.CompletionItem
		if i == m.completion.selection {
			itemStyle = m.styles.CompletionSelected
			itemText = "› " + suggestions[i].DisplayText
		}
		itemText = runewidth.Truncate(itemText, layout.ContentWidth, "")
		if delta := layout.ContentWidth - runewidth.StringWidth(itemText); delta > 0 {
			itemText += strings.Repeat(" ", delta)
		}
		lines = append(lines, itemStyle.Render(itemText))
	}
	return m.completionPopupStyle().Width(layout.PopupWidth).Render(strings.Join(lines, "\n"))
}

func (m *Model) completionAnchorColumn() int {
	runes := []rune(m.textInput.Value())
	cursor := clampInt(m.textInput.Position(), 0, len(runes))
	prefix := string(runes[:cursor])
	return runewidth.StringWidth(m.textInput.Prompt + prefix)
}

func (m *Model) completionPopupStyle() lipgloss.Style {
	if !m.completion.noBorder {
		return m.styles.CompletionPopup
	}
	return m.styles.CompletionPopup.
		Border(lipgloss.HiddenBorder(), false, false, false, false).
		Padding(0, 0)
}
