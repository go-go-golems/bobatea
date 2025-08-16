package chat

import (
	context2 "context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	geppetto_events "github.com/go-go-golems/geppetto/pkg/events"
	"github.com/google/uuid"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/filepicker"
	mode_keymap "github.com/go-go-golems/bobatea/pkg/mode-keymap"
	"github.com/go-go-golems/bobatea/pkg/textarea"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	renderers "github.com/go-go-golems/bobatea/pkg/timeline/renderers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Tracing counters for debugging recursive calls
var (
	updateCallCounter = int64(0)
	viewCallCounter   = int64(0)
)

type ErrorMsg error

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

type Status struct {
	State        State  `json:"state"`
	InputText    string `json:"inputText"`
	SelectedIdx  int    `json:"selectedIdx"`
	MessageCount int    `json:"messageCount"`
	Error        error  `json:"error,omitempty"`
}

type model struct {
	autoStartBackend bool

	viewport       viewport.Model
	scrollToBottom bool

	// not really what we want, but use this for now, we'll have to either find a normal text box,
	// or implement wrapping ourselves.
	textArea textarea.Model

	filepicker filepicker.Model

	// conversation conversationui.Model // removed in favor of timeline selection

	// Timeline controller replaces conversation view rendering
	timelineReg  *timeline.Registry
	timelineCtrl *timeline.Controller
	entityVers   map[string]int64
	// entityStart removed; engines now provide DurationMs in metadata
	// entityStart     map[string]time.Time
	timelineRegHook func(*timeline.Registry)

	help help.Model

	err    error
	keyMap KeyMap

	style  *Style
	width  int
	height int

	backend Backend

	state        State
	quitReceived bool

	title string

	status *Status

	// inputBlurred tracks whether the input field has been programmatically blurred
	// by BlurInputMsg / UnblurInputMsg actions. This flag can be used by UI tooling
	// to gate or reflect input state if needed.
	inputBlurred bool
}

type ModelOption func(*model)

func WithTitle(title string) ModelOption {
	return func(m *model) {
		m.title = title
	}
}

func WithStatus(status *Status) ModelOption {
	return func(m *model) {
		m.status = status
	}
}

func WithAutoStartBackend(autoStartBackend bool) ModelOption {
	return func(m *model) {
		m.autoStartBackend = autoStartBackend
	}
}

// WithTimelineRegister allows callers to register custom renderers on the timeline registry
func WithTimelineRegister(hook func(*timeline.Registry)) ModelOption {
	return func(m *model) {
		m.timelineRegHook = hook
	}
}

// TODO(manuel, 2024-04-07) Add options to configure filepicker

func InitialModel(backend Backend, options ...ModelOption) model {
	fp := filepicker.NewModel()

	fp.Filepicker.DirAllowed = false
	fp.Filepicker.FileAllowed = true
	dir, _ := os.Getwd()
	fp.Filepicker.CurrentDirectory = dir
	fp.Filepicker.Height = 10

	ret := model{
		filepicker:     fp,
		style:          DefaultStyles(),
		keyMap:         DefaultKeyMap,
		backend:        backend,
		viewport:       viewport.New(0, 0),
		help:           help.New(),
		scrollToBottom: true,
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

	// Initialize timeline components
	ret.timelineReg = timeline.NewRegistry()
	// Register interactive entity model factories
	ret.timelineReg.RegisterModelFactory(renderers.NewLLMTextFactory())
	ret.timelineReg.RegisterModelFactory(renderers.ToolCallsPanelFactory{})
	ret.timelineReg.RegisterModelFactory(renderers.PlainFactory{})
	if ret.timelineRegHook != nil {
		ret.timelineRegHook(ret.timelineReg)
	}
	ret.timelineCtrl = timeline.NewController(ret.timelineReg)
	ret.entityVers = map[string]int64{}
	// ret.entityStart = map[string]time.Time{}

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

	cmds = append(cmds, m.filepicker.Init(), m.viewport.Init())

	// Seed existing chat messages as timeline entities
	// Seeding from conversation is disabled; timeline should be sourced from entity events

	// Set initial timeline view content
	{
		v := m.timelineCtrl.View()
		log.Debug().Str("component", "chat").Str("when", "Init").Int("view_len", len(v)).Msg("SetContent")
		m.viewport.SetContent(v)
	}
	m.viewport.YPosition = 0
	m.viewport.GotoBottom()

	m.updateKeyBindings()
	// Select last entity if any
	m.timelineCtrl.SelectLast()

	if m.autoStartBackend {
		cmds = append(cmds, func() tea.Msg {
			return StartBackendMsg{}
		})
	}

	return tea.Batch(cmds...)
}

func (m *model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If an error is visible, allow quick dismissal with Esc or Enter
	if m.state == StateError {
		if msg.String() == "esc" || msg.String() == "enter" {
			return m.handleUserAction(DismissErrorMsg{})
		}
	}

	// When streaming, forbid entering/focusing input and submitting
	if m.state == StateStreamCompletion {
		if key.Matches(msg, m.keyMap.SubmitMessage) || key.Matches(msg, m.keyMap.FocusMessage) {
			return m, nil
		}
	}

	switch {
	case key.Matches(msg, m.keyMap.Help):
		log.Debug().Str("component", "chat").Str("key", msg.String()).Msg("Help pressed")
		cmd = func() tea.Msg { return ToggleHelpMsg{} }
	case key.Matches(msg, m.keyMap.UnfocusMessage):
		log.Debug().Str("component", "chat").Str("key", msg.String()).Msg("Unfocus (ESC) pressed")
		cmd = func() tea.Msg { return UnfocusMessageMsg{} }
	case key.Matches(msg, m.keyMap.Quit):
		log.Debug().Str("component", "chat").Str("key", msg.String()).Msg("Quit pressed")
		cmd = func() tea.Msg { return QuitMsg{} }
	case key.Matches(msg, m.keyMap.FocusMessage):
		log.Debug().Str("component", "chat").Str("key", msg.String()).Msg("Focus pressed")
		cmd = func() tea.Msg { return FocusMessageMsg{} }
	case key.Matches(msg, m.keyMap.SelectNextMessage):
		log.Debug().Str("component", "chat").Str("key", msg.String()).Msg("SelectNext pressed")
		cmd = func() tea.Msg { return SelectNextMessageMsg{} }
	case key.Matches(msg, m.keyMap.SelectPrevMessage):
		log.Debug().Str("component", "chat").Str("key", msg.String()).Msg("SelectPrev pressed")
		cmd = func() tea.Msg { return SelectPrevMessageMsg{} }
	case key.Matches(msg, m.keyMap.SubmitMessage):
		cmd = func() tea.Msg { return SubmitMessageMsg{} }
	case key.Matches(msg, m.keyMap.CopyToClipboard):
		cmd = func() tea.Msg { return CopyToClipboardMsg{} }
	case key.Matches(msg, m.keyMap.CopyLastResponseToClipboard):
		cmd = func() tea.Msg { return CopyLastResponseToClipboardMsg{} }
	case key.Matches(msg, m.keyMap.CopyLastSourceBlocksToClipboard):
		cmd = func() tea.Msg { return CopyLastSourceBlocksToClipboardMsg{} }
	case key.Matches(msg, m.keyMap.CopySourceBlocksToClipboard):
		cmd = func() tea.Msg { return CopySourceBlocksToClipboardMsg{} }
	case key.Matches(msg, m.keyMap.SaveToFile):
		cmd = func() tea.Msg { return SaveToFileMsg{} }
	case key.Matches(msg, m.keyMap.CancelCompletion):
		cmd = func() tea.Msg { return CancelCompletionMsg{} }
	case key.Matches(msg, m.keyMap.DismissError):
		cmd = func() tea.Msg { return DismissErrorMsg{} }
	default:
		switch m.state {
		case StateUserInput:
			m.textArea, cmd = m.textArea.Update(msg)
		case StateSavingToFile:
			var updatedModel tea.Model
			updatedModel, cmd = m.filepicker.Update(msg)
			m.filepicker = updatedModel.(filepicker.Model)
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
	// No conversation manager; writing viewport content as a simple fallback
	// In a real backend, this should request an export from the backend.
	content := m.timelineCtrl.View()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return m, func() tea.Msg { return ErrorMsg(err) }
	}

	m.state = StateUserInput
	m.updateKeyBindings()
	m.recomputeSize()

	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Track Update calls and timing
	updateCallID := atomic.AddInt64(&updateCallCounter, 1)
	updateStartTime := time.Now()
	msgType := fmt.Sprintf("%T", msg)

	// Consolidated logger for this update call
	logger := log.With().Int64("update_call_id", updateCallID).Logger()
	logger.Trace().
		Str("msg_type", msgType).
		Str("current_state", string(m.state)).
		Bool("scroll_to_bottom", m.scrollToBottom).
		Bool("backend_finished", m.backend.IsFinished()).
		Time("start_time", updateStartTime).
		Msg("UPDATE ENTRY")

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg_ := msg.(type) {
	case tea.KeyMsg:
		// When input is blurred, ignore key events on the input field
		if m.inputBlurred {
			return m, nil
		}
		// Entering mode and selection routing
		if m.state == StateMovingAround {
			switch msg_.String() {
			case "enter":
				m.timelineCtrl.EnterSelection()
				v := m.timelineCtrl.View()
				log.Debug().Str("component", "chat").Str("when", "enter_selection").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
				m.viewport.SetContent(v)
				return m, nil
			case "esc":
				if m.timelineCtrl.IsEntering() {
					m.timelineCtrl.ExitSelection()
					v := m.timelineCtrl.View()
					log.Debug().Str("component", "chat").Str("when", "exit_entering").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
					m.viewport.SetContent(v)
					return m, nil
				}
				// leave moving-around back to text mode
				log.Debug().Str("component", "chat").Str("transition", "moving-around->user-input").Msg("State transition")
				m.state = StateUserInput
				m.textArea.Focus()
				m.updateKeyBindings()
				// hide selection highlight and unselect entity
				m.timelineCtrl.SetSelectionVisible(false)
				m.timelineCtrl.Unselect()
				v := m.timelineCtrl.View()
				log.Debug().Str("component", "chat").Str("when", "esc_to_input").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
				m.viewport.SetContent(v)
				return m, nil
			}
			// Route all keys while entering to the selected model and log
			if m.timelineCtrl.IsEntering() {
				logger.Debug().Str("route", "entering").Str("key", msg_.String()).Msg("Routing key to selected entity model")
				cmd := m.timelineCtrl.HandleMsg(msg_)
				v := m.timelineCtrl.View()
				log.Debug().Str("component", "chat").Str("when", "entering_route_key").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
				m.viewport.SetContent(v)
				return m, cmd
			}
			// Allow entities to react to copy requests even when not entering
			if msg_.String() == "alt+c" {
				cmd := m.timelineCtrl.SendToSelected(timeline.EntityCopyTextMsg{})
				v := m.timelineCtrl.View()
				log.Debug().Str("component", "chat").Str("when", "copy_selected").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
				m.viewport.SetContent(v)
				if cmd != nil {
					return m, cmd
				}
			}
		}
		return m.handleKeyPress(msg_)

	case tea.WindowSizeMsg:
		logger.Debug().Int("width", msg_.Width).Int("height", msg_.Height).Msg("Window size changed")
		m.width = msg_.Width
		m.height = msg_.Height
		m.timelineCtrl.SetSize(m.width, m.height)
		m.recomputeSize()

	case ErrorMsg:
		logger.Trace().Str("error", msg_.Error()).Msg("Error message received")
		m.err = msg_
		m.state = StateError
		m.updateKeyBindings()
		return m, nil

	case StreamCompletionMsg,
		StreamStartMsg,
		StreamStatusMsg,
		StreamDoneMsg,
		StreamCompletionError:

		logger.Trace().Str("stream_msg_type", msgType).Msg("Stream message received - ENTERING STREAM PROCESSING")

		startTime := time.Now()

		switch streamMsg := msg.(type) {
		case StreamStartMsg:
			logger.Debug().Str("operation", "stream_start_reception").
				Str("messageID", streamMsg.ID.String()).
				Time("timestamp", startTime).
				Bool("scroll_to_bottom", m.scrollToBottom).
				Msg("StreamStartMsg details")
		case StreamCompletionMsg:
			logger.Debug().Str("operation", "stream_completion_reception").
				Str("messageID", streamMsg.ID.String()).
				Int("delta_length", len(streamMsg.Delta)).
				Int("completion_length", len(streamMsg.Completion)).
				Msg("StreamCompletionMsg details")
		case StreamDoneMsg:
			logger.Debug().Str("operation", "stream_done_reception").
				Str("messageID", streamMsg.ID.String()).
				Msg("StreamDoneMsg details")
		}

		// Translate stream messages to timeline entity lifecycle
		switch v := msg.(type) {
		case StreamStartMsg:
			id := v.ID.String()
			m.entityVers[id] = 0
			// start time tracking removed; rely on metadata.DurationMs from engines
			md := toLLMInferenceData(v.EventMetadata)
			log.Debug().
				Str("local_id", id).
				Str("model", func() string {
					if md != nil {
						return md.Model
					}
					return ""
				}()).
				Interface("usage", func() any {
					if md != nil {
						return md.Usage
					}
					return nil
				}()).
				Msg("StreamStartMsg: converted metadata")
			m.timelineCtrl.OnCreated(timeline.UIEntityCreated{
				ID:        timeline.EntityID{LocalID: id, Kind: "llm_text"},
				Renderer:  timeline.RendererDescriptor{Kind: "llm_text"},
				Props:     map[string]any{"role": "assistant", "text": "", "metadata": md},
				StartedAt: time.Now(),
			})
		case StreamCompletionMsg:
			id := v.ID.String()
			m.entityVers[id] = m.entityVers[id] + 1
			md := toLLMInferenceData(v.EventMetadata)
			log.Debug().
				Str("local_id", id).
				Int64("version", m.entityVers[id]).
				Str("model", func() string {
					if md != nil {
						return md.Model
					}
					return ""
				}()).
				Msg("StreamCompletionMsg: converted metadata")
			m.timelineCtrl.OnUpdated(timeline.UIEntityUpdated{
				ID:        timeline.EntityID{LocalID: id, Kind: "llm_text"},
				Patch:     map[string]any{"text": v.Completion, "metadata": md},
				Version:   m.entityVers[id],
				UpdatedAt: time.Now(),
			})
		case StreamDoneMsg:
			id := v.ID.String()
			md := toLLMInferenceData(v.EventMetadata)
			log.Debug().
				Str("local_id", id).
				Interface("duration_ms", func() any {
					if md != nil {
						return md.DurationMs
					}
					return nil
				}()).
				Interface("usage", func() any {
					if md != nil {
						return md.Usage
					}
					return nil
				}()).
				Str("stop_reason", func() string {
					if md != nil && md.StopReason != nil {
						return *md.StopReason
					}
					return ""
				}()).
				Msg("StreamDoneMsg: converted metadata")
			m.timelineCtrl.OnCompleted(timeline.UIEntityCompleted{
				ID:     timeline.EntityID{LocalID: id, Kind: "llm_text"},
				Result: map[string]any{"text": v.Completion, "metadata": md},
			})
			cmds = append(cmds, m.finishCompletion())
		case StreamCompletionError:
			id := v.ID.String()
			m.timelineCtrl.OnCompleted(timeline.UIEntityCompleted{
				ID:     timeline.EntityID{LocalID: id, Kind: "llm_text"},
				Result: map[string]any{"text": fmt.Sprintf("**Error**\n\n%s", v.Err)},
			})
			cmds = append(cmds, m.finishCompletion())
		}

		if m.scrollToBottom {
			v := m.timelineCtrl.View()
			log.Debug().Str("component", "chat").Str("when", "external_created").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
			m.viewport.SetContent(v)
			m.viewport.GotoBottom()
		}

		totalDuration := time.Since(startTime)
		logger.Trace().Str("operation", "stream_message_total").
			Dur("total_duration", totalDuration).
			Msg("Stream message processing completed")

		cmds = append(cmds, cmd)

	case BackendFinishedMsg:
		logger.Trace().Msg("Backend finished - calling finishCompletion()")
		cmd = m.finishCompletion()
		cmds = append(cmds, cmd)

		// Accept external timeline lifecycle messages (e.g., from backend simulating agent tool calls)
	case timeline.UIEntityCreated:
		logger.Debug().Str("lifecycle", "created").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Msg("Applying external entity event")
		m.timelineCtrl.OnCreated(msg_)
		if m.scrollToBottom {
			v := m.timelineCtrl.View()
			log.Debug().Str("component", "chat").Str("when", "external_updated").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
			m.viewport.SetContent(v)
			m.viewport.GotoBottom()
		}
		return m, nil
	case timeline.UIEntityUpdated:
		logger.Debug().Str("lifecycle", "updated").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Int64("version", msg_.Version).Msg("Applying external entity event")
		m.timelineCtrl.OnUpdated(msg_)
		if m.scrollToBottom {
			v := m.timelineCtrl.View()
			log.Debug().Str("component", "chat").Str("when", "external_completed").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
			m.viewport.SetContent(v)
			m.viewport.GotoBottom()
		}
		return m, nil
	case timeline.UIEntityCompleted:
		logger.Debug().Str("lifecycle", "completed").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Msg("Applying external entity event")
		m.timelineCtrl.OnCompleted(msg_)
		if m.scrollToBottom {
			v := m.timelineCtrl.View()
			log.Debug().Str("component", "chat").Str("when", "external_deleted").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
			m.viewport.SetContent(v)
			m.viewport.GotoBottom()
		}
		return m, nil
	case timeline.UIEntityDeleted:
		logger.Debug().Str("lifecycle", "deleted").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Msg("Applying external entity event")
		m.timelineCtrl.OnDeleted(msg_)
		if m.scrollToBottom {
			v := m.timelineCtrl.View()
			log.Debug().Str("component", "chat").Str("when", "scrollToSelected").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
			m.viewport.SetContent(v)
			m.viewport.GotoBottom()
		}
		return m, nil

	// Side-effects requested by entity models
	case timeline.CopyTextRequestedMsg:
		_ = clipboard.WriteAll(msg_.Text)
		return m, nil
	case timeline.CopyCodeRequestedMsg:
		_ = clipboard.WriteAll(msg_.Code)
		return m, nil

	case refreshMessageMsg:
		logger.Trace().Bool("go_to_bottom", msg_.GoToBottom).Bool("scroll_to_bottom", m.scrollToBottom).Msg("REFRESH MESSAGE - POTENTIAL TRIGGER FOR LOOPS")

		v := m.timelineCtrl.View()
		m.viewport.SetContent(v)
		m.recomputeSize()
		if msg_.GoToBottom || m.scrollToBottom {
			m.viewport.GotoBottom()
		}

	case filepicker.SelectFileMsg:
		logger.Trace().Str("path", msg_.Path).Msg("File selected for saving")
		return m.saveToFile(msg_.Path)

	case filepicker.CancelFilePickerMsg:
		logger.Trace().Msg("File picker cancelled")
		m.state = StateUserInput
		m.updateKeyBindings()

	case StartBackendMsg:
		logger.Trace().Msg("Starting backend - POTENTIAL COMMAND GENERATOR")
		return m, m.startBackend()

	case UserActionMsg:
		logger.Trace().Str("user_action_type", fmt.Sprintf("%T", msg_)).Msg("User action message - calling handleUserAction()")
		return m.handleUserAction(msg_)

	default:
		logger.Trace().Str("msg_type", msgType).Str("state", string(m.state)).Msg("DEFAULT CASE - updating viewport or filepicker")

		switch m.state {
		case StateUserInput, StateError, StateStreamCompletion:
			m.viewport, cmd = m.viewport.Update(msg_)
			if cmd != nil {
				logger.Trace().Str("viewport_cmd_type", fmt.Sprintf("%T", cmd)).Msg("Viewport returned command")
			}
			cmds = append(cmds, cmd)
		case StateMovingAround:
			// In moving-around mode, use timeline selection controls and scroll
			if km, ok := msg_.(tea.KeyMsg); ok {
				switch {
				case key.Matches(km, m.keyMap.SelectNextMessage):
					m.timelineCtrl.SelectNext()
					m.scrollToSelected()
				case key.Matches(km, m.keyMap.SelectPrevMessage):
					m.timelineCtrl.SelectPrev()
					m.scrollToSelected()
				case key.Matches(km, m.keyMap.ScrollDown):
					m.viewport.ScrollDown(1)
				case key.Matches(km, m.keyMap.ScrollUp):
					m.viewport.ScrollUp(1)
				}
			}
		case StateSavingToFile:
			log.Trace().
				Int64("update_call_id", updateCallID).
				Msg("Updating filepicker")
			var updatedModel tea.Model
			updatedModel, cmd = m.filepicker.Update(msg_)
			m.filepicker = updatedModel.(filepicker.Model)
			cmds = append(cmds, cmd)
		}
	}

	// Update status if it's not nil
	if m.status != nil {
		oldMessageCount := m.status.MessageCount
		m.status.State = m.state
		m.status.InputText = m.textArea.Value()
		m.status.SelectedIdx = m.timelineCtrl.SelectedIndex()
		// Fallback approximation using rendered height if entity count is not available
		m.status.MessageCount = lipgloss.Height(m.timelineCtrl.View())
		m.status.Error = m.err

		if oldMessageCount != m.status.MessageCount {
			logger.Trace().Int("old_count", oldMessageCount).Int("new_count", m.status.MessageCount).Msg("MESSAGE COUNT CHANGED")
		}
	}

	updateDuration := time.Since(updateStartTime)
	cmdCount := len(cmds)

	// Log all commands being returned
	for i, cmd := range cmds {
		if cmd != nil {
			logger.Trace().Int("cmd_index", i).Str("cmd_type", fmt.Sprintf("%T", cmd)).Msg("COMMAND BEING RETURNED - POTENTIAL LOOP SOURCE")
		}
	}

	logger.Trace().Str("msg_type", msgType).Dur("total_duration", updateDuration).Int("cmd_count", cmdCount).Str("final_state", string(m.state)).Msg("UPDATE EXIT")

	return m, tea.Batch(cmds...)
}

func (m *model) updateKeyBindings() {
	mode_keymap.EnableMode(&m.keyMap, string(m.state))
}

// scrollToSelected scrolls the viewport to keep the selected entity in view
func (m *model) scrollToSelected() {
	scrollCallID := atomic.AddInt64(&updateCallCounter, 1)
	log.Trace().
		Int64("scroll_call_id", scrollCallID).
		Int("viewport_y_offset", m.viewport.YOffset).
		Int("viewport_height", m.viewport.Height).
		Msg("SCROLL TO SELECTED ENTRY - TIMELINE")

	viewStart := time.Now()
	v, off, h := m.timelineCtrl.ViewAndSelectedPosition()
	viewDuration := time.Since(viewStart)

	log.Trace().
		Int64("scroll_call_id", scrollCallID).
		Dur("view_generation", viewDuration).
		Int("view_length", len(v)).
		Int("pos_offset", off).
		Int("pos_height", h).
		Msg("View generated for scroll calculation")

	setContentStart := time.Now()
	m.viewport.SetContent(v)
	setContentDuration := time.Since(setContentStart)

	midScreenOffset := m.viewport.YOffset + m.viewport.Height/2
	msgEndOffset := off + h
	bottomOffset := m.viewport.YOffset + m.viewport.Height

	if off > midScreenOffset && msgEndOffset > bottomOffset {
		newOffset := off - max(m.viewport.Height-h-1, m.viewport.Height/2)
		m.viewport.SetYOffset(newOffset)
		log.Trace().
			Int64("scroll_call_id", scrollCallID).
			Int("new_y_offset", newOffset).
			Msg("Scrolled down to show entity")
	} else if off < m.viewport.YOffset {
		m.viewport.SetYOffset(off)
		log.Trace().
			Int64("scroll_call_id", scrollCallID).
			Int("new_y_offset", off).
			Msg("Scrolled up to show entity")
	}

	log.Trace().
		Int64("scroll_call_id", scrollCallID).
		Dur("set_content_duration", setContentDuration).
		Msg("SCROLL TO SELECTED EXIT")
}

func (m *model) recomputeSize() {
	recomputeCallID := atomic.AddInt64(&updateCallCounter, 1)
	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Int("model_width", m.width).
		Int("model_height", m.height).
		Str("state", string(m.state)).
		Msg("RECOMPUTE SIZE ENTRY - POTENTIAL CASCADE TRIGGER")

	headerStart := time.Now()
	headerView := m.headerView()
	headerHeight := lipgloss.Height(headerView)
	if headerView == "" {
		headerHeight = 0
	}
	headerDuration := time.Since(headerStart)

	helpStart := time.Now()
	helpView := m.help.View(m.keyMap)
	helpViewHeight := lipgloss.Height(helpView)
	helpDuration := time.Since(helpStart)

	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Dur("header_duration", headerDuration).
		Dur("help_duration", helpDuration).
		Int("header_height", headerHeight).
		Int("help_height", helpViewHeight).
		Msg("Header and help views computed")

	if m.state == StateSavingToFile {
		m.filepicker.Filepicker.Height = m.height - headerHeight - helpViewHeight
		log.Trace().
			Int64("recompute_call_id", recomputeCallID).
			Int("filepicker_height", m.filepicker.Filepicker.Height).
			Msg("File picker size set")
		return
	}

	textAreaStart := time.Now()
	textAreaView := m.textAreaView()
	textAreaHeight := lipgloss.Height(textAreaView)
	textAreaDuration := time.Since(textAreaStart)

	newHeight := m.height - textAreaHeight - headerHeight - helpViewHeight
	if newHeight < 0 {
		newHeight = 0
	}

	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Dur("textarea_duration", textAreaDuration).
		Int("textarea_height", textAreaHeight).
		Int("calculated_viewport_height", newHeight).
		Msg("Text area computed, viewport height calculated")

	// Update viewport dimensions
	m.viewport.Width = m.width
	m.viewport.Height = newHeight
	m.viewport.YPosition = headerHeight + 1

	h, _ := m.style.SelectedMessage.GetFrameSize()
	m.textArea.SetWidth(m.width - h)

	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Int("viewport_width", m.viewport.Width).
		Int("viewport_height", m.viewport.Height).
		Int("viewport_y_position", m.viewport.YPosition).
		Int("textarea_width", m.width-h).
		Msg("Component dimensions updated")

	// CRITICAL: Regenerate timeline view and set content
	v := m.timelineCtrl.View()
	log.Debug().Str("component", "chat").Str("when", "recompute_size").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
	m.viewport.SetContent(v)
	m.viewport.GotoBottom()
	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Int("view_length", len(v)).
		Msg("RECOMPUTE SIZE EXIT - View regenerated and viewport updated")
}

func (m model) headerView() string {
	return ""
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
	case StateMovingAround:
		// Grey out input when in selection mode
		v = m.style.UnselectedMessage.Foreground(lipgloss.Color("240")).Render(v)
	case StateStreamCompletion:
		// Grey out and ensure blurred while streaming
		m.textArea.Blur()
		v = m.style.UnselectedMessage.Render(v)
	case StateError, StateSavingToFile:
	}

	return v
}

func (m model) View() string {
	// Track View calls and timing - CRITICAL FOR DETECTING EXCESSIVE RENDERS
	viewCallID := atomic.AddInt64(&viewCallCounter, 1)
	viewStartTime := time.Now()

	vlogger := log.With().Int64("view_call_id", viewCallID).Logger()
	vlogger.Trace().Str("state", string(m.state)).Bool("scroll_to_bottom", m.scrollToBottom).Time("start_time", viewStartTime).Msg("VIEW ENTRY - POTENTIAL EXCESSIVE CALL POINT")

	headerStart := time.Now()
	headerView := m.headerView()
	headerDuration := time.Since(headerStart)

	vlogger.Trace().Dur("header_duration", headerDuration).Bool("header_empty", headerView == "").Msg("Header view generated")

	// Generate timeline view instead of conversation view
	view := m.timelineCtrl.View()
	m.viewport.SetContent(view)

	vlogger.Trace().Int("viewport_width", m.viewport.Width).Int("viewport_height", m.viewport.Height).Int("viewport_y_position", m.viewport.YPosition).Msg("VIEWPORT CONTENT SET - POTENTIAL TRIGGER FOR UPDATES")

	viewportViewStart := time.Now()
	viewportView := m.viewport.View()
	viewportViewDuration := time.Since(viewportViewStart)

	textAreaStart := time.Now()
	textAreaView := m.textAreaView()
	textAreaDuration := time.Since(textAreaStart)

	helpStart := time.Now()
	helpView := m.help.View(m.keyMap)
	helpDuration := time.Since(helpStart)

	vlogger.Trace().Dur("viewport_view_duration", viewportViewDuration).Dur("textarea_duration", textAreaDuration).Dur("help_duration", helpDuration).Msg("UI component views generated")

	// debugging heights with trace logging
	viewportHeight := lipgloss.Height(viewportView)
	textAreaHeight := lipgloss.Height(textAreaView)
	headerHeight := lipgloss.Height(headerView)
	helpViewHeight := lipgloss.Height(helpView)

	vlogger.Trace().Int("viewport_height", viewportHeight).Int("textarea_height", textAreaHeight).Int("header_height", headerHeight).Int("help_height", helpViewHeight).Int("total_calculated_height", viewportHeight+textAreaHeight+headerHeight+helpViewHeight).Int("model_height", m.height).Msg("Height calculations")

	ret := ""
	if headerView != "" {
		ret = headerView
	}

	switch m.state {
	case StateUserInput, StateError, StateStreamCompletion:
		ret += viewportView + "\n" + textAreaView + "\n" + helpView
		vlogger.Trace().Str("combined_state", "viewport+textarea+help").Int("final_length", len(ret)).Msg("Combined view for main states")
	case StateMovingAround:
		// Keep input visible (greyed) while selecting entities
		ret += viewportView + "\n" + textAreaView + "\n" + helpView
		vlogger.Trace().Str("combined_state", "viewport+textarea+help (selection mode)").Int("final_length", len(ret)).Msg("Combined view for moving-around state")

	case StateSavingToFile:
		ret += m.filepicker.View()
		vlogger.Trace().Str("combined_state", "filepicker").Int("final_length", len(ret)).Msg("Combined view for file saving state")
	}

	viewDuration := time.Since(viewStartTime)
	vlogger.Trace().Dur("total_view_duration", viewDuration).Int("final_output_length", len(ret)).Msg("VIEW EXIT")

	return ret
}

func (m *model) startBackend() tea.Cmd {
	startCallID := atomic.AddInt64(&updateCallCounter, 1)
	log.Debug().
		Int64("start_call_id", startCallID).
		Str("previous_state", string(m.state)).
		Bool("backend_finished", m.backend.IsFinished()).
		Msg("START BACKEND ENTRY - MAJOR COMMAND GENERATOR")

	m.state = StateStreamCompletion
	m.textArea.Blur()
	m.updateKeyBindings()

	log.Debug().
		Int64("start_call_id", startCallID).
		Msg("Calling viewport.GotoBottom()")
	m.viewport.GotoBottom()

	refreshCmd := func() tea.Msg {
		log.Debug().
			Int64("start_call_id", startCallID).
			Msg("REFRESH MESSAGE FROM START BACKEND - LOOP RISK")
		return refreshMessageMsg{
			GoToBottom: true,
		}
	}

	backendCmd := func() tea.Msg {
		log.Debug().
			Int64("start_call_id", startCallID).
			Msg("BACKEND START COMMAND EXECUTING (no-op in new prompt flow)")
		return nil
	}

	log.Debug().
		Int64("start_call_id", startCallID).
		Msg("START BACKEND EXIT - returning batch of refresh + backend commands")

	return tea.Batch(refreshCmd, backendCmd)
}

func (m *model) submit() tea.Cmd {
	submitCallID := atomic.AddInt64(&updateCallCounter, 1)
	slogger := log.With().Int64("submit_call_id", submitCallID).Logger()
	slogger.Trace().
		Bool("backend_finished", m.backend.IsFinished()).
		Int("input_length", len(m.textArea.Value())).
		Msg("SUBMIT ENTRY")

	if !m.backend.IsFinished() {
		slogger.Trace().Msg("Backend not finished - returning error")
		return func() tea.Msg {
			return ErrorMsg(errors.New("already streaming"))
		}
	}

	// Filter out empty submissions (spaces/newlines only)
	rawInput := m.textArea.Value()
	userMessage := strings.TrimSpace(rawInput)
	if userMessage == "" {
		slogger.Debug().Msg("Ignoring empty submit (no message sent)")
		return nil
	}

	// Add entity to timeline
	id := uuid.New().String()
	log.Debug().Str("component", "chat").Str("when", "submit").Str("id", id).Msg("Adding user message to timeline")
	m.timelineCtrl.OnCreated(timeline.UIEntityCreated{
		ID:       timeline.EntityID{LocalID: id, Kind: "llm_text"},
		Renderer: timeline.RendererDescriptor{Kind: "llm_text"},
		Props:    map[string]any{"role": "user", "text": userMessage},
	})
	log.Debug().Str("component", "chat").Str("when", "submit").Str("id", id).Msg("Adding user message to timeline")
	m.timelineCtrl.OnCompleted(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: id, Kind: "llm_text"}})
	log.Debug().Str("component", "chat").Str("when", "submit").Str("id", id).Msg("User message added to timeline")

	m.textArea.SetValue("")

	refreshCmd := func() tea.Msg {
		log.Debug().
			Int64("submit_call_id", submitCallID).
			Msg("REFRESH COMMAND EXECUTED - POTENTIAL LOOP TRIGGER")
		return refreshMessageMsg{
			GoToBottom: true,
		}
	}

	backendCmd := func() tea.Msg {
		ctx := context2.Background()
		cmd, err := m.backend.Start(ctx, userMessage)
		if err != nil {
			return ErrorMsg(err)
		}
		return cmd()
	}
	return tea.Batch(refreshCmd, backendCmd)
}

