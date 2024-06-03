package chat

import (
	context2 "context"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	conversationui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
	"github.com/go-go-golems/bobatea/pkg/conversation"
	"github.com/go-go-golems/bobatea/pkg/filepicker"
	mode_keymap "github.com/go-go-golems/bobatea/pkg/mode-keymap"
	"github.com/go-go-golems/bobatea/pkg/textarea"
	"github.com/go-go-golems/glazed/pkg/helpers/markdown"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type errMsg error

type State string

// TODO(manuel, 2024-01-15)
// we should also have a state that we are starting a completion
// (which will only really be finished until the subjacent steps are done, but how do we know that?)

const (
	StateUserInput        State = "user-input"
	StateMovingAround     State = "moving-around"
	StateStreamCompletion State = "stream-completion"
	StateSavingToFile     State = "saving-to-file"

	StateError State = "error"
)

type model struct {
	conversationManager conversation.Manager

	viewport       viewport.Model
	scrollToBottom bool

	// not really what we want, but use this for now, we'll have to either find a normal text box,
	// or implement wrapping ourselves.
	textArea textarea.Model

	filepicker filepicker.Model

	conversation conversationui.Model

	help help.Model

	err    error
	keyMap KeyMap

	style  *conversationui.Style
	width  int
	height int

	backend Backend

	state        State
	quitReceived bool

	title string
}

type ModelOption func(*model)

func WithTitle(title string) ModelOption {
	return func(m *model) {
		m.title = title
	}
}

// TODO(manuel, 2024-04-07) Add options to configure filepicker

func InitialModel(manager conversation.Manager, backend Backend, options ...ModelOption) model {
	fp := filepicker.NewModel()

	fp.Filepicker.DirAllowed = false
	fp.Filepicker.FileAllowed = true
	dir, _ := os.Getwd()
	fp.Filepicker.CurrentDirectory = dir
	fp.Filepicker.Height = 10

	ret := model{
		conversationManager: manager,
		conversation:        conversationui.NewModel(manager),
		filepicker:          fp,
		style:               conversationui.DefaultStyles(),
		keyMap:              DefaultKeyMap,
		backend:             backend,
		viewport:            viewport.New(0, 0),
		help:                help.New(),
		scrollToBottom:      true,
	}

	for _, option := range options {
		option(&ret)
	}

	ret.textArea = textarea.New()
	ret.textArea.Placeholder = "Dear AI, answer my plight..."
	ret.textArea.Focus()
	ret.textArea.CharLimit = 20000
	ret.textArea.MaxHeight = 500
	ret.state = StateUserInput

	return ret
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		textarea.Blink,
	}
	//err := clipboard.Init()
	//if err != nil {
	//	cmds = append(cmds, func() tea.Msg {
	//		return errMsg(err)
	//	})
	//}

	cmds = append(cmds, m.filepicker.Init(), m.viewport.Init(), m.conversation.Init())

	// TODO(manuel, 2024-04-07) this probably belongs into init, and maybe a separate init message?
	messages := m.conversation.View()
	m.viewport.SetContent(messages)
	m.viewport.YPosition = 0
	m.viewport.GotoBottom()

	m.updateKeyBindings()

	return tea.Batch(cmds...)
}

