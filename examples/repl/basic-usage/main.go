package main

import (
    "context"
    "fmt"
    "log"
    "strconv"
    "strings"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/eventbus"
    "github.com/go-go-golems/bobatea/pkg/repl"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    "github.com/rs/zerolog"
    zlog "github.com/rs/zerolog/log"
    "io"
)

// EchoEvaluator is a simple evaluator that echoes back the input
type EchoEvaluator struct{}

func (e *EchoEvaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
    code = strings.TrimSpace(code)
    if code == "" {
        emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": "Empty input"}})
        return nil
    }
    if num, err := strconv.Atoi(code); err == nil {
        emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": fmt.Sprintf("Number: %d (hex: 0x%x)", num, num)}})
        return nil
    }
    if strings.HasPrefix(strings.ToLower(code), "hello") {
        emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": "Hello there! 👋"}})
        return nil
    }
    emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": fmt.Sprintf("You said: %s", code)}})
    return nil
}

func (e *EchoEvaluator) GetPrompt() string        { return "echo> " }
func (e *EchoEvaluator) GetName() string          { return "Echo" }
func (e *EchoEvaluator) SupportsMultiline() bool  { return false }
func (e *EchoEvaluator) GetFileExtension() string { return ".txt" }

func main() {
    // Initialize logging (avoid stdout noise in TUI)
    zerolog.SetGlobalLevel(zerolog.ErrorLevel)
    zlog.Logger = zerolog.New(io.Discard)

    // Create the evaluator
    evaluator := &EchoEvaluator{}

    // Create a basic configuration
    config := repl.DefaultConfig()
    config.Title = "Basic Echo REPL"
    config.Placeholder = "Type something to echo back..."
    config.Width = 80

    // Build in-memory bus and wire transformer + forwarder
    bus, err := eventbus.NewInMemoryBus()
    if err != nil { log.Fatal(err) }
    repl.RegisterReplToTimelineTransformer(bus)

    model := repl.NewModel(evaluator, config, bus.Publisher)
    p := tea.NewProgram(model, tea.WithAltScreen())
    timeline.RegisterUIForwarder(bus, p)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    errs := make(chan error, 2)
    go func() { errs <- bus.Run(ctx) }()
    go func() { _, e := p.Run(); cancel(); errs <- e }()
    if e := <-errs; e != nil { log.Fatal(e) }
}
