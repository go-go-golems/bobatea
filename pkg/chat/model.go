package chat

import (
	context2 "context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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

	scrollToBottom bool

	// not really what we want, but use this for now, we'll have to either find a normal text box,
	// or implement wrapping ourselves.
	textArea textarea.Model

	filepicker filepicker.Model

	// conversation conversationui.Model // removed in favor of timeline selection

	// Timeline controller replaces conversation view rendering
	timelineReg *timeline.Registry
	timelineSh  *timeline.Shell
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

	// externalInput, when true, hides the internal input widget and expects the host
	// to drive input via control messages (Replace/Append/Prepend/Submit). This allows
	// embedding the chat timeline in apps with their own input UX.
	externalInput bool
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

// WithExternalInput turns the chat model into a reusable timeline shell by hiding the
// built-in input widget. The host should control input text and submission via messages.
func WithExternalInput(enabled bool) ModelOption {
	return func(m *model) {
		m.externalInput = enabled
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
	// NOTE: backend-specific renderers can be registered externally via WithTimelineRegister
	ret.timelineReg.RegisterModelFactory(renderers.PlainFactory{})
	if ret.timelineRegHook != nil {
		ret.timelineRegHook(ret.timelineReg)
	}
	ret.timelineSh = timeline.NewShell(ret.timelineReg)
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

	cmds = append(cmds, m.filepicker.Init(), m.timelineSh.Init())

	// Seed existing chat messages as timeline entities
	// Seeding from conversation is disabled; timeline should be sourced from entity events

	// Set initial timeline view content
	m.timelineSh.GotoBottom()

	m.updateKeyBindings()
	// Select last entity if any
	m.timelineSh.SelectLast()

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
			if !m.externalInput {
				m.textArea, cmd = m.textArea.Update(msg)
			}
		case StateSavingToFile:
			var updatedModel tea.Model
			updatedModel, cmd = m.filepicker.Update(msg)
			m.filepicker = updatedModel.(filepicker.Model)
		case StateMovingAround, StateStreamCompletion, StateError:
			prevAtBottom := m.timelineSh.AtBottom()
			cmd = m.timelineSh.UpdateViewport(msg)
			if m.timelineSh.AtBottom() && !prevAtBottom {
				m.scrollToBottom = false
			}
		}
	}

	return m, cmd
}

