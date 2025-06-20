package chat

import (
	context2 "context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	geppetto_conversation "github.com/go-go-golems/geppetto/pkg/conversation"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	conversationui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
	"github.com/go-go-golems/bobatea/pkg/filepicker"
	mode_keymap "github.com/go-go-golems/bobatea/pkg/mode-keymap"
	"github.com/go-go-golems/bobatea/pkg/textarea"
	"github.com/go-go-golems/glazed/pkg/helpers/markdown"
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
	conversationManager geppetto_conversation.Manager
	autoStartBackend    bool

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

	status *Status
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

// TODO(manuel, 2024-04-07) Add options to configure filepicker

func InitialModel(manager geppetto_conversation.Manager, backend Backend, options ...ModelOption) model {
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

	if m.autoStartBackend {
		cmds = append(cmds, func() tea.Msg {
			return StartBackendMsg{}
		})
	}

	return tea.Batch(cmds...)
}

func (m *model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keyMap.Help):
		cmd = func() tea.Msg { return ToggleHelpMsg{} }
	case key.Matches(msg, m.keyMap.UnfocusMessage):
		cmd = func() tea.Msg { return UnfocusMessageMsg{} }
	case key.Matches(msg, m.keyMap.Quit):
		cmd = func() tea.Msg { return QuitMsg{} }
	case key.Matches(msg, m.keyMap.FocusMessage):
		cmd = func() tea.Msg { return FocusMessageMsg{} }
	case key.Matches(msg, m.keyMap.SelectNextMessage):
		cmd = func() tea.Msg { return SelectNextMessageMsg{} }
	case key.Matches(msg, m.keyMap.SelectPrevMessage):
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
			return ErrorMsg(err)
		}
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
	
	log.Trace().
		Int64("update_call_id", updateCallID).
		Str("msg_type", msgType).
		Str("current_state", string(m.state)).
		Int("message_count", len(m.conversationManager.GetConversation())).
		Bool("scroll_to_bottom", m.scrollToBottom).
		Bool("backend_finished", m.backend.IsFinished()).
		Time("start_time", updateStartTime).
		Msg("UPDATE ENTRY")

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg_ := msg.(type) {
	case tea.KeyMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("key", msg_.String()).
			Msg("Handling key press")
		return m.handleKeyPress(msg_)

	case tea.WindowSizeMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Int("width", msg_.Width).
			Int("height", msg_.Height).
			Msg("Window size changed")
		m.width = msg_.Width
		m.height = msg_.Height
		m.recomputeSize()

	case ErrorMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("error", msg_.Error()).
			Msg("Error message received")
		m.err = msg_
		return m, nil

	case conversationui.StreamCompletionMsg,
		conversationui.StreamStartMsg,
		conversationui.StreamStatusMsg,
		conversationui.StreamDoneMsg,
		conversationui.StreamCompletionError:
		
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("stream_msg_type", msgType).
			Msg("Stream message received - ENTERING STREAM PROCESSING")
		
		startTime := time.Now()
		
		switch streamMsg := msg.(type) {
		case conversationui.StreamStartMsg:
			log.Trace().
				Int64("update_call_id", updateCallID).
				Str("operation", "stream_start_reception").
				Str("messageID", streamMsg.ID.String()).
				Str("parentID", streamMsg.ParentID.String()).
				Time("timestamp", startTime).
				Int("current_message_count", len(m.conversationManager.GetConversation())).
				Bool("scroll_to_bottom", m.scrollToBottom).
				Msg("StreamStartMsg details")
		case conversationui.StreamCompletionMsg:
			log.Trace().
				Int64("update_call_id", updateCallID).
				Str("operation", "stream_completion_reception").
				Str("messageID", streamMsg.ID.String()).
				Int("delta_length", len(streamMsg.Delta)).
				Int("completion_length", len(streamMsg.Completion)).
				Msg("StreamCompletionMsg details")
		case conversationui.StreamDoneMsg:
			log.Trace().
				Int64("update_call_id", updateCallID).
				Str("operation", "stream_done_reception").
				Str("messageID", streamMsg.ID.String()).
				Msg("StreamDoneMsg details")
		}

		// Update conversation with timing and detailed logging
		log.Trace().
			Int64("update_call_id", updateCallID).
			Msg("CALLING conversation.Update() - POTENTIAL RECURSION POINT")
		
		convUpdateStart := time.Now()
		oldConversationState := m.conversation.SelectedIdx()
		m.conversation, cmd = m.conversation.Update(msg)
		convUpdateDuration := time.Since(convUpdateStart)
		newConversationState := m.conversation.SelectedIdx()
		
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("operation", "conversation_update_complete").
			Dur("duration", convUpdateDuration).
			Int("old_selected_idx", oldConversationState).
			Int("new_selected_idx", newConversationState).
			Bool("cmd_returned", cmd != nil).
			Msg("Conversation update completed")

		if cmd != nil {
			log.Trace().
				Int64("update_call_id", updateCallID).
				Str("cmd_type", fmt.Sprintf("%T", cmd)).
				Msg("Command returned from conversation.Update() - POTENTIAL LOOP SOURCE")
		}

		if m.scrollToBottom {
			log.Trace().
				Int64("update_call_id", updateCallID).
				Msg("SCROLL TO BOTTOM - Starting UI update sequence")
			
			// Log UI update operations with timing
			viewStart := time.Now()
			v, _ := m.conversation.ViewAndSelectedPosition()
			viewDuration := time.Since(viewStart)
			
			setContentStart := time.Now()
			m.viewport.SetContent(v)
			setContentDuration := time.Since(setContentStart)
			
			gotoBottomStart := time.Now()
			m.viewport.GotoBottom()
			gotoBottomDuration := time.Since(gotoBottomStart)
			
			log.Trace().
				Int64("update_call_id", updateCallID).
				Str("operation", "ui_update_timing").
				Dur("view_generation", viewDuration).
				Dur("set_content", setContentDuration).
				Dur("goto_bottom", gotoBottomDuration).
				Int("content_length", len(v)).
				Msg("UI update operations completed")
		}
		
		totalDuration := time.Since(startTime)
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("operation", "stream_message_total").
			Dur("total_duration", totalDuration).
			Msg("Stream message processing completed")
			
		cmds = append(cmds, cmd)

	case BackendFinishedMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Msg("Backend finished - calling finishCompletion()")
		cmd = m.finishCompletion()
		cmds = append(cmds, cmd)

	case refreshMessageMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Bool("go_to_bottom", msg_.GoToBottom).
			Bool("scroll_to_bottom", m.scrollToBottom).
			Msg("REFRESH MESSAGE - POTENTIAL TRIGGER FOR LOOPS")
		
		v, _ := m.conversation.ViewAndSelectedPosition()
		m.viewport.SetContent(v)
		m.recomputeSize()
		if msg_.GoToBottom || m.scrollToBottom {
			m.viewport.GotoBottom()
		}

	case filepicker.SelectFileMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("path", msg_.Path).
			Msg("File selected for saving")
		return m.saveToFile(msg_.Path)

	case filepicker.CancelFilePickerMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Msg("File picker cancelled")
		m.state = StateUserInput
		m.updateKeyBindings()

	case StartBackendMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Msg("Starting backend - POTENTIAL COMMAND GENERATOR")
		return m, m.startBackend()

	case UserActionMsg:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("user_action_type", fmt.Sprintf("%T", msg_)).
			Msg("User action message - calling handleUserAction()")
		return m.handleUserAction(msg_)

	default:
		log.Trace().
			Int64("update_call_id", updateCallID).
			Str("msg_type", msgType).
			Str("state", string(m.state)).
			Msg("DEFAULT CASE - updating viewport or filepicker")
		
		switch m.state {
		case StateUserInput, StateError, StateMovingAround, StateStreamCompletion:
			log.Trace().
				Int64("update_call_id", updateCallID).
				Msg("Updating viewport - POTENTIAL RECURSION POINT")
			m.viewport, cmd = m.viewport.Update(msg_)
			if cmd != nil {
				log.Trace().
					Int64("update_call_id", updateCallID).
					Str("viewport_cmd_type", fmt.Sprintf("%T", cmd)).
					Msg("Viewport returned command")
			}
			cmds = append(cmds, cmd)
		case StateSavingToFile:
			log.Trace().
				Int64("update_call_id", updateCallID).
				Msg("Updating filepicker")
			m.filepicker, cmd = m.filepicker.Update(msg_)
			cmds = append(cmds, cmd)
		}
	}

	// Update status if it's not nil
	if m.status != nil {
		oldMessageCount := m.status.MessageCount
		m.status.State = m.state
		m.status.InputText = m.textArea.Value()
		m.status.SelectedIdx = m.conversation.SelectedIdx()
		m.status.MessageCount = len(m.conversation.Conversation())
		m.status.Error = m.err
		
		if oldMessageCount != m.status.MessageCount {
			log.Trace().
				Int64("update_call_id", updateCallID).
				Int("old_count", oldMessageCount).
				Int("new_count", m.status.MessageCount).
				Msg("MESSAGE COUNT CHANGED")
		}
	}

	updateDuration := time.Since(updateStartTime)
	cmdCount := len(cmds)
	
	// Log all commands being returned
	for i, cmd := range cmds {
		if cmd != nil {
			log.Trace().
				Int64("update_call_id", updateCallID).
				Int("cmd_index", i).
				Str("cmd_type", fmt.Sprintf("%T", cmd)).
				Msg("COMMAND BEING RETURNED - POTENTIAL LOOP SOURCE")
		}
	}
	
	log.Trace().
		Int64("update_call_id", updateCallID).
		Str("msg_type", msgType).
		Dur("total_duration", updateDuration).
		Int("cmd_count", cmdCount).
		Str("final_state", string(m.state)).
		Msg("UPDATE EXIT")

	return m, tea.Batch(cmds...)
}

