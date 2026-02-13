package repl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	renderers "github.com/go-go-golems/bobatea/pkg/timeline/renderers"
	"github.com/rs/zerolog/log"
)

// Model is a timeline-first REPL shell: timeline transcript + input line.
type Model struct {
	evaluator Evaluator
	config    Config
	styles    Styles

	// input & history
	history   *History
	textInput textinput.Model
	multiline bool
	lines     []string

	// layout
	width, height int

	// timeline shell (viewport + controller)
	reg   *timeline.Registry
	sh    *timeline.Shell
	focus string // "input" or "timeline"

	// bus publisher
	pub     message.Publisher
	turnSeq int

	// refresh scheduling
	refreshPending   bool
	refreshScheduled bool

	// optional autocomplete capability
	completer             InputCompleter
	completionReqSeq      uint64
	completionDebounce    time.Duration
	completionReqTimeout  time.Duration
	completionTriggerKeys map[string]struct{}
	completionLastResult  CompletionResult
	completionLastError   error
	completionLastReqID   uint64
	completionLastReqKind CompletionReason
}

// NewModel constructs a new REPL shell with timeline transcript.
func NewModel(evaluator Evaluator, config Config, pub message.Publisher) *Model {
	if config.Prompt == "" {
		config.Prompt = evaluator.GetPrompt()
	}
	ti := textinput.New()
	ti.Prompt = config.Prompt
	ti.Placeholder = config.Placeholder
	ti.Focus()
	ti.Width = max(10, config.Width-10)

	reg := timeline.NewRegistry()
	// Register base widgets
	reg.RegisterModelFactory(renderers.TextFactory{})
	reg.RegisterModelFactory(renderers.NewMarkdownFactory())
	reg.RegisterModelFactory(renderers.StructuredDataFactory{})
	reg.RegisterModelFactory(renderers.LogEventFactory{})
	reg.RegisterModelFactory(renderers.StructuredLogEventFactory{})

	sh := timeline.NewShell(reg)

	var completer InputCompleter
	if c, ok := evaluator.(InputCompleter); ok {
		completer = c
	}

	return &Model{
		evaluator: evaluator,
		config:    config,
		styles:    DefaultStyles(),
		history:   NewHistory(config.MaxHistorySize),
		textInput: ti,
		multiline: config.StartMultiline,
		lines:     []string{},
		width:     config.Width,
		reg:       reg,
		sh:        sh,
		focus:     "input",
		pub:       pub,
		completer: completer,

		// These become configurable in a later task.
		completionDebounce:   120 * time.Millisecond,
		completionReqTimeout: 400 * time.Millisecond,
		completionTriggerKeys: map[string]struct{}{
			"tab": {},
		},
	}
}

// Init subscribes to evaluator events.
func (m *Model) Init() tea.Cmd {
	// no blinking on text input, because it makes copy paste impossible
	return tea.Batch(m.sh.Init())
}

