package repl

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCompleterEvaluator struct {
	result   CompletionResult
	err      error
	requests []CompletionRequest
}

func (f *fakeCompleterEvaluator) EvaluateStream(context.Context, string, func(Event)) error {
	return nil
}

func (f *fakeCompleterEvaluator) GetPrompt() string        { return "> " }
func (f *fakeCompleterEvaluator) GetName() string          { return "fake" }
func (f *fakeCompleterEvaluator) SupportsMultiline() bool  { return false }
func (f *fakeCompleterEvaluator) GetFileExtension() string { return ".txt" }
func (f *fakeCompleterEvaluator) CompleteInput(_ context.Context, req CompletionRequest) (CompletionResult, error) {
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

	m.completionReqSeq = 2
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
	assert.False(t, m.completionVisible)
	assert.Equal(t, uint64(0), m.completionLastReqID)

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
	assert.True(t, m.completionVisible)
	assert.Equal(t, 0, m.completionSelection)
	assert.Equal(t, uint64(2), m.completionLastReqID)
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
	m.completionVisible = true
	m.completionLastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "const", DisplayText: "const"},
			{Id: "2", Value: "console", DisplayText: "console"},
		},
		ReplaceFrom: 0,
		ReplaceTo:   4,
	}
	m.completionSelection = 0
	m.completionReplaceFrom = 0
	m.completionReplaceTo = 4

	m.history.Add("history-entry", "", false)

	_, _ = m.updateInput(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m.completionSelection, "down key should navigate popup when visible")
	assert.Equal(t, "cons", m.textInput.Value(), "history navigation must not run while popup handles key")

	_, _ = m.updateInput(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "console", m.textInput.Value())
	assert.False(t, m.completionVisible, "popup should close after apply")
}
