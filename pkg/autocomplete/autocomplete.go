package autocomplete

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/listbox"
)

// Span represents a range of runes to highlight within a suggestion.
type Span struct {
	Start, End int // rune offsets, half-open [Start,End)
}

// Suggestion is a completion result with associated metadata and highlight regions.
type Suggestion struct {
	Id          string // Unique identifier
	Value       string // The actual value (for programmatic use)
	DisplayText string // The displayed value (for viewing)
	Submatches  []Span // Regions to highlight
}

// ID returns the unique identifier of the suggestion (implements listbox.Item)
func (s Suggestion) ID() string { return s.Id }

// Display returns the display text of the suggestion (implements listbox.Item)
func (s Suggestion) Display() string { return s.DisplayText }

// SuggestionsMsg is a message containing new completion suggestions.
type SuggestionsMsg []Suggestion

// DoneMsg is a message indicating a suggestion was selected.
type DoneMsg Suggestion

// ErrMsg is a message containing an error from the completioner.
type ErrMsg error

// Completioner is a function that provides suggestions for a query string.
type Completioner func(ctx context.Context, query string) ([]Suggestion, error)

// Model is the Bubble Tea model for the autocomplete widget.
type Model struct {
	Input     textinput.Model
	List      listbox.ListModel
	Selection *Suggestion
	Width     int
	Height    int

	completioner Completioner
}

// New creates a new autocomplete model with the given completioner.
func New(completioner Completioner, width, height int) Model {
	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.Focus()

	// Initialize listbox with maximum visible items
	li := listbox.New(height)

	return Model{
		Input:        ti,
		List:         li,
		completioner: completioner,
		Width:        width,
		Height:       height,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case listbox.SelectedMsg:
		// Handle item selection from listbox
		if suggestion, ok := msg.Item.(Suggestion); ok {
			m.Selection = &suggestion
			return m, func() tea.Msg { return DoneMsg(suggestion) }
		}

	case tea.KeyMsg:
		// Handle input updates
		var cmd tea.Cmd
		m.Input, cmd = m.Input.Update(msg)

		// Trigger completion after input changes
		return m, tea.Batch(cmd, m.completionCmd(m.Input.Value()))

	case SuggestionsMsg:
		// Set items in the listbox
		items := make([]listbox.Item, len(msg))
		for i, s := range msg {
			items[i] = s
		}
		return m, m.List.SetItems(items)

	case ErrMsg:
		// Simply ignore errors for now
		return m, nil
	}

	// Forward unhandled messages to list
	newList, cmd := m.List.Update(msg)
	m.List = newList.(listbox.ListModel)
	return m, cmd
}

// View renders the model.
func (m Model) View() string {
	return m.Input.View() + "\n\n" + m.List.View()
}

// completionCmd creates a command that fetches completions asynchronously.
func (m Model) completionCmd(query string) tea.Cmd {
	return func() tea.Msg {
		// Simple debounce to avoid excessive API calls
		time.Sleep(120 * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
		defer cancel()

		suggestions, err := m.completioner(ctx, query)
		if err != nil {
			return ErrMsg(err)
		}
		return SuggestionsMsg(suggestions)
	}
}

// CustomizeStyles allows modifying the list's pointer style
func (m *Model) CustomizeStyles(pointer string) {
	m.List.SetPointer(pointer)
}

// ApplyHighlighting applies highlighting to all suggestions in the list
// You should call this after setting items in the list if you want to display highlighted text
func (m *Model) ApplyHighlighting() {
	// Our custom listbox doesn't support highlighting directly
	// You'd need to modify pkg/listbox to support custom rendering for highlighted text
	// or implement a delegate system similar to the original list implementation
}
