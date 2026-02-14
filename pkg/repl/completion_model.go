package repl

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/tui/widgets/suggest"
	"time"
)

type completionModel struct {
	provider InputCompleter

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
	placement  CompletionOverlayPlacement
	horizontal CompletionOverlayHorizontalGrow

	lastResult  CompletionResult
	lastError   error
	lastReqID   uint64
	lastReqKind CompletionReason

	widget *suggest.Widget
}

type completionBufferAdapter struct {
	model *Model
}

func (a completionBufferAdapter) Value() string {
	return a.model.textInput.Value()
}

func (a completionBufferAdapter) CursorByte() int {
	return a.model.textInput.Position()
}

func (a completionBufferAdapter) SetValue(value string) {
	a.model.textInput.SetValue(value)
}

func (a completionBufferAdapter) SetCursorByte(cursor int) {
	a.model.textInput.SetCursor(cursor)
}

func (m *Model) ensureCompletionWidget() {
	if m.completion.widget == nil {
		m.completion.widget = suggest.New(m.completion.provider, suggest.Config{
			Debounce:       m.completion.debounce,
			RequestTimeout: m.completion.reqTimeout,
			MaxVisible:     m.completion.maxVisible,
			PageSize:       m.completion.pageSize,
			MaxWidth:       m.completion.maxWidth,
			MaxHeight:      m.completion.maxHeight,
			MinWidth:       m.completion.minWidth,
			Margin:         m.completion.margin,
			OffsetX:        m.completion.offsetX,
			OffsetY:        m.completion.offsetY,
			NoBorder:       m.completion.noBorder,
			Placement:      suggest.Placement(m.completion.placement),
			HorizontalGrow: suggest.HorizontalGrow(m.completion.horizontal),
		})
	}
	m.syncCompletionWidgetFromLegacy()
}

func (m *Model) syncCompletionWidgetFromLegacy() {
	if m.completion.widget == nil {
		return
	}
	w := m.completion.widget
	w.SetVisible(m.completion.visible)
	w.SetSelection(m.completion.selection)
	w.SetReplaceFrom(m.completion.replaceFrom)
	w.SetReplaceTo(m.completion.replaceTo)
	w.SetScrollTop(m.completion.scrollTop)
	w.SetVisibleRows(m.completion.visibleRows)
	w.SetMaxVisible(m.completion.maxVisible)
	w.SetPageSize(m.completion.pageSize)
	w.SetMaxWidth(m.completion.maxWidth)
	w.SetMaxHeight(m.completion.maxHeight)
	w.SetMinWidth(m.completion.minWidth)
	w.SetMargin(m.completion.margin)
	w.SetOffsetX(m.completion.offsetX)
	w.SetOffsetY(m.completion.offsetY)
	w.SetNoBorder(m.completion.noBorder)
	w.SetPlacement(suggest.Placement(m.completion.placement))
	w.SetHorizontalGrow(suggest.HorizontalGrow(m.completion.horizontal))
	w.SetLastResult(m.completion.lastResult)
	w.SetRequestSeq(m.completion.reqSeq)
	w.SetRequestTimeout(m.completion.reqTimeout)
}

func (m *Model) syncCompletionLegacyFromWidget() {
	if m.completion.widget == nil {
		return
	}
	w := m.completion.widget
	m.completion.visible = w.Visible()
	m.completion.selection = w.Selection()
	m.completion.replaceFrom = w.ReplaceFrom()
	m.completion.replaceTo = w.ReplaceTo()
	m.completion.scrollTop = w.ScrollTop()
	m.completion.visibleRows = w.VisibleRows()
	m.completion.maxVisible = w.MaxVisible()
	m.completion.pageSize = w.PageSize()
	m.completion.maxWidth = w.MaxWidth()
	m.completion.maxHeight = w.MaxHeight()
	m.completion.minWidth = w.MinWidth()
	m.completion.margin = w.Margin()
	m.completion.offsetX = w.OffsetX()
	m.completion.offsetY = w.OffsetY()
	m.completion.noBorder = w.NoBorder()
	m.completion.placement = CompletionOverlayPlacement(w.Placement())
	m.completion.horizontal = CompletionOverlayHorizontalGrow(w.HorizontalGrow())
	m.completion.lastResult = w.LastResult()
	m.completion.lastError = w.LastError()
	m.completion.lastReqID = w.LastRequestID()
	m.completion.lastReqKind = CompletionReason(w.LastRequestReason())
	m.completion.reqSeq = w.RequestSeq()
	m.completion.reqTimeout = w.RequestTimeout()
}

func (m *Model) scheduleDebouncedCompletionIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return nil
	}
	cmd := m.completion.widget.OnBufferChanged(prevValue, prevCursor, m.textInput.Value(), m.textInput.Position())
	m.syncCompletionLegacyFromWidget()
	return cmd
}

func (m *Model) handleDebouncedCompletion(msg completionDebounceMsg) tea.Cmd {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return nil
	}
	cmd := m.completion.widget.HandleDebounce(m.appContext(), msg, m.textInput.Value(), m.textInput.Position())
	m.syncCompletionLegacyFromWidget()
	return cmd
}

func (m *Model) triggerCompletionFromShortcut(k tea.KeyMsg) tea.Cmd {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return nil
	}
	if !key.Matches(k, m.keyMap.CompletionTrigger) {
		return nil
	}
	cmd := m.completion.widget.TriggerShortcut(m.appContext(), m.textInput.Value(), m.textInput.Position(), k.String())
	m.syncCompletionLegacyFromWidget()
	return cmd
}

func (m *Model) handleCompletionResult(msg completionResultMsg) tea.Cmd {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return nil
	}
	m.completion.widget.HandleResult(msg)
	m.syncCompletionLegacyFromWidget()
	return nil
}

func (m *Model) handleCompletionNavigation(k tea.KeyMsg) (bool, tea.Cmd) {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return false, nil
	}
	if !m.completion.widget.Visible() {
		return false, nil
	}

	buffer := completionBufferAdapter{model: m}
	switch {
	case key.Matches(k, m.keyMap.CompletionCancel):
		handled := m.completion.widget.HandleNavigation(suggest.ActionCancel, buffer)
		m.syncCompletionLegacyFromWidget()
		return handled, nil
	case key.Matches(k, m.keyMap.CompletionPrev):
		handled := m.completion.widget.HandleNavigation(suggest.ActionPrev, buffer)
		m.syncCompletionLegacyFromWidget()
		return handled, nil
	case key.Matches(k, m.keyMap.CompletionNext):
		handled := m.completion.widget.HandleNavigation(suggest.ActionNext, buffer)
		m.syncCompletionLegacyFromWidget()
		return handled, nil
	case key.Matches(k, m.keyMap.CompletionPageUp):
		handled := m.completion.widget.HandleNavigation(suggest.ActionPageUp, buffer)
		m.syncCompletionLegacyFromWidget()
		return handled, nil
	case key.Matches(k, m.keyMap.CompletionPageDown):
		handled := m.completion.widget.HandleNavigation(suggest.ActionPageDown, buffer)
		m.syncCompletionLegacyFromWidget()
		return handled, nil
	case key.Matches(k, m.keyMap.CompletionAccept):
		handled := m.completion.widget.HandleNavigation(suggest.ActionAccept, buffer)
		m.syncCompletionLegacyFromWidget()
		return handled, nil
	}
	return false, nil
}

func (m *Model) ensureCompletionSelectionVisible() {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return
	}
	m.syncCompletionWidgetFromLegacy()
	m.completion.widget.EnsureSelectionVisible()
	m.syncCompletionLegacyFromWidget()
}