func (m *model) updateKeyBindings() {
	mode_keymap.EnableMode(&m.keyMap, string(m.state))
}

func (m *model) scrollToSelected() {
	scrollCallID := atomic.AddInt64(&updateCallCounter, 1)
	log.Trace().
		Int64("scroll_call_id", scrollCallID).
		Int("selected_idx", m.conversation.SelectedIdx()).
		Int("viewport_y_offset", m.viewport.YOffset).
		Int("viewport_height", m.viewport.Height).
		Msg("SCROLL TO SELECTED ENTRY - POTENTIAL VIEW TRIGGER")
	
	viewStart := time.Now()
	v, pos := m.conversation.ViewAndSelectedPosition()
	viewDuration := time.Since(viewStart)
	
	log.Trace().
		Int64("scroll_call_id", scrollCallID).
		Dur("view_generation", viewDuration).
		Int("view_length", len(v)).
		Int("pos_offset", pos.Offset).
		Int("pos_height", pos.Height).
		Msg("View generated for scroll calculation")
	
	setContentStart := time.Now()
	m.viewport.SetContent(v)
	setContentDuration := time.Since(setContentStart)
	
	midScreenOffset := m.viewport.YOffset + m.viewport.Height/2
	msgEndOffset := pos.Offset + pos.Height
	bottomOffset := m.viewport.YOffset + m.viewport.Height
	
	if pos.Offset > midScreenOffset && msgEndOffset > bottomOffset {
		newOffset := pos.Offset - max(m.viewport.Height-pos.Height-1, m.viewport.Height/2)
		m.viewport.SetYOffset(newOffset)
		log.Trace().
			Int64("scroll_call_id", scrollCallID).
			Int("new_y_offset", newOffset).
			Msg("Scrolled down to show message")
	} else if pos.Offset < m.viewport.YOffset {
		m.viewport.SetYOffset(pos.Offset)
		log.Trace().
			Int64("scroll_call_id", scrollCallID).
			Int("new_y_offset", pos.Offset).
			Msg("Scrolled up to show message")
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

	// Set conversation width - might trigger internal updates
	convWidthStart := time.Now()
	m.conversation.SetWidth(m.width)
	convWidthDuration := time.Since(convWidthStart)

	// Update viewport dimensions
	m.viewport.Width = m.width
	m.viewport.Height = newHeight
	m.viewport.YPosition = headerHeight + 1

	h, _ := m.style.SelectedMessage.GetFrameSize()
	m.textArea.SetWidth(m.width - h)

	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Dur("conversation_width_duration", convWidthDuration).
		Int("viewport_width", m.viewport.Width).
		Int("viewport_height", m.viewport.Height).
		Int("viewport_y_position", m.viewport.YPosition).
		Int("textarea_width", m.width-h).
		Msg("Component dimensions updated")

	// CRITICAL: This generates a new view and sets content - major cascade risk
	viewStart := time.Now()
	v, _ := m.conversation.ViewAndSelectedPosition()
	viewDuration := time.Since(viewStart)

	setContentStart := time.Now()
	m.viewport.SetContent(v)
	setContentDuration := time.Since(setContentStart)

	gotoBottomStart := time.Now()
	m.viewport.GotoBottom()
	gotoBottomDuration := time.Since(gotoBottomStart)

	log.Trace().
		Int64("recompute_call_id", recomputeCallID).
		Dur("view_generation", viewDuration).
		Dur("set_content", setContentDuration).
		Dur("goto_bottom", gotoBottomDuration).
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
	case StateMovingAround, StateStreamCompletion:
		v = m.style.UnselectedMessage.Render(v)
	case StateError, StateSavingToFile:
	}

	return v
}

