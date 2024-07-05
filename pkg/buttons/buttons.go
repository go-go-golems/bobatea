package buttons

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	buttons  []string
	active   int
	keyMap   KeyMap
	Question string
	Width    int
}

type ModelOption func(*Model)

func WithQuestion(question string) ModelOption {
	return func(m *Model) {
		m.Question = question
	}
}

func WithButtons(buttons ...string) ModelOption {
	return func(m *Model) {
		m.buttons = buttons
	}
}

func WithWidth(width int) ModelOption {
	return func(m *Model) {
		m.Width = width
	}
}

func WithActiveButton(button string) ModelOption {
	return func(m *Model) {
		for i, b := range m.buttons {
			if b == button {
				m.active = i
				return
			}
		}
	}
}

func NewModel(options ...ModelOption) Model {
	ret := Model{
		active: 0,
		Width:  100,
		keyMap: DefaultKeyMap(),
	}

	for _, option := range options {
		option(&ret)
	}

	return ret
}

type SelectedMsg struct {
	Index int
	Name  string
}

type AbortedMsg struct{}

func (m Model) Init() tea.Cmd {
	return nil
}

var (
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			Margin(0, 1)

	activeButtonStyle = buttonStyle.
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				Underline(true)

	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	_      = subtle
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.LeftButton):
			if m.active > 0 {
				m.active--
			}

		case key.Matches(msg, m.keyMap.RightButton):
			if m.active < len(m.buttons)-1 {
				m.active++
			}

		case key.Matches(msg, m.keyMap.Accept):
			cmd = func() tea.Msg {
				return SelectedMsg{
					Index: m.active,
					Name:  m.buttons[m.active],
				}
			}

		case key.Matches(msg, m.keyMap.Exit):
			cmd = func() tea.Msg {
				return AbortedMsg{}
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	buttons := []string{}
	for i, b := range m.buttons {
		if i == m.active {
			buttons = append(buttons, activeButtonStyle.Render(b))
		} else {
			buttons = append(buttons, buttonStyle.Render(b))
		}
	}
	buttons_ := lipgloss.JoinHorizontal(lipgloss.Top, buttons...)

	ui := buttons_
	if m.Question != "" {
		question := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).Render(m.Question)
		ui = lipgloss.JoinVertical(lipgloss.Center, question, buttons_)
	}

	dialog := lipgloss.Place(m.Width, 9,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
	)

	return dialog
}
