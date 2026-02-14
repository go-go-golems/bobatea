package repl

import (
	"context"
	"fmt"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

func (m *Model) completionCmd(req CompletionRequest) tea.Cmd {
	return func() tea.Msg {
		var (
			result    CompletionResult
			err       error
			recovered any
			stack     string
		)

		func() {
			defer func() {
				if r := recover(); r != nil {
					recovered = r
					stack = string(debug.Stack())
				}
			}()

			ctx, cancel := context.WithTimeout(m.appContext(), m.completion.reqTimeout)
			defer cancel()

			result, err = m.completion.provider.CompleteInput(ctx, req)
		}()

		if recovered != nil {
			log.Error().
				Interface("panic", recovered).
				Str("stack", stack).
				Uint64("request_id", req.RequestID).
				Msg("input completer panicked")
			return completionResultMsg{
				RequestID: req.RequestID,
				Err:       fmt.Errorf("input completer panic: %v", recovered),
			}
		}

		return completionResultMsg{
			RequestID: req.RequestID,
			Result:    result,
			Err:       err,
		}
	}
}

func (m *Model) helpBarCmd(req HelpBarRequest) tea.Cmd {
	return func() tea.Msg {
		var (
			payload   HelpBarPayload
			err       error
			recovered any
			stack     string
		)

		func() {
			defer func() {
				if r := recover(); r != nil {
					recovered = r
					stack = string(debug.Stack())
				}
			}()

			ctx, cancel := context.WithTimeout(m.appContext(), m.helpBar.reqTimeout)
			defer cancel()

			payload, err = m.helpBar.provider.GetHelpBar(ctx, req)
		}()

		if recovered != nil {
			log.Error().
				Interface("panic", recovered).
				Str("stack", stack).
				Uint64("request_id", req.RequestID).
				Msg("help bar provider panicked")
			return helpBarResultMsg{
				RequestID: req.RequestID,
				Err:       fmt.Errorf("help bar provider panic: %v", recovered),
			}
		}

		return helpBarResultMsg{
			RequestID: req.RequestID,
			Payload:   payload,
			Err:       err,
		}
	}
}

func (m *Model) helpDrawerCmd(req HelpDrawerRequest) tea.Cmd {
	return func() tea.Msg {
		var (
			doc       HelpDrawerDocument
			err       error
			recovered any
			stack     string
		)

		func() {
			defer func() {
				if r := recover(); r != nil {
					recovered = r
					stack = string(debug.Stack())
				}
			}()

			ctx, cancel := context.WithTimeout(m.appContext(), m.helpDrawer.reqTimeout)
			defer cancel()

			doc, err = m.helpDrawer.provider.GetHelpDrawer(ctx, req)
		}()

		if recovered != nil {
			log.Error().
				Interface("panic", recovered).
				Str("stack", stack).
				Uint64("request_id", req.RequestID).
				Msg("help drawer provider panicked")
			return helpDrawerResultMsg{
				RequestID: req.RequestID,
				Err:       fmt.Errorf("help drawer provider panic: %v", recovered),
			}
		}

		return helpDrawerResultMsg{
			RequestID: req.RequestID,
			Doc:       doc,
			Err:       err,
		}
	}
}