func (m model) View() string {
	// Track View calls and timing - CRITICAL FOR DETECTING EXCESSIVE RENDERS
	viewCallID := atomic.AddInt64(&viewCallCounter, 1)
	viewStartTime := time.Now()
	
	log.Trace().
		Int64("view_call_id", viewCallID).
		Str("state", string(m.state)).
		Int("message_count", len(m.conversationManager.GetConversation())).
		Bool("scroll_to_bottom", m.scrollToBottom).
		Time("start_time", viewStartTime).
		Msg("VIEW ENTRY - POTENTIAL EXCESSIVE CALL POINT")

	headerStart := time.Now()
	headerView := m.headerView()
	headerDuration := time.Since(headerStart)

	log.Trace().
		Int64("view_call_id", viewCallID).
		Dur("header_duration", headerDuration).
		Bool("header_empty", headerView == "").
		Msg("Header view generated")

	// CRITICAL: This call might trigger cascading updates
	conversationViewStart := time.Now()
	view, position := m.conversation.ViewAndSelectedPosition()
	conversationViewDuration := time.Since(conversationViewStart)
	
	log.Trace().
		Int64("view_call_id", viewCallID).
		Dur("conversation_view_duration", conversationViewDuration).
		Int("view_length", len(view)).
		Int("position_offset", position.Offset).
		Int("position_height", position.Height).
		Msg("CONVERSATION VIEW GENERATED - EXPENSIVE OPERATION")

	// CRITICAL: This might trigger viewport updates that cause loops
	setContentStart := time.Now()
	m.viewport.SetContent(view)
	setContentDuration := time.Since(setContentStart)
	
	log.Trace().
		Int64("view_call_id", viewCallID).
		Dur("set_content_duration", setContentDuration).
		Int("viewport_width", m.viewport.Width).
		Int("viewport_height", m.viewport.Height).
		Int("viewport_y_position", m.viewport.YPosition).
		Msg("VIEWPORT CONTENT SET - POTENTIAL TRIGGER FOR UPDATES")

	viewportViewStart := time.Now()
	viewportView := m.viewport.View()
	viewportViewDuration := time.Since(viewportViewStart)
	
	textAreaStart := time.Now()
	textAreaView := m.textAreaView()
	textAreaDuration := time.Since(textAreaStart)
	
	helpStart := time.Now()
	helpView := m.help.View(m.keyMap)
	helpDuration := time.Since(helpStart)

	log.Trace().
		Int64("view_call_id", viewCallID).
		Dur("viewport_view_duration", viewportViewDuration).
		Dur("textarea_duration", textAreaDuration).
		Dur("help_duration", helpDuration).
		Msg("UI component views generated")

	// debugging heights with trace logging
	viewportHeight := lipgloss.Height(viewportView)
	textAreaHeight := lipgloss.Height(textAreaView)
	headerHeight := lipgloss.Height(headerView)
	helpViewHeight := lipgloss.Height(helpView)
	
	log.Trace().
		Int64("view_call_id", viewCallID).
		Int("viewport_height", viewportHeight).
		Int("textarea_height", textAreaHeight).
		Int("header_height", headerHeight).
		Int("help_height", helpViewHeight).
		Int("total_calculated_height", viewportHeight+textAreaHeight+headerHeight+helpViewHeight).
		Int("model_height", m.height).
		Msg("Height calculations")

	ret := ""
	if headerView != "" {
		ret = headerView
	}

	switch m.state {
	case StateUserInput, StateError, StateMovingAround, StateStreamCompletion:
		ret += viewportView + "\n" + textAreaView + "\n" + helpView
		log.Trace().
			Int64("view_call_id", viewCallID).
			Str("combined_state", "viewport+textarea+help").
			Int("final_length", len(ret)).
			Msg("Combined view for main states")

	case StateSavingToFile:
		ret += m.filepicker.View()
		log.Trace().
			Int64("view_call_id", viewCallID).
			Str("combined_state", "filepicker").
			Int("final_length", len(ret)).
			Msg("Combined view for file saving state")
	}

	viewDuration := time.Since(viewStartTime)
	log.Trace().
		Int64("view_call_id", viewCallID).
		Dur("total_view_duration", viewDuration).
		Int("final_output_length", len(ret)).
		Msg("VIEW EXIT")

	return ret
}