type refreshMessageMsg struct {
	GoToBottom bool
}

func (m *model) finishCompletion() tea.Cmd {
	finishCallID := atomic.AddInt64(&updateCallCounter, 1)
	log.Trace().
		Int64("finish_call_id", finishCallID).
		Str("current_state", string(m.state)).
		Bool("quit_received", m.quitReceived).
		Bool("backend_finished", m.backend.IsFinished()).
		Msg("FINISH COMPLETION ENTRY")

	refreshCommand := func() tea.Msg {
		log.Trace().
			Int64("finish_call_id", finishCallID).
			Msg("REFRESH COMMAND EXECUTED - POTENTIAL LOOP TRIGGER")
		return refreshMessageMsg{
			GoToBottom: true,
		}
	}

	if m.state == StateStreamCompletion {
		log.Trace().
			Int64("finish_call_id", finishCallID).
			Msg("Stream completion state - performing cleanup")

		// WARN not sure if really necessary actually, this should only be called once at this point.
		m.backend.Kill()

		m.state = StateUserInput
		m.textArea.Focus()
		m.textArea.SetValue("")

		log.Trace().
			Int64("finish_call_id", finishCallID).
			Msg("CALLING recomputeSize() from finishCompletion - CASCADE RISK")
		m.recomputeSize()
		m.updateKeyBindings()

		if m.quitReceived {
			log.Trace().
				Int64("finish_call_id", finishCallID).
				Msg("Quit received - returning tea.Quit")
			return tea.Quit
		}
	}

	log.Trace().
		Int64("finish_call_id", finishCallID).
		Msg("FINISH COMPLETION EXIT - returning refresh command")
	return refreshCommand
}