// Update handles TUI events.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Trace().Interface("msg", msg).Interface("type", fmt.Sprintf("%T", msg)).Msg("updating repl model")
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
		m.textInput.Width = max(10, v.Width-10)
		// give most of the space to timeline shell viewport
		tlHeight := max(0, v.Height-4)
		m.sh.SetSize(v.Width, tlHeight)
		// initial refresh to fit new size
		m.sh.RefreshView(false)
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch m.focus {
		case "input":
			return m.updateInput(v)
		case "timeline":
			return m.updateTimeline(v)
		}

	case timeline.UIEntityCreated:
		m.ctrl().OnCreated(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timeline.UIEntityUpdated:
		m.ctrl().OnUpdated(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timeline.UIEntityCompleted:
		m.ctrl().OnCompleted(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timeline.UIEntityDeleted:
		m.ctrl().OnDeleted(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timelineRefreshMsg:
		m.refreshScheduled = false
		if m.refreshPending {
			m.sh.RefreshView(true)
			m.refreshPending = false
		}
		return m, nil

	case completionDebounceMsg:
		return m, m.handleDebouncedCompletion(v)
	case completionResultMsg:
		return m, m.handleCompletionResult(v)

	case cursor.BlinkMsg:
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	log.Trace().Interface("msg", msg).Msg("updating repl model default case")

	return m, nil
}

func (m *Model) updateInput(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Trace().Interface("k", k).Str("key", k.String()).Msg("updating input")
	prevValue := m.textInput.Value()
	prevCursor := m.textInput.Position()

	//nolint:exhaustive
	switch k.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	if cmd := m.triggerCompletionFromShortcut(k.String()); cmd != nil {
		return m, cmd
	}

	switch k.String() {
	case "tab":
		m.focus = "timeline"
		m.textInput.Blur()
		m.sh.SetSelectionVisible(true)
		return m, nil
	case "enter":
		input := m.textInput.Value()
		if strings.TrimSpace(input) == "" {
			return m, nil
		}
		m.textInput.Reset()
		if m.config.EnableHistory {
			m.history.Add(input, "", false)
			m.history.ResetNavigation()
		}
		return m, m.submit(input)
	case "up":
		if m.config.EnableHistory {
			if entry := m.history.NavigateUp(); entry != "" {
				m.textInput.SetValue(entry)
			}
		}
		return m, m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor)
	case "down":
		if m.config.EnableHistory {
			entry := m.history.NavigateDown()
			m.textInput.SetValue(entry)
		}
		return m, m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor)
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(k)
	return m, tea.Batch(cmd, m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor))
}

func (m *Model) updateTimeline(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "tab":
		m.focus = "input"
		m.textInput.Focus()
		m.sh.SetSelectionVisible(false)
		return m, nil
	case "up":
		m.sh.SelectPrev()
		return m, nil
	case "down":
		m.sh.SelectNext()
		return m, nil
	case "enter":
		if m.sh.IsEntering() {
			m.sh.ExitSelection()
		} else {
			m.sh.EnterSelection()
		}
		return m, nil
	case "c":
		return m, m.sh.SendToSelected(timeline.EntityCopyCodeMsg{})
	case "y":
		return m, m.sh.SendToSelected(timeline.EntityCopyTextMsg{})
	}
	// route keys to shell/controller (e.g., Tab cycles inside entity)
	cmd := m.sh.HandleMsg(k)
	return m, cmd
}

func (m *Model) View() string {
	var b strings.Builder
	title := m.config.Title
	if title == "" {
		title = fmt.Sprintf("%s REPL", m.evaluator.GetName())
	}
	b.WriteString(m.styles.Title.Render(" " + title + " "))
	b.WriteString("\n\n")
	// timeline view (viewport-wrapped)
	b.WriteString(m.sh.View())
	b.WriteString("\n")
	// input (dim when in selection mode)
	inputView := m.textInput.View()
	if m.focus == "timeline" {
		inputView = m.styles.HelpText.Render(inputView)
	}
	b.WriteString(inputView)
	b.WriteString("\n")
	// help
	help := "TAB: switch focus | Enter: submit | Up/Down: history/selection | c: copy code | y: copy text | Ctrl+C: quit"
	b.WriteString(m.styles.HelpText.Render(help))
	b.WriteString("\n")
	return b.String()
}

// submit runs evaluation and streams events to m.events
func (m *Model) submit(code string) tea.Cmd {
	return func() tea.Msg {
		turnID := newTurnID(m.turnSeq)
		m.turnSeq++
		// Create input entity directly on UI bus to guarantee ordering and avoid extra newlines
		_ = m.publishUIEntityCreated(turnID, timeline.EntityID{TurnID: turnID, LocalID: "input", Kind: "text"}, timeline.RendererDescriptor{Kind: "text"}, map[string]any{"text": code})
		// Optionally still publish the semantic input event to repl.events? We skip to avoid duplicate UI entities.
		_ = m.evaluator.EvaluateStream(context.Background(), code, func(e Event) {
			log.Trace().Str("turn_id", turnID).Interface("event", e).Msg("publishing repl event")
			_ = m.publishReplEvent(turnID, e)
		})
		return nil
	}
}

