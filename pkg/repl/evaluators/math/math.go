package math

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-go-golems/bobatea/pkg/repl"
)

// MathEvaluator is a simple math evaluator for demonstration
type MathEvaluator struct{}

func NewEvaluator() *MathEvaluator {
	return &MathEvaluator{}
}

func (m *MathEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)

	// Simple arithmetic operations
	if strings.Contains(code, "+") {
		parts := strings.Split(code, "+")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			b, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				return fmt.Sprintf("%.2f", a+b), nil
			}
		}
	}

	if strings.Contains(code, "-") {
		parts := strings.Split(code, "-")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			b, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				return fmt.Sprintf("%.2f", a-b), nil
			}
		}
	}

	if strings.Contains(code, "*") {
		parts := strings.Split(code, "*")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			b, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				return fmt.Sprintf("%.2f", a*b), nil
			}
		}
	}

	if strings.Contains(code, "/") {
		parts := strings.Split(code, "/")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			b, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				if b == 0 {
					return "", fmt.Errorf("division by zero")
				}
				return fmt.Sprintf("%.2f", a/b), nil
			}
		}
	}

	// Try to parse as number
	if val, err := strconv.ParseFloat(code, 64); err == nil {
		return fmt.Sprintf("%.2f", val), nil
	}

	return "", fmt.Errorf("unsupported expression: %s", code)
}

func (m *MathEvaluator) GetPrompt() string {
	return "math> "
}

func (m *MathEvaluator) GetName() string {
	return "Math"
}

func (m *MathEvaluator) SupportsMultiline() bool {
	return false
}

func (m *MathEvaluator) GetFileExtension() string {
	return ".txt"
}

var _ repl.Evaluator = (*MathEvaluator)(nil)
