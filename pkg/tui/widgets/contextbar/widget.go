package contextbar

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/tui/asyncprovider"
)

// Widget manages debounce/request/visibility state for a one-line context bar.
type Widget struct {
	provider Provider

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	visible bool
	payload Payload

	lastErr     error
	lastReqID   uint64
	lastReqKind Reason
}

func New(provider Provider, debounce time.Duration, reqTimeout time.Duration) *Widget {
	return &Widget{
		provider:   provider,
		debounce:   debounce,
		reqTimeout: reqTimeout,
	}
}

func (w *Widget) Provider() Provider {
	return w.provider
}

func (w *Widget) Debounce() time.Duration {
	return w.debounce
}

func (w *Widget) RequestTimeout() time.Duration {
	return w.reqTimeout
}

func (w *Widget) SetRequestTimeout(timeout time.Duration) {
	w.reqTimeout = timeout
}

func (w *Widget) RequestSeq() uint64 {
	return w.reqSeq
}

func (w *Widget) SetRequestSeq(seq uint64) {
	w.reqSeq = seq
}

func (w *Widget) LastRequestID() uint64 {
	return w.lastReqID
}

func (w *Widget) LastRequestReason() Reason {
	return w.lastReqKind
}

func (w *Widget) LastError() error {
	return w.lastErr
}

func (w *Widget) Visible() bool {
	return w.visible
}

func (w *Widget) Payload() Payload {
	return w.payload
}

func (w *Widget) OnBufferChanged(prevValue string, prevCursor int, value string, cursor int) tea.Cmd {
	if w.provider == nil {
		return nil
	}
	if prevValue == value && prevCursor == cursor {
		return nil
	}

	w.reqSeq++
	reqID := w.reqSeq
	return tea.Tick(w.debounce, func(time.Time) tea.Msg {
		return DebounceMsg{RequestID: reqID}
	})
}

func (w *Widget) HandleDebounce(ctx context.Context, msg DebounceMsg, value string, cursor int) tea.Cmd {
	if w.provider == nil {
		return nil
	}
	if msg.RequestID != w.reqSeq {
		return nil
	}

	req := Request{
		Input:      value,
		CursorByte: cursor,
		Reason:     ReasonDebounce,
		RequestID:  msg.RequestID,
	}
	return w.CommandForRequest(ctx, req)
}

func (w *Widget) TriggerNow(ctx context.Context, value string, cursor int, reason Reason, shortcut string) tea.Cmd {
	if w.provider == nil {
		return nil
	}

	w.reqSeq++
	req := Request{
		Input:      value,
		CursorByte: cursor,
		Reason:     reason,
		Shortcut:   shortcut,
		RequestID:  w.reqSeq,
	}
	return w.CommandForRequest(ctx, req)
}

func (w *Widget) CommandForRequest(ctx context.Context, req Request) tea.Cmd {
	if w.provider == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	w.lastReqID = req.RequestID
	w.lastReqKind = req.Reason
	return func() tea.Msg {
		payload, err := asyncprovider.Run(
			ctx,
			req.RequestID,
			w.reqTimeout,
			"context-bar-provider",
			"context bar provider",
			func(runCtx context.Context) (Payload, error) {
				return w.provider.GetContextBar(runCtx, req)
			},
		)

		return ResultMsg{
			RequestID: req.RequestID,
			Payload:   payload,
			Err:       err,
		}
	}
}

// HandleResult applies a provider result and reports whether visibility changed.
func (w *Widget) HandleResult(msg ResultMsg) bool {
	if msg.RequestID != w.reqSeq {
		return false
	}
	prevVisible := w.visible
	w.lastReqID = msg.RequestID
	w.lastErr = msg.Err
	if msg.Err != nil {
		w.visible = false
		return prevVisible != w.visible
	}
	if !msg.Payload.Show || strings.TrimSpace(msg.Payload.Text) == "" {
		w.visible = false
		return prevVisible != w.visible
	}

	w.payload = msg.Payload
	w.visible = true
	return prevVisible != w.visible
}

// Hide forces the widget invisible and reports whether visibility changed.
func (w *Widget) Hide() bool {
	prevVisible := w.visible
	w.visible = false
	return prevVisible != w.visible
}

func (w *Widget) Render(renderer func(severity string, text string) string) string {
	if !w.visible {
		return ""
	}
	if renderer == nil {
		return w.payload.Text
	}
	return renderer(w.payload.Severity, w.payload.Text)
}
