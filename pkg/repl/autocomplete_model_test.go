package repl

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCompleterEvaluator struct {
	result          CompletionResult
	err             error
	requests        []CompletionRequest
	panicOnComplete bool
}

func (f *fakeCompleterEvaluator) EvaluateStream(context.Context, string, func(Event)) error {
	return nil
}

func (f *fakeCompleterEvaluator) GetPrompt() string        { return "> " }
func (f *fakeCompleterEvaluator) GetName() string          { return "fake" }
func (f *fakeCompleterEvaluator) SupportsMultiline() bool  { return false }
func (f *fakeCompleterEvaluator) GetFileExtension() string { return ".txt" }
func (f *fakeCompleterEvaluator) CompleteInput(_ context.Context, req CompletionRequest) (CompletionResult, error) {
	if f.panicOnComplete {
		panic("fake completer panic")
	}
	f.requests = append(f.requests, req)
	return f.result, f.err
}

func newAutocompleteTestModel(t *testing.T, evaluator *fakeCompleterEvaluator) *Model {
	t.Helper()

	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Autocomplete = DefaultAutocompleteConfig()
	cfg.Autocomplete.Debounce = time.Nanosecond
	cfg.Autocomplete.RequestTimeout = 50 * time.Millisecond

	return NewModel(evaluator, cfg, bus.Publisher)
}

func drainModelCmds(m *Model, cmd tea.Cmd) {
	queue := []tea.Cmd{cmd}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == nil {
			continue
		}

		msg := current()
		if msg == nil {
			continue
		}

		if batch, ok := msg.(tea.BatchMsg); ok {
			queue = append(queue, []tea.Cmd(batch)...)
			continue
		}

		_, nextCmd := m.Update(msg)
		queue = append(queue, nextCmd)
	}
}

func TestCompletionDebounceCoalescesToLatestRequest(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	m.textInput.SetValue("a")
	m.textInput.SetCursor(1)
	cmd1 := m.scheduleDebouncedCompletionIfNeeded("", 0)
	require.NotNil(t, cmd1)
	msg1, ok := cmd1().(completionDebounceMsg)
	require.True(t, ok)

	m.textInput.SetValue("ab")
	m.textInput.SetCursor(2)
	cmd2 := m.scheduleDebouncedCompletionIfNeeded("a", 1)
	require.NotNil(t, cmd2)
	msg2, ok := cmd2().(completionDebounceMsg)
	require.True(t, ok)

	staleCmd := m.handleDebouncedCompletion(msg1)
	assert.Nil(t, staleCmd, "outdated debounce request must be dropped")

	activeCmd := m.handleDebouncedCompletion(msg2)
	require.NotNil(t, activeCmd)
	_, ok = activeCmd().(completionResultMsg)
	require.True(t, ok)

	require.Len(t, evaluator.requests, 1)
	assert.Equal(t, CompletionReasonDebounce, evaluator.requests[0].Reason)
	assert.Equal(t, msg2.RequestID, evaluator.requests[0].RequestID)
}

func TestCompletionResultDropsStaleResponse(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	m.completion.reqSeq = 2
	m.textInput.SetValue("console.lo")

	stale := completionResultMsg{
		RequestID: 1,
		Result: CompletionResult{
			Show: true,
			Suggestions: []autocomplete.Suggestion{
				{Id: "1", Value: "log", DisplayText: "log"},
			},
			ReplaceFrom: 8,
			ReplaceTo:   10,
		},
	}
	_ = m.handleCompletionResult(stale)
	assert.False(t, m.completion.visible)
	assert.Equal(t, uint64(0), m.completion.lastReqID)

	current := completionResultMsg{
		RequestID: 2,
		Result: CompletionResult{
			Show: true,
			Suggestions: []autocomplete.Suggestion{
				{Id: "1", Value: "log", DisplayText: "log"},
			},
			ReplaceFrom: 8,
			ReplaceTo:   10,
		},
	}
	_ = m.handleCompletionResult(current)
	assert.True(t, m.completion.visible)
	assert.Equal(t, 0, m.completion.selection)
	assert.Equal(t, uint64(2), m.completion.lastReqID)
}

func TestShortcutTriggerUsesShortcutReason(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{
		result: CompletionResult{
			Show: true,
			Suggestions: []autocomplete.Suggestion{
				{Id: "1", Value: "log", DisplayText: "log"},
			},
			ReplaceFrom: 0,
			ReplaceTo:   0,
		},
	}
	m := newAutocompleteTestModel(t, evaluator)
	m.textInput.SetValue("con")
	m.textInput.SetCursor(3)

	_, cmd := m.updateInput(tea.KeyMsg{Type: tea.KeyTab})
	require.NotNil(t, cmd)
	_, ok := cmd().(completionResultMsg)
	require.True(t, ok)

	require.Len(t, evaluator.requests, 1)
	assert.Equal(t, CompletionReasonShortcut, evaluator.requests[0].Reason)
	assert.Equal(t, "tab", evaluator.requests[0].Shortcut)
}

