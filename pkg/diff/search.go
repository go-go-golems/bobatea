package diff

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// searchModel holds a text input and state for visible search widget
type searchModel struct {
	input   textinput.Model
	visible bool
	query   string
}

// newSearchModel creates a new search model
func newSearchModel() searchModel {
	input := textinput.New()
	input.Placeholder = "Search"
	input.Prompt = ""
	input.CharLimit = 0
	return searchModel{
		input:   input,
		visible: false,
		query:   "",
	}
}

// Show makes the search widget visible and focuses the input
func (s *searchModel) Show() {
	s.visible = true
	s.input.Focus()
}

// Hide makes the search widget invisible and clears the query
func (s *searchModel) Hide() {
	s.visible = false
	s.query = ""
	s.input.Blur()
	s.input.SetValue("")
}

// Update handles tea messages for the search input
func (s *searchModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.query = strings.TrimSpace(s.input.Value())
	return cmd
}

// View renders the search input
func (s *searchModel) View() string {
	if !s.visible {
		return ""
	}
	return s.input.View()
}

// SetWidth sets the width of the search input
func (s *searchModel) SetWidth(width int) {
	if width > 4 {
		s.input.Width = width - 4
	} else {
		s.input.Width = width
	}
}

// Query returns the current search query
func (s *searchModel) Query() string {
	return s.query
}

// Visible returns whether the search widget is visible
func (s *searchModel) Visible() bool {
	return s.visible
}
