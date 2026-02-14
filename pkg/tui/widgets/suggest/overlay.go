package suggest

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (w *Widget) ComputeOverlayLayout(width int, height int, headerHeight int, timelineHeight int, prompt string, input string, cursor int, popupStyle lipgloss.Style) (OverlayLayout, bool) {
	if !w.visible || width <= 0 || height <= 0 {
		return OverlayLayout{}, false
	}
	suggestions := w.lastResult.Suggestions
	if len(suggestions) == 0 {
		return OverlayLayout{}, false
	}

	inputY := headerHeight + 1 + timelineHeight
	frameWidth := popupStyle.GetHorizontalFrameSize()
	frameHeight := popupStyle.GetVerticalFrameSize()

	contentWidth := 1
	for _, suggestion := range suggestions {
		itemWidth := runewidth.StringWidth("  " + suggestion.DisplayText)
		if itemWidth > contentWidth {
			contentWidth = itemWidth
		}
	}

	popupWidth := contentWidth + frameWidth
	if w.minWidth > 0 {
		popupWidth = maxInt(popupWidth, w.minWidth)
	}
	if w.maxWidth > 0 {
		popupWidth = minInt(popupWidth, w.maxWidth)
	}
	popupWidth = minInt(popupWidth, width)
	contentWidth = maxInt(1, popupWidth-frameWidth)

	desiredRows := len(suggestions)
	if w.maxVisible > 0 {
		desiredRows = minInt(desiredRows, w.maxVisible)
	}
	maxHeight := w.maxHeight
	if maxHeight <= 0 {
		maxHeight = height
	}
	maxHeight = minInt(maxHeight, height)
	maxRowsByConfig := maxInt(1, maxHeight-frameHeight)
	desiredRows = minInt(desiredRows, maxRowsByConfig)
	if desiredRows <= 0 {
		return OverlayLayout{}, false
	}

	margin := maxInt(0, w.margin)
	availableBelow := maxInt(0, height-(inputY+1+margin))
	availableAbove := maxInt(0, inputY-margin)
	belowRows := maxInt(0, minInt(availableBelow, maxHeight)-frameHeight)
	aboveRows := maxInt(0, minInt(availableAbove, maxHeight)-frameHeight)
	bottomRows := maxInt(0, minInt(maxHeight, height-margin)-frameHeight)
	if belowRows == 0 && aboveRows == 0 && bottomRows == 0 {
		return OverlayLayout{}, false
	}

	visibleRows := desiredRows
	popupY := inputY + 1 + margin
	switch w.placement {
	case PlacementAbove:
		visibleRows = minInt(visibleRows, aboveRows)
		popupY = inputY - margin - (visibleRows + frameHeight)
	case PlacementBelow:
		visibleRows = minInt(visibleRows, belowRows)
		popupY = inputY + 1 + margin
	case PlacementBottom:
		visibleRows = minInt(visibleRows, bottomRows)
		popupY = height - margin - (visibleRows + frameHeight)
	case PlacementAuto:
		placeBelow := belowRows >= desiredRows || belowRows >= aboveRows
		if placeBelow {
			visibleRows = minInt(visibleRows, belowRows)
		} else {
			visibleRows = minInt(visibleRows, aboveRows)
			popupY = inputY - margin - (visibleRows + frameHeight)
		}
	default:
		visibleRows = minInt(visibleRows, aboveRows)
		popupY = inputY - margin - (visibleRows + frameHeight)
	}
	if visibleRows <= 0 {
		return OverlayLayout{}, false
	}

	anchorX := completionAnchorColumn(prompt, input, cursor)
	popupX := anchorX
	if w.horizontal == HorizontalGrowLeft {
		popupX -= popupWidth
	}
	popupX += w.offsetX
	popupY += w.offsetY
	popupX = clampInt(popupX, 0, maxInt(0, width-popupWidth))
	popupY = clampInt(popupY, 0, maxInt(0, height-1))

	return OverlayLayout{
		PopupX:       popupX,
		PopupY:       popupY,
		PopupWidth:   popupWidth,
		VisibleRows:  visibleRows,
		ContentWidth: contentWidth,
	}, true
}

func completionAnchorColumn(prompt string, value string, cursor int) int {
	runes := []rune(value)
	cursor = clampInt(cursor, 0, len(runes))
	prefix := string(runes[:cursor])
	return runewidth.StringWidth(prompt + prefix)
}
