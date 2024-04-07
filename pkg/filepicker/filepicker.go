package filepicker

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/buttons"
	bubblesfilepicker "github.com/go-go-golems/bobatea/pkg/thirdparty/bubbles/filepicker"
	"os"
	"path"
	"strings"
	"time"
)

type SelectFileMsg struct {
	Path string
}
type CancelFilePickerMsg struct{}

type clearError struct{}

//goland:noinspection GoMixedReceiverTypes
type Model struct {
	Title string
	Error string

	Filepicker    bubblesfilepicker.Model
	confirmDialog buttons.Model
	help          help.Model
	filenameInput textinput.Model

	width             int
	height            int
	state             state
	creatingDirectory bool

	keyMap KeyMap

	SelectedPath string
}

type state string

const (
	stateBrowse           state = "browse"
	stateNewFile          state = "new-file"
	stateConfirmNew       state = "confirm-new"
	stateConfirmOverwrite state = "confirm-overwrite"
)

func NewModel() Model {
	fp := bubblesfilepicker.New()
	filenameInput := textinput.New()
	filenameInput.Focus()
	keyMap := DefaultKeyMap()

	return Model{
		Filepicker:    fp,
		confirmDialog: newConfirmCreateDialog(""),
		filenameInput: filenameInput,
		help:          help.New(),
		keyMap:        keyMap,
		state:         stateBrowse,
	}
}

func newConfirmCreateDialog(question string) buttons.Model {
	return buttons.NewModel(
		buttons.WithQuestion(question),
		buttons.WithButtons("No", "Yes"),
		buttons.WithActiveButton("Yes"),
	)
}

func newConfirmOverwriteDialog(filename string) buttons.Model {
	return buttons.NewModel(
		buttons.WithQuestion(fmt.Sprintf("Overwrite file %s?", filename)),
		buttons.WithButtons("No", "Yes"),
		buttons.WithActiveButton("No"),
	)
}

func (m Model) Init() tea.Cmd {
	return m.Filepicker.Init()
}

func (m *Model) resize() tea.Cmd {
	m_, cmd := m.Update(tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.height,
	})
	m.Filepicker = m_.Filepicker
	return cmd
}

func (m *Model) setError(err string) tea.Cmd {
	m.Error = err
	m.state = stateBrowse

	if m.Error != "" {
		return tea.Batch(m.resize(), func() tea.Msg {
			time.Sleep(3 * time.Second)
			return clearError{}
		})
	}

	return m.resize()
}

func (m *Model) enterNewFile(creatingDirectory bool) tea.Cmd {
	m.filenameInput.Reset()
	m.filenameInput.Focus()
	m.state = stateNewFile
	m.creatingDirectory = creatingDirectory
	return m.resize()
}

func (m *Model) enterConfirmNew() tea.Cmd {
	fileName := path.Base(m.SelectedPath)
	question := fmt.Sprintf("Create new file %s?", fileName)
	if m.creatingDirectory {
		question = fmt.Sprintf("Create new directory %s?", fileName)
	}
	m.confirmDialog = newConfirmCreateDialog(question)
	m.state = stateConfirmNew
	return m.resize()
}

func (m *Model) enterConfirmOverwrite() tea.Cmd {
	fileName := path.Base(m.SelectedPath)
	m.confirmDialog = newConfirmOverwriteDialog(fileName)
	m.state = stateConfirmOverwrite
	return m.resize()
}

func (m *Model) enterBrowse() tea.Cmd {
	m.state = stateBrowse
	return m.resize()
}

func (m Model) selectFile() tea.Cmd {
	return func() tea.Msg {
		return SelectFileMsg{
			Path: m.SelectedPath,
		}
	}
}

