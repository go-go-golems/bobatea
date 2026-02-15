package repl

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleEvaluator(t *testing.T) {
	evaluator := NewExampleEvaluator()

	// Test basic functionality
	assert.Equal(t, "Example", evaluator.GetName())
	assert.Equal(t, "example> ", evaluator.GetPrompt())
	assert.True(t, evaluator.SupportsMultiline())
	assert.Equal(t, ".txt", evaluator.GetFileExtension())

	// Test evaluation
	ctx := context.Background()

	// Test echo
	result, err := evaluator.Evaluate(ctx, "echo hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)

	// Test math
	result, err = evaluator.Evaluate(ctx, "5 + 3")
	assert.NoError(t, err)
	assert.Equal(t, "8", result)

	// Test default
	result, err = evaluator.Evaluate(ctx, "random input")
	assert.NoError(t, err)
	assert.Equal(t, "You typed: random input", result)
}

func TestHistory(t *testing.T) {
	history := NewHistory(5)

	// Test adding entries
	history.Add("input1", "output1", false)
	history.Add("input2", "output2", true)

	entries := history.GetEntries()
	assert.Len(t, entries, 2)
	assert.Equal(t, "input1", entries[0].Input)
	assert.Equal(t, "output1", entries[0].Output)
	assert.False(t, entries[0].IsErr)
	assert.Equal(t, "input2", entries[1].Input)
	assert.Equal(t, "output2", entries[1].Output)
	assert.True(t, entries[1].IsErr)

	// Test navigation
	assert.False(t, history.IsNavigating())

	up1 := history.NavigateUp()
	assert.Equal(t, "input2", up1)
	assert.True(t, history.IsNavigating())

	up2 := history.NavigateUp()
	assert.Equal(t, "input1", up2)

	down1 := history.NavigateDown()
	assert.Equal(t, "input2", down1)

	down2 := history.NavigateDown()
	assert.Equal(t, "", down2)
	assert.False(t, history.IsNavigating())

	// Test clear
	history.Clear()
	assert.Len(t, history.GetEntries(), 0)
}

func TestModel(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	assert.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	// Test initialization
	assert.Equal(t, evaluator, model.evaluator)
	assert.Equal(t, config, model.config)
	assert.Equal(t, config.Width, model.help.Width)
}

func TestModelWindowSizeUpdatesHelpWidth(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	_, _ = model.Update(tea.WindowSizeMsg{Width: 42, Height: 16})
	assert.Equal(t, 42, model.help.Width)
}

func TestModelToggleHelpReflowsTimelineHeight(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	_, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	baseTimelineHeight := model.timelineHeight
	baseHelpLines := strings.Count(model.renderHelp(), "\n") + 1

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})

	fullHelp := model.renderHelp()
	fullHelpLines := strings.Count(fullHelp, "\n") + 1
	assert.True(t, model.help.ShowAll)
	assert.Greater(t, fullHelpLines, baseHelpLines)
	assert.Less(t, model.timelineHeight, baseTimelineHeight)
}

func TestModelFullHelpUsesAvailableWidth(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	_, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	narrowHelp := model.renderHelp()
	narrowLines := strings.Count(narrowHelp, "\n") + 1

	_, _ = model.Update(tea.WindowSizeMsg{Width: 140, Height: 24})
	wideHelp := model.renderHelp()
	wideLines := strings.Count(wideHelp, "\n") + 1

	assert.NotContains(t, narrowHelp, "…")
	assert.NotContains(t, wideHelp, "…")
	assert.Less(t, wideLines, narrowLines)
}

func TestModelFullHelpDoesNotDropBindingsAtMediumWidth(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	_, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	fullHelp := model.renderHelp()

	assert.NotContains(t, fullHelp, "…")
	assert.Contains(t, fullHelp, "open palette")
	assert.Contains(t, fullHelp, "completion page down")
}

func TestModelShortHelpShowsHelpDrawerToggle(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	_, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shortHelp := model.renderHelp()
	assert.Contains(t, shortHelp, "toggle drawer")
}

func TestLayoutAccountsForHelpBarHeight(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	_, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	baseTimelineHeight := model.timelineHeight

	_ = model.helpBar.widget.HandleResult(helpBarResultMsg{
		RequestID: model.helpBar.widget.RequestSeq(),
		Payload: HelpBarPayload{
			Show: true,
			Text: "symbol: fn(x): y",
		},
	})
	model.applyLayoutFromState()

	assert.Less(t, model.timelineHeight, baseTimelineHeight)
}

