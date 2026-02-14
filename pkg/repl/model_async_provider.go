package repl

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/tui/asyncprovider"
)

func (m *Model) completionCmd(req CompletionRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := asyncprovider.Run(
			m.appContext(),
			req.RequestID,
			m.completion.reqTimeout,
			"input-completer",
			"input completer",
			func(ctx context.Context) (CompletionResult, error) {
				return m.completion.provider.CompleteInput(ctx, req)
			},
		)

		return completionResultMsg{
			RequestID: req.RequestID,
			Result:    result,
			Err:       err,
		}
	}
}

func (m *Model) helpDrawerCmd(req HelpDrawerRequest) tea.Cmd {
	return func() tea.Msg {
		doc, err := asyncprovider.Run(
			m.appContext(),
			req.RequestID,
			m.helpDrawer.reqTimeout,
			"help-drawer-provider",
			"help drawer provider",
			func(ctx context.Context) (HelpDrawerDocument, error) {
				return m.helpDrawer.provider.GetHelpDrawer(ctx, req)
			},
		)

		return helpDrawerResultMsg{
			RequestID: req.RequestID,
			Doc:       doc,
			Err:       err,
		}
	}
}