func (m model) handleUserAction(msg UserActionMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg_ := msg.(type) {
	case ToggleHelpMsg:
		m.help.ShowAll = !m.help.ShowAll

	case UnfocusMessageMsg:
		if m.state == StateUserInput {
			m.textArea.Blur()
			m.state = StateMovingAround
			// Enter moving around; select last entity and show selection highlight
			m.timelineCtrl.SelectLast()
			m.timelineCtrl.SetSelectionVisible(true)
			v := m.timelineCtrl.View()
			log.Debug().Str("component", "chat").Str("when", "handleUserAction_search").Int("view_len", len(v)).Int("y_offset", m.viewport.YOffset).Msg("SetContent")
			m.viewport.SetContent(v)
			m.updateKeyBindings()
			log.Debug().Str("component", "chat").Str("transition", "user-input->moving-around").Int("selected_index", m.timelineCtrl.SelectedIndex()).Msg("State transition")
		}
	case BlurInputMsg:
		// Blur the input and prevent further typing until UnblurInputMsg is received
		m.textArea.Blur()
		m.inputBlurred = true
		return m, nil
	case UnblurInputMsg:
		// Restore focus to the input
		m.inputBlurred = false
		m.textArea.Focus()
		return m, nil

	case QuitMsg:
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

	case FocusMessageMsg:
		// TODO(manuel, 2024-01-06) This could potentially focus on a previous message
		// and allow us to regenerate.
		cmd = m.textArea.Focus()

		m.scrollToBottom = true
		m.viewport.GotoBottom()

		m.state = StateUserInput
		// Hide highlight in input mode
		m.timelineCtrl.SetSelectionVisible(false)
		m.updateKeyBindings()
		log.Debug().Str("component", "chat").Str("transition", "moving-around->user-input").Msg("State transition")

	case SelectNextMessageMsg:
		m.timelineCtrl.SelectNext()
		m.scrollToSelected()

	case SelectPrevMessageMsg:
		m.timelineCtrl.SelectPrev()
		m.scrollToBottom = false
		m.scrollToSelected()

	case SubmitMessageMsg:
		if m.state == StateStreamCompletion {
			// Ignore submits while streaming
			break
		}
		log.Debug().Str("component", "chat").Msg("SubmitMessageMsg received - calling submit()")
		cmd = m.submit()

	case CopyToClipboardMsg:
		// If selecting, request copy from selected entity model; else copy last assistant reply
		if m.state == StateMovingAround {
			cmd = func() tea.Msg { return timeline.EntityCopyTextMsg{} }
		} else {
			if _, props, ok := m.timelineCtrl.GetLastLLMByRole("assistant"); ok {
				if txt, _ := props["text"].(string); txt != "" {
					return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: txt} }
				}
			}
		}

	case CopyLastResponseToClipboardMsg:
		if _, props, ok := m.timelineCtrl.GetLastLLMByRole("assistant"); ok {
			if txt, _ := props["text"].(string); txt != "" {
				return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: txt} }
			}
		}

	case CopyLastSourceBlocksToClipboardMsg:
		// Future: extract from last assistant text

	case CopySourceBlocksToClipboardMsg:
		// Future: entity models can implement code extraction

	case SaveToFileMsg:
		m.state = StateSavingToFile
		cmd = m.filepicker.Init()
		m.recomputeSize()
		m.updateKeyBindings()

	case CancelCompletionMsg:
		if m.state == StateStreamCompletion {
			m.backend.Interrupt()
		}

	case DismissErrorMsg:
		m.err = nil
		m.state = StateUserInput
		m.updateKeyBindings()

	case ReplaceInputTextMsg:
		m.replaceInputText(msg_.Text)
		m.state = StateUserInput
		m.updateKeyBindings()
		m.recomputeSize()
		return m, nil

	case AppendInputTextMsg:
		m.appendInputText(msg_.Text)
		m.state = StateUserInput
		m.updateKeyBindings()
		m.recomputeSize()
		return m, nil

	case PrependInputTextMsg:
		m.prependInputText(msg_.Text)
		m.state = StateUserInput
		m.updateKeyBindings()
		m.recomputeSize()
		return m, nil

	case GetInputTextMsg:
		// This should be handled in the UserBackend, not here
		// But we'll return the current input text just in case
		return m, func() tea.Msg {
			return m.getInputText()
		}

	}

	return m, cmd
}

