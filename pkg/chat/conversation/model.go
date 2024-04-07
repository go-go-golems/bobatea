package conversation

import (
	"fmt"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/conversation"
	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"strings"
	"time"
)

type cacheEntry struct {
	msg         *conversation.Message
	rendered    string
	selected    bool
	lastUpdated time.Time
}

type Model struct {
	manager conversation.Manager
	style   *Style
	active  bool
	width   int

	cache map[conversation.NodeID]cacheEntry

	selectedIdx int
	selectedID  conversation.NodeID
}

func NewModel(manager conversation.Manager) Model {
	return Model{
		manager:    manager,
		style:      DefaultStyles(),
		cache:      make(map[conversation.NodeID]cacheEntry),
		selectedID: conversation.NullNode,
	}
}

func (m Model) Conversation() conversation.Conversation {
	return m.manager.GetConversation()
}

func (m Model) Init() tea.Cmd {
	c := m.manager.GetConversation()
	m.updateCache(c...)

	return nil
}

func (m Model) updateCache(c ...*conversation.Message) {
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

func (m *Model) SetWidth(width int) {
	m.width = width

	m.cache = make(map[conversation.NodeID]cacheEntry)
}

func (m *Model) SelectedIdx() int {
	return m.selectedIdx
}

func (m *Model) SetSelectedIdx(idx int) {
	m.selectedID = conversation.NullNode
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

func (m Model) renderMessage(selected bool, msg *conversation.Message) string {
	v := msg.Content.View()
	v = strings.TrimRight(v, "\n")

	style := m.style.UnselectedMessage
	if selected {
		style = m.style.SelectedMessage
	}
	w, _ := style.GetFrameSize()
	v_ := wrapWords(v, m.width-w-style.GetHorizontalPadding())

	v_ = style.
		Width(m.width - style.GetHorizontalPadding()).
		Render(v_)

	return v_

}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// handle chat streaming messages
	case StreamCompletionMsg:
		// update the respective message's text content with the new completion
		msg_, ok := m.manager.GetMessage(msg.ID)
		if !ok {
			// TODO not sure if we need to handle the error here, or do some warning?
			return m, nil
		}

		textMsg, ok := msg_.Content.(*conversation.ChatMessageContent)
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

		textMsg, ok := msg_.Content.(*conversation.ChatMessageContent)
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
		metadata := map[string]interface{}{
			"id":        uuid.UUID(msg.ID).String(),
			"parent_id": uuid.UUID(msg.ParentID).String(),
		}
		if msg.Step != nil {
			metadata["step"] = msg.Step.ToMap()
		}
		msg_ := conversation.NewChatMessage(
			conversation.RoleAssistant, "",
			conversation.WithID(msg.ID),
			conversation.WithParentID(msg.ParentID),
			conversation.WithMetadata(metadata))
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

// StepMetadata represents metadata about the step that issues the streaming messages.
// There is not a real definition of what a streaming message right now, this will need to be
// cleaned up as the agent framework is built out.
// NOTE(manuel, 2024-01-17) This is a copy of the StepMetadata in geppetto, and we might want to extract this out into a separate steps package.
type StepMetadata struct {
	StepID     uuid.UUID `json:"step_id"`
	Type       string    `json:"type"`
	InputType  string    `json:"input_type"`
	OutputType string    `json:"output_type"`

	Metadata map[string]interface{} `json:"meta"`
}

func (sm *StepMetadata) ToMap() map[string]interface{} {
	ret := map[string]interface{}{
		"step_id":     sm.StepID,
		"type":        sm.Type,
		"input_type":  sm.InputType,
		"output_type": sm.OutputType,
	}

	for k, v := range sm.Metadata {
		ret[k] = v
	}

	return ret
}

type StreamMetadata struct {
	ID       conversation.NodeID    `json:"id" yaml:"id"`
	ParentID conversation.NodeID    `json:"parent_id" yaml:"parent_id"`
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
	Step     *StepMetadata          `json:"step_metadata,omitempty"`
}

// StreamStartMsg is sent by the backend when a streaming operation begins.
// The UI uses this message to append a new message to the conversation,
// indicating that the assistant has started processing. The conversation
// manager is responsible for adding this message to the conversation tree.
type StreamStartMsg struct {
	StreamMetadata
}

// StreamStatusMsg is sent by the backend to provide status updates during
// a streaming operation. It includes the current text of the stream along
// with the stream metadata.
//
// The UI typically does not need to update the
// conversation view in response to this message, but it could be used to
// show a loading indicator or similar temporary status.
type StreamStatusMsg struct {
	StreamMetadata
	Text string
}

// StreamCompletionMsg is sent by the backend when new data, such as a message
// completion, is available.
//
// The UI uses this message to update the text content
// of the respective message in the conversation. The conversation manager
// updates the message content and the last update timestamp.
type StreamCompletionMsg struct {
	StreamMetadata
	Delta      string
	Completion string
}

// StreamDoneMsg is sent by the backend to signal the successful completion
// of the streaming operation.
//
// The UI uses this message to finalize the content of the
// message in the conversation. The conversation manager updates the message
// content and the last update timestamp to reflect the final text.
type StreamDoneMsg struct {
	StreamMetadata
	Completion string
}

// StreamCompletionError is sent by the backend when an error occurs during
// the streaming operation.
//
// The UI uses this message to display an error state or
// message to the user.
type StreamCompletionError struct {
	StreamMetadata
	Err error
}
