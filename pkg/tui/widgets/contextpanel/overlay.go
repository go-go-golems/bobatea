package contextpanel

func (w *Widget) ComputeOverlayLayout(width int, height int, headerHeight int, timelineHeight int) (OverlayLayout, bool) {
	if !w.visible || width <= 0 || height <= 0 {
		return OverlayLayout{}, false
	}

	widthPercent := clampInt(w.widthPercent, 20, 90)
	heightPercent := clampInt(w.heightPercent, 20, 90)
	panelWidth := maxInt(32, width*widthPercent/100)
	panelHeight := maxInt(8, height*heightPercent/100)
	panelWidth = minInt(panelWidth, maxInt(20, width-2))
	panelHeight = minInt(panelHeight, maxInt(6, height-2))

	panelStyle := panelStyleDefault()
	frameWidth := panelStyle.GetHorizontalFrameSize()
	frameHeight := panelStyle.GetVerticalFrameSize()
	contentWidth := maxInt(1, panelWidth-frameWidth)
	contentHeight := maxInt(1, panelHeight-frameHeight)

	margin := maxInt(0, w.margin)
	inputY := headerHeight + 1 + timelineHeight

	panelX := 0
	panelY := 0
	switch w.dock {
	case DockRight:
		panelX = width - margin - panelWidth
		panelY = headerHeight + 1 + margin
	case DockLeft:
		panelX = margin
		panelY = headerHeight + 1 + margin
	case DockBottom:
		panelX = (width - panelWidth) / 2
		panelY = height - margin - panelHeight
	case DockAboveRepl:
		fallthrough
	default:
		panelX = (width - panelWidth) / 2
		panelY = inputY - margin - panelHeight
	}
	panelX = clampInt(panelX, 0, maxInt(0, width-panelWidth))
	panelY = clampInt(panelY, 0, maxInt(0, height-panelHeight))

	return OverlayLayout{
		PanelX:        panelX,
		PanelY:        panelY,
		PanelWidth:    panelWidth,
		PanelHeight:   panelHeight,
		ContentWidth:  contentWidth,
		ContentHeight: contentHeight,
	}, true
}

func clampInt(value int, lower int, upper int) int {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