func TestPopupKeyRoutingConsumesNavigationAndApply(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	m.textInput.SetValue("cons")
	m.textInput.SetCursor(4)
	m.completion.visible = true
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "const", DisplayText: "const"},
			{Id: "2", Value: "console", DisplayText: "console"},
		},
		ReplaceFrom: 0,
		ReplaceTo:   4,
	}
	m.completion.selection = 0
	m.completion.replaceFrom = 0
	m.completion.replaceTo = 4

	m.history.Add("history-entry", "", false)

	_, _ = m.updateInput(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m.completion.selection, "down key should navigate popup when visible")
	assert.Equal(t, "cons", m.textInput.Value(), "history navigation must not run while popup handles key")

	_, _ = m.updateInput(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "console", m.textInput.Value())
	assert.False(t, m.completion.visible, "popup should close after apply")
}

func TestAutocompleteEndToEndTypingToApplyFlow(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{
		result: CompletionResult{
			Show: true,
			Suggestions: []autocomplete.Suggestion{
				{Id: "1", Value: "const", DisplayText: "const"},
				{Id: "2", Value: "console", DisplayText: "console"},
			},
			ReplaceFrom: 0,
			ReplaceTo:   1,
		},
	}
	m := newAutocompleteTestModel(t, evaluator)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	drainModelCmds(m, cmd)

	require.Len(t, evaluator.requests, 1)
	assert.Equal(t, CompletionReasonDebounce, evaluator.requests[0].Reason)
	assert.Equal(t, "c", evaluator.requests[0].Input)
	assert.True(t, m.completion.visible)

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	drainModelCmds(m, cmd)
	assert.Equal(t, 1, m.completion.selection)

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drainModelCmds(m, cmd)
	assert.Equal(t, "console", m.textInput.Value())
	assert.False(t, m.completion.visible)
}

func TestCompletionCmdRecoversFromCompleterPanic(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{panicOnComplete: true}
	m := newAutocompleteTestModel(t, evaluator)

	req := CompletionRequest{
		Input:      "co",
		CursorByte: 2,
		Reason:     CompletionReasonShortcut,
		RequestID:  1,
	}
	msg, ok := m.completionCmd(req)().(completionResultMsg)
	require.True(t, ok)
	assert.Equal(t, req.RequestID, msg.RequestID)
	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "input completer panic")
}

func TestComputeCompletionOverlayLayoutClampsToBounds(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	m.width = 40
	m.height = 10
	longInput := strings.Repeat("x", 140)
	m.textInput.SetValue(longInput)
	m.textInput.SetCursor(len(longInput))
	m.completion.visible = true
	m.completion.maxVisible = 10
	m.completion.maxHeight = 6
	m.completion.maxWidth = 28
	m.completion.minWidth = 16
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
			{Id: "2", Value: "constructor", DisplayText: strings.Repeat("X", 90)},
			{Id: "3", Value: "copyWithin", DisplayText: "copyWithin"},
			{Id: "4", Value: "count", DisplayText: "count"},
			{Id: "5", Value: "countReset", DisplayText: "countReset"},
		},
	}

	layout, ok := m.computeCompletionOverlayLayout("title", "timeline\nline")
	require.True(t, ok)
	assert.GreaterOrEqual(t, layout.PopupX, 0)
	assert.GreaterOrEqual(t, layout.PopupY, 0)
	assert.LessOrEqual(t, layout.PopupWidth, m.width)
	assert.LessOrEqual(t, layout.PopupWidth, m.completion.maxWidth)
	assert.GreaterOrEqual(t, layout.PopupWidth, m.completion.minWidth)
	assert.GreaterOrEqual(t, layout.VisibleRows, 1)
	assert.LessOrEqual(t, layout.VisibleRows, 4, "max height 6 with frame should cap visible rows")
}

func TestRenderCompletionPopupUsesScrollWindow(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	m.completion.visible = true
	m.completion.selection = 4
	m.completion.scrollTop = 3
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "alpha", DisplayText: "alpha"},
			{Id: "2", Value: "beta", DisplayText: "beta"},
			{Id: "3", Value: "gamma", DisplayText: "gamma"},
			{Id: "4", Value: "delta", DisplayText: "delta"},
			{Id: "5", Value: "epsilon", DisplayText: "epsilon"},
		},
	}

	popup := m.renderCompletionPopup(completionOverlayLayout{
		PopupWidth:   24,
		VisibleRows:  2,
		ContentWidth: 18,
	})
	assert.Contains(t, popup, "delta")
	assert.Contains(t, popup, "epsilon")
	assert.NotContains(t, popup, "alpha")
	assert.NotContains(t, popup, "beta")
}

