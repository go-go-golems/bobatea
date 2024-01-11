package chat

import (
	context2 "context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/chat/conversation"
	mode_keymap "github.com/go-go-golems/bobatea/pkg/mode-keymap"
	"github.com/go-go-golems/bobatea/pkg/textarea"
	"github.com/go-go-golems/glazed/pkg/helpers/markdown"
	"github.com/pkg/errors"
	"golang.design/x/clipboard"
	"strings"
	"time"
)

type errMsg error

type State string

const (
	StateUserInput        State = "user-input"
	StateMovingAround     State = "moving-around"
	StateStreamCompletion State = "stream-completion"
	StateError            State = "error"
)

type model struct {
	contextManager conversation.ConversationManager

	viewport viewport.Model

	// not really what we want, but use this for now, we'll have to either find a normal text box,
	// or implement wrapping ourselves.
	textArea textarea.Model

	help help.Model

	// currently selected message, always valid
	selectedIdx int
	err         error
	keyMap      KeyMap

	style  *Style
	width  int
	height int

	backend Backend
	//step chat.Step
	//// if not nil, streaming is going on
	//stepResult steps.StepResult[string]

	currentResponse        string
	previousResponseHeight int

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

func InitialModel(manager conversation.ConversationManager, backend Backend, options ...ModelOption) model {
	ret := model{
		contextManager: manager,
		style:          DefaultStyles(),
		keyMap:         DefaultKeyMap,
		backend:        backend,
		viewport:       viewport.New(0, 0),
		help:           help.New(),
	}

	for _, option := range options {
		option(&ret)
	}

	ret.textArea = textarea.New()
	ret.textArea.Placeholder = "Dear AI, answer my plight..."
	ret.textArea.Focus()
	ret.state = StateUserInput

	ret.selectedIdx = len(ret.contextManager.GetMessages()) - 1

	messages, _ := ret.messageView()
	ret.viewport.SetContent(messages)
	ret.viewport.YPosition = 0
	ret.viewport.GotoBottom()

	ret.updateKeyBindings()

	return ret
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		textarea.Blink,
	}
	err := clipboard.Init()
	if err != nil {
		cmds = append(cmds, func() tea.Msg {
			return errMsg(err)
		})
	}

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
			m.selectedIdx = len(m.contextManager.GetMessages()) - 1
			if m.selectedIdx < 0 {
				m.selectedIdx = 0
			}
			v, height := m.messageView()
			m.viewport.SetContent(v)
			m.viewport.SetYOffset(height)
			m.updateKeyBindings()
		}

	case key.Matches(msg, m.keyMap.Quit):
		if !m.quitReceived {
			m.quitReceived = true
			// on first quit, try to cancel completion if running
			m.backend.Interrupt()
		}

		// force save completion before quitting
		m.finishCompletion()

		cmd = tea.Quit

	case key.Matches(msg, m.keyMap.FocusMessage):
		// TODO(manuel, 2024-01-06) This could potentially focus on a previous message
		// and allow us to regenerate.
		cmd = m.textArea.Focus()

		m.state = StateUserInput
		m.updateKeyBindings()
		v, _ := m.messageView()
		m.viewport.SetContent(v)

	case key.Matches(msg, m.keyMap.SelectNextMessage):
		messages := m.contextManager.GetMessages()
		if m.selectedIdx < len(messages)-1 {
			m.selectedIdx++
			v, height := m.messageView()
			m.viewport.SetYOffset(height)
			m.viewport.SetContent(v)
		} else {
			m.viewport, cmd = m.viewport.Update(msg)
		}

	case key.Matches(msg, m.keyMap.SelectPrevMessage):
		if m.selectedIdx > 0 {
			m.selectedIdx--
			v, height := m.messageView()
			m.viewport.SetYOffset(height)
			m.viewport.SetContent(v)
		}

	case key.Matches(msg, m.keyMap.SubmitMessage):
		cmd = m.submit()

	case key.Matches(msg, m.keyMap.CopyToClipboard):
		msgs := m.contextManager.GetMessages()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				if m.selectedIdx < len(msgs) && m.selectedIdx >= 0 {
					msg_ := msgs[m.selectedIdx]
					clipboard.Write(clipboard.FmtText, []byte(msg_.Text))
				}
			} else {
				text := ""
				for _, m := range msgs {
					if m.Role == conversation.RoleAssistant {
						text += m.Text + "\n"
					}
				}
				clipboard.Write(clipboard.FmtText, []byte(text))
			}
		}

	case key.Matches(msg, m.keyMap.CopyLastResponseToClipboard):
		msgs := m.contextManager.GetMessages()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				if m.selectedIdx < len(msgs) && m.selectedIdx >= 0 {
					msg_ := msgs[m.selectedIdx]
					clipboard.Write(clipboard.FmtText, []byte(msg_.Text))
				}
			} else {
				if m.state == StateUserInput {
					lastMsg := msgs[len(msgs)-1]
					clipboard.Write(clipboard.FmtText, []byte(lastMsg.Text))
				}
			}
		}

	case key.Matches(msg, m.keyMap.CopyLastSourceBlocksToClipboard):
		msgs := m.contextManager.GetMessages()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				if m.selectedIdx < len(msgs) && m.selectedIdx >= 0 {
					msg_ := msgs[m.selectedIdx]
					code := markdown.ExtractQuotedBlocks(msg_.Text, false)
					clipboard.Write(clipboard.FmtText, []byte(strings.Join(code, "\n")))
				}
			} else {
				if m.state == StateUserInput {
					text := ""
					for _, m := range msgs {
						if m.Role == conversation.RoleAssistant {
							text += m.Text + "\n"
						}
					}
					code := markdown.ExtractQuotedBlocks(text, false)
					clipboard.Write(clipboard.FmtText, []byte(strings.Join(code, "\n")))
				}
			}
		}

	case key.Matches(msg, m.keyMap.CopySourceBlocksToClipboard):
		msgs := m.contextManager.GetMessages()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				if m.selectedIdx < len(msgs) && m.selectedIdx >= 0 {
					msg_ := msgs[m.selectedIdx]
					code := markdown.ExtractQuotedBlocks(msg_.Text, false)
					clipboard.Write(clipboard.FmtText, []byte(strings.Join(code, "\n")))
				}
			} else {
				text := ""
				for _, m := range msgs {
					if m.Role == conversation.RoleAssistant {
						text += m.Text + "\n"
					}
				}
				code := markdown.ExtractQuotedBlocks(text, false)
				clipboard.Write(clipboard.FmtText, []byte(strings.Join(code, "\n")))
			}
		}

	case key.Matches(msg, m.keyMap.SaveToFile):
		// TODO(manuel, 2023-11-14) Implement file chosing dialog
		err := m.contextManager.SaveToFile("/tmp/output.json")
		if err != nil {
			cmd = func() tea.Msg {
				return errMsg(err)
			}
		}

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
		case StateMovingAround, StateStreamCompletion, StateError:
			m.viewport, cmd = m.viewport.Update(msg)
		}
	}

	return m, cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.recomputeSize()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil

	case StreamCompletionMsg,
		StreamStartMsg,
		StreamStatusMsg,
		StreamDoneMsg,
		StreamCompletionError:
		return m.handleStreamMessage(msg)

	case refreshMessageMsg:
		v, _ := m.messageView()
		m.viewport.SetContent(v)
		m.recomputeSize()
		if msg.GoToBottom {
			m.viewport.GotoBottom()
		}

	default:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateKeyBindings() {
	mode_keymap.EnableMode(&m.keyMap, string(m.state))
}

