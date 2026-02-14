package repl

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
}

func (m *Model) scheduleDebouncedCompletionIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	if m.completion.provider == nil {
		return nil
	}
	if prevValue == m.textInput.Value() && prevCursor == m.textInput.Position() {
		return nil
	}

	m.completion.reqSeq++
	reqID := m.completion.reqSeq
	return tea.Tick(m.completion.debounce, func(time.Time) tea.Msg {
		return completionDebounceMsg{RequestID: reqID}
	})
}

func (m *Model) handleDebouncedCompletion(msg completionDebounceMsg) tea.Cmd {
	if m.completion.provider == nil {
		return nil
	}
	if msg.RequestID != m.completion.reqSeq {
		return nil
	}

	req := CompletionRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     CompletionReasonDebounce,
		RequestID:  msg.RequestID,
	}
	m.completion.lastReqID = req.RequestID
	m.completion.lastReqKind = req.Reason
	return m.completionCmd(req)
}

func (m *Model) triggerCompletionFromShortcut(k tea.KeyMsg) tea.Cmd {
	if m.completion.provider == nil {
		return nil
	}
	if !key.Matches(k, m.keyMap.CompletionTrigger) {
		return nil
	}
	keyStr := k.String()

	m.completion.reqSeq++
	req := CompletionRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     CompletionReasonShortcut,
		Shortcut:   keyStr,
		RequestID:  m.completion.reqSeq,
	}
	m.completion.lastReqID = req.RequestID
	m.completion.lastReqKind = req.Reason
	return m.completionCmd(req)
}

func (m *Model) handleCompletionResult(msg completionResultMsg) tea.Cmd {
	if msg.RequestID != m.completion.reqSeq {
		return nil
	}
	m.completion.lastReqID = msg.RequestID
	m.completion.lastResult = msg.Result
	m.completion.lastError = msg.Err
	if msg.Err != nil || !msg.Result.Show || len(msg.Result.Suggestions) == 0 {
		m.hideCompletionPopup()
		return nil
	}

	m.completion.selection = 0
	m.completion.visible = true
	m.completion.scrollTop = 0
	m.completion.visibleRows = 0
	m.completion.replaceFrom = clampInt(msg.Result.ReplaceFrom, 0, len(m.textInput.Value()))
	m.completion.replaceTo = clampInt(msg.Result.ReplaceTo, m.completion.replaceFrom, len(m.textInput.Value()))
	m.ensureCompletionSelectionVisible()
	return nil
}

func (m *Model) handleCompletionNavigation(k tea.KeyMsg) (bool, tea.Cmd) {
	if !m.completion.visible {
		return false, nil
	}

	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		m.hideCompletionPopup()
		return false, nil
	}

	switch {
	case key.Matches(k, m.keyMap.CompletionCancel):
		m.hideCompletionPopup()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionPrev):
		if m.completion.selection > 0 {
			m.completion.selection--
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionNext):
		if m.completion.selection < len(suggestions)-1 {
			m.completion.selection++
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionPageUp):
		if m.completion.selection > 0 {
			m.completion.selection = max(0, m.completion.selection-m.completionPageStep())
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionPageDown):
		if m.completion.selection < len(suggestions)-1 {
			m.completion.selection = min(len(suggestions)-1, m.completion.selection+m.completionPageStep())
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionAccept):
		m.applySelectedCompletion()
		return true, nil
	}
	return false, nil
}

func (m *Model) applySelectedCompletion() {
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 || m.completion.selection >= len(suggestions) {
		m.hideCompletionPopup()
		return
	}

	selected := suggestions[m.completion.selection]
	input := m.textInput.Value()
	from := clampInt(m.completion.replaceFrom, 0, len(input))
	to := clampInt(m.completion.replaceTo, from, len(input))
	newInput := input[:from] + selected.Value + input[to:]

	m.textInput.SetValue(newInput)
	m.textInput.SetCursor(from + len(selected.Value))
	m.hideCompletionPopup()
}

func (m *Model) hideCompletionPopup() {
	m.completion.visible = false
	m.completion.selection = 0
	m.completion.replaceFrom = 0
	m.completion.replaceTo = 0
	m.completion.scrollTop = 0
	m.completion.visibleRows = 0
}

func (m *Model) completionVisibleLimit() int {
	if m.completion.visibleRows > 0 {
		return max(1, m.completion.visibleRows)
	}
	if m.completion.maxVisible > 0 {
		return m.completion.maxVisible
	}
	return 1
}

func (m *Model) completionPageStep() int {
	if m.completion.pageSize > 0 {
		return max(1, m.completion.pageSize)
	}
	return m.completionVisibleLimit()
}

func (m *Model) ensureCompletionSelectionVisible() {
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		m.completion.scrollTop = 0
		return
	}

	m.completion.selection = clampInt(m.completion.selection, 0, len(suggestions)-1)
	limit := m.completionVisibleLimit()
	maxTop := max(0, len(suggestions)-limit)
	m.completion.scrollTop = clampInt(m.completion.scrollTop, 0, maxTop)
	if m.completion.selection < m.completion.scrollTop {
		m.completion.scrollTop = m.completion.selection
	}
	visibleEnd := m.completion.scrollTop + limit - 1
	if m.completion.selection > visibleEnd {
		m.completion.scrollTop = m.completion.selection - limit + 1
	}
	m.completion.scrollTop = clampInt(m.completion.scrollTop, 0, maxTop)
}
