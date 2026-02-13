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

type fakeHelpDrawerEvaluator struct {
	doc        HelpDrawerDocument
	err        error
	requests   []HelpDrawerRequest
	delay      time.Duration
	panicOnGet bool
}

func (f *fakeHelpDrawerEvaluator) EvaluateStream(context.Context, string, func(Event)) error {
	return nil
}
func (f *fakeHelpDrawerEvaluator) GetPrompt() string        { return "> " }
func (f *fakeHelpDrawerEvaluator) GetName() string          { return "fake-help-drawer" }
func (f *fakeHelpDrawerEvaluator) SupportsMultiline() bool  { return false }
func (f *fakeHelpDrawerEvaluator) GetFileExtension() string { return ".txt" }

func (f *fakeHelpDrawerEvaluator) GetHelpDrawer(ctx context.Context, req HelpDrawerRequest) (HelpDrawerDocument, error) {
	if f.panicOnGet {
		panic("fake help drawer panic")
	}
	if f.delay > 0 {
		timer := time.NewTimer(f.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return HelpDrawerDocument{}, ctx.Err()
		case <-timer.C:
		}
	}
	f.requests = append(f.requests, req)
	return f.doc, f.err
}

func newHelpDrawerTestModel(t *testing.T, evaluator *fakeHelpDrawerEvaluator) *Model {
	t.Helper()

	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Autocomplete.Enabled = false
	cfg.HelpBar.Enabled = false
	cfg.HelpDrawer = DefaultHelpDrawerConfig()
	cfg.HelpDrawer.Debounce = time.Nanosecond
	cfg.HelpDrawer.RequestTimeout = 50 * time.Millisecond
	cfg.HelpDrawer.WidthPercent = 50
	cfg.HelpDrawer.HeightPercent = 40

	m := NewModel(evaluator, cfg, bus.Publisher)
	m.width = 120
	m.height = 40
	return m
}

func TestHelpDrawerToggleOpenClose(t *testing.T) {
	evaluator := &fakeHelpDrawerEvaluator{
		doc: HelpDrawerDocument{
			Show:     true,
			Title:    "console.log",
			Subtitle: "function",
			Markdown: "Writes to stdout.",
		},
	}
	m := newHelpDrawerTestModel(t, evaluator)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	drainModelCmds(m, cmd)

	assert.True(t, m.helpDrawerVisible)
	assert.False(t, m.helpDrawerLoading)
	require.Len(t, evaluator.requests, 1)
	assert.Equal(t, HelpDrawerTriggerToggleOpen, evaluator.requests[0].Trigger)

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	drainModelCmds(m, cmd)
	assert.False(t, m.helpDrawerVisible)
}

func TestHelpDrawerAdaptiveTypingWhenVisible(t *testing.T) {
	evaluator := &fakeHelpDrawerEvaluator{
		doc: HelpDrawerDocument{
			Show:     true,
			Title:    "console",
			Subtitle: "object",
			Markdown: "Console API.",
		},
	}
	m := newHelpDrawerTestModel(t, evaluator)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	drainModelCmds(m, cmd)
	require.Len(t, evaluator.requests, 1)

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	drainModelCmds(m, cmd)
	require.Len(t, evaluator.requests, 2)
	assert.Equal(t, HelpDrawerTriggerTyping, evaluator.requests[1].Trigger)
}

func TestHelpDrawerResultDropsStaleResponse(t *testing.T) {
	evaluator := &fakeHelpDrawerEvaluator{}
	m := newHelpDrawerTestModel(t, evaluator)
	m.helpDrawerVisible = true
	m.helpDrawerReqSeq = 2
	m.helpDrawerDoc = HelpDrawerDocument{Show: true, Title: "current"}

	_ = m.handleHelpDrawerResult(helpDrawerResultMsg{
		RequestID: 1,
		Doc:       HelpDrawerDocument{Show: true, Title: "stale"},
	})
	assert.Equal(t, "current", m.helpDrawerDoc.Title)

	_ = m.handleHelpDrawerResult(helpDrawerResultMsg{
		RequestID: 2,
		Doc:       HelpDrawerDocument{Show: true, Title: "fresh"},
	})
	assert.Equal(t, "fresh", m.helpDrawerDoc.Title)
}

func TestHelpDrawerCloseKey(t *testing.T) {
	evaluator := &fakeHelpDrawerEvaluator{
		doc: HelpDrawerDocument{
			Show:     true,
			Title:    "console",
			Subtitle: "object",
		},
	}
	m := newHelpDrawerTestModel(t, evaluator)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	drainModelCmds(m, cmd)
	require.True(t, m.helpDrawerVisible)

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	drainModelCmds(m, cmd)
	assert.False(t, m.helpDrawerVisible)
}
