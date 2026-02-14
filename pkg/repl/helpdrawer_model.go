package repl

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type helpDrawerModel struct {
	provider HelpDrawerProvider

	visible bool
	doc     HelpDrawerDocument

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	loading bool
	err     error
	pinned  bool

	prefetch      bool
	dock          HelpDrawerDock
	widthPercent  int
	heightPercent int
	margin        int
}

func (m *Model) scheduleDebouncedHelpDrawerIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	if m.helpDrawer.provider == nil {
		return nil
	}
	if !m.helpDrawer.visible && !m.helpDrawer.prefetch {
		return nil
	}
	if m.helpDrawer.pinned {
		return nil
	}
	if prevValue == m.textInput.Value() && prevCursor == m.textInput.Position() {
		return nil
	}

	m.helpDrawer.reqSeq++
	reqID := m.helpDrawer.reqSeq
	return tea.Tick(m.helpDrawer.debounce, func(time.Time) tea.Msg {
		return helpDrawerDebounceMsg{RequestID: reqID}
	})
}

func (m *Model) handleDebouncedHelpDrawer(msg helpDrawerDebounceMsg) tea.Cmd {
	if m.helpDrawer.provider == nil {
		return nil
	}
	if msg.RequestID != m.helpDrawer.reqSeq {
		return nil
	}
	if !m.helpDrawer.visible && !m.helpDrawer.prefetch {
		return nil
	}

	req := HelpDrawerRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		RequestID:  msg.RequestID,
		Trigger:    HelpDrawerTriggerTyping,
	}
	m.helpDrawer.loading = true
	return m.helpDrawerCmd(req)
}

func (m *Model) handleHelpDrawerShortcuts(k tea.KeyMsg) (bool, tea.Cmd) {
	if m.helpDrawer.provider == nil {
		return false, nil
	}

	switch {
	case key.Matches(k, m.keyMap.HelpDrawerToggle):
		return true, m.toggleHelpDrawer()
	case m.helpDrawer.visible && key.Matches(k, m.keyMap.HelpDrawerClose):
		if m.completion.visible && key.Matches(k, m.keyMap.CompletionCancel) {
			return false, nil
		}
		m.closeHelpDrawer()
		return true, nil
	case m.helpDrawer.visible && key.Matches(k, m.keyMap.HelpDrawerRefresh):
		return true, m.requestHelpDrawerNow(HelpDrawerTriggerManualRefresh)
	case m.helpDrawer.visible && key.Matches(k, m.keyMap.HelpDrawerPin):
		m.helpDrawer.pinned = !m.helpDrawer.pinned
		return true, nil
	}

	return false, nil
}

func (m *Model) toggleHelpDrawer() tea.Cmd {
	if m.helpDrawer.visible {
		m.closeHelpDrawer()
		return nil
	}

	m.helpDrawer.visible = true
	m.helpDrawer.err = nil
	return m.requestHelpDrawerNow(HelpDrawerTriggerToggleOpen)
}

func (m *Model) closeHelpDrawer() {
	m.helpDrawer.visible = false
	m.helpDrawer.loading = false
}

func (m *Model) requestHelpDrawerNow(trigger HelpDrawerTrigger) tea.Cmd {
	if m.helpDrawer.provider == nil {
		return nil
	}
	m.helpDrawer.loading = true
	m.helpDrawer.err = nil
	m.helpDrawer.reqSeq++
	req := HelpDrawerRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		RequestID:  m.helpDrawer.reqSeq,
		Trigger:    trigger,
	}
	return m.helpDrawerCmd(req)
}

func (m *Model) handleHelpDrawerResult(msg helpDrawerResultMsg) tea.Cmd {
	if msg.RequestID != m.helpDrawer.reqSeq {
		return nil
	}
	m.helpDrawer.loading = false
	m.helpDrawer.err = msg.Err
	if msg.Err != nil {
		return nil
	}
	m.helpDrawer.doc = msg.Doc
	return nil
}
