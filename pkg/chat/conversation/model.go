package conversation

import (
	"fmt"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/geppetto/pkg/steps"
	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"strings"
	"time"
)

type cacheEntry struct {
	msg         *Message
	rendered    string
	selected    bool
	lastUpdated time.Time
}

type Model struct {
	manager Manager
	style   *Style
	active  bool
	Width   int

	cache map[NodeID]cacheEntry

	selectedIdx int
	selectedID  NodeID
}

func NewModel(manager Manager) Model {
	return Model{
		manager:    manager,
		style:      DefaultStyles(),
		cache:      make(map[NodeID]cacheEntry),
		selectedID: NullNode,
	}
}

func (m Model) Conversation() Conversation {
	return m.manager.GetConversation()
}

func (m Model) Init() tea.Cmd {
	c := m.manager.GetConversation()
	m.updateCache(c...)

	return nil
}

func (m Model) updateCache(c ...*Message) {
	for idx, msg := range c {
		c_, ok := m.cache[msg.ID]
		selected := idx == m.selectedIdx && m.active

		if ok {
			if c_.lastUpdated.After(msg.LastUpdate) && c_.selected == selected {
				continue
			}
		}

		v_ := m.renderMessage(selected, msg)
		c_ = cacheEntry{
			msg:         msg,
			rendered:    v_,
			lastUpdated: time.Now(),
			selected:    selected,
		}

		m.cache[msg.ID] = c_
	}
}

func (m *Model) SetActive(active bool) {
	m.active = active
}

func (m *Model) SelectedIdx() int {
	return m.selectedIdx
}

func (m *Model) SetSelectedIdx(idx int) {
	m.selectedID = NullNode
	m.selectedIdx = idx
	conversation := m.manager.GetConversation()
	if idx > len(conversation) {
		m.selectedIdx = len(conversation) - 1
	}

	if m.selectedIdx < 0 {
		return
	}

	m.selectedID = conversation[m.selectedIdx].ID
}

func (m Model) renderMessage(selected bool, msg *Message) string {
	v := msg.Content.View()
	v = strings.TrimRight(v, "\n")

	style := m.style.UnselectedMessage
	if selected {
		style = m.style.SelectedMessage
	}
	w, _ := style.GetFrameSize()
	v_ := wrapWords(v, m.Width-w-style.GetHorizontalPadding())

	v_ = style.
		Width(m.Width - style.GetHorizontalPadding()).
		Render(v_)

	return v_

}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// handle chat streaming messages
	case StreamCompletionMsg:
		msg_, ok := m.manager.GetMessage(msg.ID)
		if !ok {
			// TODO not sure if we need to handle the error here, or do some warning?
			return m, nil
		}

		textMsg, ok := msg_.Content.(*ChatMessageContent)
		if !ok {
			return m, nil
		}

		textMsg.Text = msg.Completion
		msg_.LastUpdate = time.Now()

	case StreamDoneMsg:
		// I don't think there is anything to do here, for now at least
		msg_, ok := m.manager.GetMessage(msg.ID)
		if !ok {
			return m, nil
		}

		textMsg, ok := msg_.Content.(*ChatMessageContent)
		if !ok {
			return m, nil
		}

		// update with the final text
		textMsg.Text = msg.Completion
		msg_.LastUpdate = time.Now()

	case StreamCompletionError:
		// TODO(manuel, 2024-01-15) Update error view...
		//cmd = m.setError(msg.Err)

	case StreamStartMsg:
		msg_ := NewChatMessage(
			RoleAssistant, "",
			WithID(msg.ID),
			WithParentID(msg.ParentID),
			WithMetadata(map[string]interface{}{
				"id":        uuid.UUID(msg.ID).String(),
				"parent_id": uuid.UUID(msg.ParentID).String(),
				"step":      msg.Step.ToMap(),
			}))
		m.manager.AppendMessages(msg_)

		m.updateCache(msg_)

	case StreamStatusMsg:
	// TODO(manuel, 2024-01-15) Implement message status view

	//TODO(manuel, 2024-01-15) implement keyboard navigation and copy paste and all that

	default:
	}

	return m, cmd
}

func wrapWords(text string, w int) string {
	w_ := wordwrap.NewWriter(w)
	_, err := fmt.Fprint(w_, text)
	if err != nil {
		panic(err)
	}
	_ = w_.Close()
	v := w_.String()
	return v
}

func (m Model) View() string {
	v, _ := m.ViewAndSelectedPosition()
	return v
}

type MessagePosition struct {
	Offset int
	Height int
}

func (m Model) ViewAndSelectedPosition() (string, MessagePosition) {
	ret := ""
	height := 0
	selectedOffset := 0
	selectedHeight := 0

	msgs_ := m.manager.GetConversation()

	m.updateCache(msgs_...)
	for _, msg := range msgs_ {
		c_, ok := m.cache[msg.ID]
		if !ok {
			continue
		}
		h := lipgloss.Height(c_.rendered)
		if m.selectedID == msg.ID {
			selectedOffset = height
			selectedHeight = h
		}
		ret += c_.rendered
		ret += "\n"
		height += h
	}

	return ret, MessagePosition{
		Offset: selectedOffset,
		Height: selectedHeight,
	}
}

type StreamMetadata struct {
	ID       NodeID                 `json:"id" yaml:"id"`
	ParentID NodeID                 `json:"parent_id" yaml:"parent_id"`
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
	Step     *steps.StepMetadata    `json:"step_metadata,omitempty"`
}

type StreamStartMsg struct {
	StreamMetadata
}

type StreamStatusMsg struct {
	StreamMetadata
	Text string
}

type StreamDoneMsg struct {
	StreamMetadata
	Completion string
}

type StreamCompletionMsg struct {
	StreamMetadata
	Delta      string
	Completion string
}

// StreamCompletionError does not imply that the stream finished
type StreamCompletionError struct {
	StreamMetadata
	Err error
}
