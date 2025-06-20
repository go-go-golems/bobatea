package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
)

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

type model struct {
	autocomplete autocomplete.Model
	done         bool
	selection    *autocomplete.Suggestion
}

func newModel() model {
	return model{
		autocomplete: autocomplete.New(demoCompletioner, 40, 10),
		done:         false,
	}
}

func (m model) Init() tea.Cmd {
	return m.autocomplete.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

	case autocomplete.DoneMsg:
		m.done = true
		m.selection = (*autocomplete.Suggestion)(&msg)
		return m, nil
	}

	var cmd tea.Cmd
	aModel, cmd := m.autocomplete.Update(msg)
	m.autocomplete = aModel.(autocomplete.Model)
	return m, cmd
}

func (m model) View() string {
	if m.done && m.selection != nil {
		return fmt.Sprintf(
			"You selected: ID=%s, Value=%s, Display=%s\n\nPress Ctrl+C to quit.",
			m.selection.ID(),
			m.selection.Value,
			m.selection.Display(),
		)
	}

	return "Autocomplete Demo\n\n" +
		"Type to search fruits (try 'app' or 'berry')\n\n" +
		m.autocomplete.View() + "\n\n" +
		"Up/Down: Navigate  Tab/Enter: Select  Ctrl+C: Quit"
}

// Sample data for the demo
var fruits = []string{
	"Apple", "Apricot", "Banana", "Blackberry",
	"Blueberry", "Cherry", "Cranberry", "Date",
	"Grape", "Grapefruit", "Kiwi", "Lemon",
	"Lime", "Mango", "Orange", "Papaya",
	"Peach", "Pear", "Pineapple", "Plum",
	"Raspberry", "Strawberry", "Watermelon",
}

// demoCompletioner provides simple substring matching with highlighting
func demoCompletioner(_ context.Context, query string) ([]autocomplete.Suggestion, error) {
	query = strings.ToLower(query)
	if query == "" {
		return nil, nil
	}

	var suggestions []autocomplete.Suggestion

	for i, fruit := range fruits {
		fruitLower := strings.ToLower(fruit)
		if strings.Contains(fruitLower, query) {
			// Find all occurrences of query in the fruit name
			var submatches []autocomplete.Span
			start := 0
			for {
				index := strings.Index(fruitLower[start:], query)
				if index == -1 {
					break
				}
				submatches = append(submatches, autocomplete.Span{
					Start: start + index,
					End:   start + index + len(query),
				})
				start += index + len(query)
			}

			suggestions = append(suggestions, autocomplete.Suggestion{
				Id:          fmt.Sprintf("fruit-%d", i),
				Value:       fruit,
				DisplayText: fruit,
				Submatches:  submatches,
			})
		}
	}

	return suggestions, nil
}
