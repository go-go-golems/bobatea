package repl

import "github.com/go-go-golems/bobatea/pkg/commandpalette"

type commandPaletteModel struct {
	ui commandpalette.Model

	enabled      bool
	openKeys     []string
	closeKeys    []string
	slashEnabled bool
	slashPolicy  CommandPaletteSlashPolicy
	maxVisible   int
}