func TestCompletionPageNavigationMovesSelectionByViewport(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	m.completion.visible = true
	m.completion.visibleRows = 3
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "a", DisplayText: "a"},
			{Id: "2", Value: "b", DisplayText: "b"},
			{Id: "3", Value: "c", DisplayText: "c"},
			{Id: "4", Value: "d", DisplayText: "d"},
			{Id: "5", Value: "e", DisplayText: "e"},
			{Id: "6", Value: "f", DisplayText: "f"},
			{Id: "7", Value: "g", DisplayText: "g"},
		},
	}

	handled, _ := m.handleCompletionNavigation(tea.KeyMsg{Type: tea.KeyPgDown})
	require.True(t, handled)
	assert.Equal(t, 3, m.completion.selection)
	assert.GreaterOrEqual(t, m.completion.scrollTop, 1)

	handled, _ = m.handleCompletionNavigation(tea.KeyMsg{Type: tea.KeyPgUp})
	require.True(t, handled)
	assert.Equal(t, 0, m.completion.selection)
	assert.Equal(t, 0, m.completion.scrollTop)
}

func TestDebounceInputChangeKeepsPopupVisible(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	m.completion.visible = true
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
		},
		ReplaceFrom: 0,
		ReplaceTo:   2,
	}
	m.textInput.SetValue("co")
	m.textInput.SetCursor(2)

	m.textInput.SetValue("con")
	m.textInput.SetCursor(3)
	cmd := m.scheduleDebouncedCompletionIfNeeded("co", 2)
	require.NotNil(t, cmd)
	assert.True(t, m.completion.visible, "popup should remain visible while debounce request is pending")
}

func TestCompletionOverlayLayoutAppliesOffsets(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)
	m.width = 80
	m.height = 24
	m.completion.visible = true
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
		},
	}
	m.textInput.SetValue("cons")
	m.textInput.SetCursor(4)

	baseLayout, ok := m.computeCompletionOverlayLayout("title", "timeline")
	require.True(t, ok)

	m.completion.offsetX = 3
	m.completion.offsetY = 2
	shiftedLayout, ok := m.computeCompletionOverlayLayout("title", "timeline")
	require.True(t, ok)
	assert.Equal(t, baseLayout.PopupX+3, shiftedLayout.PopupX)
	assert.Equal(t, baseLayout.PopupY+2, shiftedLayout.PopupY)
}

func TestCompletionOverlayLayoutBottomPlacementAnchorsToBottom(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)
	m.width = 80
	m.height = 24
	m.completion.visible = true
	m.completion.placement = CompletionOverlayPlacementBottom
	m.completion.margin = 1
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
			{Id: "2", Value: "const", DisplayText: "const"},
			{Id: "3", Value: "continue", DisplayText: "continue"},
		},
	}
	m.textInput.SetValue("cons")
	m.textInput.SetCursor(4)

	layout, ok := m.computeCompletionOverlayLayout("title", "timeline")
	require.True(t, ok)
	frameHeight := m.completionPopupStyle().GetVerticalFrameSize()
	expectedY := m.height - m.completion.margin - (layout.VisibleRows + frameHeight)
	assert.Equal(t, expectedY, layout.PopupY)
}

func TestCompletionOverlayLayoutGrowsLeftFromAnchor(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)
	m.width = 120
	m.height = 24
	m.completion.visible = true
	m.completion.placement = CompletionOverlayPlacementBottom
	m.completion.margin = 1
	m.textInput.SetValue("console.log")
	m.textInput.SetCursor(len("console.log"))
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
		},
	}

	m.completion.horizontal = CompletionOverlayHorizontalGrowRight
	rightLayout, ok := m.computeCompletionOverlayLayout("title", "timeline")
	require.True(t, ok)

	m.completion.horizontal = CompletionOverlayHorizontalGrowLeft
	leftLayout, ok := m.computeCompletionOverlayLayout("title", "timeline")
	require.True(t, ok)

	expectedLeftX := max(0, rightLayout.PopupX-leftLayout.PopupWidth)
	assert.Equal(t, expectedLeftX, leftLayout.PopupX)
}

func TestNormalizeAutocompleteConfigSanitizesOverlayPlacementAndGrow(t *testing.T) {
	cfg := DefaultAutocompleteConfig()
	cfg.OverlayPlacement = CompletionOverlayPlacement("nonsense")
	cfg.OverlayHorizontalGrow = CompletionOverlayHorizontalGrow("sideways")

	normalized := normalizeAutocompleteConfig(cfg)
	assert.Equal(t, CompletionOverlayPlacementAuto, normalized.OverlayPlacement)
	assert.Equal(t, CompletionOverlayHorizontalGrowRight, normalized.OverlayHorizontalGrow)
}

func TestCompletionPopupStyleNoBorderRemovesFrame(t *testing.T) {
	evaluator := &fakeCompleterEvaluator{}
	m := newAutocompleteTestModel(t, evaluator)

	defaultStyle := m.completionPopupStyle()
	assert.Greater(t, defaultStyle.GetHorizontalFrameSize(), 0)
	assert.Greater(t, defaultStyle.GetVerticalFrameSize(), 0)

	m.completion.noBorder = true
	noBorder := m.completionPopupStyle()
	assert.Equal(t, 0, noBorder.GetHorizontalFrameSize())
	assert.Equal(t, 0, noBorder.GetVerticalFrameSize())
}