func (m *model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	_ = k

	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keyMap.Help):
		m.help.ShowAll = !m.help.ShowAll

	case key.Matches(msg, m.keyMap.UnfocusMessage):
		if m.state == StateUserInput {
			m.textArea.Blur()
			m.state = StateMovingAround
			m.conversation.SetActive(true)
			if m.scrollToBottom {
				m.conversation.SetSelectedIdx(len(m.conversation.Conversation()) - 1)
			}
			m.scrollToSelected()
			m.updateKeyBindings()
		}

	case key.Matches(msg, m.keyMap.Quit):
		if !m.quitReceived {
			m.quitReceived = true
			// on first quit, try to cancel completion if running.
			// NOTE(manuel, 2024-01-15) Maybe we should also check for the state here, add some invariants.
			if !m.backend.IsFinished() {
				m.backend.Interrupt()
			}
		}

		// force save completion before quitting
		// TODO(manuel, 2024-01-15) Actually we just need to kill and then append the current response, right?
		// But if we kill we might get another completion response and then we would have two messages.
		// Maybe we should just do the right thing and implementing a Quitting state...
		m.finishCompletion()

		cmd = tea.Quit

	case key.Matches(msg, m.keyMap.FocusMessage):
		// TODO(manuel, 2024-01-06) This could potentially focus on a previous message
		// and allow us to regenerate.
		cmd = m.textArea.Focus()

		m.scrollToBottom = true
		m.viewport.GotoBottom()

		m.conversation.SetActive(false)
		m.state = StateUserInput
		m.updateKeyBindings()

	case key.Matches(msg, m.keyMap.SelectNextMessage):
		messages := m.conversation.Conversation()
		if m.conversation.SelectedIdx() < len(messages)-1 {
			m.conversation.SetSelectedIdx(m.conversation.SelectedIdx() + 1)
			m.scrollToSelected()
		} else if m.conversation.SelectedIdx() == len(messages)-1 {
			m.scrollToBottom = true
			m.viewport.GotoBottom()
		}

	case key.Matches(msg, m.keyMap.SelectPrevMessage):
		if m.conversation.SelectedIdx() > 0 {
			m.conversation.SetSelectedIdx(m.conversation.SelectedIdx() - 1)
			m.scrollToSelected()
			m.scrollToBottom = false
		}

	case key.Matches(msg, m.keyMap.SubmitMessage):
		cmd = m.submit()

	case key.Matches(msg, m.keyMap.CopyToClipboard):
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					err := clipboard.WriteAll(msg_.Content.String())
					if err != nil {
						cmd = func() tea.Msg {
							return errMsg(err)
						}
					}
				}
			} else {
				text := ""
				for _, m := range msgs {
					if content, ok := m.Content.(*conversation.ChatMessageContent); ok {
						if content.Role == conversation.RoleAssistant {
							text += content.Text + "\n"
						}
					}
				}
				err := clipboard.WriteAll(text)
				if err != nil {
					cmd = func() tea.Msg {
						return errMsg(err)
					}
				}
			}
		}

	case key.Matches(msg, m.keyMap.CopyLastResponseToClipboard):
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					if content, ok := msg_.Content.(*conversation.ChatMessageContent); ok {
						err := clipboard.WriteAll(content.Text)
						if err != nil {
							cmd = func() tea.Msg {
								return errMsg(err)
							}
						}
					}
				}
			} else {
				if m.state == StateUserInput {
					lastMsg := msgs[len(msgs)-1]
					if content, ok := lastMsg.Content.(*conversation.ChatMessageContent); ok {
						err := clipboard.WriteAll(content.Text)
						if err != nil {
							cmd = func() tea.Msg {
								return errMsg(err)
							}
						}
					}
				}
			}
		}

	case key.Matches(msg, m.keyMap.CopyLastSourceBlocksToClipboard):
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					if content, ok := msg_.Content.(*conversation.ChatMessageContent); ok {
						code := markdown.ExtractQuotedBlocks(content.Text, false)
						err := clipboard.WriteAll(strings.Join(code, "\n"))
						if err != nil {
							cmd = func() tea.Msg {
								return errMsg(err)
							}
						}
					}
				}
			} else {
				if m.state == StateUserInput {
					text := ""
					for _, m := range msgs {
						if content, ok := m.Content.(*conversation.ChatMessageContent); ok {
							if content.Role == conversation.RoleAssistant {
								text += content.Text + "\n"
							}
						}
					}
					code := markdown.ExtractQuotedBlocks(text, false)
					err := clipboard.WriteAll(strings.Join(code, "\n"))
					if err != nil {
						cmd = func() tea.Msg {
							return errMsg(err)
						}
					}
				}
			}
		}

	case key.Matches(msg, m.keyMap.CopySourceBlocksToClipboard):
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					if content, ok := msg_.Content.(*conversation.ChatMessageContent); ok {
						code := markdown.ExtractQuotedBlocks(content.Text, false)
						err := clipboard.WriteAll(strings.Join(code, "\n"))
						if err != nil {
							cmd = func() tea.Msg {
								return errMsg(err)
							}
						}
					}
				}
			} else {
				text := ""
				for _, m := range msgs {
					if content, ok := m.Content.(*conversation.ChatMessageContent); ok {
						if content.Role == conversation.RoleAssistant {
							text += content.Text + "\n"
						}
					}
				}
				code := markdown.ExtractQuotedBlocks(text, false)
				err := clipboard.WriteAll(strings.Join(code, "\n"))
				if err != nil {
					cmd = func() tea.Msg {
						return errMsg(err)
					}
				}
			}
		}

	case key.Matches(msg, m.keyMap.SaveToFile):
		m.state = StateSavingToFile
		// need to reload the directory the filepicker is in
		cmd = m.filepicker.Init()
		m.recomputeSize()
		m.updateKeyBindings()

	// same keybinding for both
	case key.Matches(msg, m.keyMap.CancelCompletion):
		if m.state == StateStreamCompletion {
			m.backend.Interrupt()
		}

	case key.Matches(msg, m.keyMap.DismissError):
		if m.state == StateError {
			m.err = nil
			m.state = StateUserInput
			m.updateKeyBindings()
		}

	default:
		switch m.state {
		case StateUserInput:
			m.textArea, cmd = m.textArea.Update(msg)
		case StateSavingToFile:
			m.filepicker, cmd = m.filepicker.Update(msg)
		case StateMovingAround, StateStreamCompletion, StateError:
			prevAtBottom := m.viewport.AtBottom()
			m.viewport, cmd = m.viewport.Update(msg)
			if m.viewport.AtBottom() && !prevAtBottom {
				m.scrollToBottom = false
			}
		}
	}

	return m, cmd
}

