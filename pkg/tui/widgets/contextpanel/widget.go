package contextpanel

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/tui/asyncprovider"
)

type Widget struct {
	provider Provider

	visible bool
	doc     Document

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	loading bool
	err     error
	pinned  bool

	prefetch      bool
	dock          Dock
	widthPercent  int
	heightPercent int
	margin        int
}

func New(provider Provider, cfg Config) *Widget {
	return &Widget{
		provider:      provider,
		debounce:      cfg.Debounce,
		reqTimeout:    cfg.RequestTimeout,
		prefetch:      cfg.PrefetchWhenHidden,
		dock:          cfg.Dock,
		widthPercent:  cfg.WidthPercent,
		heightPercent: cfg.HeightPercent,
		margin:        cfg.Margin,
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

func (w *Widget) Visible() bool {
	return w.visible
}

func (w *Widget) SetVisible(visible bool) {
	w.visible = visible
}

func (w *Widget) Loading() bool {
	return w.loading
}

func (w *Widget) SetLoading(loading bool) {
	w.loading = loading
}

func (w *Widget) Err() error {
	return w.err
}

func (w *Widget) SetErr(err error) {
	w.err = err
}

func (w *Widget) Pinned() bool {
	return w.pinned
}

func (w *Widget) SetPinned(pinned bool) {
	w.pinned = pinned
}

func (w *Widget) Document() Document {
	return w.doc
}

func (w *Widget) SetDocument(doc Document) {
	w.doc = doc
}

func (w *Widget) PrefetchWhenHidden() bool {
	return w.prefetch
}

func (w *Widget) SetPrefetchWhenHidden(prefetch bool) {
	w.prefetch = prefetch
}

func (w *Widget) Dock() Dock {
	return w.dock
}

func (w *Widget) SetDock(dock Dock) {
	w.dock = dock
}

func (w *Widget) WidthPercent() int {
	return w.widthPercent
}

func (w *Widget) SetWidthPercent(widthPercent int) {
	w.widthPercent = widthPercent
}

func (w *Widget) HeightPercent() int {
	return w.heightPercent
}

func (w *Widget) SetHeightPercent(heightPercent int) {
	w.heightPercent = heightPercent
}

func (w *Widget) Margin() int {
	return w.margin
}

func (w *Widget) SetMargin(margin int) {
	w.margin = margin
}

func (w *Widget) OnBufferChanged(prevValue string, prevCursor int, value string, cursor int) tea.Cmd {
	if w.provider == nil {
		return nil
	}
	if !w.visible && !w.prefetch {
		return nil
	}
	if w.pinned {
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
	if !w.visible && !w.prefetch {
		return nil
	}

	req := Request{
		Input:      value,
		CursorByte: cursor,
		RequestID:  msg.RequestID,
		Trigger:    TriggerTyping,
	}
	w.loading = true
	return w.CommandForRequest(ctx, req)
}

func (w *Widget) Toggle(ctx context.Context, value string, cursor int) tea.Cmd {
	if w.visible {
		w.Close()
		return nil
	}
	w.visible = true
	w.err = nil
	return w.RequestNow(ctx, value, cursor, TriggerToggleOpen)
}

func (w *Widget) Close() {
	w.visible = false
	w.loading = false
}

func (w *Widget) TogglePin() bool {
	w.pinned = !w.pinned
	return w.pinned
}

func (w *Widget) RequestNow(ctx context.Context, value string, cursor int, trigger Trigger) tea.Cmd {
	if w.provider == nil {
		return nil
	}
	w.loading = true
	w.err = nil
	w.reqSeq++

	req := Request{
		Input:      value,
		CursorByte: cursor,
		RequestID:  w.reqSeq,
		Trigger:    trigger,
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
	return func() tea.Msg {
		doc, err := asyncprovider.Run(
			ctx,
			req.RequestID,
			w.reqTimeout,
			"context-panel-provider",
			"context panel provider",
			func(runCtx context.Context) (Document, error) {
				return w.provider.GetContextPanel(runCtx, req)
			},
		)

		return ResultMsg{
			RequestID: req.RequestID,
			Doc:       doc,
			Err:       err,
		}
	}
}

func (w *Widget) HandleResult(msg ResultMsg) {
	if msg.RequestID != w.reqSeq {
		return
	}
	w.loading = false
	w.err = msg.Err
	if msg.Err != nil {
		return
	}
	w.doc = msg.Doc
}
