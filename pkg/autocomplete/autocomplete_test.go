package autocomplete_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/go-go-golems/bobatea/pkg/listbox"
	"github.com/stretchr/testify/assert"
)

func TestAutocompleteModel(t *testing.T) {
	// Create a test completioner that returns predetermined results
	completioner := func(ctx context.Context, query string) ([]autocomplete.Suggestion, error) {
		if query == "test" {
			return []autocomplete.Suggestion{
				{
					Id:          "test-1",
					Value:       "test-value-1",
					DisplayText: "Test Value 1",
					Submatches: []autocomplete.Span{
						{Start: 0, End: 4},
					},
				},
				{
					Id:          "test-2",
					Value:       "test-value-2",
					DisplayText: "Test Value 2",
					Submatches: []autocomplete.Span{
						{Start: 0, End: 4},
					},
				},
			}, nil
		}
		return nil, nil
	}

	// Create the model
	m := autocomplete.New(completioner, 40, 10)

	// Test the initial state
	assert.Nil(t, m.Selection)

	// Test typing into the input
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test")})
	m = model.(autocomplete.Model)

	// Wait for the completioner to be called
	// This is a bit hacky for testing but works for this simple test
	// In a real application, we'd use a proper testing strategy for async code
	suggestions := []autocomplete.Suggestion{
		{
			Id:          "test-1",
			Value:       "test-value-1",
			DisplayText: "Test Value 1",
			Submatches: []autocomplete.Span{
				{Start: 0, End: 4},
			},
		},
		{
			Id:          "test-2",
			Value:       "test-value-2",
			DisplayText: "Test Value 2",
			Submatches: []autocomplete.Span{
				{Start: 0, End: 4},
			},
		},
	}

	// Simulate getting suggestions
	model, cmd := m.Update(autocomplete.SuggestionsMsg(suggestions))
	m = model.(autocomplete.Model)

	// Execute the SetItems command if it was returned
	if cmd != nil {
		msg := cmd()
		if setItemsMsg, ok := msg.(interface{}); ok {
			// Process the setItems message
			model, _ = m.Update(setItemsMsg)
			m = model.(autocomplete.Model)
		}
	}

	// Verify we can select an item by simulating Enter key
	// First, let's make sure we have something to select
	selected := m.List.Selected()
	if selected != nil {
		// Simulate the selection by directly creating a SelectedMsg
		selectedMsg := listbox.SelectedMsg{Item: selected}
		model, _ = m.Update(selectedMsg)
		m = model.(autocomplete.Model)

		// Verify we selected something
		assert.NotNil(t, m.Selection)
		assert.Equal(t, "test-1", m.Selection.ID())
	} else {
		t.Fatal("No items available for selection")
	}
}
