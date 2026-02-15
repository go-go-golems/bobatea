package repl

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCommandPaletteEvaluator struct {
	commands []PaletteCommand
	listErr  error

	slashOpen bool
	slashErr  error
	slashReqs []CommandPaletteSlashRequest

	helpDoc HelpDrawerDocument
	helpErr error
}

func (f *fakeCommandPaletteEvaluator) EvaluateStream(context.Context, string, func(Event)) error {
	return nil
}

func (f *fakeCommandPaletteEvaluator) GetPrompt() string        { return "> " }
func (f *fakeCommandPaletteEvaluator) GetName() string          { return "fake-command-palette" }
func (f *fakeCommandPaletteEvaluator) SupportsMultiline() bool  { return false }
func (f *fakeCommandPaletteEvaluator) GetFileExtension() string { return ".txt" }

func (f *fakeCommandPaletteEvaluator) ListPaletteCommands(context.Context) ([]PaletteCommand, error) {
	return f.commands, f.listErr
}

func (f *fakeCommandPaletteEvaluator) ShouldOpenCommandPaletteOnSlash(_ context.Context, req CommandPaletteSlashRequest) (bool, error) {
	f.slashReqs = append(f.slashReqs, req)
	return f.slashOpen, f.slashErr
}

func (f *fakeCommandPaletteEvaluator) GetHelpDrawer(context.Context, HelpDrawerRequest) (HelpDrawerDocument, error) {
	return f.helpDoc, f.helpErr
}

func newCommandPaletteTestModel(t *testing.T, evaluator Evaluator, mutateConfig func(*Config)) *Model {
	t.Helper()

	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Autocomplete.Enabled = false
	cfg.HelpBar.Enabled = false
	cfg.HelpDrawer.Enabled = false
	cfg.CommandPalette = DefaultCommandPaletteConfig()
	if mutateConfig != nil {
		mutateConfig(&cfg)
	}

	m := NewModel(evaluator, cfg, bus.Publisher)
	m.width = 120
	m.height = 40
	m.palette.ui.SetSize(m.width, m.height)
	return m
}

func TestCommandPaletteConfigNormalizationBounds(t *testing.T) {
	cfg := normalizeCommandPaletteConfig(CommandPaletteConfig{
		Enabled:          true,
		OpenKeys:         []string{},
		CloseKeys:        []string{},
		SlashOpenEnabled: true,
		SlashPolicy:      CommandPaletteSlashPolicy("invalid"),
		MaxVisibleItems:  0,
	})

	assert.Equal(t, []string{"ctrl+p"}, cfg.OpenKeys)
	assert.Equal(t, []string{"esc", "ctrl+p"}, cfg.CloseKeys)
	assert.Equal(t, CommandPaletteSlashPolicyEmptyInput, cfg.SlashPolicy)
	assert.Equal(t, 8, cfg.MaxVisibleItems)
}

func TestCommandPaletteRoutingTakesPrecedenceOverCompletionNavigation(t *testing.T) {
	m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, nil)
	m.completion.visible = true
	m.completion.lastResult = CompletionResult{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "const", DisplayText: "const"},
			{Id: "2", Value: "continue", DisplayText: "continue"},
		},
	}
	m.completion.selection = 0

	_, openCmd := m.updateInput(tea.KeyMsg{Type: tea.KeyCtrlP})
	require.Nil(t, openCmd)
	require.True(t, m.palette.ui.IsVisible())

	_, navCmd := m.updateInput(tea.KeyMsg{Type: tea.KeyDown})
	require.Nil(t, navCmd)
	assert.Equal(t, 0, m.completion.selection, "completion navigation must not run while palette is open")
}

func TestCommandPaletteRoutingTakesPrecedenceOverHelpDrawerShortcuts(t *testing.T) {
	evaluator := &fakeCommandPaletteEvaluator{
		helpDoc: HelpDrawerDocument{
			Show:     true,
			Title:    "console.log",
			Subtitle: "function",
		},
	}
	m := newCommandPaletteTestModel(t, evaluator, func(cfg *Config) {
		cfg.HelpDrawer.Enabled = true
	})

	require.False(t, m.helpDrawer.visible)
	_, _ = m.updateInput(tea.KeyMsg{Type: tea.KeyCtrlP})
	require.True(t, m.palette.ui.IsVisible())

	_, drawerCmd := m.updateInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}, Alt: true})
	require.Nil(t, drawerCmd)
	assert.False(t, m.helpDrawer.visible, "drawer shortcut should be ignored while palette is open")
}

func TestCommandPaletteSlashPolicyEmptyInputOpensAndConsumesSlash(t *testing.T) {
	m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, func(cfg *Config) {
		cfg.CommandPalette.SlashPolicy = CommandPaletteSlashPolicyEmptyInput
	})
	m.textInput.SetValue("")
	m.textInput.SetCursor(0)

	_, cmd := m.updateInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	require.Nil(t, cmd)
	assert.True(t, m.palette.ui.IsVisible())
	assert.Equal(t, "", m.textInput.Value(), "slash should be consumed when opening the palette")
}

