package filepicker

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type SelectFileMsg struct {
	Path string
}
type CancelFilePickerMsg struct{}

type Model struct {
	Filepicker    filepicker.Model
	Title         string
	Help          help.Model
	filenameInput textinput.Model
	state         state

	keyMap KeyMap
}

var (
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)
)

type state int

const (
	stateList state = iota
	stateConfirm
)

func NewModel() Model {
	fp := filepicker.New()
	filenameInput := textinput.New()
	filenameInput.Focus()
	keyMap := DefaultKeyMap()

	return Model{
		Filepicker:    fp,
		filenameInput: filenameInput,
		Help:          help.New(),
		state:         stateList,
		keyMap:        keyMap,
	}
}

func (m Model) Init() tea.Cmd {
	return m.Filepicker.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == stateList {
			switch {
			case key.Matches(msg,
				m.keyMap.GoToTop, m.keyMap.GoToLast,
				m.keyMap.Down, m.keyMap.Up,
				m.keyMap.PageUp, m.keyMap.PageDown,
				m.keyMap.Back):
				m.Filepicker, cmd = m.Filepicker.Update(msg)
				cmds = append(cmds, cmd)

			case key.Matches(msg, m.keyMap.ResetFileInput):
				m.filenameInput.Reset()

			case key.Matches(msg, m.keyMap.Accept):
				if m.filenameInput.Value() != "" {
					m.state = stateConfirm
				} else {
					m.Filepicker, cmd = m.Filepicker.Update(msg)
					if ok, path := m.Filepicker.DidSelectFile(msg); ok {
						return m, func() tea.Msg {
							return SelectFileMsg{
								Path: path,
							}
						}
					}
					cmds = append(cmds, cmd)
				}

			case key.Matches(msg, m.keyMap.Exit):
				return m, func() tea.Msg {
					return CancelFilePickerMsg{}
				}

			case key.Matches(msg, m.keyMap.Help):
				m.Help, cmd = m.Help.Update(msg)

			default:
				m.filenameInput, cmd = m.filenameInput.Update(msg)
			}
		}

	case tea.WindowSizeMsg:
		helpView := m.Help.View(m.keyMap)
		helpViewHeight := lipgloss.Height(helpView)
		fpHeight := msg.Height - helpViewHeight

		switch m.state {
		case stateList:
			inputView := m.filenameInput.View()
			inputViewHeight := lipgloss.Height(inputView)
			fpHeight -= inputViewHeight

		case stateConfirm:

		}
	default:
		m.Filepicker, cmd = m.Filepicker.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.keyMap.UpdateKeyBindings(m.state, m.filenameInput.Value())

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.state == stateList {
		return m.viewList()
	}
	return m.viewConfirm()
}

func (m Model) viewList() string {
	helpView := m.Help.View(m.keyMap)
	listView := m.Filepicker.View()
	inputView := m.filenameInput.View()

	return strings.Join([]string{
		listView, inputView, helpView,
	}, "\n")
}

func (m Model) viewConfirm() string {
	return ""
}
