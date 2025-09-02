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
)

// MathEvaluator is a simple math evaluator
type MathEvaluator struct{}

func (e *MathEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)

	// Simple arithmetic operations
	if strings.Contains(code, "+") {
		return e.evaluateOperation(code, "+", func(a, b int) int { return a + b })
	}
	if strings.Contains(code, "-") {
		return e.evaluateOperation(code, "-", func(a, b int) int { return a - b })
	}
	if strings.Contains(code, "*") {
		return e.evaluateOperation(code, "*", func(a, b int) int { return a * b })
	}
	if strings.Contains(code, "/") {
		return e.evaluateOperation(code, "/", func(a, b int) int { return a / b })
	}

	// Try to parse as single number
	if num, err := strconv.Atoi(code); err == nil {
		return strconv.Itoa(num), nil
	}

	return "", fmt.Errorf("invalid expression: %s", code)
}

// Implement streaming by adapting Evaluate
func (e *MathEvaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
	out, err := e.Evaluate(ctx, code)
	if err != nil {
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": fmt.Sprintf("Error: %v", err)}})
		return nil
	}
	emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": out}})
	return nil
}

func (e *MathEvaluator) evaluateOperation(code, op string, fn func(int, int) int) (string, error) {
	parts := strings.Split(code, op)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid %s operation", op)
	}

	a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err1 != nil || err2 != nil {
		return "", fmt.Errorf("invalid numbers in %s operation", op)
	}

	if op == "/" && b == 0 {
		return "", fmt.Errorf("division by zero")
	}

	result := fn(a, b)
	return strconv.Itoa(result), nil
}

func (e *MathEvaluator) GetPrompt() string {
	return "math> "
}

func (e *MathEvaluator) GetName() string {
	return "Math"
}

func (e *MathEvaluator) SupportsMultiline() bool {
	return false
}

func (e *MathEvaluator) GetFileExtension() string {
	return ".math"
}

func main() {
	// Create the evaluator
	evaluator := &MathEvaluator{}

	// Create configuration
	config := repl.DefaultConfig()
	config.Title = "Math REPL"
	config.Placeholder = "Enter math expression (e.g., 5 + 3)"

	// Wire bus and forwarders
	bus, err := eventbus.NewInMemoryBus()
	if err != nil {
		log.Fatal(err)
	}
	repl.RegisterReplToTimelineTransformer(bus)

	// Create the timeline-based model and program
	model := repl.NewModel(evaluator, config, bus.Publisher)
	p := tea.NewProgram(model, tea.WithAltScreen())
	timeline.RegisterUIForwarder(bus, p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 2)
	go func() { errs <- bus.Run(ctx) }()
	go func() { _, e := p.Run(); cancel(); errs <- e }()
	if e := <-errs; e != nil {
		log.Fatal(e)
	}
}
