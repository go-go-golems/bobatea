package conversation

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	conversation2 "github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/geppetto/pkg/events"
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

// logMemoryUsage logs current memory statistics
func logMemoryUsage(operation string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Trace().
		Str("operation", operation).
		Uint64("alloc_mb", m.Alloc/1024/1024).
		Uint64("total_alloc_mb", m.TotalAlloc/1024/1024).
		Uint64("sys_mb", m.Sys/1024/1024).
		Uint32("num_gc", m.NumGC).
		Msg("Memory usage")
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
	cacheStart := time.Now()
	logMemoryUsage("cache_update_start")
	totalCacheHits := 0
	totalCacheMisses := 0
	totalRedraws := 0

	log.Trace().
		Str("operation", "cache_update_start").
		Int("message_count", len(c)).
		Int("current_cache_size", len(m.cache)).
		Msg("Starting cache update")

	for idx, msg := range c {
		msgStart := time.Now()
		c_, ok := m.cache[msg.ID]
		selected := idx == m.selectedIdx && m.active

		cacheHit := false
		if ok {
			totalCacheHits++
			cacheHit = true
			if c_.lastUpdated.After(msg.LastUpdate) && c_.selected == selected {
				log.Trace().
					Str("operation", "cache_skip").
					Str("messageID", msg.ID.String()).
					Time("cached_time", c_.lastUpdated).
					Time("msg_time", msg.LastUpdate).
					Bool("selected_match", c_.selected == selected).
					Msg("Skipping cache update - cached version is newer")
				continue
			}
		} else {
			totalCacheMisses++
		}

		// Log cache entry overwrite
		if ok {
			log.Trace().
				Str("operation", "cache_overwrite").
				Str("messageID", msg.ID.String()).
				Time("old_cached_time", c_.lastUpdated).
				Time("msg_time", msg.LastUpdate).
				Bool("selected_changed", c_.selected != selected).
				Msg("Overwriting existing cache entry")
		}

		// Expensive rendering operation
		renderStart := time.Now()
		v_ := m.renderMessage(selected, msg)
		renderDuration := time.Since(renderStart)
		totalRedraws++

		log.Trace().
			Str("operation", "message_render").
			Str("messageID", msg.ID.String()).
			Dur("render_duration", renderDuration).
			Int("content_length", len(msg.Content.String())).
			Int("rendered_length", len(v_)).
			Bool("was_cached", cacheHit).
			Bool("selected", selected).
			Msg("Message rendered")

		c_ = cacheEntry{
			msg:         msg,
			rendered:    v_,
			lastUpdated: time.Now(),
			selected:    selected,
		}

		m.cache[msg.ID] = c_

		msgDuration := time.Since(msgStart)
		log.Trace().
			Str("operation", "cache_entry_update").
			Str("messageID", msg.ID.String()).
			Dur("total_msg_duration", msgDuration).
			Msg("Cache entry updated")
	}

	totalDuration := time.Since(cacheStart)
	logMemoryUsage("cache_update_complete")

	log.Trace().
		Str("operation", "cache_update_complete").
		Dur("total_duration", totalDuration).
		Int("cache_hits", totalCacheHits).
		Int("cache_misses", totalCacheMisses).
		Int("redraws_performed", totalRedraws).
		Int("final_cache_size", len(m.cache)).
		Msg("Cache update completed")
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

	var style lipgloss.Style

	// Check if this is an error message and apply special styling
	if msg.Content.ContentType() == conversation2.ContentTypeError {
		if selected {
			style = m.style.ErrorSelected
		} else {
			style = m.style.ErrorMessage
		}
		// Add error icon/prefix
		v = "⚠️  " + v
	} else {
		style = m.style.UnselectedMessage
		if selected {
			style = m.style.SelectedMessage
		}
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
		startTime := time.Now()
		logMemoryUsage("stream_start_begin")

		// Check if this is a duplicate message
		existingMsg, isDuplicate := m.manager.GetMessage(msg.ID)

		log.Debug().
			Str("operation", "stream_start_processing").
			Str("messageID", msg.ID.String()).
			Str("parentID", msg.ParentID.String()).
			Bool("is_duplicate", isDuplicate).
			Int("conversation_size", len(m.manager.GetConversation())).
			Time("timestamp", startTime).
			Msg("Processing StreamStartMsg in conversation model")

		if isDuplicate {
			log.Warn().
				Str("operation", "duplicate_message_detection").
				Str("messageID", msg.ID.String()).
				Time("existing_last_update", existingMsg.LastUpdate).
				Str("existing_content", existingMsg.Content.String()).
				Msg("Duplicate StreamStartMsg detected - same ID already exists")

			// Skip duplicate processing to prevent tree corruption
			log.Debug().
				Str("messageID", msg.ID.String()).
				Msg("Skipping duplicate StreamStartMsg to prevent tree corruption")
			// return m, nil
		}

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

		// Create new message (even if duplicate exists)
		msgCreateStart := time.Now()
		msg_ := conversation2.NewChatMessage(
			conversation2.RoleAssistant, "",
			conversation2.WithID(msg.ID),
			conversation2.WithParentID(msg.ParentID),
			conversation2.WithMetadata(metadata))
		msgCreateDuration := time.Since(msgCreateStart)

		log.Debug().
			Str("operation", "message_creation").
			Str("messageID", msg.ID.String()).
			Dur("duration", msgCreateDuration).
			Msg("New message object created")

		// Append to manager
		appendStart := time.Now()
		if err := m.manager.AppendMessages(msg_); err != nil {
			log.Error().
				Err(err).
				Str("messageID", msg.ID.String()).
				Msg("Failed to append message - creating error message")

			// Create an error message to display to the user
			errorContent := conversation2.NewErrorContentWithDetails(
				conversation2.ErrorTypeTreeCorruption,
				"Failed to add message to conversation",
				fmt.Sprintf("Error details: %s", err.Error()),
				false, // not recoverable
			)
			errorMsg := conversation2.NewMessage(errorContent,
				conversation2.WithMetadata(map[string]interface{}{
					"original_message_id": msg.ID.String(),
					"parent_id":           msg.ParentID.String(),
					"error_type":          "append_failure",
				}),
			)

			// Try to append the error message (this should succeed as it's a different ID)
			if errAppend := m.manager.AppendMessages(errorMsg); errAppend != nil {
				log.Error().
					Err(errAppend).
					Msg("Failed to append error message - critical error")
			} else {
				m.updateCache(errorMsg)
			}
			return m, nil
		}
		appendDuration := time.Since(appendStart)

		log.Debug().
			Str("operation", "message_append").
			Str("messageID", msg.ID.String()).
			Dur("duration", appendDuration).
			Int("new_conversation_size", len(m.manager.GetConversation())).
			Msg("Message appended to manager")

		// Update cache
		cacheStart := time.Now()
		m.updateCache(msg_)
		cacheDuration := time.Since(cacheStart)

		log.Debug().
			Str("operation", "cache_update").
			Str("messageID", msg.ID.String()).
			Dur("duration", cacheDuration).
			Int("cache_size", len(m.cache)).
			Msg("Cache updated for new message")

		totalDuration := time.Since(startTime)
		logMemoryUsage("stream_start_complete")

		log.Debug().
			Str("operation", "stream_start_complete").
			Str("messageID", msg.ID.String()).
			Dur("total_duration", totalDuration).
			Bool("was_duplicate", isDuplicate).
			Msg("StreamStartMsg processing completed")

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
	log.Trace().
		Str("operation", "view_generation_start").
		Int("cache_size", len(m.cache)).
		Msg("Starting view generation")

	ret := ""
	height := 0
	selectedOffset := 0
	selectedHeight := 0

	msgs_ := m.manager.GetConversation()

	log.Trace().
		Str("operation", "conversation_retrieval").
		Int("message_count", len(msgs_)).
		Msg("Retrieved conversation messages")

	// This triggers cache update for ALL messages
	cacheUpdateStart := time.Now()
	m.updateCache(msgs_...)
	cacheUpdateDuration := time.Since(cacheUpdateStart)

	log.Trace().
		Str("operation", "full_cache_update").
		Dur("duration", cacheUpdateDuration).
		Int("processed_messages", len(msgs_)).
		Msg("Full cache update completed in view generation")

	// Assemble the view
	renderedMessages := 0
	skippedMessages := 0
	totalContentLength := 0

	for _, msg := range msgs_ {
		c_, ok := m.cache[msg.ID]
		if !ok {
			skippedMessages++
			log.Trace().
				Str("operation", "view_skip_message").
				Str("messageID", msg.ID.String()).
				Msg("Skipping message - not in cache")
			continue
		}

		renderedMessages++
		h := lipgloss.Height(c_.rendered)
		contentLen := len(c_.rendered)
		totalContentLength += contentLen

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
	StepMetadata  *events.StepMetadata  `json:"step_metadata,omitempty"`
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
