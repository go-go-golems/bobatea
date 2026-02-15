package repl

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/commandpalette"
	"github.com/rs/zerolog/log"
)

type commandPaletteModel struct {
	ui commandpalette.Model

	enabled          bool
	openKeys         []string
	closeKeys        []string
	slashEnabled     bool
	slashPolicy      CommandPaletteSlashPolicy
	maxVisible       int
	overlayPlacement CommandPaletteOverlayPlacement
	overlayMargin    int
	overlayOffsetX   int
	overlayOffsetY   int
}

func (m *Model) handleCommandPaletteInput(k tea.KeyMsg) (bool, tea.Cmd) {
	if !m.palette.enabled {
		return false, nil
	}

	if m.palette.ui.IsVisible() {
		if key.Matches(k, m.keyMap.CommandPaletteClose) || key.Matches(k, m.keyMap.CommandPaletteOpen) {
			m.palette.ui.Hide()
			return true, nil
		}
		var cmd tea.Cmd
		m.palette.ui, cmd = m.palette.ui.Update(k)
		return true, cmd
	}

	if key.Matches(k, m.keyMap.CommandPaletteOpen) {
		m.openCommandPalette()
		return true, nil
	}
	if isSlashOpenKey(k) && m.shouldOpenCommandPaletteFromSlash() {
		m.openCommandPalette()
		return true, nil
	}

	return false, nil
}

func (m *Model) openCommandPalette() {
	commands := m.listPaletteCommands(m.appContext())
	m.palette.ui.SetCommands(commands)
	m.palette.ui.Show()
}

func (m *Model) listPaletteCommands(ctx context.Context) []commandpalette.Command {
	paletteCommands := m.builtinPaletteCommands()

	if provider, ok := m.evaluator.(PaletteCommandProvider); ok {
		extra, err := provider.ListPaletteCommands(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("failed to list evaluator palette commands")
		} else {
			paletteCommands = mergePaletteCommands(paletteCommands, extra)
		}
	}

	ret := make([]commandpalette.Command, 0, len(paletteCommands))
	for _, pc := range paletteCommands {
		if pc.Name == "" || pc.Action == nil {
			continue
		}
		if pc.Enabled != nil && !pc.Enabled(m) {
			continue
		}
		paletteCommand := pc
		ret = append(ret, commandpalette.Command{
			Name:        paletteCommand.Name,
			Description: paletteCommand.Description,
			Action: func() tea.Cmd {
				return paletteCommand.Action(m)
			},
		})
	}
	return ret
}

func (m *Model) builtinPaletteCommands() []PaletteCommand {
	commands := []PaletteCommand{
		{
			ID:          "repl.clear-input",
			Name:        "Clear Input",
			Description: "Reset the current input line",
			Category:    "repl",
			Keywords:    []string{"reset", "line"},
			Action: func(m *Model) tea.Cmd {
				m.textInput.Reset()
				m.hideHelpBar()
				return nil
			},
		},
		{
			ID:          "repl.clear-history",
			Name:        "Clear History",
			Description: "Remove REPL history entries",
			Category:    "history",
			Keywords:    []string{"history", "reset"},
			Enabled: func(m *Model) bool {
				return m.config.EnableHistory
			},
			Action: func(m *Model) tea.Cmd {
				m.history.Clear()
				return nil
			},
		},
		{
			ID:          "repl.toggle-help",
			Name:        "Toggle Help",
			Description: "Toggle key help visibility",
			Category:    "help",
			Keywords:    []string{"help", "keys"},
			Action: func(m *Model) tea.Cmd {
				m.help.ShowAll = !m.help.ShowAll
				return nil
			},
		},
		{
			ID:          "repl.toggle-focus",
			Name:        "Toggle Focus",
			Description: "Switch focus between input and timeline",
			Category:    "repl",
			Keywords:    []string{"focus", "timeline", "input"},
			Action: func(m *Model) tea.Cmd {
				if m.focus == "input" {
					m.focus = "timeline"
					m.textInput.Blur()
					m.sh.SetSelectionVisible(true)
				} else {
					m.focus = "input"
					m.textInput.Focus()
					m.sh.SetSelectionVisible(false)
				}
				m.updateKeyBindings()
				return nil
			},
		},
		{
			ID:          "repl.quit",
			Name:        "Quit REPL",
			Description: "Exit the REPL application",
			Category:    "repl",
			Keywords:    []string{"exit", "quit"},
			Action: func(m *Model) tea.Cmd {
				m.cancelAppContext()
				return tea.Quit
			},
		},
	}

	if m.helpDrawer.provider != nil {
		commands = append(commands, PaletteCommand{
			ID:          "helpdrawer.toggle",
			Name:        "Toggle Help Drawer",
			Description: "Open or close contextual help drawer",
			Category:    "help",
			Keywords:    []string{"drawer", "help"},
			Action: func(m *Model) tea.Cmd {
				return m.toggleHelpDrawer()
			},
		})
	}

	return commands
}

func mergePaletteCommands(base, extra []PaletteCommand) []PaletteCommand {
	ret := make([]PaletteCommand, 0, len(base)+len(extra))
	seen := make(map[string]struct{}, len(base)+len(extra))

	add := func(c PaletteCommand) {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			id = strings.ToLower(strings.TrimSpace(c.Name))
		}
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		c.ID = id
		seen[id] = struct{}{}
		ret = append(ret, c)
	}

	for _, c := range base {
		add(c)
	}
	for _, c := range extra {
		add(c)
	}

	return ret
}

func isSlashOpenKey(k tea.KeyMsg) bool {
	if k.Paste || k.Alt {
		return false
	}
	return k.Type == tea.KeyRunes && len(k.Runes) == 1 && k.Runes[0] == '/'
}

func (m *Model) shouldOpenCommandPaletteFromSlash() bool {
	if !m.palette.enabled || !m.palette.slashEnabled {
		return false
	}
	if m.completion.visible {
		return false
	}

	input := m.textInput.Value()
	cursor := m.textInput.Position()

	switch m.palette.slashPolicy {
	case CommandPaletteSlashPolicyEmptyInput:
		return cursor == 0 && strings.TrimSpace(input) == ""
	case CommandPaletteSlashPolicyColumnZero:
		return cursor == 0
	case CommandPaletteSlashPolicyProvider:
		provider, ok := m.evaluator.(CommandPaletteSlashOpenProvider)
		if !ok {
			return false
		}
		okOpen, err := provider.ShouldOpenCommandPaletteOnSlash(m.appContext(), CommandPaletteSlashRequest{
			Input:      input,
			CursorByte: cursor,
		})
		if err != nil {
			log.Warn().Err(err).Msg("slash policy provider failed")
			return false
		}
		return okOpen
	default:
		return false
	}
}