func (m model) saveToFile(path string) (tea.Model, tea.Cmd) {
	err := m.conversationManager.SaveToFile(path)
	if err != nil {
		return m, func() tea.Msg {
			return errMsg(err)
		}
	}

	m.state = StateUserInput
	m.updateKeyBindings()
	m.recomputeSize()

	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg_ := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg_)

	case tea.WindowSizeMsg:
		m.width = msg_.Width
		m.height = msg_.Height

		m.recomputeSize()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg_
		return m, nil

	case conversationui.StreamCompletionMsg,
		conversationui.StreamStartMsg,
		conversationui.StreamStatusMsg,
		conversationui.StreamDoneMsg,
		conversationui.StreamCompletionError:
		// is CompletionMsg, we need to getNextCompletion
		m.conversation, cmd = m.conversation.Update(msg)
		if m.scrollToBottom {
			v, _ := m.conversation.ViewAndSelectedPosition()
			m.viewport.SetContent(v)
			m.viewport.GotoBottom()
		}
		cmds = append(cmds, cmd)

	case BackendFinishedMsg:
		cmd = m.finishCompletion()
		cmds = append(cmds, cmd)

	case refreshMessageMsg:
		v, _ := m.conversation.ViewAndSelectedPosition()
		m.viewport.SetContent(v)
		m.recomputeSize()
		if msg_.GoToBottom || m.scrollToBottom {
			m.viewport.GotoBottom()
		}

	case filepicker.SelectFileMsg:
		return m.saveToFile(msg_.Path)

	case filepicker.CancelFilePickerMsg:
		m.state = StateUserInput
		m.updateKeyBindings()

	default:
		switch m.state {
		case StateUserInput, StateError, StateMovingAround, StateStreamCompletion:
			m.viewport, cmd = m.viewport.Update(msg_)
			cmds = append(cmds, cmd)
		case StateSavingToFile:
			m.filepicker, cmd = m.filepicker.Update(msg_)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateKeyBindings() {
	mode_keymap.EnableMode(&m.keyMap, string(m.state))
}

func (m *model) scrollToSelected() {
	v, pos := m.conversation.ViewAndSelectedPosition()
	m.viewport.SetContent(v)
	midScreenOffset := m.viewport.YOffset + m.viewport.Height/2
	msgEndOffset := pos.Offset + pos.Height
	bottomOffset := m.viewport.YOffset + m.viewport.Height
	if pos.Offset > midScreenOffset && msgEndOffset > bottomOffset {
		m.viewport.SetYOffset(pos.Offset - max(m.viewport.Height-pos.Height-1, m.viewport.Height/2))
	} else if pos.Offset < m.viewport.YOffset {
		m.viewport.SetYOffset(pos.Offset)
	}
}

func (m *model) recomputeSize() {
	headerView := m.headerView()
	headerHeight := lipgloss.Height(headerView)
	if headerView == "" {
		headerHeight = 0
	}

	helpView := m.help.View(m.keyMap)
	helpViewHeight := lipgloss.Height(helpView)

	if m.state == StateSavingToFile {
		m.filepicker.Filepicker.Height = m.height - headerHeight - helpViewHeight
		return
	}

	textAreaView := m.textAreaView()
	textAreaHeight := lipgloss.Height(textAreaView)

	newHeight := m.height - textAreaHeight - headerHeight - helpViewHeight
	if newHeight < 0 {
		newHeight = 0
	}

	m.conversation.SetWidth(m.width)

	m.viewport.Width = m.width
	m.viewport.Height = newHeight
	m.viewport.YPosition = headerHeight + 1

	h, _ := m.style.SelectedMessage.GetFrameSize()

	m.textArea.SetWidth(m.width - h)

	v, _ := m.conversation.ViewAndSelectedPosition()
	m.viewport.SetContent(v)

	// TODO(manuel, 2023-09-21) Keep the current position by trying to match it to some message
	// This is probably going to be tricky
	m.viewport.GotoBottom()
}

func (m model) headerView() string {
	return m.title
}

func (m model) textAreaView() string {
	if m.err != nil {
		// TODO(manuel, 2023-09-21) Use a proper error style
		w, _ := m.style.SelectedMessage.GetFrameSize()
		v := wrapWords(m.err.Error(), m.width-w)
		return m.style.SelectedMessage.Render(v)
	}

	v := m.textArea.View()
	switch m.state {
	case StateUserInput:
		v = m.style.FocusedMessage.Render(v)
	case StateMovingAround, StateStreamCompletion:
		v = m.style.UnselectedMessage.Render(v)
	case StateError, StateSavingToFile:
	}

	return v
}

func (m model) View() string {
	headerView := m.headerView()

	view, _ := m.conversation.ViewAndSelectedPosition()
	m.viewport.SetContent(view)

	viewportView := m.viewport.View()
	textAreaView := m.textAreaView()
	helpView := m.help.View(m.keyMap)

	// debugging heights
	viewportHeight := lipgloss.Height(viewportView)
	_ = viewportHeight
	textAreaHeight := lipgloss.Height(textAreaView)
	_ = textAreaHeight
	headerHeight := lipgloss.Height(headerView)
	_ = headerHeight
	helpViewHeight := lipgloss.Height(helpView)
	_ = helpViewHeight

	ret := ""
	if headerView != "" {
		ret = headerView + "\n" + helpView
	}

	switch m.state {
	case StateUserInput, StateError, StateMovingAround, StateStreamCompletion:
		ret += viewportView + "\n" + textAreaView + "\n" + helpView

	case StateSavingToFile:
		ret += m.filepicker.View()
	}

	return ret
}

func (m *model) submit() tea.Cmd {
	if !m.backend.IsFinished() {
		return func() tea.Msg {
			return errMsg(errors.New("already streaming"))
		}
	}

	m.conversationManager.AppendMessages(
		conversation.NewChatMessage(conversation.RoleUser, m.textArea.Value()))
	m.textArea.SetValue("")

	m.state = StateStreamCompletion
	m.updateKeyBindings()

	m.textArea.SetValue("")
	m.viewport.GotoBottom()
	cmds := []tea.Cmd{
		func() tea.Msg {
			return refreshMessageMsg{
				GoToBottom: true,
			}
		},
	}
	ctx := context2.Background()
	cmd, err := m.backend.Start(ctx, m.conversationManager.GetConversation())
	if err != nil {
		cmds = append(cmds, func() tea.Msg {
			return errMsg(err)
		})
	} else {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

type refreshMessageMsg struct {
	GoToBottom bool
}

func (m *model) finishCompletion() tea.Cmd {
	refreshCommand := func() tea.Msg {
		return refreshMessageMsg{
			GoToBottom: true,
		}
	}

	if m.state == StateStreamCompletion {
		// WARN not sure if really necessary actually, this should only be called once at this point.
		m.backend.Kill()

		m.state = StateUserInput
		m.textArea.Focus()
		m.textArea.SetValue("")

		m.recomputeSize()
		m.updateKeyBindings()

		if m.quitReceived {
			return tea.Quit
		}
	}

	return refreshCommand
}