func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	if key.Matches(msg, m.keyMap.Help) {
		m.help.ShowAll = !m.help.ShowAll
		return cmd
	}

	switch m.state {
	case stateBrowse:
		switch {
		case key.Matches(msg, m.keyMap.CreateFile):
			return m.enterNewFile(false)

		case key.Matches(msg, m.keyMap.CreateDirectory):
			return m.enterNewFile(true)

		case key.Matches(msg, m.keyMap.Exit):
			return func() tea.Msg {
				return CancelFilePickerMsg{}
			}

		default:
			m.Filepicker, cmd = m.Filepicker.Update(msg)
			return cmd
		}
	case stateNewFile:
		switch {
		case key.Matches(msg, m.keyMap.Accept):
			if m.filenameInput.Value() != "" {
				m.SelectedPath = path.Join(m.Filepicker.CurrentDirectory, m.filenameInput.Value())
				if m.creatingDirectory {
					return m.enterConfirmNew()
				} else {
					if fi, err := os.Stat(m.SelectedPath); err == nil && fi.IsDir() {
						return tea.Batch(m.setError("File is a directory"), m.enterBrowse())
					}
					return m.enterConfirmNew()
				}
			}
			return m.enterBrowse()

		case key.Matches(msg, m.keyMap.Exit):
			return m.enterBrowse()

		case key.Matches(msg, m.keyMap.ResetFileInput):
			m.filenameInput.Reset()
			return nil

		default:
			m.filenameInput, cmd = m.filenameInput.Update(msg)
			return cmd
		}

	case stateConfirmNew, stateConfirmOverwrite:
		switch {
		case key.Matches(msg, m.keyMap.Exit):
			return m.enterBrowse()
		default:
			m.confirmDialog, cmd = m.confirmDialog.Update(msg)
		}
		return cmd
	}

	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case clearError:
		cmds = append(cmds, m.setError(""))

	case buttons.AbortedMsg:
		cmds = append(cmds, m.enterBrowse())

	case buttons.SelectedMsg:
		switch m.state {
		case stateConfirmNew:
			switch msg.Name {
			case "No":
				cmds = append(cmds, m.enterBrowse())
			case "Yes":
				cmd = func() tea.Cmd {
					fi, err := os.Stat(m.SelectedPath)

					if m.creatingDirectory {
						if err != nil {
							if os.IsNotExist(err) {
								err = os.Mkdir(m.SelectedPath, 0755)
								if err != nil {
									return tea.Batch(
										m.setError(err.Error()),
										m.enterBrowse(),
									)
								}
								return tea.Batch(
									m.enterDir(m.SelectedPath),
									m.enterBrowse(),
								)
							} else {
								return tea.Batch(
									m.setError(err.Error()),
									m.enterBrowse(),
								)
							}
						}

						if fi.IsDir() {
							return tea.Batch(
								m.enterDir(m.SelectedPath),
								m.enterBrowse(),
							)
						}

						return tea.Batch(
							m.setError("File or Directory already exists"),
							m.enterBrowse(),
						)
					} else {
						if err != nil {
							if os.IsNotExist(err) {
								return tea.Batch(
									m.selectFile(),
									m.enterBrowse(),
								)
							} else {
								return tea.Batch(
									m.setError(err.Error()),
									m.enterBrowse(),
								)
							}
						}
						if fi.IsDir() {
							return tea.Batch(
								m.setError("File is a directory"),
								m.enterBrowse(),
							)
						}

						return m.enterConfirmOverwrite()
					}

				}()
				cmds = append(cmds, cmd)
			}

		case stateConfirmOverwrite:
			switch msg.Name {
			case "No":
				cmds = append(cmds, m.enterBrowse())
			case "Yes":
				cmds = append(cmds,
					m.selectFile(),
					m.enterBrowse(),
				)
			}

		case stateBrowse, stateNewFile:
		}

	case tea.KeyMsg:
		cmds = append(cmds, m.handleKey(msg))

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.help.Width = msg.Width
		m.filenameInput.Width = msg.Width

		errorView := m.viewError()
		errorViewHeight := lipgloss.Height(errorView)
		if errorViewHeight > 0 {
			errorViewHeight++
		}

		helpView := m.help.View(m.keyMap)
		helpViewHeight := lipgloss.Height(helpView)
		titleViewHeight := lipgloss.Height(m.viewTitle(""))

		switch m.state {
		case stateBrowse:
			fpHeight := msg.Height - helpViewHeight - errorViewHeight - titleViewHeight

			m.Filepicker, cmd = m.Filepicker.Update(tea.WindowSizeMsg{
				Width:  msg.Width,
				Height: fpHeight,
			})
			cmds = append(cmds, cmd)

		case stateConfirmNew, stateConfirmOverwrite, stateNewFile:
			m.confirmDialog, cmd = m.confirmDialog.Update(tea.WindowSizeMsg{
				Width:  msg.Width,
				Height: msg.Height - helpViewHeight - errorViewHeight,
			})
			cmds = append(cmds, cmd)
		}

	case bubblesfilepicker.NavigateToDirMsg:
		m.Filepicker, cmd = m.Filepicker.Update(msg)
		cmds = append(cmds, cmd)

	default:
		switch m.state {
		case stateBrowse:
			m.Filepicker, cmd = m.Filepicker.Update(msg)
			cmds = append(cmds, cmd)
		case stateNewFile:
			m.filenameInput, cmd = m.filenameInput.Update(msg)
			cmds = append(cmds, cmd)
		case stateConfirmNew, stateConfirmOverwrite:
			m.confirmDialog, cmd = m.confirmDialog.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.state == stateBrowse {
		if ok, path := m.Filepicker.DidSelectFile(msg); ok {
			m.SelectedPath = path
			m.confirmDialog = newConfirmOverwriteDialog(m.SelectedPath)
			m.state = stateConfirmOverwrite
		}
	}

	m.keyMap.UpdateKeyBindings(m.state, m.filenameInput.Value())

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.state {
	case stateConfirmNew, stateConfirmOverwrite:
		return m.viewConfirm()
	case stateBrowse:
		return m.viewBrowse()

	case stateNewFile:
		return m.viewNewFile()
	}

	return ""
}

