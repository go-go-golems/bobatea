package conversation

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/events"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	conversation2 "github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/geppetto/pkg/steps"
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
	manager  conversation2.Manager
	style    *Style
	active   bool
	width    int
	renderer *glamour.TermRenderer

	cache map[conversation2.NodeID]cacheEntry

	selectedIdx int
	selectedID  conversation2.NodeID
}

func NewModel(manager conversation2.Manager) Model {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create initial markdown renderer")
	}

	return Model{
		manager:    manager,
		style:      DefaultStyles(),
		cache:      make(map[conversation2.NodeID]cacheEntry),
		selectedID: conversation2.NullNode,
		renderer:   renderer,
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
	if m.width == width {
		return
	}
	m.width = width

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.getRendererContentWidth()),
	)
	if err != nil {
		log.Error().Err(err).Int("width", m.width).Msg("Failed to recreate markdown renderer on SetWidth")
		m.renderer = nil
	} else {
		m.renderer = renderer
	}

	m.cache = make(map[conversation2.NodeID]cacheEntry)
}

func (m Model) getRendererContentWidth() int {
	style := m.style.UnselectedMessage
	w, _ := style.GetFrameSize()
	return m.width - w - style.GetHorizontalPadding()
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
	contentWidth := m.width - w - style.GetHorizontalPadding()

	var v_ string
	if m.renderer == nil {
		log.Warn().Msg("Markdown renderer is nil, falling back to basic wrapping")
		v_ = wrapWords(v, contentWidth)
	} else {
		rendered, err := m.renderer.Render(v)
		if err != nil {
			log.Error().Err(err).Msg("Failed to render markdown")
			v_ = wrapWords(v, contentWidth)
		} else {
			v_ = strings.TrimSpace(rendered)
		}
	}

	if msg.LLMMessageMetadata != nil {
		metadataStr := ""
		if msg.LLMMessageMetadata.Engine != "" {
			metadataStr += msg.LLMMessageMetadata.Engine
		}
		if msg.LLMMessageMetadata.Temperature != nil {
			if metadataStr != "" {
				metadataStr += " "
			}
			metadataStr += fmt.Sprintf("t: %.2f", *msg.LLMMessageMetadata.Temperature)
		}
		if msg.LLMMessageMetadata.Usage != nil {
			if metadataStr != "" {
				metadataStr += " "
			}
			metadataStr += fmt.Sprintf("in: %d out: %d", msg.LLMMessageMetadata.Usage.InputTokens, msg.LLMMessageMetadata.Usage.OutputTokens)
		}
		if metadataStr != "" {
			v_ += "\n\n" + m.style.MetadataStyle.Width(contentWidth).Render(metadataStr)
		}
	}

	v_ = style.
		Width(m.width - style.GetHorizontalPadding()).
		Render(v_)

	return v_
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case StreamCompletionMsg:
		msg_, ok := m.manager.GetMessage(msg.ID)
		if !ok {
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
		if msg.EventMetadata != nil {
			msg_.LLMMessageMetadata = &msg.EventMetadata.LLMMessageMetadata
		}
		m.updateCache(msg_)

	case StreamDoneMsg:
		msg_, ok := m.manager.GetMessage(msg.ID)
		if !ok {
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
		if msg.EventMetadata != nil {
			msg_.LLMMessageMetadata = &msg.EventMetadata.LLMMessageMetadata
		}
		m.updateCache(msg_)

	case StreamCompletionError:
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
	ID            conversation2.NodeID  `json:"id" yaml:"id"`
	ParentID      conversation2.NodeID  `json:"parent_id" yaml:"parent_id"`
	EventMetadata *events.EventMetadata `json:"metadata" yaml:"metadata"`
	StepMetadata  *steps.StepMetadata   `json:"step_metadata,omitempty"`
}

type StreamStartMsg struct {
	StreamMetadata
}

type StreamStatusMsg struct {
	StreamMetadata
	Text string
}

type StreamCompletionMsg struct {
	StreamMetadata
	Delta      string
	Completion string
}

type StreamDoneMsg struct {
	StreamMetadata
	Completion string
}

type StreamCompletionError struct {
	StreamMetadata
	Err error
}