func TestCommandPaletteSlashPolicyEmptyInputFallsThroughWhenInputNotEmpty(t *testing.T) {
	m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, func(cfg *Config) {
		cfg.CommandPalette.SlashPolicy = CommandPaletteSlashPolicyEmptyInput
	})
	m.textInput.SetValue("co")
	m.textInput.SetCursor(2)

	_, cmd := m.updateInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	drainModelCmds(m, cmd)
	assert.False(t, m.palette.ui.IsVisible())
	assert.Equal(t, "co/", m.textInput.Value())
}

func TestCommandPaletteSlashPolicyColumnZeroOpensAtStart(t *testing.T) {
	m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, func(cfg *Config) {
		cfg.CommandPalette.SlashPolicy = CommandPaletteSlashPolicyColumnZero
	})
	m.textInput.SetValue("console")
	m.textInput.SetCursor(0)

	_, cmd := m.updateInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	require.Nil(t, cmd)
	assert.True(t, m.palette.ui.IsVisible())
	assert.Equal(t, "console", m.textInput.Value())
}

func TestCommandPaletteSlashPolicyProviderDelegates(t *testing.T) {
	evaluator := &fakeCommandPaletteEvaluator{slashOpen: true}
	m := newCommandPaletteTestModel(t, evaluator, func(cfg *Config) {
		cfg.CommandPalette.SlashPolicy = CommandPaletteSlashPolicyProvider
	})
	m.textInput.SetValue(".co")
	m.textInput.SetCursor(3)

	_, cmd := m.updateInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	require.Nil(t, cmd)
	assert.True(t, m.palette.ui.IsVisible())
	require.Len(t, evaluator.slashReqs, 1)
	assert.Equal(t, ".co", evaluator.slashReqs[0].Input)
	assert.Equal(t, 3, evaluator.slashReqs[0].CursorByte)
}

func TestCommandPaletteExecutesSelectedCommandAndCloses(t *testing.T) {
	evaluator := &fakeCommandPaletteEvaluator{}
	m := newCommandPaletteTestModel(t, evaluator, nil)

	executed := false
	evaluator.commands = []PaletteCommand{
		{
			ID:          "custom.zzz",
			Name:        "ZZZ Custom Action",
			Description: "test command",
			Action: func(m *Model) tea.Cmd {
				executed = true
				m.textInput.SetValue("custom-action-ran")
				return nil
			},
		},
	}

	m.textInput.SetValue("seed")
	m.textInput.SetCursor(len("seed"))

	_, _ = m.updateInput(tea.KeyMsg{Type: tea.KeyCtrlP})
	require.True(t, m.palette.ui.IsVisible())

	_, _ = m.updateInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, cmd := m.updateInput(tea.KeyMsg{Type: tea.KeyEnter})
	drainModelCmds(m, cmd)

	assert.True(t, executed)
	assert.False(t, m.palette.ui.IsVisible(), "palette should close after command execution")
	assert.Equal(t, "custom-action-ran", m.textInput.Value())
}

func TestCommandPaletteListCommandsDoesNotTruncateByVisibleLimit(t *testing.T) {
	evaluator := &fakeCommandPaletteEvaluator{
		commands: []PaletteCommand{
			{
				ID:          "custom.alpha",
				Name:        "Alpha Command",
				Description: "alpha",
				Action: func(m *Model) tea.Cmd {
					return nil
				},
			},
			{
				ID:          "custom.beta",
				Name:        "Beta Command",
				Description: "beta",
				Action: func(m *Model) tea.Cmd {
					return nil
				},
			},
			{
				ID:          "custom.zzz",
				Name:        "ZZZ Tail Command",
				Description: "tail",
				Action: func(m *Model) tea.Cmd {
					return nil
				},
			},
		},
	}
	m := newCommandPaletteTestModel(t, evaluator, func(cfg *Config) {
		cfg.CommandPalette.MaxVisibleItems = 2
	})

	commands := m.listPaletteCommands(context.Background())
	assert.Greater(t, len(commands), m.palette.maxVisible, "visible-row limit must not remove searchable commands")

	foundTail := false
	for _, cmd := range commands {
		if cmd.Name == "ZZZ Tail Command" {
			foundTail = true
			break
		}
	}
	assert.True(t, foundTail, "commands beyond visible rows must remain discoverable")
}

func TestCommandPaletteSlashOpenDisabledFallsThrough(t *testing.T) {
	m := newCommandPaletteTestModel(t, &fakeCommandPaletteEvaluator{}, func(cfg *Config) {
		cfg.CommandPalette.SlashOpenEnabled = false
		cfg.CommandPalette.SlashPolicy = CommandPaletteSlashPolicyEmptyInput
	})
	m.textInput.SetValue("")
	m.textInput.SetCursor(0)

	_, cmd := m.updateInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	drainModelCmds(m, cmd)

	assert.False(t, m.palette.ui.IsVisible())
	assert.Equal(t, "/", m.textInput.Value())
}