func (m Model) viewBrowse() string {
	helpView := m.help.View(m.keyMap)
	errorView := m.viewError()
	headerView := m.viewTitle(fmt.Sprintf("Current Directory: %s", m.Filepicker.CurrentDirectory))
	listView := m.Filepicker.View()
	listViewHeight := lipgloss.Height(listView)
	_ = listViewHeight

	return strings.Join([]string{
		headerView, listView, errorView, helpView,
	}, "\n")
}

func (m Model) viewInput() string {
	inputTitle := "New file name:"
	if m.creatingDirectory {
		inputTitle = "New directory name:"
	}
	inputView := m.filenameInput.View()

	return strings.Join([]string{
		inputTitle, inputView,
	}, "\n")
}

func (m Model) viewError() string {
	if m.Error == "" {
		return ""
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F25D94")).
		Render(m.Error)
}

func (m Model) viewTitle(title string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFCC66")).
		Render(title) + "\n"
}

func (m Model) viewNewFile() string {
	helpView := m.help.View(m.keyMap)
	listView := m.Filepicker.View()
	listViewHeight := lipgloss.Height(listView)
	_ = listViewHeight
	inputView := m.viewInput()
	errorView := m.viewError()

	return strings.Join([]string{
		listView, inputView, errorView, helpView,
	}, "\n")
}

func (m Model) viewConfirm() string {
	helpView := m.help.View(m.keyMap)
	confirmView := m.confirmDialog.View()
	errorView := m.viewError()

	return strings.Join([]string{
		confirmView, errorView, helpView,
	}, "\n")
}

func (m Model) enterDir(selectedPath string) tea.Cmd {
	return func() tea.Msg {
		return bubblesfilepicker.NavigateToDirMsg{
			Path: selectedPath,
		}
	}
}
