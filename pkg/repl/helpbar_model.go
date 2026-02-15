package repl

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/tui/widgets/contextbar"
)

type helpBarModel struct {
	provider HelpBarProvider

	debounce   time.Duration
	reqTimeout time.Duration

	// Legacy mirrored state keeps existing tests/host access stable while
	// routing runtime behavior through the extracted widget.
	visible bool
	payload HelpBarPayload

	widget *contextbar.Widget
}

type helpBarProviderAdapter struct {
	provider HelpBarProvider
}

func (a helpBarProviderAdapter) GetContextBar(ctx context.Context, req contextbar.Request) (contextbar.Payload, error) {
	return a.provider.GetHelpBar(ctx, req)
}

func newHelpBarModel(provider HelpBarProvider, cfg HelpBarConfig) helpBarModel {
	var adapted contextbar.Provider
	if provider != nil {
		adapted = helpBarProviderAdapter{provider: provider}
	}
	return helpBarModel{
		provider:   provider,
		debounce:   cfg.Debounce,
		reqTimeout: cfg.RequestTimeout,
		widget:     contextbar.New(adapted, cfg.Debounce, cfg.RequestTimeout),
	}
}

func (m *Model) ensureHelpBarWidget() {
	if m.helpBar.widget != nil {
		m.syncHelpBarLegacyState()
		return
	}
	var adapted contextbar.Provider
	if m.helpBar.provider != nil {
		adapted = helpBarProviderAdapter{provider: m.helpBar.provider}
	}
	m.helpBar.widget = contextbar.New(adapted, m.helpBar.debounce, m.helpBar.reqTimeout)
	if m.helpBar.visible {
		m.helpBar.widget.HandleResult(contextbar.ResultMsg{
			RequestID: m.helpBar.widget.RequestSeq(),
			Payload:   m.helpBar.payload,
		})
	}
	m.syncHelpBarLegacyState()
}

func (m *Model) syncHelpBarLegacyState() {
	if m.helpBar.widget == nil {
		return
	}
	m.helpBar.visible = m.helpBar.widget.Visible()
	m.helpBar.payload = m.helpBar.widget.Payload()
}

func (m *Model) scheduleDebouncedHelpBarIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	m.ensureHelpBarWidget()
	if m.helpBar.widget == nil {
		return nil
	}
	return m.helpBar.widget.OnBufferChanged(prevValue, prevCursor, m.textInput.Value(), m.textInput.Position())
}

func (m *Model) handleDebouncedHelpBar(msg helpBarDebounceMsg) tea.Cmd {
	m.ensureHelpBarWidget()
	if m.helpBar.widget == nil {
		return nil
	}
	return m.helpBar.widget.HandleDebounce(m.appContext(), msg, m.textInput.Value(), m.textInput.Position())
}

func (m *Model) handleHelpBarResult(msg helpBarResultMsg) tea.Cmd {
	m.ensureHelpBarWidget()
	if m.helpBar.widget == nil {
		return nil
	}
	visibilityChanged := m.helpBar.widget.HandleResult(msg)
	m.syncHelpBarLegacyState()
	if visibilityChanged {
		m.applyLayoutAndRefresh()
	}
	return nil
}

func (m *Model) renderHelpBar() string {
	m.ensureHelpBarWidget()
	if m.helpBar.widget == nil {
		return ""
	}
	ret := m.helpBar.widget.Render(func(severity string, text string) string {
		return m.helpBarStyleForSeverity(severity).Render(text)
	})
	m.syncHelpBarLegacyState()
	return ret
}

func (m *Model) hideHelpBar() {
	m.ensureHelpBarWidget()
	if m.helpBar.widget == nil {
		return
	}
	visibilityChanged := m.helpBar.widget.Hide()
	m.syncHelpBarLegacyState()
	if visibilityChanged {
		m.applyLayoutAndRefresh()
	}
}

func (m *Model) helpBarCmd(req HelpBarRequest) tea.Cmd {
	m.ensureHelpBarWidget()
	if m.helpBar.widget == nil {
		return nil
	}
	return m.helpBar.widget.CommandForRequest(m.appContext(), req)
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