func (m *model) startBackend() tea.Cmd {
	startCallID := atomic.AddInt64(&updateCallCounter, 1)
	log.Trace().
		Int64("start_call_id", startCallID).
		Str("previous_state", string(m.state)).
		Bool("backend_finished", m.backend.IsFinished()).
		Int("conversation_length", len(m.conversationManager.GetConversation())).
		Msg("START BACKEND ENTRY - MAJOR COMMAND GENERATOR")

	m.state = StateStreamCompletion
	m.updateKeyBindings()

	log.Trace().
		Int64("start_call_id", startCallID).
		Msg("Calling viewport.GotoBottom()")
	m.viewport.GotoBottom()

	refreshCmd := func() tea.Msg {
		log.Trace().
			Int64("start_call_id", startCallID).
			Msg("REFRESH MESSAGE FROM START BACKEND - LOOP RISK")
		return refreshMessageMsg{
			GoToBottom: true,
		}
	}

	backendCmd := func() tea.Msg {
		log.Trace().
			Int64("start_call_id", startCallID).
			Msg("BACKEND START COMMAND EXECUTING")
		ctx := context2.Background()
		cmd, err := m.backend.Start(ctx, m.conversationManager.GetConversation())
		if err != nil {
			log.Trace().
				Int64("start_call_id", startCallID).
				Err(err).
				Msg("Backend start error")
			return ErrorMsg(err)
		}
		log.Trace().
			Int64("start_call_id", startCallID).
			Msg("Backend started successfully, executing returned command")
		return cmd()
	}

	log.Trace().
		Int64("start_call_id", startCallID).
		Msg("START BACKEND EXIT - returning batch of refresh + backend commands")

	return tea.Batch(refreshCmd, backendCmd)
}

