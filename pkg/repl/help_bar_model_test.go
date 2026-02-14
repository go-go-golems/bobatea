package repl

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeHelpBarEvaluator struct {
	payload    HelpBarPayload
	err        error
	requests   []HelpBarRequest
	panicOnGet bool
	delay      time.Duration
}

func (f *fakeHelpBarEvaluator) EvaluateStream(context.Context, string, func(Event)) error { return nil }
func (f *fakeHelpBarEvaluator) GetPrompt() string                                         { return "> " }
func (f *fakeHelpBarEvaluator) GetName() string                                           { return "fake-help-bar" }
func (f *fakeHelpBarEvaluator) SupportsMultiline() bool                                   { return false }
func (f *fakeHelpBarEvaluator) GetFileExtension() string                                  { return ".txt" }

func (f *fakeHelpBarEvaluator) GetHelpBar(ctx context.Context, req HelpBarRequest) (HelpBarPayload, error) {
	if f.panicOnGet {
		panic("fake help bar panic")
	}
	if f.delay > 0 {
		timer := time.NewTimer(f.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return HelpBarPayload{}, ctx.Err()
		case <-timer.C:
		}
	}
	f.requests = append(f.requests, req)
	return f.payload, f.err
}

func newHelpBarTestModel(t *testing.T, evaluator *fakeHelpBarEvaluator) *Model {
	t.Helper()

	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Autocomplete.Enabled = false
	cfg.HelpBar = DefaultHelpBarConfig()
	cfg.HelpBar.Debounce = time.Nanosecond
	cfg.HelpBar.RequestTimeout = 50 * time.Millisecond

	return NewModel(evaluator, cfg, bus.Publisher)
}

func TestHelpBarDebounceCoalescesToLatestRequest(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{payload: HelpBarPayload{Show: true, Text: "console.log"}}
	m := newHelpBarTestModel(t, evaluator)

	m.textInput.SetValue("c")
	m.textInput.SetCursor(1)
	cmd1 := m.scheduleDebouncedHelpBarIfNeeded("", 0)
	require.NotNil(t, cmd1)
	msg1, ok := cmd1().(helpBarDebounceMsg)
	require.True(t, ok)

	m.textInput.SetValue("co")
	m.textInput.SetCursor(2)
	cmd2 := m.scheduleDebouncedHelpBarIfNeeded("c", 1)
	require.NotNil(t, cmd2)
	msg2, ok := cmd2().(helpBarDebounceMsg)
	require.True(t, ok)

	stale := m.handleDebouncedHelpBar(msg1)
	assert.Nil(t, stale)

	active := m.handleDebouncedHelpBar(msg2)
	require.NotNil(t, active)
	_, ok = active().(helpBarResultMsg)
	require.True(t, ok)

	require.Len(t, evaluator.requests, 1)
	assert.Equal(t, msg2.RequestID, evaluator.requests[0].RequestID)
	assert.Equal(t, HelpBarReasonDebounce, evaluator.requests[0].Reason)
}

func TestHelpBarResultDropsStaleResponse(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{}
	m := newHelpBarTestModel(t, evaluator)
	m.helpBar.reqSeq = 2

	stale := helpBarResultMsg{
		RequestID: 1,
		Payload:   HelpBarPayload{Show: true, Text: "stale"},
	}
	_ = m.handleHelpBarResult(stale)
	assert.False(t, m.helpBar.visible)

	current := helpBarResultMsg{
		RequestID: 2,
		Payload:   HelpBarPayload{Show: true, Text: "active", Severity: "info"},
	}
	_ = m.handleHelpBarResult(current)
	assert.True(t, m.helpBar.visible)
	assert.Equal(t, "active", m.helpBar.payload.Text)
}

func TestHelpBarResultHidesOnShowFalseOrEmptyText(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{}
	m := newHelpBarTestModel(t, evaluator)
	m.helpBar.reqSeq = 1

	_ = m.handleHelpBarResult(helpBarResultMsg{
		RequestID: 1,
		Payload:   HelpBarPayload{Show: true, Text: "signature foo(x)"},
	})
	assert.True(t, m.helpBar.visible)

	_ = m.handleHelpBarResult(helpBarResultMsg{
		RequestID: 1,
		Payload:   HelpBarPayload{Show: false, Text: "hidden"},
	})
	assert.False(t, m.helpBar.visible)

	_ = m.handleHelpBarResult(helpBarResultMsg{
		RequestID: 1,
		Payload:   HelpBarPayload{Show: true, Text: "   "},
	})
	assert.False(t, m.helpBar.visible)
}

