package listbox_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/listbox"
	"github.com/stretchr/testify/assert"
)

// Simple test item implementation
type testItem struct {
	id      string
	display string
}

func (t testItem) ID() string      { return t.id }
func (t testItem) Display() string { return t.display }

func TestListModel(t *testing.T) {
	// Create test items
	items := []listbox.Item{
		testItem{id: "1", display: "Item 1"},
		testItem{id: "2", display: "Item 2"},
		testItem{id: "3", display: "Item 3"},
	}

	// Create list model with 2 visible items
	m := listbox.New(2)

	// Test initial state
	assert.Equal(t, 0, m.Cursor())
	assert.Nil(t, m.Selected())

	// Set items
	cmd := m.SetItems(items)
	assert.NotNil(t, cmd)

	// Apply the command
	model, _ := m.Update(cmd())
	m = model.(listbox.ListModel)

	// Check selection after setting items
	assert.Equal(t, 0, m.Cursor())
	assert.Equal(t, "1", m.Selected().ID())

	// Test down key
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(listbox.ListModel)
	assert.Equal(t, 1, m.Cursor())
	assert.Equal(t, "2", m.Selected().ID())

	// Test down key again
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(listbox.ListModel)
	assert.Equal(t, 2, m.Cursor())
	assert.Equal(t, "3", m.Selected().ID())

	// Test down key at boundary (shouldn't move)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(listbox.ListModel)
	assert.Equal(t, 2, m.Cursor())
	assert.Equal(t, "3", m.Selected().ID())

	// Test up key
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(listbox.ListModel)
	assert.Equal(t, 1, m.Cursor())
	assert.Equal(t, "2", m.Selected().ID())

	// Test selection
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(listbox.ListModel)

	// Check selection message
	msg := cmd()
	selectedMsg, ok := msg.(listbox.SelectedMsg)
	assert.True(t, ok)
	assert.Equal(t, "2", selectedMsg.Item.ID())

	// Test window size handling
	model, _ = m.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	m = model.(listbox.ListModel)

	// Test rendering with truncation
	view := m.View()
	assert.Contains(t, view, "Item 1")
	assert.Contains(t, view, "Item 2")
	assert.NotContains(t, view, "Item 3") // Should not be visible (maxVisible = 2)
}
