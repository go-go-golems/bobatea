package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// AdvancedCalculatorEvaluator is a more sophisticated calculator
type AdvancedCalculatorEvaluator struct {
	variables map[string]float64
	functions map[string]func(float64) float64
	history   []string
}

func NewAdvancedCalculatorEvaluator() *AdvancedCalculatorEvaluator {
	calc := &AdvancedCalculatorEvaluator{
		variables: make(map[string]float64),
		functions: make(map[string]func(float64) float64),
		history:   make([]string, 0),
	}

	// Add built-in functions
	calc.functions["sin"] = math.Sin
	calc.functions["cos"] = math.Cos
	calc.functions["tan"] = math.Tan
	calc.functions["sqrt"] = math.Sqrt
	calc.functions["log"] = math.Log
	calc.functions["abs"] = math.Abs
	calc.functions["ceil"] = math.Ceil
	calc.functions["floor"] = math.Floor

	// Add some constants
	calc.variables["pi"] = math.Pi
	calc.variables["e"] = math.E

	return calc
}

func (e *AdvancedCalculatorEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "Empty expression", nil
	}

	// Add to history
	e.history = append(e.history, code)

	// Handle variable assignment: x = 42
	if strings.Contains(code, "=") && !strings.Contains(code, "==") {
		return e.handleAssignment(code)
	}

	// Handle special commands
	if strings.HasPrefix(code, "vars") {
		return e.showVariables(), nil
	}

	if strings.HasPrefix(code, "funcs") {
		return e.showFunctions(), nil
	}

	if strings.HasPrefix(code, "hist") {
		return e.showHistory(), nil
	}

	// Evaluate mathematical expression
	result, err := e.evaluateExpression(code)
	if err != nil {
		return "", err
	}

	// Store result in special variable 'ans'
	e.variables["ans"] = result

	return fmt.Sprintf("%.6g", result), nil
}

func (e *AdvancedCalculatorEvaluator) handleAssignment(code string) (string, error) {
	parts := strings.Split(code, "=")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid assignment syntax")
	}

	varName := strings.TrimSpace(parts[0])
	expr := strings.TrimSpace(parts[1])

	// Validate variable name
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(varName) {
		return "", fmt.Errorf("invalid variable name: %s", varName)
	}

	// Evaluate the expression
	result, err := e.evaluateExpression(expr)
	if err != nil {
		return "", fmt.Errorf("error evaluating expression: %w", err)
	}

	e.variables[varName] = result
	return fmt.Sprintf("%s = %.6g", varName, result), nil
}

func (e *AdvancedCalculatorEvaluator) evaluateExpression(expr string) (float64, error) {
	// Replace variables with their values
	for varName, value := range e.variables {
		expr = strings.ReplaceAll(expr, varName, fmt.Sprintf("%.15g", value))
	}

	// Handle function calls
	for funcName, fn := range e.functions {
		pattern := regexp.MustCompile(funcName + `\(([^)]+)\)`)
		expr = pattern.ReplaceAllStringFunc(expr, func(match string) string {
			// Extract the argument
			arg := strings.TrimPrefix(match, funcName+"(")
			arg = strings.TrimSuffix(arg, ")")

			// Evaluate the argument
			argValue, err := e.evaluateBasicExpression(arg)
			if err != nil {
				return match // Return original if we can't evaluate
			}

			// Apply the function
			result := fn(argValue)
			return fmt.Sprintf("%.15g", result)
		})
	}

	// Evaluate the remaining expression
	return e.evaluateBasicExpression(expr)
}

