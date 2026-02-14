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

// HelpBarConfig controls contextual help bar request and rendering behavior.
type HelpBarConfig struct {
	// Enabled toggles REPL help bar integration.
	Enabled bool
	// Debounce is the delay after input edits before a help bar request is sent.
	Debounce time.Duration
	// RequestTimeout bounds a single help bar request.
	RequestTimeout time.Duration
}

type HelpDrawerDock string

const (
	HelpDrawerDockAboveRepl HelpDrawerDock = "above-repl"
	HelpDrawerDockRight     HelpDrawerDock = "right"
	HelpDrawerDockLeft      HelpDrawerDock = "left"
	HelpDrawerDockBottom    HelpDrawerDock = "bottom"
)

// HelpDrawerConfig controls keyboard-toggled contextual help drawer behavior.
type HelpDrawerConfig struct {
	// Enabled toggles REPL help drawer integration.
	Enabled bool
	// ToggleKeys open/close the drawer while typing.
	ToggleKeys []string
	// CloseKeys close the drawer while it is visible.
	CloseKeys []string
	// RefreshShortcuts trigger an immediate drawer refresh while visible.
	RefreshShortcuts []string
	// PinShortcuts toggle pin mode (freeze typing-triggered refresh).
	PinShortcuts []string
	// Debounce is the delay after input edits before a drawer refresh is sent.
	Debounce time.Duration
	// RequestTimeout bounds a single drawer request.
	RequestTimeout time.Duration
	// Dock controls where the drawer is anchored.
	// Supported values: above-repl, right, left, bottom.
	Dock HelpDrawerDock
	// WidthPercent controls drawer width as percentage of terminal width.
	WidthPercent int
	// HeightPercent controls drawer height as percentage of terminal height.
	HeightPercent int
	// Margin keeps space between the drawer and the terminal edge/anchor.
	Margin int
	// PrefetchWhenHidden keeps requests running while drawer is hidden.
	PrefetchWhenHidden bool
}

type CommandPaletteSlashPolicy string

const (
	CommandPaletteSlashPolicyEmptyInput CommandPaletteSlashPolicy = "empty-input"
	CommandPaletteSlashPolicyColumnZero CommandPaletteSlashPolicy = "column-zero"
	CommandPaletteSlashPolicyProvider   CommandPaletteSlashPolicy = "provider"
)

// CommandPaletteConfig controls REPL command palette behavior.
type CommandPaletteConfig struct {
	// Enabled toggles command palette integration.
	Enabled bool
	// OpenKeys open the palette from input mode.
	OpenKeys []string
	// CloseKeys close the palette while it is open.
	CloseKeys []string
	// SlashOpenEnabled enables slash-triggered palette open behavior.
	SlashOpenEnabled bool
	// SlashPolicy decides when slash should open the palette.
	// Supported values: empty-input, column-zero, provider.
	SlashPolicy CommandPaletteSlashPolicy
	// MaxVisibleItems limits visible rows rendered by the palette model.
	MaxVisibleItems int
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

// DefaultHelpBarConfig returns default contextual help bar settings.
func DefaultHelpBarConfig() HelpBarConfig {
	return HelpBarConfig{
		Enabled:        true,
		Debounce:       120 * time.Millisecond,
		RequestTimeout: 300 * time.Millisecond,
	}
}

// DefaultHelpDrawerConfig returns default help drawer settings.
func DefaultHelpDrawerConfig() HelpDrawerConfig {
	return HelpDrawerConfig{
		Enabled:            true,
		ToggleKeys:         []string{"ctrl+h"},
		CloseKeys:          []string{"esc", "ctrl+h"},
		RefreshShortcuts:   []string{"ctrl+r"},
		PinShortcuts:       []string{"ctrl+g"},
		Debounce:           140 * time.Millisecond,
		RequestTimeout:     500 * time.Millisecond,
		Dock:               HelpDrawerDockAboveRepl,
		WidthPercent:       52,
		HeightPercent:      46,
		Margin:             1,
		PrefetchWhenHidden: false,
	}
}

// DefaultCommandPaletteConfig returns default command palette settings.
func DefaultCommandPaletteConfig() CommandPaletteConfig {
	return CommandPaletteConfig{
		Enabled:          true,
		OpenKeys:         []string{"ctrl+p"},
		CloseKeys:        []string{"esc", "ctrl+p"},
		SlashOpenEnabled: true,
		SlashPolicy:      CommandPaletteSlashPolicyEmptyInput,
		MaxVisibleItems:  8,
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
	// HelpBar controls contextual in-line help updates while typing.
	HelpBar HelpBarConfig
	// HelpDrawer controls keyboard-toggle contextual panel behavior.
	HelpDrawer HelpDrawerConfig
	// CommandPalette controls command discovery/dispatch overlay behavior.
	CommandPalette CommandPaletteConfig
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
		HelpBar:              DefaultHelpBarConfig(),
		HelpDrawer:           DefaultHelpDrawerConfig(),
		CommandPalette:       DefaultCommandPaletteConfig(),
	}
}