func TestNewModelWiresFeatureSubmodelsFromConfig(t *testing.T) {
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Autocomplete = DefaultAutocompleteConfig()
	cfg.Autocomplete.Debounce = 75 * time.Millisecond
	cfg.Autocomplete.RequestTimeout = 250 * time.Millisecond
	cfg.Autocomplete.MaxSuggestions = 11
	cfg.Autocomplete.OverlayPlacement = CompletionOverlayPlacementBottom
	cfg.Autocomplete.OverlayHorizontalGrow = CompletionOverlayHorizontalGrowLeft
	cfg.HelpBar = DefaultHelpBarConfig()
	cfg.HelpBar.Debounce = 95 * time.Millisecond
	cfg.HelpBar.RequestTimeout = 280 * time.Millisecond
	cfg.HelpDrawer = DefaultHelpDrawerConfig()
	cfg.HelpDrawer.Debounce = 105 * time.Millisecond
	cfg.HelpDrawer.RequestTimeout = 350 * time.Millisecond
	cfg.HelpDrawer.Dock = HelpDrawerDockRight
	cfg.HelpDrawer.WidthPercent = 61
	cfg.HelpDrawer.HeightPercent = 45
	cfg.HelpDrawer.Margin = 2

	model := NewModel(&cancellableProviderEvaluator{}, cfg, bus.Publisher)

	require.NotNil(t, model.completion.provider)
	assert.Equal(t, cfg.Autocomplete.Debounce, model.completion.debounce)
	assert.Equal(t, cfg.Autocomplete.RequestTimeout, model.completion.reqTimeout)
	assert.Equal(t, cfg.Autocomplete.MaxSuggestions, model.completion.maxVisible)
	assert.Equal(t, cfg.Autocomplete.OverlayPlacement, model.completion.placement)
	assert.Equal(t, cfg.Autocomplete.OverlayHorizontalGrow, model.completion.horizontal)

	require.NotNil(t, model.helpBar.provider)
	assert.Equal(t, cfg.HelpBar.Debounce, model.helpBar.debounce)
	assert.Equal(t, cfg.HelpBar.RequestTimeout, model.helpBar.reqTimeout)

	require.NotNil(t, model.helpDrawer.provider)
	assert.Equal(t, cfg.HelpDrawer.Debounce, model.helpDrawer.debounce)
	assert.Equal(t, cfg.HelpDrawer.RequestTimeout, model.helpDrawer.reqTimeout)
	assert.Equal(t, cfg.HelpDrawer.Dock, model.helpDrawer.dock)
	assert.Equal(t, cfg.HelpDrawer.WidthPercent, model.helpDrawer.widthPercent)
	assert.Equal(t, cfg.HelpDrawer.HeightPercent, model.helpDrawer.heightPercent)
	assert.Equal(t, cfg.HelpDrawer.Margin, model.helpDrawer.margin)
}

func TestNewModelDisablesFeatureProvidersWhenConfigDisabled(t *testing.T) {
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Autocomplete = DefaultAutocompleteConfig()
	cfg.Autocomplete.Enabled = false
	cfg.HelpBar = DefaultHelpBarConfig()
	cfg.HelpBar.Enabled = false
	cfg.HelpDrawer = DefaultHelpDrawerConfig()
	cfg.HelpDrawer.Enabled = false

	model := NewModel(&cancellableProviderEvaluator{}, cfg, bus.Publisher)

	assert.Nil(t, model.completion.provider)
	assert.Nil(t, model.helpBar.provider)
	assert.Nil(t, model.helpDrawer.provider)
}

func TestModelWithContext(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	assert.NoError(t, err)

	parentCtx, cancel := context.WithCancel(context.Background())
	model := NewModelWithContext(parentCtx, evaluator, config, bus.Publisher)

	cancel()
	select {
	case <-model.appCtx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Fatal("model app context should be canceled when parent context is canceled")
	}
}

type cancellableProviderEvaluator struct{}

func (e *cancellableProviderEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	<-ctx.Done()
	return "", ctx.Err()
}

func (e *cancellableProviderEvaluator) EvaluateStream(ctx context.Context, _ string, _ func(Event)) error {
	<-ctx.Done()
	return ctx.Err()
}

func (e *cancellableProviderEvaluator) GetPrompt() string        { return "> " }
func (e *cancellableProviderEvaluator) GetName() string          { return "cancellable" }
func (e *cancellableProviderEvaluator) SupportsMultiline() bool  { return false }
func (e *cancellableProviderEvaluator) GetFileExtension() string { return ".txt" }