func (e *AdvancedCalculatorEvaluator) evaluateBasicExpression(expr string) (float64, error) {
	// Remove spaces
	expr = strings.ReplaceAll(expr, " ", "")

	// Handle simple cases first
	if val, err := strconv.ParseFloat(expr, 64); err == nil {
		return val, nil
	}

	// Handle basic arithmetic operations
	// This is a simplified parser - in a real implementation you'd want a proper parser

	// Handle addition and subtraction
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			left, err1 := e.evaluateBasicExpression(parts[0])
			right, err2 := e.evaluateBasicExpression(parts[1])
			if err1 == nil && err2 == nil {
				return left + right, nil
			}
		}
	}

	if strings.Contains(expr, "-") && !strings.HasPrefix(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			left, err1 := e.evaluateBasicExpression(parts[0])
			right, err2 := e.evaluateBasicExpression(parts[1])
			if err1 == nil && err2 == nil {
				return left - right, nil
			}
		}
	}

	// Handle multiplication and division
	if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		if len(parts) == 2 {
			left, err1 := e.evaluateBasicExpression(parts[0])
			right, err2 := e.evaluateBasicExpression(parts[1])
			if err1 == nil && err2 == nil {
				return left * right, nil
			}
		}
	}

	if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		if len(parts) == 2 {
			left, err1 := e.evaluateBasicExpression(parts[0])
			right, err2 := e.evaluateBasicExpression(parts[1])
			if err1 == nil && err2 == nil {
				if right == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return left / right, nil
			}
		}
	}

	// Handle power operator
	if strings.Contains(expr, "^") {
		parts := strings.Split(expr, "^")
		if len(parts) == 2 {
			left, err1 := e.evaluateBasicExpression(parts[0])
			right, err2 := e.evaluateBasicExpression(parts[1])
			if err1 == nil && err2 == nil {
				return math.Pow(left, right), nil
			}
		}
	}

	return 0, fmt.Errorf("cannot evaluate expression: %s", expr)
}

func (e *AdvancedCalculatorEvaluator) showVariables() string {
	if len(e.variables) == 0 {
		return "No variables defined"
	}

	var result strings.Builder
	result.WriteString("Variables:\n")
	for name, value := range e.variables {
		result.WriteString(fmt.Sprintf("  %s = %.6g\n", name, value))
	}

	return strings.TrimSpace(result.String())
}

func (e *AdvancedCalculatorEvaluator) showFunctions() string {
	var result strings.Builder
	result.WriteString("Available functions:\n")
	for name := range e.functions {
		result.WriteString(fmt.Sprintf("  %s(x)\n", name))
	}

	return strings.TrimSpace(result.String())
}

func (e *AdvancedCalculatorEvaluator) showHistory() string {
	if len(e.history) == 0 {
		return "No history"
	}

	var result strings.Builder
	result.WriteString("Recent expressions:\n")

	// Show last 10 expressions
	start := len(e.history) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(e.history); i++ {
		result.WriteString(fmt.Sprintf("  %d: %s\n", i+1, e.history[i]))
	}

	return strings.TrimSpace(result.String())
}

func (e *AdvancedCalculatorEvaluator) GetPrompt() string {
	return "calc> "
}

func (e *AdvancedCalculatorEvaluator) GetName() string {
	return "Advanced Calculator"
}

func (e *AdvancedCalculatorEvaluator) SupportsMultiline() bool {
	return false
}

func (e *AdvancedCalculatorEvaluator) GetFileExtension() string {
	return ".calc"
}

func main() {
	// Create the advanced calculator evaluator
	evaluator := NewAdvancedCalculatorEvaluator()

	// Create configuration with custom settings
	config := repl.Config{
		Title:                "Advanced Calculator REPL",
		Placeholder:          "Enter mathematical expression...",
		Width:                100,
		StartMultiline:       false,
		EnableExternalEditor: true,
		EnableHistory:        true,
		MaxHistorySize:       500,
	}

	// Create the REPL model
	model := repl.NewModel(evaluator, config)

	// Use the dark theme for better visibility
	model.SetTheme(repl.BuiltinThemes["dark"])

	// Add some custom commands
	model.AddCustomCommand("clear-vars", func(args []string) tea.Cmd {
		return func() tea.Msg {
			// Clear all variables except constants
			for name := range evaluator.variables {
				if name != "pi" && name != "e" {
					delete(evaluator.variables, name)
				}
			}
			return repl.EvaluationCompleteMsg{
				Input:  "/clear-vars",
				Output: "All variables cleared (except pi and e)",
				Error:  nil,
			}
		}
	})

	model.AddCustomCommand("reset", func(args []string) tea.Cmd {
		return func() tea.Msg {
			// Reset calculator state
			evaluator.variables = make(map[string]float64)
			evaluator.variables["pi"] = math.Pi
			evaluator.variables["e"] = math.E
			evaluator.history = make([]string, 0)

			return repl.EvaluationCompleteMsg{
				Input:  "/reset",
				Output: "Calculator state reset",
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
