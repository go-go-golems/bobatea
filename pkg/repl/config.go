package repl

import "time"

type CompletionOverlayPlacement string

const (
	CompletionOverlayPlacementAuto   CompletionOverlayPlacement = "auto"
	CompletionOverlayPlacementAbove  CompletionOverlayPlacement = "above"
	CompletionOverlayPlacementBelow  CompletionOverlayPlacement = "below"
	CompletionOverlayPlacementBottom CompletionOverlayPlacement = "bottom"
)

type CompletionOverlayHorizontalGrow string

const (
	CompletionOverlayHorizontalGrowRight CompletionOverlayHorizontalGrow = "right"
	CompletionOverlayHorizontalGrowLeft  CompletionOverlayHorizontalGrow = "left"
)

// AutocompleteConfig controls autocomplete request and key-routing behavior.
type AutocompleteConfig struct {
	// Enabled toggles REPL autocomplete integration.
	Enabled bool
	// Debounce is the delay after input edits before a completion request is sent.
	Debounce time.Duration
	// RequestTimeout bounds a single completion request.
	RequestTimeout time.Duration
	// TriggerKeys are optional explicit shortcut keys (for example "tab").
	TriggerKeys []string
	// AcceptKeys apply the selected completion while the popup is visible.
	AcceptKeys []string
	// FocusToggleKey switches between input and timeline focus.
	FocusToggleKey string
	// MaxSuggestions limits the number of rendered popup items.
	MaxSuggestions int
	// OverlayMaxWidth limits popup width in terminal cells.
	OverlayMaxWidth int
	// OverlayMaxHeight limits popup height in terminal rows.
	OverlayMaxHeight int
	// OverlayMinWidth keeps the popup readable when suggestions are short.
	OverlayMinWidth int
	// OverlayMargin leaves a gap between input anchor and popup.
	OverlayMargin int
	// OverlayPageSize controls page up/down steps; 0 means visible rows.
	OverlayPageSize int
	// OverlayOffsetX shifts popup horizontally before clamping to terminal bounds.
	OverlayOffsetX int
	// OverlayOffsetY shifts popup vertically before clamping to terminal bounds.
	OverlayOffsetY int
	// OverlayNoBorder renders completion popup without border chrome.
	OverlayNoBorder bool
	// OverlayPlacement controls vertical popup placement strategy.
	// Supported values: auto, above, below, bottom.
	OverlayPlacement CompletionOverlayPlacement
	// OverlayHorizontalGrow controls horizontal growth direction from anchor.
	// Supported values: right, left.
	OverlayHorizontalGrow CompletionOverlayHorizontalGrow
}

// DefaultAutocompleteConfig returns default autocomplete settings.
func DefaultAutocompleteConfig() AutocompleteConfig {
	return AutocompleteConfig{
		Enabled:               true,
		Debounce:              120 * time.Millisecond,
		RequestTimeout:        400 * time.Millisecond,
		TriggerKeys:           []string{"tab"},
		AcceptKeys:            []string{"enter", "tab"},
		FocusToggleKey:        "",
		MaxSuggestions:        8,
		OverlayMaxWidth:       56,
		OverlayMaxHeight:      12,
		OverlayMinWidth:       24,
		OverlayMargin:         1,
		OverlayPageSize:       0,
		OverlayOffsetX:        0,
		OverlayOffsetY:        0,
		OverlayNoBorder:       false,
		OverlayPlacement:      CompletionOverlayPlacementAuto,
		OverlayHorizontalGrow: CompletionOverlayHorizontalGrowRight,
	}
}

// Config holds REPL shell configuration.
type Config struct {
	Title                string
	Prompt               string
	Placeholder          string
	Width                int
	StartMultiline       bool
	EnableExternalEditor bool
	EnableHistory        bool
	MaxHistorySize       int
	// Autocomplete controls completion scheduling, shortcuts, and popup behavior.
	Autocomplete AutocompleteConfig
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Title:                "REPL",
		Prompt:               "> ",
		Placeholder:          "Enter code or /command",
		Width:                80,
		StartMultiline:       false,
		EnableExternalEditor: false,
		EnableHistory:        true,
		MaxHistorySize:       1000,
		Autocomplete:         DefaultAutocompleteConfig(),
	}
}