func (e *cancellableProviderEvaluator) CompleteInput(ctx context.Context, _ CompletionRequest) (CompletionResult, error) {
	<-ctx.Done()
	return CompletionResult{}, ctx.Err()
}

func (e *cancellableProviderEvaluator) GetHelpBar(ctx context.Context, _ HelpBarRequest) (HelpBarPayload, error) {
	<-ctx.Done()
	return HelpBarPayload{}, ctx.Err()
}

func (e *cancellableProviderEvaluator) GetHelpDrawer(ctx context.Context, _ HelpDrawerRequest) (HelpDrawerDocument, error) {
	<-ctx.Done()
	return HelpDrawerDocument{}, ctx.Err()
}

func TestProviderCommandsUseAppContextCancellation(t *testing.T) {
	bus, err := eventbus.NewInMemoryBus()
	assert.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Autocomplete.Enabled = true
	cfg.HelpBar.Enabled = true
	cfg.HelpDrawer.Enabled = true
	cfg.Autocomplete.RequestTimeout = 2 * time.Second
	cfg.HelpBar.RequestTimeout = 2 * time.Second
	cfg.HelpDrawer.RequestTimeout = 2 * time.Second

	parentCtx, cancel := context.WithCancel(context.Background())
	model := NewModelWithContext(parentCtx, &cancellableProviderEvaluator{}, cfg, bus.Publisher)
	cancel()

	compMsg, ok := model.completionCmd(CompletionRequest{RequestID: 1})().(completionResultMsg)
	assert.True(t, ok)
	assert.ErrorIs(t, compMsg.Err, context.Canceled)

	helpBarMsg, ok := model.helpBarCmd(HelpBarRequest{RequestID: 1})().(helpBarResultMsg)
	assert.True(t, ok)
	assert.ErrorIs(t, helpBarMsg.Err, context.Canceled)

	helpDrawerMsg, ok := model.helpDrawerCmd(HelpDrawerRequest{RequestID: 1})().(helpDrawerResultMsg)
	assert.True(t, ok)
	assert.ErrorIs(t, helpDrawerMsg.Err, context.Canceled)
}

func TestQuitCancelsModelAppContext(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	assert.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	select {
	case <-model.appCtx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected app context to be canceled after quit")
	}
}

func TestConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "REPL", config.Title)
	assert.Equal(t, "> ", config.Prompt)
	assert.Equal(t, "Enter code or /command", config.Placeholder)
	assert.Equal(t, 80, config.Width)
	assert.False(t, config.StartMultiline)
	// Default config for external editor is disabled in timeline-based REPL
	assert.False(t, config.EnableExternalEditor)
	assert.True(t, config.EnableHistory)
	assert.Equal(t, 1000, config.MaxHistorySize)
	assert.True(t, config.Autocomplete.Enabled)
	assert.Equal(t, 120*time.Millisecond, config.Autocomplete.Debounce)
	assert.Equal(t, 400*time.Millisecond, config.Autocomplete.RequestTimeout)
	assert.Equal(t, []string{"tab"}, config.Autocomplete.TriggerKeys)
	assert.Equal(t, []string{"enter", "tab"}, config.Autocomplete.AcceptKeys)
	assert.Equal(t, 8, config.Autocomplete.MaxSuggestions)
	assert.Equal(t, 56, config.Autocomplete.OverlayMaxWidth)
	assert.Equal(t, 12, config.Autocomplete.OverlayMaxHeight)
	assert.Equal(t, 24, config.Autocomplete.OverlayMinWidth)
	assert.Equal(t, 1, config.Autocomplete.OverlayMargin)
	assert.Equal(t, 0, config.Autocomplete.OverlayPageSize)
	assert.Equal(t, 0, config.Autocomplete.OverlayOffsetX)
	assert.Equal(t, 0, config.Autocomplete.OverlayOffsetY)
	assert.False(t, config.Autocomplete.OverlayNoBorder)
	assert.Equal(t, CompletionOverlayPlacementAuto, config.Autocomplete.OverlayPlacement)
	assert.Equal(t, CompletionOverlayHorizontalGrowRight, config.Autocomplete.OverlayHorizontalGrow)
	assert.True(t, config.HelpBar.Enabled)
	assert.Equal(t, 120*time.Millisecond, config.HelpBar.Debounce)
	assert.Equal(t, 300*time.Millisecond, config.HelpBar.RequestTimeout)
	assert.True(t, config.HelpDrawer.Enabled)
	assert.Equal(t, []string{"alt+h"}, config.HelpDrawer.ToggleKeys)
	assert.Equal(t, []string{"esc", "alt+h"}, config.HelpDrawer.CloseKeys)
	assert.Equal(t, []string{"ctrl+r"}, config.HelpDrawer.RefreshShortcuts)
	assert.Equal(t, []string{"ctrl+g"}, config.HelpDrawer.PinShortcuts)
	assert.Equal(t, 140*time.Millisecond, config.HelpDrawer.Debounce)
	assert.Equal(t, 500*time.Millisecond, config.HelpDrawer.RequestTimeout)
	assert.Equal(t, HelpDrawerDockAboveRepl, config.HelpDrawer.Dock)
	assert.Equal(t, 52, config.HelpDrawer.WidthPercent)
	assert.Equal(t, 46, config.HelpDrawer.HeightPercent)
	assert.Equal(t, 1, config.HelpDrawer.Margin)
	assert.False(t, config.HelpDrawer.PrefetchWhenHidden)
	assert.True(t, config.CommandPalette.Enabled)
	assert.Equal(t, []string{"ctrl+p"}, config.CommandPalette.OpenKeys)
	assert.Equal(t, []string{"esc", "ctrl+p"}, config.CommandPalette.CloseKeys)
	assert.True(t, config.CommandPalette.SlashOpenEnabled)
	assert.Equal(t, CommandPaletteSlashPolicyEmptyInput, config.CommandPalette.SlashPolicy)
	assert.Equal(t, 8, config.CommandPalette.MaxVisibleItems)
	assert.Equal(t, CommandPaletteOverlayPlacementCenter, config.CommandPalette.OverlayPlacement)
	assert.Equal(t, 1, config.CommandPalette.OverlayMargin)
	assert.Equal(t, 0, config.CommandPalette.OverlayOffsetX)
	assert.Equal(t, 0, config.CommandPalette.OverlayOffsetY)
}

