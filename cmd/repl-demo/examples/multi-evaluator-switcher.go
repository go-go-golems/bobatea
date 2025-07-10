package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
)

// MultiEvaluatorSwitcher demonstrates switching between different evaluators
type MultiEvaluatorSwitcher struct {
	evaluators   map[string]repl.Evaluator
	currentEval  string
	currentModel *repl.Model
}

func NewMultiEvaluatorSwitcher() *MultiEvaluatorSwitcher {
	// Create evaluators
	jsEval, err := javascript.NewWithDefaults()
	if err != nil {
		log.Fatal(err)
	}

	evaluators := map[string]repl.Evaluator{
		"js":      jsEval,
		"example": repl.NewExampleEvaluator(),
	}

	return &MultiEvaluatorSwitcher{
		evaluators:  evaluators,
		currentEval: "js",
	}
}

func (m *MultiEvaluatorSwitcher) createModelForEvaluator(evalName string) repl.Model {
	eval, ok := m.evaluators[evalName]
	if !ok {
		eval = m.evaluators["example"] // fallback
		evalName = "example"
	}

	config := repl.DefaultConfig()
	config.Title = fmt.Sprintf("Multi-Evaluator REPL - Current: %s", eval.GetName())

	model := repl.NewModel(eval, config)

	// Add switch command
	model.AddCustomCommand("switch", func(args []string) tea.Cmd {
		if len(args) == 0 {
			available := make([]string, 0, len(m.evaluators))
			for name := range m.evaluators {
				available = append(available, name)
			}
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/switch",
					Output: fmt.Sprintf("Available evaluators: %v\nCurrent: %s", available, evalName),
					Error:  nil,
				}
			}
		}

		newEval := args[0]
		if _, ok := m.evaluators[newEval]; ok {
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/switch " + newEval,
					Output: fmt.Sprintf("Switched to %s evaluator. Note: History is preserved per evaluator.", m.evaluators[newEval].GetName()),
					Error:  nil,
				}
			}
		}

		return func() tea.Msg {
			return repl.EvaluationCompleteMsg{
				Input:  "/switch " + newEval,
				Output: fmt.Sprintf("Unknown evaluator: %s", newEval),
				Error:  fmt.Errorf("unknown evaluator: %s", newEval),
			}
		}
	})

	// Add list command
	model.AddCustomCommand("list", func(args []string) tea.Cmd {
		return func() tea.Msg {
			var result string
			for name, eval := range m.evaluators {
				marker := ""
				if name == evalName {
					marker = " (current)"
				}
				result += fmt.Sprintf("- %s: %s%s\n", name, eval.GetName(), marker)
			}
			return repl.EvaluationCompleteMsg{
				Input:  "/list",
				Output: "Available evaluators:\n" + result,
				Error:  nil,
			}
		}
	})

	return model
}

func main() {
	switcher := NewMultiEvaluatorSwitcher()

	// Create initial model
	model := switcher.createModelForEvaluator(switcher.currentEval)

	// Start the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