func (m model) saveToFile(path string) (tea.Model, tea.Cmd) {
	// No conversation manager; writing viewport content as a simple fallback
	// In a real backend, this should request an export from the backend.
	content := m.timelineSh.View()
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
				m.timelineSh.EnterSelection()
				v := m.timelineSh.View()
				log.Debug().Str("component", "chat").Str("when", "enter_selection").Int("view_len", len(v)).Msg("SetContent")
				return m, nil
			case "esc":
				if m.timelineSh.IsEntering() {
					m.timelineSh.ExitSelection()
					v := m.timelineSh.View()
					log.Debug().Str("component", "chat").Str("when", "exit_entering").Int("view_len", len(v)).Msg("SetContent")
					return m, nil
				}
				// leave moving-around back to text mode
				log.Debug().Str("component", "chat").Str("transition", "moving-around->user-input").Msg("State transition")
				m.state = StateUserInput
				m.textArea.Focus()
				m.updateKeyBindings()
				// hide selection highlight and unselect entity
				m.timelineSh.SetSelectionVisible(false)
				m.timelineSh.Unselect()
				v := m.timelineSh.View()
				log.Debug().Str("component", "chat").Str("when", "esc_to_input").Int("view_len", len(v)).Msg("SetContent")
				return m, nil
			}
			// While entering, allow scrolling keys to control the timeline viewport
			if m.timelineSh.IsEntering() {
				if key.Matches(msg_, m.keyMap.ScrollDown) || msg_.String() == "down" {
					log.Debug().Str("component", "chat_model").Str("op", "scroll_down_entering").Str("key", msg_.String()).Msg("pass-through to timeline shell")
					m.timelineSh.ScrollDown(1)
					m.scrollToBottom = false
					_ = m.timelineSh.View()
					return m, nil
				}
				if key.Matches(msg_, m.keyMap.ScrollUp) || msg_.String() == "up" {
					log.Debug().Str("component", "chat_model").Str("op", "scroll_up_entering").Str("key", msg_.String()).Msg("pass-through to timeline shell")
					m.timelineSh.ScrollUp(1)
					m.scrollToBottom = false
					_ = m.timelineSh.View()
					return m, nil
				}
				logger.Debug().Str("route", "entering").Str("key", msg_.String()).Msg("Routing key to selected entity model")
				cmd := m.timelineSh.HandleMsg(msg_)
				v := m.timelineSh.View()
				log.Debug().Str("component", "chat").Str("when", "entering_route_key").Int("view_len", len(v)).Msg("SetContent")
				return m, cmd
			}
			// Route TAB / shift+TAB to the selected entity even when not entering
			if msg_.String() == "tab" || msg_.String() == "shift+tab" {
				logger.Debug().Str("route", "non-entering").Str("key", msg_.String()).Msg("Routing TAB to selected entity model")
				cmd := m.timelineSh.HandleMsg(msg_)
				v := m.timelineSh.View()
				log.Debug().Str("component", "chat").Str("when", "non_entering_route_tab").Int("view_len", len(v)).Msg("SetContent")
				return m, cmd
			}
			// Allow entities to react to copy requests even when not entering
			if msg_.String() == "alt+c" {
				cmd := m.timelineSh.SendToSelected(timeline.EntityCopyTextMsg{})
				v := m.timelineSh.View()
				log.Debug().Str("component", "chat").Str("when", "copy_selected").Int("view_len", len(v)).Msg("SetContent")
				if cmd != nil {
					return m, cmd
				}
			}
			if msg_.String() == "alt+d" {
				cmd := m.timelineSh.SendToSelected(timeline.EntityCopyCodeMsg{})
				v := m.timelineSh.View()
				log.Debug().Str("component", "chat").Str("when", "copy_selected_code").Int("view_len", len(v)).Msg("SetContent")
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
		m.timelineSh.SetSize(m.width, m.height)
		m.recomputeSize()

	case ErrorMsg:
		logger.Trace().Str("error", msg_.Error()).Msg("Error message received")
		m.err = msg_
		m.state = StateError
		m.updateKeyBindings()
		return m, nil

	case BackendFinishedMsg:
		logger.Trace().Msg("Backend finished - calling finishCompletion()")
		cmd = m.finishCompletion()
		cmds = append(cmds, cmd)

		// Accept external timeline lifecycle messages (e.g., from backend simulating agent tool calls)
	case timeline.UIEntityCreated:
		logger.Debug().Str("lifecycle", "created").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Msg("Applying external entity event")
		m.timelineSh.OnCreated(msg_)
		if m.scrollToBottom {
			m.timelineSh.GotoBottom()
		}
		return m, nil
	case timeline.UIEntityUpdated:
		logger.Debug().Str("lifecycle", "updated").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Int64("version", msg_.Version).Msg("Applying external entity event")
		m.timelineSh.OnUpdated(msg_)
		if m.scrollToBottom {
			m.timelineSh.GotoBottom()
		}
		return m, nil
	case timeline.UIEntityCompleted:
		logger.Debug().Str("lifecycle", "completed").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Msg("Applying external entity event")
		m.timelineSh.OnCompleted(msg_)
		if m.scrollToBottom {
			m.timelineSh.GotoBottom()
		}
		return m, nil
	case timeline.UIEntityDeleted:
		logger.Debug().Str("lifecycle", "deleted").Str("kind", msg_.ID.Kind).Str("local_id", msg_.ID.LocalID).Msg("Applying external entity event")
		m.timelineSh.OnDeleted(msg_)
		if m.scrollToBottom {
			m.timelineSh.GotoBottom()
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

		m.recomputeSize()
		if msg_.GoToBottom || m.scrollToBottom {
			m.timelineSh.GotoBottom()
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
			cmd = m.timelineSh.UpdateViewport(msg_)
			if cmd != nil {
				logger.Trace().Str("viewport_cmd_type", fmt.Sprintf("%T", cmd)).Msg("Shell viewport returned command")
			}
			cmds = append(cmds, cmd)
		case StateMovingAround:
			// In moving-around mode, use timeline selection controls and scroll
			if km, ok := msg_.(tea.KeyMsg); ok {
				switch {
				case key.Matches(km, m.keyMap.SelectNextMessage):
					m.timelineSh.SelectNext()
					m.scrollToSelected()
				case key.Matches(km, m.keyMap.SelectPrevMessage):
					m.timelineSh.SelectPrev()
					m.scrollToSelected()
				case key.Matches(km, m.keyMap.ScrollDown) || km.String() == "down":
					m.timelineSh.ScrollDown(1)
				case key.Matches(km, m.keyMap.ScrollUp) || km.String() == "up":
					m.timelineSh.ScrollUp(1)
				}
			}
			// Allow non-key messages (e.g., mouse wheel) to reach the shell viewport
			m.timelineSh.UpdateViewport(msg_)
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
		m.status.SelectedIdx = m.timelineSh.SelectedIndex()
		// Fallback approximation using rendered height if entity count is not available
		m.status.MessageCount = lipgloss.Height(m.timelineSh.View())
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
func (m *model) scrollToSelected() { m.timelineSh.ScrollToSelected() }

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

	newHeight := m.height - headerHeight - helpViewHeight
	if !m.externalInput {
		newHeight = m.height - textAreaHeight - headerHeight - helpViewHeight
	}
	if newHeight < 0 {
		newHeight = 0
	}

	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Dur("textarea_duration", textAreaDuration).
		Int("textarea_height", textAreaHeight).
		Int("calculated_viewport_height", newHeight).
		Msg("Text area computed, viewport height calculated")

	// Update shell (timeline) dimensions and position
	m.timelineSh.SetSize(m.width, newHeight)

	h, _ := m.style.SelectedMessage.GetFrameSize()
	m.textArea.SetWidth(m.width - h)

	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Int("textarea_width", m.width-h).
		Msg("Component dimensions updated")

	// CRITICAL: Regenerate timeline view and set content
	v := m.timelineSh.View()
	log.Debug().Str("component", "chat").Str("when", "recompute_size").Int("view_len", len(v)).Msg("SetContent")
	m.timelineSh.GotoBottom()
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

	if m.externalInput {
		return "" // host renders its own input; we render nothing here
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

	// Generate timeline view via shell (no outer viewport wrapping)
	viewportViewStart := time.Now()
	viewportView := m.timelineSh.View()
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
		if m.externalInput {
			ret += viewportView + "\n" + helpView
		} else {
			ret += viewportView + "\n" + textAreaView + "\n" + helpView
		}
		vlogger.Trace().Str("combined_state", "viewport+textarea+help").Int("final_length", len(ret)).Msg("Combined view for main states")
	case StateMovingAround:
		// Keep input visible (greyed) while selecting entities; if external, omit
		if m.externalInput {
			ret += viewportView + "\n" + helpView
		} else {
			ret += viewportView + "\n" + textAreaView + "\n" + helpView
		}
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
	m.timelineSh.GotoBottom()

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

	// Transition to streaming: blur input immediately
	m.state = StateStreamCompletion
	if m.externalInput {
		m.inputBlurred = true
	} else {
		m.textArea.Blur()
	}
	m.updateKeyBindings()

	// Add entity to timeline
	id := uuid.New().String()
	log.Debug().Str("component", "chat").Str("when", "submit").Str("id", id).Msg("Adding user message to timeline")
	m.timelineSh.OnCreated(timeline.UIEntityCreated{
		ID:       timeline.EntityID{LocalID: id, Kind: "llm_text"},
		Renderer: timeline.RendererDescriptor{Kind: "llm_text"},
		Props:    map[string]any{"role": "user", "text": userMessage},
	})
	m.timelineSh.OnCompleted(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: id, Kind: "llm_text"}})
	log.Debug().Str("component", "chat").Str("when", "submit").Str("id", id).Msg("User message added to timeline")

	if !m.externalInput {
		m.textArea.SetValue("")
	}

	refreshCmd := func() tea.Msg {
		log.Debug().
			Int64("submit_call_id", submitCallID).
			Msg("REFRESH COMMAND EXECUTED - POTENTIAL LOOP TRIGGER")
		return refreshMessageMsg{GoToBottom: true}
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
		m.inputBlurred = false
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
			m.timelineSh.SelectLast()
			m.timelineSh.SetSelectionVisible(true)
			v := m.timelineSh.View()
			log.Debug().Str("component", "chat").Str("when", "handleUserAction_search").Int("view_len", len(v)).Msg("SetContent")
			m.updateKeyBindings()
			log.Debug().Str("component", "chat").Str("transition", "user-input->moving-around").Int("selected_index", m.timelineSh.SelectedIndex()).Msg("State transition")
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
		m.timelineSh.GotoBottom()

		m.state = StateUserInput
		// Hide highlight in input mode
		m.timelineSh.SetSelectionVisible(false)
		m.updateKeyBindings()
		log.Debug().Str("component", "chat").Str("transition", "moving-around->user-input").Msg("State transition")

	case SelectNextMessageMsg:
		m.timelineSh.SelectNext()
		m.scrollToSelected()

	case SelectPrevMessageMsg:
		m.timelineSh.SelectPrev()
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
			if _, props, ok := m.timelineSh.GetLastLLMByRole("assistant"); ok {
				if txt, _ := props["text"].(string); txt != "" {
					return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: txt} }
				}
			}
		}

	case CopyLastResponseToClipboardMsg:
		if _, props, ok := m.timelineSh.GetLastLLMByRole("assistant"); ok {
			if txt, _ := props["text"].(string); txt != "" {
				return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: txt} }
			}
		}

	case CopyLastSourceBlocksToClipboardMsg:
		// Future: extract from last assistant text

	case CopySourceBlocksToClipboardMsg:
		// Ask the selected timeline entity to provide code blocks
		cmd := m.timelineSh.SendToSelected(timeline.EntityCopyCodeMsg{})
		if cmd != nil {
			return m, cmd
		}

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