func (m *model) submit() tea.Cmd {
	submitCallID := atomic.AddInt64(&updateCallCounter, 1)
	log.Trace().
		Int64("submit_call_id", submitCallID).
		Bool("backend_finished", m.backend.IsFinished()).
		Int("input_length", len(m.textArea.Value())).
		Int("current_message_count", len(m.conversationManager.GetConversation())).
		Msg("SUBMIT ENTRY")

	if !m.backend.IsFinished() {
		log.Trace().
			Int64("submit_call_id", submitCallID).
			Msg("Backend not finished - returning error")
		return func() tea.Msg {
			return ErrorMsg(errors.New("already streaming"))
		}
	}

	userMessage := m.textArea.Value()
	m.conversationManager.AppendMessages(
		geppetto_conversation.NewChatMessage(geppetto_conversation.RoleUser, userMessage))
	m.textArea.SetValue("")

	log.Trace().
		Int64("submit_call_id", submitCallID).
		Str("user_message", userMessage).
		Int("new_message_count", len(m.conversationManager.GetConversation())).
		Msg("User message appended, calling startBackend()")

	return m.startBackend()
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
			m.conversation.SetActive(true)
			if m.scrollToBottom {
				m.conversation.SetSelectedIdx(len(m.conversation.Conversation()) - 1)
			}
			m.scrollToSelected()
			m.updateKeyBindings()
		}

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

		m.conversation.SetActive(false)
		m.state = StateUserInput
		m.updateKeyBindings()

	case SelectNextMessageMsg:
		messages := m.conversation.Conversation()
		if m.conversation.SelectedIdx() < len(messages)-1 {
			m.conversation.SetSelectedIdx(m.conversation.SelectedIdx() + 1)
			m.scrollToSelected()
		} else if m.conversation.SelectedIdx() == len(messages)-1 {
			m.scrollToBottom = true
			m.viewport.GotoBottom()
		}

	case SelectPrevMessageMsg:
		if m.conversation.SelectedIdx() > 0 {
			m.conversation.SetSelectedIdx(m.conversation.SelectedIdx() - 1)
			m.scrollToSelected()
			m.scrollToBottom = false
		}

	case SubmitMessageMsg:
		cmd = m.submit()

	case CopyToClipboardMsg:
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					err := clipboard.WriteAll(msg_.Content.String())
					if err != nil {
						cmd = func() tea.Msg {
							return ErrorMsg(err)
						}
					}
				}
			} else {
				text := ""
				for _, m := range msgs {
					if content, ok := m.Content.(*geppetto_conversation.ChatMessageContent); ok {
						if content.Role == geppetto_conversation.RoleAssistant {
							text += content.Text + "\n"
						}
					}
				}
				err := clipboard.WriteAll(text)
				if err != nil {
					cmd = func() tea.Msg {
						return ErrorMsg(err)
					}
				}
			}
		}

	case CopyLastResponseToClipboardMsg:
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					if content, ok := msg_.Content.(*geppetto_conversation.ChatMessageContent); ok {
						err := clipboard.WriteAll(content.Text)
						if err != nil {
							cmd = func() tea.Msg {
								return ErrorMsg(err)
							}
						}
					}
				}
			} else {
				if m.state == StateUserInput {
					lastMsg := msgs[len(msgs)-1]
					if content, ok := lastMsg.Content.(*geppetto_conversation.ChatMessageContent); ok {
						err := clipboard.WriteAll(content.Text)
						if err != nil {
							cmd = func() tea.Msg {
								return ErrorMsg(err)
							}
						}
					}
				}
			}
		}

	case CopyLastSourceBlocksToClipboardMsg:
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					if content, ok := msg_.Content.(*geppetto_conversation.ChatMessageContent); ok {
						code := markdown.ExtractQuotedBlocks(content.Text, false)
						err := clipboard.WriteAll(strings.Join(code, "\n"))
						if err != nil {
							cmd = func() tea.Msg {
								return ErrorMsg(err)
							}
						}
					}
				}
			} else {
				if m.state == StateUserInput {
					text := ""
					for _, m := range msgs {
						if content, ok := m.Content.(*geppetto_conversation.ChatMessageContent); ok {
							if content.Role == geppetto_conversation.RoleAssistant {
								text += content.Text + "\n"
							}
						}
					}
					code := markdown.ExtractQuotedBlocks(text, false)
					err := clipboard.WriteAll(strings.Join(code, "\n"))
					if err != nil {
						cmd = func() tea.Msg {
							return ErrorMsg(err)
						}
					}
				}
			}
		}

	case CopySourceBlocksToClipboardMsg:
		msgs := m.conversation.Conversation()
		if len(msgs) > 0 {
			if m.state == StateMovingAround {
				selectedIdx := m.conversation.SelectedIdx()
				if selectedIdx < len(msgs) && selectedIdx >= 0 {
					msg_ := msgs[selectedIdx]
					if content, ok := msg_.Content.(*geppetto_conversation.ChatMessageContent); ok {
						code := markdown.ExtractQuotedBlocks(content.Text, false)
						err := clipboard.WriteAll(strings.Join(code, "\n"))
						if err != nil {
							cmd = func() tea.Msg {
								return ErrorMsg(err)
							}
						}
					}
				}
			} else {
				text := ""
				for _, m := range msgs {
					if content, ok := m.Content.(*geppetto_conversation.ChatMessageContent); ok {
						if content.Role == geppetto_conversation.RoleAssistant {
							text += content.Text + "\n"
						}
					}
				}
				code := markdown.ExtractQuotedBlocks(text, false)
				err := clipboard.WriteAll(strings.Join(code, "\n"))
				if err != nil {
					cmd = func() tea.Msg {
						return ErrorMsg(err)
					}
				}
			}
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
		if m.state == StateError {
			m.err = nil
			m.state = StateUserInput
			m.updateKeyBindings()
		}

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
