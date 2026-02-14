package repl

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

func runProvider[T any](
	baseCtx context.Context,
	requestID uint64,
	timeout time.Duration,
	providerName string,
	panicPrefix string,
	fn func(context.Context) (T, error),
) (T, error) {
	ctx, cancel := context.WithTimeout(baseCtx, timeout)
	defer cancel()

	var (
		out T
		err error
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Uint64("request_id", requestID).
					Str("provider", providerName).
					Msg("provider panicked")
				err = fmt.Errorf("%s panic: %v", panicPrefix, r)
			}
		}()
		out, err = fn(ctx)
	}()

	return out, err
}

func (m *Model) completionCmd(req CompletionRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := runProvider(
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

func (m *Model) helpBarCmd(req HelpBarRequest) tea.Cmd {
	return func() tea.Msg {
		payload, err := runProvider(
			m.appContext(),
			req.RequestID,
			m.helpBar.reqTimeout,
			"help-bar-provider",
			"help bar provider",
			func(ctx context.Context) (HelpBarPayload, error) {
				return m.helpBar.provider.GetHelpBar(ctx, req)
			},
		)

		return helpBarResultMsg{
			RequestID: req.RequestID,
			Payload:   payload,
			Err:       err,
		}
	}
}

func (m *Model) helpDrawerCmd(req HelpDrawerRequest) tea.Cmd {
	return func() tea.Msg {
		doc, err := runProvider(
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
