package conversation

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/events"

	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	conversation2 "github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/geppetto/pkg/steps"
	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

type cacheEntry struct {
	msg         *conversation2.Message
	rendered    string
	selected    bool
	lastUpdated time.Time
}

type Model struct {
	manager         conversation2.Manager
	style           *Style
	active          bool
	width           int
	renderer        *glamour.TermRenderer
	determinedStyle string

	cache map[conversation2.NodeID]cacheEntry

	selectedIdx int
	selectedID  conversation2.NodeID
}

func NewModel(manager conversation2.Manager) Model {
	log.Debug().Msg("Creating initial glamour renderer in NewModel...")
	start := time.Now()

	// Determine the style once
	var determinedStyle string
	var initialStyleOption glamour.TermRendererOption

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		determinedStyle = "notty"
		initialStyleOption = glamour.WithStandardStyle(determinedStyle)
	} else if termenv.HasDarkBackground() {
		determinedStyle = "dark"
		initialStyleOption = glamour.WithStandardStyle(determinedStyle)
	} else {
		determinedStyle = "light"
		initialStyleOption = glamour.WithStandardStyle(determinedStyle)
	}
	log.Debug().Str("determinedStyle", determinedStyle).Msg("Determined initial style")

	// Create renderer with the determined style
	renderer, err := glamour.NewTermRenderer(
		initialStyleOption, // Use the determined style
	)
	duration := time.Since(start)
	log.Debug().Dur("duration", duration).Msg("Initial glamour renderer creation complete")

	if err != nil {
		log.Error().Err(err).Msg("Failed to create initial markdown renderer")
	}

	return Model{
		manager:         manager,
		style:           DefaultStyles(),
		cache:           make(map[conversation2.NodeID]cacheEntry),
		selectedID:      conversation2.NullNode,
		renderer:        renderer,
		determinedStyle: determinedStyle,
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
	log.Debug().Int("newWidth", width).Int("currentWidth", m.width).Msg("SetWidth called")
	startSetWidth := time.Now()

	if m.width == width {
		log.Debug().Int("width", width).Msg("SetWidth: Width unchanged, returning early")
		return // No change
	}
	log.Debug().Int("newWidth", width).Int("oldWidth", m.width).Msg("SetWidth: Width changed, proceeding")
	m.width = width

	// Recreate renderer with the new width and the *pre-determined* style
	log.Debug().Int("width", width).Str("style", m.determinedStyle).Msg("SetWidth: Preparing to recreate glamour renderer with determined style...")
	startRenderer := time.Now()
	// XXX: Restore WithWordWrap - NO, keep it here
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(m.determinedStyle),      // Use determined style
		glamour.WithWordWrap(m.getRendererContentWidth()), // Use calculated width
	)
	durationRenderer := time.Since(startRenderer)
	log.Debug().Dur("duration", durationRenderer).Int("width", width).Msg("SetWidth: glamour.NewTermRenderer call finished") // Restored log message

	if err != nil {
		log.Error().Err(err).Int("width", m.width).Msg("SetWidth: Failed to recreate markdown renderer")
		m.renderer = nil // Set to nil on error
	} else {
		log.Debug().Int("width", m.width).Msg("SetWidth: Successfully recreated renderer")
		m.renderer = renderer
	}

	// Invalidate cache as rendering depends on width
	log.Debug().Int("width", m.width).Msg("SetWidth: Invalidating cache...")
	startCache := time.Now()
	m.cache = make(map[conversation2.NodeID]cacheEntry)
	durationCache := time.Since(startCache)
	log.Debug().Dur("duration", durationCache).Int("width", m.width).Msg("SetWidth: Cache invalidated")

	durationSetWidth := time.Since(startSetWidth)
	log.Debug().Dur("totalDuration", durationSetWidth).Int("width", m.width).Msg("SetWidth finished")
}

// Helper to calculate the actual content width for the renderer
func (m Model) getRendererContentWidth() int {
	log.Debug().Msg("getRendererContentWidth called")
	start := time.Now()
	// Use UnselectedMessage style for width calculation consistency
	style := m.style.UnselectedMessage

	w, _ := style.GetFrameSize()
	padding := style.GetHorizontalPadding()
	contentWidth := m.width - w - padding
	duration := time.Since(start)
	log.Debug().Dur("duration", duration).Int("frameWidth", w).Int("padding", padding).Int("resultWidth", contentWidth).Msg("getRendererContentWidth finished")
	return contentWidth
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
		rendered, err := m.renderer.Render(v + "\n")
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
		msg_, ok := m.manager.GetMessage(msg.ID)
		if !ok {
			return m, nil
		}

		textMsg, ok := msg_.Content.(*conversation2.ChatMessageContent)
		if !ok {
			return m, nil
		}

		textMsg.Text = "**Error**\n\n" + msg.Err.Error()
		msg_.LastUpdate = time.Now()
		msg_.Metadata["step_metadata"] = msg.StepMetadata
		msg_.Metadata["event_metadata"] = msg.EventMetadata
		if msg.EventMetadata != nil {
			msg_.LLMMessageMetadata = &msg.EventMetadata.LLMMessageMetadata
		}

		m.updateCache(msg_)

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
