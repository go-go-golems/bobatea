package repl

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// PaletteCommand describes a REPL command that can be executed from the command palette.
type PaletteCommand struct {
	ID          string
	Name        string
	Description string
	Category    string
	Keywords    []string
	Enabled     func(*Model) bool
	Action      func(*Model) tea.Cmd
}

// PaletteCommandProvider allows evaluators to contribute commands to the REPL command palette.
type PaletteCommandProvider interface {
	ListPaletteCommands(ctx context.Context) ([]PaletteCommand, error)
}

// PaletteCommandRegistry composes built-in and evaluator-contributed commands.
type PaletteCommandRegistry interface {
	List(ctx context.Context, m *Model) ([]PaletteCommand, error)
}
