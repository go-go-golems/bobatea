package main

import "github.com/charmbracelet/bubbletea"
import "github.com/go-go-golems/bobatea/pkg/buttons"

type Model struct {
	buttons buttons.Model
}

func (m Model) Init() tea.Cmd {
	return m.buttons.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.buttons.Update(msg)
}

func (m Model) View() string {
	return m.buttons.View()
}

var _ tea.Model = Model{}

func NewModel() Model {
	model := buttons.NewModel([]string{
		"button1",
		"button2",
		"button3",
	})
	model.Question = "What is your favorite color?"
	ret := Model{
		buttons: model,
	}

	return ret
}

func main() {
	b := NewModel()

	p := tea.NewProgram(b)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