// toLLMInferenceData converts Geppetto EventMetadata into unified LLMInferenceData for UI/storage.
func toLLMInferenceData(em *geppetto_events.EventMetadata) *geppetto_events.LLMInferenceData {
	if em == nil {
		return nil
	}
	out := &geppetto_events.LLMInferenceData{
		Model:       em.Model,
		Temperature: em.Temperature,
		TopP:        em.TopP,
		MaxTokens:   em.MaxTokens,
		StopReason:  em.StopReason,
		DurationMs:  em.DurationMs,
		Usage:       em.Usage,
	}
	return out
}

// Add these new methods to the model struct

func (m *model) replaceInputText(text string) {
	m.textArea.SetValue(text)
}

func (m *model) appendInputText(text string) {
	currentText := m.textArea.Value()
	m.textArea.SetValue(currentText + text)
}

func (m *model) prependInputText(text string) {
	currentText := m.textArea.Value()
	m.textArea.SetValue(text + currentText)
}

func (m *model) getInputText() string {
	return m.textArea.Value()
}

func (m *model) GetUIState() map[string]interface{} {
	if m.status != nil {
		return map[string]interface{}{
			"state":        m.status.State,
			"inputText":    m.status.InputText,
			"selectedIdx":  m.status.SelectedIdx,
			"messageCount": m.status.MessageCount,
			"error":        m.status.Error,
		}
	}
	return nil
}
