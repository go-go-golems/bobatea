package repl

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func normalizeHelpBarConfig(cfg HelpBarConfig) HelpBarConfig {
	if cfg.Debounce == 0 && cfg.RequestTimeout == 0 && !cfg.Enabled {
		return DefaultHelpBarConfig()
	}
	merged := DefaultHelpBarConfig()
	merged.Enabled = cfg.Enabled
	if cfg.Debounce > 0 {
		merged.Debounce = cfg.Debounce
	}
	if cfg.RequestTimeout > 0 {
		merged.RequestTimeout = cfg.RequestTimeout
	}
	return merged
}

func normalizeHelpDrawerConfig(cfg HelpDrawerConfig) HelpDrawerConfig {
	if cfg.Debounce == 0 &&
		cfg.RequestTimeout == 0 &&
		len(cfg.ToggleKeys) == 0 &&
		len(cfg.CloseKeys) == 0 &&
		len(cfg.RefreshShortcuts) == 0 &&
		len(cfg.PinShortcuts) == 0 &&
		cfg.Dock == "" &&
		cfg.WidthPercent == 0 &&
		cfg.HeightPercent == 0 &&
		cfg.Margin == 0 &&
		!cfg.PrefetchWhenHidden &&
		!cfg.Enabled {
		return DefaultHelpDrawerConfig()
	}

	merged := DefaultHelpDrawerConfig()
	merged.Enabled = cfg.Enabled
	if len(cfg.ToggleKeys) > 0 {
		merged.ToggleKeys = cfg.ToggleKeys
	}
	if len(cfg.CloseKeys) > 0 {
		merged.CloseKeys = cfg.CloseKeys
	}
	if len(cfg.RefreshShortcuts) > 0 {
		merged.RefreshShortcuts = cfg.RefreshShortcuts
	}
	if len(cfg.PinShortcuts) > 0 {
		merged.PinShortcuts = cfg.PinShortcuts
	}
	if cfg.Debounce > 0 {
		merged.Debounce = cfg.Debounce
	}
	if cfg.RequestTimeout > 0 {
		merged.RequestTimeout = cfg.RequestTimeout
	}
	if cfg.Dock != "" {
		merged.Dock = cfg.Dock
	}
	if cfg.WidthPercent > 0 {
		merged.WidthPercent = cfg.WidthPercent
	}
	if cfg.HeightPercent > 0 {
		merged.HeightPercent = cfg.HeightPercent
	}
	if cfg.Margin > 0 {
		merged.Margin = cfg.Margin
	}
	if cfg.PrefetchWhenHidden {
		merged.PrefetchWhenHidden = true
	}

	merged.Dock = normalizeHelpDrawerDock(merged.Dock)
	merged.WidthPercent = clampInt(merged.WidthPercent, 20, 90)
	merged.HeightPercent = clampInt(merged.HeightPercent, 20, 90)
	merged.Margin = max(0, merged.Margin)
	return merged
}

func normalizeHelpDrawerDock(v HelpDrawerDock) HelpDrawerDock {
	switch v {
	case HelpDrawerDockAboveRepl, HelpDrawerDockRight, HelpDrawerDockLeft, HelpDrawerDockBottom:
		return v
	default:
		return HelpDrawerDockAboveRepl
	}
}

func normalizeAutocompleteConfig(cfg AutocompleteConfig) AutocompleteConfig {
	if cfg.Debounce == 0 &&
		cfg.RequestTimeout == 0 &&
		len(cfg.TriggerKeys) == 0 &&
		len(cfg.AcceptKeys) == 0 &&
		cfg.FocusToggleKey == "" &&
		cfg.MaxSuggestions == 0 &&
		cfg.OverlayMaxWidth == 0 &&
		cfg.OverlayMaxHeight == 0 &&
		cfg.OverlayMinWidth == 0 &&
		cfg.OverlayMargin == 0 &&
		cfg.OverlayPageSize == 0 &&
		cfg.OverlayOffsetX == 0 &&
		cfg.OverlayOffsetY == 0 &&
		!cfg.OverlayNoBorder &&
		cfg.OverlayPlacement == "" &&
		cfg.OverlayHorizontalGrow == "" &&
		!cfg.Enabled {
		return DefaultAutocompleteConfig()
	}

	merged := DefaultAutocompleteConfig()
	merged.Enabled = cfg.Enabled
	if cfg.Debounce > 0 {
		merged.Debounce = cfg.Debounce
	}
	if cfg.RequestTimeout > 0 {
		merged.RequestTimeout = cfg.RequestTimeout
	}
	if len(cfg.TriggerKeys) > 0 {
		merged.TriggerKeys = cfg.TriggerKeys
	}
	if len(cfg.AcceptKeys) > 0 {
		merged.AcceptKeys = cfg.AcceptKeys
	}
	if cfg.FocusToggleKey != "" {
		merged.FocusToggleKey = cfg.FocusToggleKey
	}
	if cfg.MaxSuggestions > 0 {
		merged.MaxSuggestions = cfg.MaxSuggestions
	}
	if cfg.OverlayMaxWidth > 0 {
		merged.OverlayMaxWidth = cfg.OverlayMaxWidth
	}
	if cfg.OverlayMaxHeight > 0 {
		merged.OverlayMaxHeight = cfg.OverlayMaxHeight
	}
	if cfg.OverlayMinWidth > 0 {
		merged.OverlayMinWidth = cfg.OverlayMinWidth
	}
	if cfg.OverlayMargin > 0 {
		merged.OverlayMargin = cfg.OverlayMargin
	}
	if cfg.OverlayPageSize > 0 {
		merged.OverlayPageSize = cfg.OverlayPageSize
	}
	if cfg.OverlayOffsetX != 0 {
		merged.OverlayOffsetX = cfg.OverlayOffsetX
	}
	if cfg.OverlayOffsetY != 0 {
		merged.OverlayOffsetY = cfg.OverlayOffsetY
	}
	if cfg.OverlayNoBorder {
		merged.OverlayNoBorder = true
	}
	if cfg.OverlayPlacement != "" {
		merged.OverlayPlacement = cfg.OverlayPlacement
	}
	if cfg.OverlayHorizontalGrow != "" {
		merged.OverlayHorizontalGrow = cfg.OverlayHorizontalGrow
	}
	merged.OverlayPlacement = normalizeOverlayPlacement(merged.OverlayPlacement)
	merged.OverlayHorizontalGrow = normalizeOverlayHorizontalGrow(merged.OverlayHorizontalGrow)
	return merged
}

func normalizeOverlayPlacement(v CompletionOverlayPlacement) CompletionOverlayPlacement {
	switch v {
	case CompletionOverlayPlacementAuto,
		CompletionOverlayPlacementAbove,
		CompletionOverlayPlacementBelow,
		CompletionOverlayPlacementBottom:
		return v
	default:
		return CompletionOverlayPlacementAuto
	}
}

func normalizeOverlayHorizontalGrow(v CompletionOverlayHorizontalGrow) CompletionOverlayHorizontalGrow {
	switch v {
	case CompletionOverlayHorizontalGrowRight, CompletionOverlayHorizontalGrowLeft:
		return v
	default:
		return CompletionOverlayHorizontalGrowRight
	}
}
