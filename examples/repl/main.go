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

	// Create the model
	model := repl.NewModel(evaluator, config)

	// Set dark theme
	model.SetTheme(repl.BuiltinThemes["dark"])

	// Add a custom command
	model.AddCustomCommand("square", func(args []string) tea.Cmd {
		return func() tea.Msg {
			if len(args) != 1 {
				return repl.EvaluationCompleteMsg{
					Input:  "/square",
					Output: "Usage: /square <number>",
					Error:  fmt.Errorf("invalid usage"),
				}
			}

			num, err := strconv.Atoi(args[0])
			if err != nil {
				return repl.EvaluationCompleteMsg{
					Input:  "/square " + args[0],
					Output: "Invalid number",
					Error:  err,
				}
			}

			result := num * num
			return repl.EvaluationCompleteMsg{
				Input:  "/square " + args[0],
				Output: strconv.Itoa(result),
				Error:  nil,
			}
		}
	})

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
