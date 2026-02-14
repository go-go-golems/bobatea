package suggest

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/tui/asyncprovider"
)

type Widget struct {
	provider Provider

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	visible     bool
	selection   int
	replaceFrom int
	replaceTo   int
	scrollTop   int
	visibleRows int

	maxVisible int
	pageSize   int
	maxWidth   int
	maxHeight  int
	minWidth   int
	margin     int
	offsetX    int
	offsetY    int
	noBorder   bool
	placement  Placement
	horizontal HorizontalGrow

	lastResult   Result
	lastError    error
	lastReqID    uint64
	lastReqKind  Reason
	lastInputLen int
}

func New(provider Provider, cfg Config) *Widget {
	return &Widget{
		provider:   provider,
		debounce:   cfg.Debounce,
		reqTimeout: cfg.RequestTimeout,
		maxVisible: cfg.MaxVisible,
		pageSize:   cfg.PageSize,
		maxWidth:   cfg.MaxWidth,
		maxHeight:  cfg.MaxHeight,
		minWidth:   cfg.MinWidth,
		margin:     cfg.Margin,
		offsetX:    cfg.OffsetX,
		offsetY:    cfg.OffsetY,
		noBorder:   cfg.NoBorder,
		placement:  cfg.Placement,
		horizontal: cfg.HorizontalGrow,
	}
}

func (w *Widget) Provider() Provider                          { return w.provider }
func (w *Widget) Debounce() time.Duration                     { return w.debounce }
func (w *Widget) RequestTimeout() time.Duration               { return w.reqTimeout }
func (w *Widget) SetRequestTimeout(timeout time.Duration)     { w.reqTimeout = timeout }
func (w *Widget) RequestSeq() uint64                          { return w.reqSeq }
func (w *Widget) SetRequestSeq(seq uint64)                    { w.reqSeq = seq }
func (w *Widget) Visible() bool                               { return w.visible }
func (w *Widget) SetVisible(visible bool)                     { w.visible = visible }
func (w *Widget) Selection() int                              { return w.selection }
func (w *Widget) SetSelection(selection int)                  { w.selection = selection }
func (w *Widget) ReplaceFrom() int                            { return w.replaceFrom }
func (w *Widget) SetReplaceFrom(replaceFrom int)              { w.replaceFrom = replaceFrom }
func (w *Widget) ReplaceTo() int                              { return w.replaceTo }
func (w *Widget) SetReplaceTo(replaceTo int)                  { w.replaceTo = replaceTo }
func (w *Widget) ScrollTop() int                              { return w.scrollTop }
func (w *Widget) SetScrollTop(scrollTop int)                  { w.scrollTop = scrollTop }
func (w *Widget) VisibleRows() int                            { return w.visibleRows }
func (w *Widget) SetVisibleRows(visibleRows int)              { w.visibleRows = visibleRows }
func (w *Widget) MaxVisible() int                             { return w.maxVisible }
func (w *Widget) SetMaxVisible(maxVisible int)                { w.maxVisible = maxVisible }
func (w *Widget) PageSize() int                               { return w.pageSize }
func (w *Widget) SetPageSize(pageSize int)                    { w.pageSize = pageSize }
func (w *Widget) MaxWidth() int                               { return w.maxWidth }
func (w *Widget) SetMaxWidth(maxWidth int)                    { w.maxWidth = maxWidth }
func (w *Widget) MaxHeight() int                              { return w.maxHeight }
func (w *Widget) SetMaxHeight(maxHeight int)                  { w.maxHeight = maxHeight }
func (w *Widget) MinWidth() int                               { return w.minWidth }
func (w *Widget) SetMinWidth(minWidth int)                    { w.minWidth = minWidth }
func (w *Widget) Margin() int                                 { return w.margin }
func (w *Widget) SetMargin(margin int)                        { w.margin = margin }
func (w *Widget) OffsetX() int                                { return w.offsetX }
func (w *Widget) SetOffsetX(offsetX int)                      { w.offsetX = offsetX }
func (w *Widget) OffsetY() int                                { return w.offsetY }
func (w *Widget) SetOffsetY(offsetY int)                      { w.offsetY = offsetY }
func (w *Widget) NoBorder() bool                              { return w.noBorder }
func (w *Widget) SetNoBorder(noBorder bool)                   { w.noBorder = noBorder }
func (w *Widget) Placement() Placement                        { return w.placement }
func (w *Widget) SetPlacement(placement Placement)            { w.placement = placement }
func (w *Widget) HorizontalGrow() HorizontalGrow              { return w.horizontal }
func (w *Widget) SetHorizontalGrow(horizontal HorizontalGrow) { w.horizontal = horizontal }
func (w *Widget) LastResult() Result                          { return w.lastResult }
func (w *Widget) SetLastResult(result Result)                 { w.lastResult = result }
func (w *Widget) LastError() error                            { return w.lastError }
func (w *Widget) LastRequestID() uint64                       { return w.lastReqID }
func (w *Widget) LastRequestReason() Reason                   { return w.lastReqKind }

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
	w.lastReqID = req.RequestID
	w.lastReqKind = req.Reason
	w.lastInputLen = len(req.Input)
	return w.CommandForRequest(ctx, req)
}

