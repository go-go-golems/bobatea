package repl

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/tui/widgets/contextpanel"
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

	widget *contextpanel.Widget
}

type helpDrawerProviderAdapter struct {
	provider HelpDrawerProvider
}

func (a helpDrawerProviderAdapter) GetContextPanel(ctx context.Context, req contextpanel.Request) (contextpanel.Document, error) {
	return a.provider.GetHelpDrawer(ctx, req)
}

func (m *Model) ensureHelpDrawerWidget() {
	if m.helpDrawer.widget == nil {
		var adapted contextpanel.Provider
		if m.helpDrawer.provider != nil {
			adapted = helpDrawerProviderAdapter{provider: m.helpDrawer.provider}
		}
		m.helpDrawer.widget = contextpanel.New(adapted, contextpanel.Config{
			Debounce:           m.helpDrawer.debounce,
			RequestTimeout:     m.helpDrawer.reqTimeout,
			PrefetchWhenHidden: m.helpDrawer.prefetch,
			Dock:               contextpanel.Dock(m.helpDrawer.dock),
			WidthPercent:       m.helpDrawer.widthPercent,
			HeightPercent:      m.helpDrawer.heightPercent,
			Margin:             m.helpDrawer.margin,
		})
	}
	m.syncHelpDrawerWidgetFromLegacy()
}

func (m *Model) syncHelpDrawerWidgetFromLegacy() {
	if m.helpDrawer.widget == nil {
		return
	}
	m.helpDrawer.widget.SetVisible(m.helpDrawer.visible)
	m.helpDrawer.widget.SetLoading(m.helpDrawer.loading)
	m.helpDrawer.widget.SetErr(m.helpDrawer.err)
	m.helpDrawer.widget.SetPinned(m.helpDrawer.pinned)
	m.helpDrawer.widget.SetDocument(m.helpDrawer.doc)
	m.helpDrawer.widget.SetRequestSeq(m.helpDrawer.reqSeq)
	m.helpDrawer.widget.SetPrefetchWhenHidden(m.helpDrawer.prefetch)
	m.helpDrawer.widget.SetDock(contextpanel.Dock(m.helpDrawer.dock))
	m.helpDrawer.widget.SetWidthPercent(m.helpDrawer.widthPercent)
	m.helpDrawer.widget.SetHeightPercent(m.helpDrawer.heightPercent)
	m.helpDrawer.widget.SetMargin(m.helpDrawer.margin)
	m.helpDrawer.widget.SetRequestTimeout(m.helpDrawer.reqTimeout)
}

func (m *Model) syncHelpDrawerLegacyFromWidget() {
	if m.helpDrawer.widget == nil {
		return
	}
	m.helpDrawer.visible = m.helpDrawer.widget.Visible()
	m.helpDrawer.loading = m.helpDrawer.widget.Loading()
	m.helpDrawer.err = m.helpDrawer.widget.Err()
	m.helpDrawer.pinned = m.helpDrawer.widget.Pinned()
	m.helpDrawer.doc = m.helpDrawer.widget.Document()
	m.helpDrawer.reqSeq = m.helpDrawer.widget.RequestSeq()
	m.helpDrawer.prefetch = m.helpDrawer.widget.PrefetchWhenHidden()
	m.helpDrawer.dock = HelpDrawerDock(m.helpDrawer.widget.Dock())
	m.helpDrawer.widthPercent = m.helpDrawer.widget.WidthPercent()
	m.helpDrawer.heightPercent = m.helpDrawer.widget.HeightPercent()
	m.helpDrawer.margin = m.helpDrawer.widget.Margin()
	m.helpDrawer.reqTimeout = m.helpDrawer.widget.RequestTimeout()
}

func (m *Model) scheduleDebouncedHelpDrawerIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil {
		return nil
	}
	cmd := m.helpDrawer.widget.OnBufferChanged(prevValue, prevCursor, m.textInput.Value(), m.textInput.Position())
	m.syncHelpDrawerLegacyFromWidget()
	return cmd
}

func (m *Model) handleDebouncedHelpDrawer(msg helpDrawerDebounceMsg) tea.Cmd {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil {
		return nil
	}
	cmd := m.helpDrawer.widget.HandleDebounce(m.appContext(), msg, m.textInput.Value(), m.textInput.Position())
	m.syncHelpDrawerLegacyFromWidget()
	return cmd
}

func (m *Model) handleHelpDrawerShortcuts(k tea.KeyMsg) (bool, tea.Cmd) {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil || m.helpDrawer.provider == nil {
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
		m.helpDrawer.widget.TogglePin()
		m.syncHelpDrawerLegacyFromWidget()
		return true, nil
	}

	return false, nil
}

func (m *Model) toggleHelpDrawer() tea.Cmd {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil {
		return nil
	}
	cmd := m.helpDrawer.widget.Toggle(m.appContext(), m.textInput.Value(), m.textInput.Position())
	m.syncHelpDrawerLegacyFromWidget()
	return cmd
}

func (m *Model) closeHelpDrawer() {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil {
		return
	}
	m.helpDrawer.widget.Close()
	m.syncHelpDrawerLegacyFromWidget()
}

func (m *Model) requestHelpDrawerNow(trigger HelpDrawerTrigger) tea.Cmd {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil || m.helpDrawer.provider == nil {
		return nil
	}
	cmd := m.helpDrawer.widget.RequestNow(m.appContext(), m.textInput.Value(), m.textInput.Position(), contextpanel.Trigger(trigger))
	m.syncHelpDrawerLegacyFromWidget()
	return cmd
}

func (m *Model) handleHelpDrawerResult(msg helpDrawerResultMsg) tea.Cmd {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil {
		return nil
	}
	m.helpDrawer.widget.HandleResult(msg)
	m.syncHelpDrawerLegacyFromWidget()
	return nil
}

func (m *Model) helpDrawerCmd(req HelpDrawerRequest) tea.Cmd {
	m.ensureHelpDrawerWidget()
	if m.helpDrawer.widget == nil {
		return nil
	}
	return m.helpDrawer.widget.CommandForRequest(m.appContext(), req)
}
