package main

import (
	"fmt"
	"github.com/charmbracelet/bubbletea"
)
import "github.com/go-go-golems/bobatea/pkg/filepicker"

type Model struct {
	fp           filepicker.Model
	selectedPath string
}

func NewModel() Model {
	fp := filepicker.NewModel()
	fp.Filepicker.DirAllowed = false
	fp.Filepicker.FileAllowed = true
	fp.Filepicker.CurrentDirectory = "/home/manuel"
	fp.Filepicker.Height = 10

	return Model{
		fp: fp,
	}
}

func (m Model) Init() tea.Cmd {
	return m.fp.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case filepicker.SelectFileMsg:
		m.selectedPath = msg.Path
		return m, tea.Quit

	case filepicker.CancelFilePickerMsg:
		fmt.Println("Cancelled")
		return m, tea.Quit

	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		default:
			m.fp, cmd = m.fp.Update(msg)
			return m, cmd
		}

	default:
		m.fp, cmd = m.fp.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {
	return m.fp.View()
}

func main() {
	b := NewModel()

	p := tea.NewProgram(b)
	var err error
	var m tea.Model
	if m, err = p.Run(); err != nil {
		panic(err)
	}

	fmt.Println("Selected path:", m.(Model).selectedPath)
}
