package repl_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/repl"
	jsrepl "github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSHelpBarVisibleAfterDebouncedTyping(t *testing.T) {
	evaluator, err := jsrepl.NewWithDefaults()
	require.NoError(t, err)

	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := repl.DefaultConfig()
	cfg.Autocomplete.Enabled = false
	cfg.HelpBar.Enabled = true
	cfg.HelpBar.Debounce = time.Nanosecond
	cfg.HelpBar.RequestTimeout = 100 * time.Millisecond

	model := repl.NewModel(evaluator, cfg, bus.Publisher)
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 28})

	typeAndDrain(t, model, "cons")

	view := model.View()
	assert.Contains(t, view, "console: object")
}

func TestJSHelpBarRemainsHiddenForSingleCharDebounce(t *testing.T) {
	evaluator, err := jsrepl.NewWithDefaults()
	require.NoError(t, err)

	bus, err := eventbus.NewInMemoryBus()
	require.NoError(t, err)

	cfg := repl.DefaultConfig()
	cfg.Autocomplete.Enabled = false
	cfg.HelpBar.Enabled = true
	cfg.HelpBar.Debounce = time.Nanosecond
	cfg.HelpBar.RequestTimeout = 100 * time.Millisecond

	model := repl.NewModel(evaluator, cfg, bus.Publisher)
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 28})

	typeAndDrain(t, model, "c")

	view := model.View()
	assert.NotContains(t, view, "console: object")
}

func typeAndDrain(t *testing.T, model *repl.Model, text string) {
	t.Helper()
	for _, r := range text {
		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		drainModelCmds(model, cmd)
	}
}

func drainModelCmds(model *repl.Model, cmd tea.Cmd) {
	queue := []tea.Cmd{cmd}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == nil {
			continue
		}

		msg := current()
		if msg == nil {
			continue
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			queue = append(queue, []tea.Cmd(batch)...)
			continue
		}

		_, next := model.Update(msg)
		queue = append(queue, next)
	}
}