func TestHelpBarDebounceInputChangeKeepsBarVisible(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{}
	m := newHelpBarTestModel(t, evaluator)
	m.helpBar.visible = true
	m.helpBar.payload = HelpBarPayload{Show: true, Text: "old context"}
	m.textInput.SetValue("co")
	m.textInput.SetCursor(2)

	m.textInput.SetValue("con")
	m.textInput.SetCursor(3)
	cmd := m.scheduleDebouncedHelpBarIfNeeded("co", 2)
	require.NotNil(t, cmd)
	assert.True(t, m.helpBar.visible, "help bar should remain visible while debounce request is pending")
}

func TestHelpBarStyleSeverityMapping(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{}
	m := newHelpBarTestModel(t, evaluator)

	assert.Equal(t, m.styles.Error.Render("boom"), m.helpBarStyleForSeverity("error").Render("boom"))
	assert.Equal(t, m.styles.HelpText.Render("warn"), m.helpBarStyleForSeverity("warning").Render("warn"))
	assert.Equal(t, m.styles.Info.Render("info"), m.helpBarStyleForSeverity("info").Render("info"))
	assert.Equal(t, m.styles.Info.Render("default"), m.helpBarStyleForSeverity("").Render("default"))
}

func TestHelpBarEndToEndTypingFlow(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{
		payload: HelpBarPayload{
			Show:     true,
			Text:     "console.log(...args): void",
			Severity: "info",
			Kind:     "signature",
		},
	}
	m := newHelpBarTestModel(t, evaluator)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	drainModelCmds(m, cmd)

	require.Len(t, evaluator.requests, 1)
	assert.Equal(t, HelpBarReasonDebounce, evaluator.requests[0].Reason)
	assert.True(t, m.helpBar.visible)
	assert.Equal(t, "console.log(...args): void", m.helpBar.payload.Text)
}

func TestHelpBarCmdRecoversFromProviderPanic(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{panicOnGet: true}
	m := newHelpBarTestModel(t, evaluator)

	req := HelpBarRequest{
		Input:      "co",
		CursorByte: 2,
		Reason:     HelpBarReasonDebounce,
		RequestID:  1,
	}
	msg, ok := m.helpBarCmd(req)().(helpBarResultMsg)
	require.True(t, ok)
	assert.Equal(t, req.RequestID, msg.RequestID)
	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "help bar provider panic")
}

func TestHelpBarCmdTimesOutSlowProvider(t *testing.T) {
	evaluator := &fakeHelpBarEvaluator{
		delay:   10 * time.Millisecond,
		payload: HelpBarPayload{Show: true, Text: "late"},
	}
	m := newHelpBarTestModel(t, evaluator)
	m.helpBar.reqTimeout = time.Millisecond

	req := HelpBarRequest{
		Input:      "co",
		CursorByte: 2,
		Reason:     HelpBarReasonDebounce,
		RequestID:  1,
	}
	msg, ok := m.helpBarCmd(req)().(helpBarResultMsg)
	require.True(t, ok)
	require.Error(t, msg.Err)
	assert.ErrorIs(t, msg.Err, context.DeadlineExceeded)
}

type noHelpBarEvaluator struct{}

func (n *noHelpBarEvaluator) EvaluateStream(context.Context, string, func(Event)) error { return nil }
func (n *noHelpBarEvaluator) GetPrompt() string                                         { return "> " }
func (n *noHelpBarEvaluator) GetName() string                                           { return "no-help-bar" }
func (n *noHelpBarEvaluator) SupportsMultiline() bool                                   { return false }
func (n *noHelpBarEvaluator) GetFileExtension() string                                  { return ".txt" }

func TestHelpBarNoProviderIsInert(t *testing.T) {
	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.HelpBar = DefaultHelpBarConfig()
	m := NewModel(&noHelpBarEvaluator{}, cfg, bus.Publisher)
	require.Nil(t, m.helpBar.provider)

	prevValue := m.textInput.Value()
	prevCursor := m.textInput.Position()
	m.textInput.SetValue("c")
	m.textInput.SetCursor(1)
	cmd := m.scheduleDebouncedHelpBarIfNeeded(prevValue, prevCursor)
	assert.Nil(t, cmd)
	assert.False(t, m.helpBar.visible)
	assert.Equal(t, uint64(0), m.helpBar.reqSeq)
}