func TestNormalizeCommandPaletteConfigDefaults(t *testing.T) {
	normalized := normalizeCommandPaletteConfig(CommandPaletteConfig{})
	assert.Equal(t, DefaultCommandPaletteConfig(), normalized)
}

func TestNormalizeCommandPaletteConfigSanitizesValues(t *testing.T) {
	cfg := CommandPaletteConfig{
		Enabled:          true,
		OpenKeys:         []string{"f1"},
		CloseKeys:        []string{"esc"},
		SlashOpenEnabled: true,
		SlashPolicy:      CommandPaletteSlashPolicy("unknown"),
		MaxVisibleItems:  200,
		OverlayPlacement: CommandPaletteOverlayPlacement("weird"),
		OverlayMargin:    -5,
		OverlayOffsetX:   7,
		OverlayOffsetY:   -3,
	}
	normalized := normalizeCommandPaletteConfig(cfg)
	assert.True(t, normalized.Enabled)
	assert.Equal(t, []string{"f1"}, normalized.OpenKeys)
	assert.Equal(t, []string{"esc"}, normalized.CloseKeys)
	assert.True(t, normalized.SlashOpenEnabled)
	assert.Equal(t, CommandPaletteSlashPolicyEmptyInput, normalized.SlashPolicy)
	assert.Equal(t, 50, normalized.MaxVisibleItems)
	assert.Equal(t, CommandPaletteOverlayPlacementCenter, normalized.OverlayPlacement)
	assert.Equal(t, 1, normalized.OverlayMargin)
	assert.Equal(t, 7, normalized.OverlayOffsetX)
	assert.Equal(t, -3, normalized.OverlayOffsetY)
}

func TestNormalizeCommandPaletteConfigHonorsSlashOpenDisabled(t *testing.T) {
	cfg := DefaultCommandPaletteConfig()
	cfg.SlashOpenEnabled = false

	normalized := normalizeCommandPaletteConfig(cfg)
	assert.False(t, normalized.SlashOpenEnabled)
}

func TestStyles(t *testing.T) {
	styles := DefaultStyles()

	// Just test that styles are created
	assert.NotNil(t, styles.Title)
	assert.NotNil(t, styles.Prompt)
	assert.NotNil(t, styles.Result)
	assert.NotNil(t, styles.Error)
	assert.NotNil(t, styles.Info)
	assert.NotNil(t, styles.HelpText)

	// Test themes
	assert.Contains(t, BuiltinThemes, "default")
	assert.Contains(t, BuiltinThemes, "dark")
	assert.Contains(t, BuiltinThemes, "light")
}
