package main

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/textarea"
)

type Model struct {
	ta textarea.Model
}

func NewModel() Model {
	ta := textarea.New()
	ta.Focus()
	ta.CharLimit = 8000

	return Model{
		ta: ta,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		//exhaustive:ignore
		switch msg.Type {
		case tea.KeyEsc:
			if m.ta.Focused() {
				m.ta.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.ta.Focused() {
				cmd = m.ta.Focus()
				cmds = append(cmds, cmd)
			}

			m.ta, cmd = m.ta.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.ta.SetWidth(msg.Width)
		m.ta.SetHeight(msg.Height)

	default:
		m.ta, cmd = m.ta.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.ta.View()
}

func main() {
	b := NewModel()

	p := tea.NewProgram(b)
	var err error
	var m tea.Model
	if m, err = p.Run(); err != nil {
		panic(err)
	}

	fmt.Println("Typed text:", m.(Model).ta.Value())
}