func (w *Widget) TriggerShortcut(ctx context.Context, value string, cursor int, shortcut string) tea.Cmd {
	if w.provider == nil {
		return nil
	}
	w.reqSeq++
	req := Request{
		Input:      value,
		CursorByte: cursor,
		Reason:     ReasonShortcut,
		Shortcut:   shortcut,
		RequestID:  w.reqSeq,
	}
	w.lastReqID = req.RequestID
	w.lastReqKind = req.Reason
	w.lastInputLen = len(req.Input)
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
	w.lastInputLen = len(req.Input)

	return func() tea.Msg {
		result, err := asyncprovider.Run(
			ctx,
			req.RequestID,
			w.reqTimeout,
			"suggest-provider",
			"suggest provider",
			func(runCtx context.Context) (Result, error) {
				return w.provider.CompleteInput(runCtx, req)
			},
		)

		return ResultMsg{
			RequestID: req.RequestID,
			Result:    result,
			Err:       err,
		}
	}
}

func (w *Widget) HandleResult(msg ResultMsg) {
	if msg.RequestID != w.reqSeq {
		return
	}
	w.lastReqID = msg.RequestID
	w.lastResult = msg.Result
	w.lastError = msg.Err
	if msg.Err != nil || !msg.Result.Show || len(msg.Result.Suggestions) == 0 {
		w.Hide()
		return
	}

	w.selection = 0
	w.visible = true
	w.scrollTop = 0
	w.visibleRows = 0
	w.replaceFrom = clampInt(msg.Result.ReplaceFrom, 0, w.lastInputLen)
	w.replaceTo = clampInt(msg.Result.ReplaceTo, w.replaceFrom, w.lastInputLen)
	w.ensureSelectionVisible()
}

func (w *Widget) HandleNavigation(action Action, buffer Buffer) bool {
	if !w.visible {
		return false
	}
	suggestions := w.lastResult.Suggestions
	if len(suggestions) == 0 {
		w.Hide()
		return false
	}

	switch action {
	case ActionCancel:
		w.Hide()
		return true
	case ActionPrev:
		if w.selection > 0 {
			w.selection--
		}
		w.ensureSelectionVisible()
		return true
	case ActionNext:
		if w.selection < len(suggestions)-1 {
			w.selection++
		}
		w.ensureSelectionVisible()
		return true
	case ActionPageUp:
		if w.selection > 0 {
			w.selection = maxInt(0, w.selection-w.pageStep())
		}
		w.ensureSelectionVisible()
		return true
	case ActionPageDown:
		if w.selection < len(suggestions)-1 {
			w.selection = minInt(len(suggestions)-1, w.selection+w.pageStep())
		}
		w.ensureSelectionVisible()
		return true
	case ActionAccept:
		w.ApplySelected(buffer)
		return true
	default:
		return false
	}
}

func (w *Widget) ApplySelected(buffer Buffer) {
	suggestions := w.lastResult.Suggestions
	if len(suggestions) == 0 || w.selection >= len(suggestions) {
		w.Hide()
		return
	}
	selected := suggestions[w.selection]
	input := buffer.Value()
	from := clampInt(w.replaceFrom, 0, len(input))
	to := clampInt(w.replaceTo, from, len(input))
	newInput := input[:from] + selected.Value + input[to:]
	buffer.SetValue(newInput)
	buffer.SetCursorByte(from + len(selected.Value))
	w.Hide()
}

func (w *Widget) Hide() {
	w.visible = false
	w.selection = 0
	w.replaceFrom = 0
	w.replaceTo = 0
	w.scrollTop = 0
	w.visibleRows = 0
}

func (w *Widget) EnsureSelectionVisible() {
	w.ensureSelectionVisible()
}

func (w *Widget) visibleLimit() int {
	if w.visibleRows > 0 {
		return maxInt(1, w.visibleRows)
	}
	if w.maxVisible > 0 {
		return w.maxVisible
	}
	return 1
}

func (w *Widget) pageStep() int {
	if w.pageSize > 0 {
		return maxInt(1, w.pageSize)
	}
	return w.visibleLimit()
}

func (w *Widget) ensureSelectionVisible() {
	suggestions := w.lastResult.Suggestions
	if len(suggestions) == 0 {
		w.scrollTop = 0
		return
	}

	w.selection = clampInt(w.selection, 0, len(suggestions)-1)
	limit := w.visibleLimit()
	maxTop := maxInt(0, len(suggestions)-limit)
	w.scrollTop = clampInt(w.scrollTop, 0, maxTop)
	if w.selection < w.scrollTop {
		w.scrollTop = w.selection
	}
	visibleEnd := w.scrollTop + limit - 1
	if w.selection > visibleEnd {
		w.scrollTop = w.selection - limit + 1
	}
	w.scrollTop = clampInt(w.scrollTop, 0, maxTop)
}

func clampInt(value int, lower int, upper int) int {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
