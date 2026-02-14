package contextbar

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	payload Payload
	err     error

	requests   []Request
	delay      time.Duration
	panicOnGet bool
}

func (f *fakeProvider) GetContextBar(ctx context.Context, req Request) (Payload, error) {
	if f.panicOnGet {
		panic("fake context bar panic")
	}
	if f.delay > 0 {
		timer := time.NewTimer(f.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return Payload{}, ctx.Err()
		case <-timer.C:
		}
	}
	f.requests = append(f.requests, req)
	return f.payload, f.err
}

func newTestWidget(provider *fakeProvider) *Widget {
	return New(provider, time.Nanosecond, 50*time.Millisecond)
}

func TestDebounceCoalescesToLatestRequest(t *testing.T) {
	provider := &fakeProvider{payload: Payload{Show: true, Text: "console.log"}}
	w := newTestWidget(provider)

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

	require.Len(t, provider.requests, 1)
	assert.Equal(t, ReasonDebounce, provider.requests[0].Reason)
	assert.Equal(t, msg2.RequestID, provider.requests[0].RequestID)
}

func TestHandleResultDropsStaleResponse(t *testing.T) {
	provider := &fakeProvider{}
	w := newTestWidget(provider)

	w.OnBufferChanged("", 0, "a", 1)
	w.OnBufferChanged("a", 1, "ab", 2)
	require.Equal(t, uint64(2), w.RequestSeq())

	changed := w.HandleResult(ResultMsg{
		RequestID: 1,
		Payload: Payload{
			Show: true,
			Text: "stale",
		},
	})
	assert.False(t, changed)
	assert.False(t, w.Visible())
	assert.Equal(t, uint64(0), w.LastRequestID())

	changed = w.HandleResult(ResultMsg{
		RequestID: 2,
		Payload: Payload{
			Show: true,
			Text: "fresh",
		},
	})
	assert.True(t, changed)
	assert.True(t, w.Visible())
	assert.Equal(t, "fresh", w.Payload().Text)
}

func TestHandleResultHidesOnShowFalseOrEmptyText(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.OnBufferChanged("", 0, "x", 1)

	changed := w.HandleResult(ResultMsg{
		RequestID: 1,
		Payload:   Payload{Show: true, Text: "signature foo(x)"},
	})
	assert.True(t, changed)
	assert.True(t, w.Visible())

	changed = w.HandleResult(ResultMsg{
		RequestID: 1,
		Payload:   Payload{Show: false, Text: "hidden"},
	})
	assert.True(t, changed)
	assert.False(t, w.Visible())

	changed = w.HandleResult(ResultMsg{
		RequestID: 1,
		Payload:   Payload{Show: true, Text: "   "},
	})
	assert.False(t, changed)
	assert.False(t, w.Visible())
}

func TestBufferChangeDoesNotHideVisibleBar(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.OnBufferChanged("", 0, "x", 1)
	w.HandleResult(ResultMsg{
		RequestID: 1,
		Payload: Payload{
			Show: true,
			Text: "old context",
		},
	})
	assert.True(t, w.Visible())

	cmd := w.OnBufferChanged("x", 1, "xy", 2)
	require.NotNil(t, cmd)
	assert.True(t, w.Visible())
}

func TestCommandRecoversFromProviderPanic(t *testing.T) {
	w := newTestWidget(&fakeProvider{panicOnGet: true})

	req := Request{
		Input:      "co",
		CursorByte: 2,
		Reason:     ReasonShortcut,
		RequestID:  1,
	}
	msg, ok := w.CommandForRequest(context.Background(), req)().(ResultMsg)
	require.True(t, ok)
	assert.Equal(t, req.RequestID, msg.RequestID)
	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "context bar provider panic")
}

func TestCommandTimesOutSlowProvider(t *testing.T) {
	provider := &fakeProvider{
		delay:   10 * time.Millisecond,
		payload: Payload{Show: true, Text: "late"},
	}
	w := New(provider, time.Nanosecond, time.Millisecond)

	req := Request{
		Input:      "co",
		CursorByte: 2,
		Reason:     ReasonDebounce,
		RequestID:  1,
	}
	msg, ok := w.CommandForRequest(context.Background(), req)().(ResultMsg)
	require.True(t, ok)
	require.Error(t, msg.Err)
	assert.ErrorIs(t, msg.Err, context.DeadlineExceeded)
}

func TestRenderUsesRendererAndVisibility(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	assert.Equal(t, "", w.Render(nil))

	w.OnBufferChanged("", 0, "x", 1)
	w.HandleResult(ResultMsg{
		RequestID: 1,
		Payload: Payload{
			Show:     true,
			Text:     "console.log",
			Severity: "info",
		},
	})

	rendered := w.Render(func(severity string, text string) string {
		return severity + ":" + text
	})
	assert.Equal(t, "info:console.log", rendered)
}

func TestTriggerNowUsesShortcutReason(t *testing.T) {
	provider := &fakeProvider{payload: Payload{Show: true, Text: "ok"}}
	w := newTestWidget(provider)

	cmd := w.TriggerNow(context.Background(), "co", 2, ReasonShortcut, "f2")
	require.NotNil(t, cmd)
	msg, ok := cmd().(ResultMsg)
	require.True(t, ok)
	assert.Equal(t, uint64(1), msg.RequestID)
	require.Len(t, provider.requests, 1)
	assert.Equal(t, ReasonShortcut, provider.requests[0].Reason)
	assert.Equal(t, "f2", provider.requests[0].Shortcut)
}

func TestDebounceMsgType(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	cmd := w.OnBufferChanged("", 0, "a", 1)
	require.NotNil(t, cmd)
	_, ok := cmd().(DebounceMsg)
	assert.True(t, ok)
}

func TestHandleDebounceNilProvider(t *testing.T) {
	w := New(nil, time.Nanosecond, time.Millisecond)
	cmd := w.HandleDebounce(context.Background(), DebounceMsg{RequestID: 1}, "a", 1)
	assert.Nil(t, cmd)
}

func TestCommandForRequestReturnsResultMsg(t *testing.T) {
	w := newTestWidget(&fakeProvider{payload: Payload{Show: true, Text: "ok"}})
	msg := w.CommandForRequest(context.Background(), Request{
		Input:      "x",
		CursorByte: 1,
		Reason:     ReasonManual,
		RequestID:  1,
	})()
	_, ok := msg.(ResultMsg)
	assert.True(t, ok)
}
