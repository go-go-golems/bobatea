package commandpalette

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// Command represents a command that can be executed
type Command struct {
	Name        string
	Description string
	Action      func() tea.Cmd
}

// Model represents the command palette state
type Model struct {
	commands     []Command
	filteredCmds []Command
	query        string
	selected     int
	visible      bool
	maxVisible   int
	width        int
	height       int
	styles       Styles
}

// New creates a new command palette model
func New() Model {
	return Model{
		commands:     []Command{},
		filteredCmds: []Command{},
		query:        "",
		selected:     0,
		visible:      false,
		maxVisible:   8,
		styles:       DefaultStyles(),
	}
}

// WithStyles sets custom styles for the command palette
func (m Model) WithStyles(styles Styles) Model {
	m.styles = styles
	return m
}

// RegisterCommand adds a new command to the palette
func (m *Model) RegisterCommand(name, description string, action func() tea.Cmd) {
	m.commands = append(m.commands, Command{
		Name:        name,
		Description: description,
		Action:      action,
	})
	m.updateFiltered()
}

// SetCommands replaces all commands in the palette.
func (m *Model) SetCommands(commands []Command) {
	m.commands = append([]Command(nil), commands...)
	m.query = ""
	m.selected = 0
	m.updateFiltered()
}

// Show makes the command palette visible
func (m *Model) Show() {
	m.visible = true
	m.query = ""
	m.selected = 0
	m.updateFiltered()
}

// Hide makes the command palette invisible
func (m *Model) Hide() {
	m.visible = false
	m.query = ""
	m.selected = 0
}

// IsVisible returns whether the command palette is visible
func (m Model) IsVisible() bool {
	return m.visible
}

// SetSize sets the dimensions for the command palette
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetMaxVisible sets the maximum number of commands rendered in the list.
func (m *Model) SetMaxVisible(maxVisible int) {
	if maxVisible <= 0 {
		m.maxVisible = 8
		return
	}
	m.maxVisible = maxVisible
}

// updateFiltered updates the filtered commands based on the current query
func (m *Model) updateFiltered() {
	if m.query == "" {
		m.filteredCmds = m.commands
	} else {
		// Use fuzzy matching
		var targets []string
		for _, cmd := range m.commands {
			targets = append(targets, cmd.Name)
		}

		matches := fuzzy.Find(m.query, targets)
		m.filteredCmds = []Command{}

		for _, match := range matches {
			m.filteredCmds = append(m.filteredCmds, m.commands[match.Index])
		}
	}

	// Reset selection if out of bounds
	if m.selected >= len(m.filteredCmds) {
		m.selected = 0
	}
}

// Init initializes the command palette model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles command palette updates
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "escape", "ctrl+p":
			m.visible = false
			m.query = ""
			m.selected = 0
			return m, nil

		case "enter":
			if len(m.filteredCmds) > 0 && m.selected < len(m.filteredCmds) {
				cmd := m.filteredCmds[m.selected].Action()
				m.visible = false
				m.query = ""
				m.selected = 0
				return m, cmd
			}
			return m, nil

		case "up", "ctrl+k":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case "down", "ctrl+j":
			if m.selected < len(m.filteredCmds)-1 {
				m.selected++
			}
			return m, nil

		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				if m.query == "" {
					m.filteredCmds = m.commands
				} else {
					// Use fuzzy matching
					var targets []string
					for _, cmd := range m.commands {
						targets = append(targets, cmd.Name)
					}

					matches := fuzzy.Find(m.query, targets)
					m.filteredCmds = []Command{}

					for _, match := range matches {
						m.filteredCmds = append(m.filteredCmds, m.commands[match.Index])
					}
				}
				// Reset selection if out of bounds
				if m.selected >= len(m.filteredCmds) {
					m.selected = 0
				}
			}
			return m, nil

		default:
			if len(msg.String()) == 1 {
				m.query += msg.String()
				if m.query == "" {
					m.filteredCmds = m.commands
				} else {
					// Use fuzzy matching
					var targets []string
					for _, cmd := range m.commands {
						targets = append(targets, cmd.Name)
					}

					matches := fuzzy.Find(m.query, targets)
					m.filteredCmds = []Command{}

					for _, match := range matches {
						m.filteredCmds = append(m.filteredCmds, m.commands[match.Index])
					}
				}
				// Reset selection if out of bounds
				if m.selected >= len(m.filteredCmds) {
					m.selected = 0
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the command palette
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	// Header
	header := m.styles.Header.Render("Command Palette")

	// Query input
	queryPrompt := "> " + m.query
	if m.query == "" {
		queryPrompt = "> Type to search commands..."
	}
	query := m.styles.Query.Width(m.width - 12).Render(queryPrompt)

	// Commands list
	var commandLines []string
	maxCommands := m.maxVisible
	if maxCommands <= 0 {
		maxCommands = 8
	}

	for i, cmd := range m.filteredCmds {
		if i >= maxCommands {
			break
		}

		name := m.styles.CommandName.Render(cmd.Name)
		desc := m.styles.CommandDescription.Render(" - " + cmd.Description)
		line := name + desc

		if i == m.selected {
			line = m.styles.SelectedCommand.Width(m.width - 12).Render(line)
		} else {
			line = m.styles.Command.Width(m.width - 12).Render(line)
		}

		commandLines = append(commandLines, line)
	}

	if len(commandLines) == 0 {
		commandLines = append(commandLines, m.styles.Command.Render("No commands found"))
	}

	// Footer with navigation help
	footer := m.styles.Help.Render("↑↓ navigate • Enter select • Esc close")

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		query,
		strings.Join(commandLines, "\n"),
		"",
		footer,
	)

	return m.styles.Palette.Width(m.width - 8).Render(content)
}
