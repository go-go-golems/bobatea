package contextpanel

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	doc        Document
	err        error
	requests   []Request
	delay      time.Duration
	panicOnGet bool
}

func (f *fakeProvider) GetContextPanel(ctx context.Context, req Request) (Document, error) {
	if f.panicOnGet {
		panic("fake context panel panic")
	}
	if f.delay > 0 {
		timer := time.NewTimer(f.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return Document{}, ctx.Err()
		case <-timer.C:
		}
	}
	f.requests = append(f.requests, req)
	return f.doc, f.err
}

func newTestWidget(provider *fakeProvider) *Widget {
	return New(provider, Config{
		Debounce:           time.Nanosecond,
		RequestTimeout:     50 * time.Millisecond,
		PrefetchWhenHidden: false,
		Dock:               DockAboveRepl,
		WidthPercent:       50,
		HeightPercent:      40,
		Margin:             1,
	})
}

func TestToggleOpenClose(t *testing.T) {
	w := newTestWidget(&fakeProvider{
		doc: Document{
			Show:     true,
			Title:    "console.log",
			Subtitle: "function",
		},
	})
	cmd := w.Toggle(context.Background(), "con", 3)
	require.NotNil(t, cmd)
	assert.True(t, w.Visible())

	msg, ok := cmd().(ResultMsg)
	require.True(t, ok)
	w.HandleResult(msg)
	assert.False(t, w.Loading())

	cmd = w.Toggle(context.Background(), "con", 3)
	assert.Nil(t, cmd)
	assert.False(t, w.Visible())
}

func TestDebounceCoalescesToLatestRequest(t *testing.T) {
	w := newTestWidget(&fakeProvider{
		doc: Document{
			Show:  true,
			Title: "console",
		},
	})
	w.SetVisible(true)
	cmd1 := w.OnBufferChanged("", 0, "c", 1)
	require.NotNil(t, cmd1)
	msg1, ok := cmd1().(DebounceMsg)
	require.True(t, ok)

	cmd2 := w.OnBufferChanged("c", 1, "co", 2)
	require.NotNil(t, cmd2)
	msg2, ok := cmd2().(DebounceMsg)
	require.True(t, ok)

	assert.Nil(t, w.HandleDebounce(context.Background(), msg1, "co", 2))

	active := w.HandleDebounce(context.Background(), msg2, "co", 2)
	require.NotNil(t, active)
	_, ok = active().(ResultMsg)
	require.True(t, ok)
}

func TestPinPreventsTypingRefresh(t *testing.T) {
	provider := &fakeProvider{
		doc: Document{
			Show:  true,
			Title: "console",
		},
	}
	w := newTestWidget(provider)

	cmd := w.Toggle(context.Background(), "con", 3)
	require.NotNil(t, cmd)
	w.HandleResult(cmd().(ResultMsg))
	require.Len(t, provider.requests, 1)

	w.TogglePin()
	require.True(t, w.Pinned())
	cmd = w.OnBufferChanged("co", 2, "con", 3)
	assert.Nil(t, cmd)
	assert.Len(t, provider.requests, 1)
}

func TestResultDropsStaleResponse(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetRequestSeq(2)
	w.SetDocument(Document{Show: true, Title: "current"})

	w.HandleResult(ResultMsg{RequestID: 1, Doc: Document{Show: true, Title: "stale"}})
	assert.Equal(t, "current", w.Document().Title)

	w.HandleResult(ResultMsg{RequestID: 2, Doc: Document{Show: true, Title: "fresh"}})
	assert.Equal(t, "fresh", w.Document().Title)
}

func TestRefreshNowUsesManualTrigger(t *testing.T) {
	provider := &fakeProvider{
		doc: Document{
			Show:  true,
			Title: "console",
		},
	}
	w := newTestWidget(provider)
	w.SetVisible(true)

	cmd := w.RequestNow(context.Background(), "con", 3, TriggerManualRefresh)
	require.NotNil(t, cmd)
	result, ok := cmd().(ResultMsg)
	require.True(t, ok)
	w.HandleResult(result)

	require.Len(t, provider.requests, 1)
	assert.Equal(t, TriggerManualRefresh, provider.requests[0].Trigger)
}

func TestOverlayLayoutDockAboveRepl(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetDock(DockAboveRepl)

	headerHeight := 1
	timelineHeight := strings.Count(strings.Repeat("line\n", 30), "\n")
	layout, ok := w.ComputeOverlayLayout(100, 40, headerHeight, timelineHeight)
	require.True(t, ok)
	inputY := headerHeight + 1 + timelineHeight
	assert.LessOrEqual(t, layout.PanelY+layout.PanelHeight, inputY)
	assert.LessOrEqual(t, layout.PanelX+layout.PanelWidth, 100)
}

func TestOverlayLayoutDockRightNoCutoff(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetDock(DockRight)

	layout, ok := w.ComputeOverlayLayout(80, 20, 1, 1)
	require.True(t, ok)
	assert.GreaterOrEqual(t, layout.PanelX, 0)
	assert.LessOrEqual(t, layout.PanelX+layout.PanelWidth, 80)
	assert.LessOrEqual(t, layout.PanelY+layout.PanelHeight, 20)
}

func TestRenderKeepsLastDocumentWhileRefreshing(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetLoading(true)
	w.SetDocument(Document{
		Show:     true,
		Title:    "console.log",
		Subtitle: "function",
		Markdown: "Writes to stdout.",
	})
	layout, ok := w.ComputeOverlayLayout(100, 40, 1, 1)
	require.True(t, ok)

	panel := w.RenderPanel(layout, RenderOptions{
		ToggleBinding:  "alt+h",
		RefreshBinding: "ctrl+r",
		PinBinding:     "ctrl+g",
	})
	assert.Contains(t, panel, "console.log")
	assert.NotContains(t, panel, "No contextual help provider content yet")
}

func TestDebounceMsgType(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	cmd := w.OnBufferChanged("", 0, "a", 1)
	require.NotNil(t, cmd)
	_, ok := cmd().(DebounceMsg)
	assert.True(t, ok)
}

func TestHandleDebounceNilProvider(t *testing.T) {
	w := New(nil, Config{Debounce: time.Nanosecond, RequestTimeout: time.Millisecond})
	cmd := w.HandleDebounce(context.Background(), DebounceMsg{RequestID: 1}, "a", 1)
	assert.Nil(t, cmd)
}

func TestCommandForRequestReturnsResultMsg(t *testing.T) {
	w := newTestWidget(&fakeProvider{
		doc: Document{
			Show:  true,
			Title: "ok",
		},
	})
	msg := w.CommandForRequest(context.Background(), Request{
		Input:      "x",
		CursorByte: 1,
		Trigger:    TriggerManualRefresh,
		RequestID:  1,
	})()
	_, ok := msg.(ResultMsg)
	assert.True(t, ok)
}