func (m *Model) publishReplEvent(turnID string, e Event) error {
	payload, _ := json.Marshal(struct {
		TurnID string    `json:"turn_id"`
		Event  Event     `json:"event"`
		Time   time.Time `json:"time"`
	}{TurnID: turnID, Event: e, Time: time.Now()})
	log.Trace().Str("turn_id", turnID).Interface("event", e).Msg("publishing repl event")
	return m.pub.Publish(eventbus.TopicReplEvents, message.NewMessage(watermill.NewUUID(), payload))
}

func (m *Model) publishUIEntityCreated(turnID string, id timeline.EntityID, rd timeline.RendererDescriptor, props map[string]any) error {
	// Envelope must match timeline.RegisterUIForwarder expectations
	created := timeline.UIEntityCreated{ID: id, Renderer: rd, Props: props, StartedAt: time.Now()}
	b, _ := json.Marshal(created)
	env, _ := json.Marshal(struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}{Type: "timeline.created", Payload: b})
	return m.pub.Publish(eventbus.TopicUIEntities, message.NewMessage(watermill.NewUUID(), env))
}

func ensureAppendPatch(props map[string]any) map[string]any {
	if props == nil {
		return map[string]any{}
	}
	if _, ok := props["append"]; ok {
		return props
	}
	if s, ok := props["text"].(string); ok {
		return map[string]any{"append": s}
	}
	return props
}

// internal helpers
type timelineRefreshMsg struct{}
type completionDebounceMsg struct {
	RequestID uint64
}

type completionResultMsg struct {
	RequestID uint64
	Result    CompletionResult
	Err       error
}

func (m *Model) scheduleRefresh() tea.Cmd {
	if m.refreshScheduled {
		return nil
	}
	m.refreshScheduled = true
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg { return timelineRefreshMsg{} })
}

func (m *Model) scheduleDebouncedCompletionIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	if m.completer == nil {
		return nil
	}
	if prevValue == m.textInput.Value() && prevCursor == m.textInput.Position() {
		return nil
	}

	m.completionReqSeq++
	reqID := m.completionReqSeq
	return tea.Tick(m.completionDebounce, func(time.Time) tea.Msg {
		return completionDebounceMsg{RequestID: reqID}
	})
}

func (m *Model) handleDebouncedCompletion(msg completionDebounceMsg) tea.Cmd {
	if m.completer == nil {
		return nil
	}
	if msg.RequestID != m.completionReqSeq {
		return nil
	}

	req := CompletionRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     CompletionReasonDebounce,
		RequestID:  msg.RequestID,
	}
	m.completionLastReqID = req.RequestID
	m.completionLastReqKind = req.Reason
	return m.completionCmd(req)
}

func (m *Model) triggerCompletionFromShortcut(key string) tea.Cmd {
	if m.completer == nil {
		return nil
	}
	if _, ok := m.completionTriggerKeys[key]; !ok {
		return nil
	}

	m.completionReqSeq++
	req := CompletionRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     CompletionReasonShortcut,
		Shortcut:   key,
		RequestID:  m.completionReqSeq,
	}
	m.completionLastReqID = req.RequestID
	m.completionLastReqKind = req.Reason
	return m.completionCmd(req)
}

func (m *Model) completionCmd(req CompletionRequest) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.completionReqTimeout)
		defer cancel()

		result, err := m.completer.CompleteInput(ctx, req)
		return completionResultMsg{
			RequestID: req.RequestID,
			Result:    result,
			Err:       err,
		}
	}
}

func (m *Model) handleCompletionResult(msg completionResultMsg) tea.Cmd {
	if msg.RequestID != m.completionReqSeq {
		return nil
	}
	m.completionLastReqID = msg.RequestID
	m.completionLastResult = msg.Result
	m.completionLastError = msg.Err
	return nil
}

func (m *Model) ctrl() *timeline.Controller { return m.sh.Controller() }

func newTurnID(seq int) string {
	return timeNow().Format("20060102-150405.000000000") + ":" + fmt.Sprintf("%d", seq)
}

func timeNow() time.Time { return time.Now() }
