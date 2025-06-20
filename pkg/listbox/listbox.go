package listbox

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

// Item is the minimal interface the list needs.
type Item interface {
	Display() string // what will be displayed
	ID() string      // unique identifier
}

// SelectedMsg is emitted when an item is selected
type SelectedMsg struct {
	Item Item
}

// setItemsMsg is used internally to update items
type setItemsMsg []Item

// ListModel represents a scrollable list component
type ListModel struct {
	items       []Item
	cursor      int    // current selection position
	offset      int    // top of visible window
	maxVisible  int    // maximum rows to display
	width       int    // current terminal width
	truncateW   int    // width available for text
	pointerRune string // selection indicator
}

// New creates a new ListModel with the specified maximum visible items
func New(maxVisible int) ListModel {
	return ListModel{
		maxVisible:  maxVisible,
		pointerRune: "› ",
	}
}

// Init initializes the model
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model accordingly
func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Internal message to set items
	case setItemsMsg:
		m.items = msg
		m.cursor, m.offset = 0, 0
		return m, nil

	// Key handling for navigation
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
				// Adjust offset if cursor moves out of visible window
				if m.cursor < m.offset {
					m.offset--
				}
			}
			return m, nil

		case "down", "j", "ctrl+n":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				// Adjust offset if cursor moves out of visible window
				if m.cursor >= m.offset+m.maxVisible {
					m.offset++
				}
			}
			return m, nil

		case "enter", "tab":
			// Only emit selection if we have items
			if len(m.items) > 0 {
				return m, func() tea.Msg {
					return SelectedMsg{Item: m.items[m.cursor]}
				}
			}
			return m, nil

		default:
			// Not consumed, bubble up to parent
			return m, nil
		}

	// Handle window size changes
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.truncateW = maxInt(0, m.width-len(m.pointerRune)-2) // Account for pointer and padding
		return m, nil
	}

	// Unhandled message, let parent handle it
	return m, nil
}

// View renders the list
func (m ListModel) View() string {
	if len(m.items) == 0 {
		return "(no items)"
	}

	var b strings.Builder
	limit := minInt(m.maxVisible, len(m.items))

	for i := 0; i < limit; i++ {
		idx := m.offset + i
		if idx >= len(m.items) {
			break
		}

		prefix := "  "
		if idx == m.cursor {
			prefix = m.pointerRune
		}

		// Get item text and truncate if needed
		text := m.items[idx].Display()
		if m.truncateW > 0 {
			text = runewidth.Truncate(text, m.truncateW, "…")
		}

		b.WriteString(prefix + text + "\n")
	}

	return b.String()
}

// SetItems updates the items in the list
func (m *ListModel) SetItems(items []Item) tea.Cmd {
	return func() tea.Msg {
		return setItemsMsg(items)
	}
}

// SetPointer changes the pointer rune used to indicate selection
func (m *ListModel) SetPointer(pointer string) {
	m.pointerRune = pointer
}

// Cursor returns the current cursor position
func (m ListModel) Cursor() int {
	return m.cursor
}

// Selected returns the currently selected item, or nil if there are no items
func (m ListModel) Selected() Item {
	if len(m.items) == 0 {
		return nil
	}
	return m.items[m.cursor]
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
