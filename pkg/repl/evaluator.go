package repl

import (
	"context"
)

// Evaluator represents the interface for pluggable evaluators
type Evaluator interface {
	// Evaluate executes the given code and returns the result
	Evaluate(ctx context.Context, code string) (string, error)

	// GetPrompt returns the prompt string for this evaluator
	GetPrompt() string

	// GetName returns the name of this evaluator (for display)
	GetName() string

	// SupportsMultiline returns true if this evaluator supports multiline input
	SupportsMultiline() bool

	// GetFileExtension returns the file extension for external editor
	GetFileExtension() string
}

// EvaluatorResult represents the result of an evaluation
type EvaluatorResult struct {
	Output string
	Error  error
}

// Config holds configuration for the REPL
type Config struct {
	Title                string
	Prompt               string
	Placeholder          string
	Width                int
	StartMultiline       bool
	EnableExternalEditor bool
	EnableHistory        bool
	MaxHistorySize       int
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Title:                "REPL",
		Prompt:               "> ",
		Placeholder:          "Enter code or /command",
		Width:                80,
		StartMultiline:       false,
		EnableExternalEditor: true,
		EnableHistory:        true,
		MaxHistorySize:       1000,
	}
}
