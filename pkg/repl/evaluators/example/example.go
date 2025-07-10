package example

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-go-golems/bobatea/pkg/repl"
)

// ExampleEvaluator is a simple evaluator for demonstration
type ExampleEvaluator struct {
	name string
}

// NewEvaluator creates a new example evaluator
func NewEvaluator() *ExampleEvaluator {
	return &ExampleEvaluator{
		name: "Example",
	}
}

// Evaluate implements the Evaluator interface
func (e *ExampleEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)

	// Simple math evaluation
	if strings.Contains(code, "+") {
		parts := strings.Split(code, "+")
		if len(parts) == 2 {
			a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				return strconv.Itoa(a + b), nil
			}
		}
	}

	// Echo functionality
	if strings.HasPrefix(code, "echo ") {
		return strings.TrimPrefix(code, "echo "), nil
	}

	// Default response
	return fmt.Sprintf("You typed: %s", code), nil
}

// GetPrompt returns the prompt string
func (e *ExampleEvaluator) GetPrompt() string {
	return "example> "
}

// GetName returns the evaluator name
func (e *ExampleEvaluator) GetName() string {
	return e.name
}

// SupportsMultiline returns true if multiline is supported
func (e *ExampleEvaluator) SupportsMultiline() bool {
	return true
}

// GetFileExtension returns the file extension for external editor
func (e *ExampleEvaluator) GetFileExtension() string {
	return ".txt"
}

var _ repl.Evaluator = (*ExampleEvaluator)(nil)