func (m *model) recomputeSize() {
	headerView := m.headerView()
	headerHeight := lipgloss.Height(headerView)
	if headerView == "" {
		headerHeight = 0
	}
	textAreaView := m.textAreaView()
	textAreaHeight := lipgloss.Height(textAreaView)

	helpView := m.help.View(m.keyMap)
	helpViewHeight := lipgloss.Height(helpView)

	m.previousResponseHeight = textAreaHeight
	newHeight := m.height - textAreaHeight - headerHeight - helpViewHeight
	if newHeight < 0 {
		newHeight = 0
	}
	m.viewport.Width = m.width
	m.viewport.Height = newHeight
	m.viewport.YPosition = headerHeight + 1

	h, _ := m.style.SelectedMessage.GetFrameSize()

	m.textArea.SetWidth(m.width - h)

	v, _ := m.messageView()
	m.viewport.SetContent(v)

	// TODO(manuel, 2023-09-21) Keep the current position by trying to match it to some message
	// This is probably going to be tricky
	m.viewport.GotoBottom()
}

func (m model) headerView() string {
	return m.title
}

// messageView computes the content of the message view and returns the y offset of the selected message or 0
func (m model) messageView() (string, int) {
	ret := ""
	height := 0
	selectedHeight := 0

	for idx := range m.contextManager.GetMessages() {
		message := m.contextManager.GetMessages()[idx]
		v := fmt.Sprintf("[%s]: %s", message.Role, message.Text)

		style := m.style.UnselectedMessage
		if idx == m.selectedIdx && m.state == StateMovingAround {
			style = m.style.SelectedMessage
			selectedHeight = height
		}
		w, _ := style.GetFrameSize()

		v_ := wrapWords(v, m.width-w-style.GetHorizontalPadding())
		v_ = style.
			Width(m.width - style.GetHorizontalPadding()).
			Render(v_)
		ret += v_
		ret += "\n"
		height += lipgloss.Height(v_)
	}

	return ret, selectedHeight
}

