package conversation

import (
	"fmt"
	"strings"
	"time"

	conversation2 "github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/geppetto/pkg/steps"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/chat"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"github.com/rs/zerolog/log"
)

type cacheEntry struct {
	msg         *conversation2.Message
	rendered    string
	selected    bool
	lastUpdated time.Time
}

type Model struct {
	manager conversation2.Manager
	style   *Style
	active  bool
	width   int

	cache map[conversation2.NodeID]cacheEntry

	selectedIdx int
	selectedID  conversation2.NodeID
}

func NewModel(manager conversation2.Manager) Model {
	return Model{
		manager:    manager,
		style:      DefaultStyles(),
		cache:      make(map[conversation2.NodeID]cacheEntry),
		selectedID: conversation2.NullNode,
	}
}

func (m Model) Conversation() conversation2.Conversation {
	return m.manager.GetConversation()
}

func (m Model) Init() tea.Cmd {
	c := m.manager.GetConversation()
	m.updateCache(c...)

	return nil
}

func (m Model) updateCache(c ...*conversation2.Message) {
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

	m.cache = make(map[conversation2.NodeID]cacheEntry)
}

func (m *Model) SelectedIdx() int {
	return m.selectedIdx
}

func (m *Model) SetSelectedIdx(idx int) {
	m.selectedID = conversation2.NullNode
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

func (m Model) renderMessage(selected bool, msg *conversation2.Message) string {
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

		textMsg, ok := msg_.Content.(*conversation2.ChatMessageContent)
		if !ok {
			return m, nil
		}

		textMsg.Text = msg.Completion
		msg_.LastUpdate = time.Now()
		msg_.Metadata["step_metadata"] = msg.StepMetadata
		msg_.Metadata["event_metadata"] = msg.EventMetadata
		m.updateCache(msg_)

	case StreamDoneMsg:
		// I don't think there is anything to do here, for now at least
		msg_, ok := m.manager.GetMessage(msg.ID)
		if !ok {
			return m, nil
		}

		textMsg, ok := msg_.Content.(*conversation2.ChatMessageContent)
		if !ok {
			return m, nil
		}

		// update with the final text
		textMsg.Text = msg.Completion
		msg_.LastUpdate = time.Now()
		msg_.Metadata["step_metadata"] = msg.StepMetadata
		msg_.Metadata["event_metadata"] = msg.EventMetadata
		m.updateCache(msg_)

	case StreamCompletionError:
		// TODO(manuel, 2024-01-15) Update error view...
		//cmd = m.setError(msg.Err)
		log.Error().Err(msg.Err).Msg("StreamCompletionError")

	case StreamStartMsg:
		metadata := map[string]interface{}{
			"id":        uuid.UUID(msg.ID).String(),
			"parent_id": uuid.UUID(msg.ParentID).String(),
		}
		if msg.StepMetadata != nil {
			metadata["step_metadata"] = msg.StepMetadata
		}
		if msg.EventMetadata != nil {
			metadata["event_metadata"] = msg.EventMetadata
		}

		msg_ := conversation2.NewChatMessage(
			conversation2.RoleAssistant, "",
			conversation2.WithID(msg.ID),
			conversation2.WithParentID(msg.ParentID),
			conversation2.WithMetadata(metadata))
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
	ID            conversation2.NodeID `json:"id" yaml:"id"`
	ParentID      conversation2.NodeID `json:"parent_id" yaml:"parent_id"`
	EventMetadata *chat.EventMetadata  `json:"metadata" yaml:"metadata"`
	StepMetadata  *steps.StepMetadata  `json:"step_metadata,omitempty"`
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
	// Delta is the delta that was added to the message
	Delta string
	// Completion is the full completion text
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
	// Completion is the full completion text
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
