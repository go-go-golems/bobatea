package repl

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) completionCmd(req CompletionRequest) tea.Cmd {
	m.ensureCompletionWidget()
	if m.completion.widget == nil {
		return nil
	}
	cmd := m.completion.widget.CommandForRequest(m.appContext(), req)
	m.syncCompletionLegacyFromWidget()
	return cmd
}
