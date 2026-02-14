package repl

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpBarModel struct {
	provider HelpBarProvider

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	visible bool
	payload HelpBarPayload

	lastErr     error
	lastReqID   uint64
	lastReqKind HelpBarReason
}

func (m *Model) scheduleDebouncedHelpBarIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	if m.helpBar.provider == nil {
		return nil
	}
	if prevValue == m.textInput.Value() && prevCursor == m.textInput.Position() {
		return nil
	}

	m.helpBar.reqSeq++
	reqID := m.helpBar.reqSeq
	return tea.Tick(m.helpBar.debounce, func(time.Time) tea.Msg {
		return helpBarDebounceMsg{RequestID: reqID}
	})
}

func (m *Model) handleDebouncedHelpBar(msg helpBarDebounceMsg) tea.Cmd {
	if m.helpBar.provider == nil {
		return nil
	}
	if msg.RequestID != m.helpBar.reqSeq {
		return nil
	}

	req := HelpBarRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     HelpBarReasonDebounce,
		RequestID:  msg.RequestID,
	}
	m.helpBar.lastReqID = req.RequestID
	m.helpBar.lastReqKind = req.Reason
	return m.helpBarCmd(req)
}

func (m *Model) handleHelpBarResult(msg helpBarResultMsg) tea.Cmd {
	if msg.RequestID != m.helpBar.reqSeq {
		return nil
	}
	m.helpBar.lastReqID = msg.RequestID
	m.helpBar.lastErr = msg.Err
	if msg.Err != nil {
		m.helpBar.visible = false
		return nil
	}
	if !msg.Payload.Show || strings.TrimSpace(msg.Payload.Text) == "" {
		m.helpBar.visible = false
		return nil
	}

	m.helpBar.payload = msg.Payload
	m.helpBar.visible = true
	return nil
}

func (m *Model) renderHelpBar() string {
	if !m.helpBar.visible {
		return ""
	}
	return m.helpBarStyleForSeverity(m.helpBar.payload.Severity).Render(m.helpBar.payload.Text)
}

func (m *Model) helpBarStyleForSeverity(severity string) lipgloss.Style {
	switch severity {
	case "error":
		return m.styles.Error
	case "warning":
		return m.styles.HelpText
	default:
		return m.styles.Info
	}
}
