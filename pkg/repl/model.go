package repl

import (
	lipglossv2 "charm.land/lipgloss/v2"
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	mode_keymap "github.com/go-go-golems/bobatea/pkg/mode-keymap"
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
	reg    *timeline.Registry
	sh     *timeline.Shell
	focus  string // "input" or "timeline"
	help   help.Model
	keyMap KeyMap

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
	completionVisible     bool
	completionSelection   int
	completionReplaceFrom int
	completionReplaceTo   int
	completionScrollTop   int
	completionVisibleRows int
	completionMaxVisible  int
	completionPageSize    int
	completionMaxWidth    int
	completionMaxHeight   int
	completionMinWidth    int
	completionMargin      int
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
	autocompleteCfg := normalizeAutocompleteConfig(config.Autocomplete)
	if !autocompleteCfg.Enabled {
		completer = nil
	}

	focusToggleKey := autocompleteCfg.FocusToggleKey
	if focusToggleKey == "" {
		if completer != nil {
			focusToggleKey = "ctrl+t"
		} else {
			focusToggleKey = "tab"
		}
	}

	ret := &Model{
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
		help:      help.New(),
		keyMap:    NewKeyMap(autocompleteCfg, focusToggleKey),
		pub:       pub,
		completer: completer,

		completionDebounce:   autocompleteCfg.Debounce,
		completionReqTimeout: autocompleteCfg.RequestTimeout,
		completionMaxVisible: autocompleteCfg.MaxSuggestions,
		completionPageSize:   autocompleteCfg.OverlayPageSize,
		completionMaxWidth:   autocompleteCfg.OverlayMaxWidth,
		completionMaxHeight:  autocompleteCfg.OverlayMaxHeight,
		completionMinWidth:   autocompleteCfg.OverlayMinWidth,
		completionMargin:     autocompleteCfg.OverlayMargin,
	}
	ret.updateKeyBindings()
	return ret
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
		helpHeight := lipgloss.Height(m.help.View(m.keyMap))
		// reserve room for title, input, and help rows
		tlHeight := max(0, v.Height-helpHeight-4)
		m.sh.SetSize(v.Width, tlHeight)
		// initial refresh to fit new size
		m.sh.RefreshView(false)
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(v, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(v, m.keyMap.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

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

	if handled, cmd := m.handleCompletionNavigation(k); handled {
		return m, cmd
	}

	if cmd := m.triggerCompletionFromShortcut(k); cmd != nil {
		return m, cmd
	}

	switch {
	case key.Matches(k, m.keyMap.ToggleFocus):
		m.focus = "timeline"
		m.textInput.Blur()
		m.sh.SetSelectionVisible(true)
		m.updateKeyBindings()
		return m, nil
	case key.Matches(k, m.keyMap.Submit):
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
	case key.Matches(k, m.keyMap.HistoryPrev):
		if m.config.EnableHistory {
			if entry := m.history.NavigateUp(); entry != "" {
				m.textInput.SetValue(entry)
			}
		}
		return m, m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor)
	case key.Matches(k, m.keyMap.HistoryNext):
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
	switch {
	case key.Matches(k, m.keyMap.ToggleFocus):
		m.focus = "input"
		m.textInput.Focus()
		m.sh.SetSelectionVisible(false)
		m.updateKeyBindings()
		return m, nil
	case key.Matches(k, m.keyMap.TimelinePrev):
		m.sh.SelectPrev()
		return m, nil
	case key.Matches(k, m.keyMap.TimelineNext):
		m.sh.SelectNext()
		return m, nil
	case key.Matches(k, m.keyMap.TimelineEnterExit):
		if m.sh.IsEntering() {
			m.sh.ExitSelection()
		} else {
			m.sh.EnterSelection()
		}
		return m, nil
	case key.Matches(k, m.keyMap.CopyCode):
		return m, m.sh.SendToSelected(timeline.EntityCopyCodeMsg{})
	case key.Matches(k, m.keyMap.CopyText):
		return m, m.sh.SendToSelected(timeline.EntityCopyTextMsg{})
	}
	// route keys to shell/controller (e.g., Tab cycles inside entity)
	cmd := m.sh.HandleMsg(k)
	return m, cmd
}

func (m *Model) View() string {
	title := m.config.Title
	if title == "" {
		title = fmt.Sprintf("%s REPL", m.evaluator.GetName())
	}

	header := m.styles.Title.Render(" " + title + " ")
	timelineView := m.sh.View()

	inputView := m.textInput.View()
	if m.focus == "timeline" {
		inputView = m.styles.HelpText.Render(inputView)
	}

	helpView := m.help.View(m.keyMap)
	baseSections := []string{
		header,
		"",
		timelineView,
		inputView,
		helpView,
	}
	base := lipgloss.JoinVertical(lipgloss.Left, baseSections...)

	popup := m.renderCompletionPopup()
	if popup == "" || m.width <= 0 || m.height <= 0 {
		return base
	}

	inputY := lipgloss.Height(header) + 1 + lipgloss.Height(timelineView)
	popupHeight := lipgloss.Height(popup)
	overlayX := 0
	overlayY := inputY + 1 + m.completionMargin
	if overlayY+popupHeight > m.height {
		overlayY = inputY - popupHeight - m.completionMargin
	}
	overlayY = clampInt(overlayY, 0, max(0, m.height-1))

	comp := lipglossv2.NewCompositor(
		lipglossv2.NewLayer(base).X(0).Y(0).Z(0).ID("repl-base"),
		lipglossv2.NewLayer(popup).X(overlayX).Y(overlayY).Z(20).ID("completion-overlay"),
	)
	canvas := lipglossv2.NewCanvas(max(1, m.width), max(1, m.height))
	canvas.Compose(comp)
	return canvas.Render()
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
	m.hideCompletionPopup()

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

func (m *Model) triggerCompletionFromShortcut(k tea.KeyMsg) tea.Cmd {
	if m.completer == nil {
		return nil
	}
	if !key.Matches(k, m.keyMap.CompletionTrigger) {
		return nil
	}
	keyStr := k.String()

	m.completionReqSeq++
	req := CompletionRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     CompletionReasonShortcut,
		Shortcut:   keyStr,
		RequestID:  m.completionReqSeq,
	}
	m.completionLastReqID = req.RequestID
	m.completionLastReqKind = req.Reason
	return m.completionCmd(req)
}

func (m *Model) completionCmd(req CompletionRequest) tea.Cmd {
	return func() tea.Msg {
		var (
			result    CompletionResult
			err       error
			recovered any
			stack     string
		)

		func() {
			defer func() {
				if r := recover(); r != nil {
					recovered = r
					stack = string(debug.Stack())
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), m.completionReqTimeout)
			defer cancel()

			result, err = m.completer.CompleteInput(ctx, req)
		}()

		if recovered != nil {
			log.Error().
				Interface("panic", recovered).
				Str("stack", stack).
				Uint64("request_id", req.RequestID).
				Msg("input completer panicked")
			return completionResultMsg{
				RequestID: req.RequestID,
				Err:       fmt.Errorf("input completer panic: %v", recovered),
			}
		}

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
	if msg.Err != nil || !msg.Result.Show || len(msg.Result.Suggestions) == 0 {
		m.hideCompletionPopup()
		return nil
	}

	m.completionSelection = 0
	m.completionVisible = true
	m.completionScrollTop = 0
	m.completionVisibleRows = 0
	m.completionReplaceFrom = clampInt(msg.Result.ReplaceFrom, 0, len(m.textInput.Value()))
	m.completionReplaceTo = clampInt(msg.Result.ReplaceTo, m.completionReplaceFrom, len(m.textInput.Value()))
	m.ensureCompletionSelectionVisible()
	return nil
}

func (m *Model) handleCompletionNavigation(k tea.KeyMsg) (bool, tea.Cmd) {
	if !m.completionVisible {
		return false, nil
	}

	suggestions := m.completionLastResult.Suggestions
	if len(suggestions) == 0 {
		m.hideCompletionPopup()
		return false, nil
	}

	switch {
	case key.Matches(k, m.keyMap.CompletionCancel):
		m.hideCompletionPopup()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionPrev):
		if m.completionSelection > 0 {
			m.completionSelection--
		}
		return true, nil
	case key.Matches(k, m.keyMap.CompletionNext):
		if m.completionSelection < len(suggestions)-1 {
			m.completionSelection++
		}
		return true, nil
	case key.Matches(k, m.keyMap.CompletionAccept):
		m.applySelectedCompletion()
		return true, nil
	}
	return false, nil
}

func (m *Model) applySelectedCompletion() {
	suggestions := m.completionLastResult.Suggestions
	if len(suggestions) == 0 || m.completionSelection >= len(suggestions) {
		m.hideCompletionPopup()
		return
	}

	selected := suggestions[m.completionSelection]
	input := m.textInput.Value()
	from := clampInt(m.completionReplaceFrom, 0, len(input))
	to := clampInt(m.completionReplaceTo, from, len(input))
	newInput := input[:from] + selected.Value + input[to:]

	m.textInput.SetValue(newInput)
	m.textInput.SetCursor(from + len(selected.Value))
	m.hideCompletionPopup()
}

func (m *Model) hideCompletionPopup() {
	m.completionVisible = false
	m.completionSelection = 0
	m.completionReplaceFrom = 0
	m.completionReplaceTo = 0
	m.completionScrollTop = 0
	m.completionVisibleRows = 0
}

func (m *Model) renderCompletionPopup() string {
	if !m.completionVisible {
		return ""
	}
	suggestions := m.completionLastResult.Suggestions
	if len(suggestions) == 0 {
		return ""
	}

	var lines []string
	limit := min(len(suggestions), m.completionMaxVisible)
	for i := 0; i < limit; i++ {
		itemText := suggestions[i].DisplayText
		itemStyle := m.styles.CompletionItem
		if i == m.completionSelection {
			itemStyle = m.styles.CompletionSelected
			itemText = "› " + itemText
		} else {
			itemText = "  " + itemText
		}
		lines = append(lines, itemStyle.Render(itemText))
	}
	return m.styles.CompletionPopup.Render(strings.Join(lines, "\n"))
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func (m *Model) completionVisibleLimit() int {
	if m.completionVisibleRows > 0 {
		return max(1, m.completionVisibleRows)
	}
	if m.completionMaxVisible > 0 {
		return m.completionMaxVisible
	}
	return 1
}

func (m *Model) completionPageStep() int {
	if m.completionPageSize > 0 {
		return max(1, m.completionPageSize)
	}
	return m.completionVisibleLimit()
}

func (m *Model) ensureCompletionSelectionVisible() {
	suggestions := m.completionLastResult.Suggestions
	if len(suggestions) == 0 {
		m.completionScrollTop = 0
		return
	}

	m.completionSelection = clampInt(m.completionSelection, 0, len(suggestions)-1)
	limit := m.completionVisibleLimit()
	maxTop := max(0, len(suggestions)-limit)
	m.completionScrollTop = clampInt(m.completionScrollTop, 0, maxTop)
	if m.completionSelection < m.completionScrollTop {
		m.completionScrollTop = m.completionSelection
	}
	visibleEnd := m.completionScrollTop + limit - 1
	if m.completionSelection > visibleEnd {
		m.completionScrollTop = m.completionSelection - limit + 1
	}
	m.completionScrollTop = clampInt(m.completionScrollTop, 0, maxTop)
}

func (m *Model) updateKeyBindings() { mode_keymap.EnableMode(&m.keyMap, m.focus) }

func normalizeAutocompleteConfig(cfg AutocompleteConfig) AutocompleteConfig {
	if cfg.Debounce == 0 &&
		cfg.RequestTimeout == 0 &&
		len(cfg.TriggerKeys) == 0 &&
		len(cfg.AcceptKeys) == 0 &&
		cfg.FocusToggleKey == "" &&
		cfg.MaxSuggestions == 0 &&
		cfg.OverlayMaxWidth == 0 &&
		cfg.OverlayMaxHeight == 0 &&
		cfg.OverlayMinWidth == 0 &&
		cfg.OverlayMargin == 0 &&
		cfg.OverlayPageSize == 0 &&
		!cfg.Enabled {
		return DefaultAutocompleteConfig()
	}

	merged := DefaultAutocompleteConfig()
	merged.Enabled = cfg.Enabled
	if cfg.Debounce > 0 {
		merged.Debounce = cfg.Debounce
	}
	if cfg.RequestTimeout > 0 {
		merged.RequestTimeout = cfg.RequestTimeout
	}
	if len(cfg.TriggerKeys) > 0 {
		merged.TriggerKeys = cfg.TriggerKeys
	}
	if len(cfg.AcceptKeys) > 0 {
		merged.AcceptKeys = cfg.AcceptKeys
	}
	if cfg.FocusToggleKey != "" {
		merged.FocusToggleKey = cfg.FocusToggleKey
	}
	if cfg.MaxSuggestions > 0 {
		merged.MaxSuggestions = cfg.MaxSuggestions
	}
	if cfg.OverlayMaxWidth > 0 {
		merged.OverlayMaxWidth = cfg.OverlayMaxWidth
	}
	if cfg.OverlayMaxHeight > 0 {
		merged.OverlayMaxHeight = cfg.OverlayMaxHeight
	}
	if cfg.OverlayMinWidth > 0 {
		merged.OverlayMinWidth = cfg.OverlayMinWidth
	}
	if cfg.OverlayMargin > 0 {
		merged.OverlayMargin = cfg.OverlayMargin
	}
	if cfg.OverlayPageSize > 0 {
		merged.OverlayPageSize = cfg.OverlayPageSize
	}
	return merged
}

func (m *Model) ctrl() *timeline.Controller { return m.sh.Controller() }

func newTurnID(seq int) string {
	return timeNow().Format("20060102-150405.000000000") + ":" + fmt.Sprintf("%d", seq)
}

func timeNow() time.Time { return time.Now() }
