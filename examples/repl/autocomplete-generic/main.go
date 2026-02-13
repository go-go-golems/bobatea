package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

type GenericEvaluator struct {
	symbols []string
}

func newGenericEvaluator() *GenericEvaluator {
	return &GenericEvaluator{
		symbols: []string{
			"console",
			"const",
			"context",
			"continue",
			"count",
			"contains",
			"concat",
		},
	}
}

func (e *GenericEvaluator) EvaluateStream(_ context.Context, code string, emit func(repl.Event)) error {
	code = strings.TrimSpace(code)
	if code == "" {
		emit(repl.Event{
			Kind:  repl.EventResultMarkdown,
			Props: map[string]any{"markdown": "Type a symbol prefix (for example `co`) and wait for debounce, or press `tab` to trigger completion immediately."},
		})
		return nil
	}

	emit(repl.Event{
		Kind:  repl.EventResultMarkdown,
		Props: map[string]any{"markdown": fmt.Sprintf("Echo: `%s`", code)},
	})
	return nil
}

func (e *GenericEvaluator) CompleteInput(_ context.Context, req repl.CompletionRequest) (repl.CompletionResult, error) {
	token, from, to := currentToken(req.Input, req.CursorByte)
	show := false

	switch req.Reason {
	case repl.CompletionReasonDebounce:
		show = len(token) >= 2
	case repl.CompletionReasonShortcut:
		show = true
	case repl.CompletionReasonManual:
		show = len(token) > 0
	}

	if !show {
		return repl.CompletionResult{
			Show:        false,
			ReplaceFrom: from,
			ReplaceTo:   to,
		}, nil
	}

	tokenLower := strings.ToLower(token)
	suggestions := make([]autocomplete.Suggestion, 0, len(e.symbols))
	for _, symbol := range e.symbols {
		if tokenLower != "" && !strings.HasPrefix(strings.ToLower(symbol), tokenLower) {
			continue
		}

		suggestions = append(suggestions, autocomplete.Suggestion{
			Id:          symbol,
			Value:       symbol,
			DisplayText: symbol,
		})
	}

	return repl.CompletionResult{
		Show:        len(suggestions) > 0,
		Suggestions: suggestions,
		ReplaceFrom: from,
		ReplaceTo:   to,
	}, nil
}

func currentToken(input string, cursor int) (string, int, int) {
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(input) {
		cursor = len(input)
	}

	isTokenChar := func(r byte) bool {
		return (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_'
	}

	start := cursor
	for start > 0 && isTokenChar(input[start-1]) {
		start--
	}
	end := cursor
	for end < len(input) && isTokenChar(input[end]) {
		end++
	}

	return input[start:end], start, end
}

func (e *GenericEvaluator) GetPrompt() string        { return "generic> " }
func (e *GenericEvaluator) GetName() string          { return "Generic" }
func (e *GenericEvaluator) SupportsMultiline() bool  { return false }
func (e *GenericEvaluator) GetFileExtension() string { return ".txt" }

func main() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	zlog.Logger = zerolog.New(io.Discard)

	evaluator := newGenericEvaluator()
	config := repl.DefaultConfig()
	config.Title = "Generic Autocomplete REPL"
	config.Placeholder = "Type 'co' and wait, or press Tab for explicit trigger"
	config.Autocomplete.Enabled = true
	config.Autocomplete.TriggerKeys = []string{"tab"}
	config.Autocomplete.AcceptKeys = []string{"enter", "tab"}
	config.Autocomplete.FocusToggleKey = "ctrl+t"

	bus, err := eventbus.NewInMemoryBus()
	if err != nil {
		log.Fatal(err)
	}
	repl.RegisterReplToTimelineTransformer(bus)

	model := repl.NewModel(evaluator, config, bus.Publisher)
	p := tea.NewProgram(model, tea.WithAltScreen())
	timeline.RegisterUIForwarder(bus, p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 2)
	go func() { errs <- bus.Run(ctx) }()
	go func() {
		_, runErr := p.Run()
		cancel()
		errs <- runErr
	}()
	if runErr := <-errs; runErr != nil {
		log.Fatal(runErr)
	}
}
