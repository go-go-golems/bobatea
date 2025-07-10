package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// EchoEvaluator is a simple evaluator that echoes back the input
type EchoEvaluator struct{}

func (e *EchoEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	// Simple echo with some formatting
	code = strings.TrimSpace(code)
	if code == "" {
		return "Empty input", nil
	}

	// If it's a number, show it as both decimal and hex
	if num, err := strconv.Atoi(code); err == nil {
		return fmt.Sprintf("Number: %d (hex: 0x%x)", num, num), nil
	}

	// If it's a greeting, respond nicely
	if strings.HasPrefix(strings.ToLower(code), "hello") {
		return "Hello there! 👋", nil
	}

	// Otherwise, just echo it back
	return fmt.Sprintf("You said: %s", code), nil
}

func (e *EchoEvaluator) GetPrompt() string {
	return "echo> "
}

func (e *EchoEvaluator) GetName() string {
	return "Echo"
}

func (e *EchoEvaluator) SupportsMultiline() bool {
	return false
}

func (e *EchoEvaluator) GetFileExtension() string {
	return ".txt"
}

func main() {
	// Create the evaluator
	evaluator := &EchoEvaluator{}

	// Create a basic configuration
	config := repl.DefaultConfig()
	config.Title = "Basic Echo REPL"
	config.Placeholder = "Type something to echo back..."
	config.Width = 80

	// Create the REPL model
	model := repl.NewModel(evaluator, config)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
