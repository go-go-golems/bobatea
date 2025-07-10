package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// CustomEvaluator demonstrates creating a custom evaluator
type CustomEvaluator struct {
	variables map[string]string
}

func NewCustomEvaluator() *CustomEvaluator {
	return &CustomEvaluator{
		variables: make(map[string]string),
	}
}

func (c *CustomEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)

	// Handle variable assignment
	if strings.Contains(code, "=") {
		parts := strings.SplitN(code, "=", 2)
		if len(parts) == 2 {
			varName := strings.TrimSpace(parts[0])
			varValue := strings.TrimSpace(parts[1])
			c.variables[varName] = varValue
			return fmt.Sprintf("Set %s = %s", varName, varValue), nil
		}
	}

	// Handle variable lookup
	if value, exists := c.variables[code]; exists {
		return fmt.Sprintf("%s = %s", code, value), nil
	}

	// Handle special commands
	switch code {
	case "time":
		return time.Now().Format(time.RFC3339), nil
	case "vars":
		if len(c.variables) == 0 {
			return "No variables defined", nil
		}
		var result strings.Builder
		result.WriteString("Variables:\n")
		for name, value := range c.variables {
			result.WriteString(fmt.Sprintf("  %s = %s\n", name, value))
		}
		return result.String(), nil
	case "clear":
		c.variables = make(map[string]string)
		return "Variables cleared", nil
	default:
		return fmt.Sprintf("Unknown command or variable: %s", code), fmt.Errorf("unknown: %s", code)
	}
}

func (c *CustomEvaluator) GetPrompt() string {
	return "custom> "
}

func (c *CustomEvaluator) GetName() string {
	return "Custom Variable Store"
}

func (c *CustomEvaluator) SupportsMultiline() bool {
	return false
}

func (c *CustomEvaluator) GetFileExtension() string {
	return ".txt"
}

// Ensure interface compliance
var _ repl.Evaluator = (*CustomEvaluator)(nil)

func main() {
	// Create custom evaluator
	evaluator := NewCustomEvaluator()

	// Create configuration
	config := repl.DefaultConfig()
	config.Title = "Custom Evaluator Example"
	config.Placeholder = "Enter var=value, var, time, vars, or clear"

	// Create model
	model := repl.NewModel(evaluator, config)

	// Add custom slash command
	model.AddCustomCommand("info", func(args []string) tea.Cmd {
		return func() tea.Msg {
			return repl.EvaluationCompleteMsg{
				Input:  "/info",
				Output: "Custom evaluator with variable storage:\n- Set variables: var=value\n- Get variables: var\n- Show time: time\n- List variables: vars\n- Clear variables: clear",
				Error:  nil,
			}
		}
	})

	// Start the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
