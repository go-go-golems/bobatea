package main

import (
	"log"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// SimpleEmbeddingExample demonstrates the simplest way to embed a REPL
func main() {
	// Create a simple evaluator
	evaluator := repl.NewExampleEvaluator()

	// Create a default configuration
	config := repl.DefaultConfig()
	config.Title = "Simple Embedding Example"

	// Create the REPL model
	model := repl.NewModel(evaluator, config)

	// Start the TUI program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