func (m model) textAreaView() string {
	if m.err != nil {
		// TODO(manuel, 2023-09-21) Use a proper error style
		w, _ := m.style.SelectedMessage.GetFrameSize()
		v := wrapWords(m.err.Error(), m.width-w)
		return m.style.SelectedMessage.Render(v)
	}

	// we are currently streaming
	if !m.backend.IsFinished() {
		w, _ := m.style.SelectedMessage.GetFrameSize()
		v := wrapWords(m.currentResponse, m.width-w-m.style.SelectedMessage.GetHorizontalPadding())
		// TODO(manuel, 2023-09-21) this is where we'd add the spinner
		return m.style.SelectedMessage.Width(m.width - m.style.SelectedMessage.GetHorizontalPadding()).Render(v)
	}

	v := m.textArea.View()
	switch m.state {
	case StateUserInput:
		v = m.style.FocusedMessage.Render(v)
	case StateMovingAround, StateStreamCompletion:
		v = m.style.UnselectedMessage.Render(v)
	case StateError:
	}

	return v
}

func (m model) View() string {
	headerView := m.headerView()
	viewportView := m.viewport.View()
	textAreaView := m.textAreaView()
	helpView := m.help.View(m.keyMap)

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
		ret = headerView + "\n"
	}
	ret += viewportView + "\n" + textAreaView + "\n" + helpView

	return ret
}

// Chat completion messages
func (m *model) submit() tea.Cmd {
	if !m.backend.IsFinished() {
		return func() tea.Msg {
			return errMsg(errors.New("already streaming"))
		}
	}

	m.contextManager.AddMessages(&conversation.Message{
		Role: conversation.RoleUser,
		Text: m.textArea.Value(),
		Time: time.Now(),
	})

	ctx := context2.Background()
	err := m.backend.Start(ctx, m.contextManager.GetMessages())
	if err != nil {
		return func() tea.Msg {
			return errMsg(err)
		}
	}

	return m.startCompletion()
}

func (m *model) startCompletion() tea.Cmd {
	m.state = StateStreamCompletion
	m.updateKeyBindings()
	m.currentResponse = ""
	m.previousResponseHeight = 0

	m.viewport.GotoBottom()

	return tea.Batch(func() tea.Msg {
		return refreshMessageMsg{
			GoToBottom: true,
		}
	},
		m.getNextCompletion(),
	)

}

func (m model) getNextCompletion() tea.Cmd {
	return m.backend.GetNextCompletion()
}

type refreshMessageMsg struct {
	GoToBottom bool
}

func (m *model) finishCompletion() tea.Cmd {
	// completion already finished, happens when error and completion finish or cancellation happen
	if m.backend.IsFinished() {
		return nil
	}

	m.contextManager.AddMessages(&conversation.Message{
		Role: conversation.RoleAssistant,
		Text: m.currentResponse,
		Time: time.Now(),
	})
	m.currentResponse = ""
	m.previousResponseHeight = 0
	m.backend.Kill()

	m.state = StateUserInput
	m.textArea.Focus()
	m.textArea.SetValue("")

	m.recomputeSize()
	m.updateKeyBindings()

	if m.quitReceived {
		return tea.Quit
	}

	return func() tea.Msg {
		return refreshMessageMsg{
			GoToBottom: true,
		}
	}
}

func (m *model) setError(err error) tea.Cmd {
	cmd := m.finishCompletion()
	m.err = err
	m.state = StateError
	m.updateKeyBindings()
	return cmd
}

func (m model) handleStreamMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// handle chat streaming messages
	case StreamCompletionMsg:
		m.currentResponse += msg.Completion
		newTextAreaView := m.textAreaView()
		newHeight := lipgloss.Height(newTextAreaView)
		if newHeight != m.previousResponseHeight {
			m.recomputeSize()
			m.previousResponseHeight = newHeight
		}
		//cmds = append(cmds, func() tea.Msg {
		//	return refreshMessageMsg{}
		//})
		// TODO(manuel, 2024-01-11) I don't think we need this anymore now that we are not streaming partial events over the completion channel
		cmd = m.getNextCompletion()

	case StreamDoneMsg:
		cmd = m.finishCompletion()

	case StreamCompletionError:
		// does this imply the completion is finished, or do we get a separate StreamDoneMsg
		cmd = m.setError(msg.Err)

	case StreamStartMsg:
	case StreamStatusMsg:
	}

	return m, cmd
}
