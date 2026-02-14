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
	"github.com/mattn/go-runewidth"
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
	appCtx  context.Context
	appStop context.CancelFunc

	// refresh scheduling
	refreshPending   bool
	refreshScheduled bool

	completion completionModel
	helpBar    helpBarModel
	helpDrawer helpDrawerModel
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
	var helpBarProvider HelpBarProvider
	if p, ok := evaluator.(HelpBarProvider); ok {
		helpBarProvider = p
	}
	var helpDrawerProvider HelpDrawerProvider
	if p, ok := evaluator.(HelpDrawerProvider); ok {
		helpDrawerProvider = p
	}
	autocompleteCfg := normalizeAutocompleteConfig(config.Autocomplete)
	if !autocompleteCfg.Enabled {
		completer = nil
	}
	helpBarCfg := normalizeHelpBarConfig(config.HelpBar)
	if !helpBarCfg.Enabled {
		helpBarProvider = nil
	}
	helpDrawerCfg := normalizeHelpDrawerConfig(config.HelpDrawer)
	if !helpDrawerCfg.Enabled {
		helpDrawerProvider = nil
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
		keyMap:    NewKeyMap(autocompleteCfg, helpDrawerCfg, focusToggleKey),
		pub:       pub,
		completion: completionModel{
			provider:   completer,
			debounce:   autocompleteCfg.Debounce,
			reqTimeout: autocompleteCfg.RequestTimeout,
			maxVisible: autocompleteCfg.MaxSuggestions,
			pageSize:   autocompleteCfg.OverlayPageSize,
			maxWidth:   autocompleteCfg.OverlayMaxWidth,
			maxHeight:  autocompleteCfg.OverlayMaxHeight,
			minWidth:   autocompleteCfg.OverlayMinWidth,
			margin:     autocompleteCfg.OverlayMargin,
			offsetX:    autocompleteCfg.OverlayOffsetX,
			offsetY:    autocompleteCfg.OverlayOffsetY,
			noBorder:   autocompleteCfg.OverlayNoBorder,
			placement:  autocompleteCfg.OverlayPlacement,
			horizontal: autocompleteCfg.OverlayHorizontalGrow,
		},
		helpBar: helpBarModel{
			provider:   helpBarProvider,
			debounce:   helpBarCfg.Debounce,
			reqTimeout: helpBarCfg.RequestTimeout,
		},
		helpDrawer: helpDrawerModel{
			provider:      helpDrawerProvider,
			debounce:      helpDrawerCfg.Debounce,
			reqTimeout:    helpDrawerCfg.RequestTimeout,
			prefetch:      helpDrawerCfg.PrefetchWhenHidden,
			dock:          helpDrawerCfg.Dock,
			widthPercent:  helpDrawerCfg.WidthPercent,
			heightPercent: helpDrawerCfg.HeightPercent,
			margin:        helpDrawerCfg.Margin,
		},
	}
	ret.appCtx, ret.appStop = context.WithCancel(context.Background())
	if ret.helpDrawer.provider == nil {
		ret.keyMap.HelpDrawerToggle.SetEnabled(false)
		ret.keyMap.HelpDrawerClose.SetEnabled(false)
		ret.keyMap.HelpDrawerRefresh.SetEnabled(false)
	}
	ret.updateKeyBindings()
	return ret
}

// NewModelWithContext constructs a REPL model whose internal app context derives from ctx.
// Passing nil uses context.Background().
func NewModelWithContext(ctx context.Context, evaluator Evaluator, config Config, pub message.Publisher) *Model {
	ret := NewModel(evaluator, config, pub)
	ret.cancelAppContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ret.appCtx, ret.appStop = context.WithCancel(ctx)
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
			m.cancelAppContext()
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
	case helpBarDebounceMsg:
		return m, m.handleDebouncedHelpBar(v)
	case helpBarResultMsg:
		return m, m.handleHelpBarResult(v)
	case helpDrawerDebounceMsg:
		return m, m.handleDebouncedHelpDrawer(v)
	case helpDrawerResultMsg:
		return m, m.handleHelpDrawerResult(v)

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

	if handled, cmd := m.handleHelpDrawerShortcuts(k); handled {
		return m, cmd
	}

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
		m.helpBar.visible = false
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
		return m, tea.Batch(
			m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpBarIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpDrawerIfNeeded(prevValue, prevCursor),
		)
	case key.Matches(k, m.keyMap.HistoryNext):
		if m.config.EnableHistory {
			entry := m.history.NavigateDown()
			m.textInput.SetValue(entry)
		}
		return m, tea.Batch(
			m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpBarIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpDrawerIfNeeded(prevValue, prevCursor),
		)
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(k)
	return m, tea.Batch(
		cmd,
		m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor),
		m.scheduleDebouncedHelpBarIfNeeded(prevValue, prevCursor),
		m.scheduleDebouncedHelpDrawerIfNeeded(prevValue, prevCursor),
	)
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
	}
	if helpBarView := m.renderHelpBar(); helpBarView != "" {
		baseSections = append(baseSections, helpBarView)
	}
	baseSections = append(baseSections, helpView)
	base := lipgloss.JoinVertical(lipgloss.Left, baseSections...)

	if m.width <= 0 || m.height <= 0 {
		return base
	}

	completionLayout, completionOK := m.computeCompletionOverlayLayout(header, timelineView)
	completionPopup := ""
	if completionOK {
		completionPopup = m.renderCompletionPopup(completionLayout)
		if completionPopup == "" {
			m.completion.visibleRows = 0
			completionOK = false
		} else {
			m.completion.visibleRows = completionLayout.VisibleRows
			m.ensureCompletionSelectionVisible()
		}
	} else {
		m.completion.visibleRows = 0
	}

	drawerLayout, drawerOK := m.computeHelpDrawerOverlayLayout(header, timelineView)
	drawerPanel := ""
	if drawerOK {
		drawerPanel = m.renderHelpDrawerPanel(drawerLayout)
		if drawerPanel == "" {
			drawerOK = false
		}
	}

	if !completionOK && !drawerOK {
		return base
	}

	layers := []*lipglossv2.Layer{
		lipglossv2.NewLayer(base).X(0).Y(0).Z(0).ID("repl-base"),
	}
	if drawerOK {
		layers = append(layers,
			lipglossv2.NewLayer(drawerPanel).X(drawerLayout.PanelX).Y(drawerLayout.PanelY).Z(15).ID("help-drawer-overlay"),
		)
	}
	if completionOK {
		layers = append(layers,
			lipglossv2.NewLayer(completionPopup).X(completionLayout.PopupX).Y(completionLayout.PopupY).Z(20).ID("completion-overlay"),
		)
	}

	comp := lipglossv2.NewCompositor(layers...)
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

type helpBarDebounceMsg struct {
	RequestID uint64
}

type helpBarResultMsg struct {
	RequestID uint64
	Payload   HelpBarPayload
	Err       error
}

type helpDrawerDebounceMsg struct {
	RequestID uint64
}

type helpDrawerResultMsg struct {
	RequestID uint64
	Doc       HelpDrawerDocument
	Err       error
}

type completionOverlayLayout struct {
	PopupX       int
	PopupY       int
	PopupWidth   int
	VisibleRows  int
	ContentWidth int
}

type helpDrawerOverlayLayout struct {
	PanelX        int
	PanelY        int
	PanelWidth    int
	PanelHeight   int
	ContentWidth  int
	ContentHeight int
}

func (m *Model) scheduleRefresh() tea.Cmd {
	if m.refreshScheduled {
		return nil
	}
	m.refreshScheduled = true
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg { return timelineRefreshMsg{} })
}

func (m *Model) scheduleDebouncedCompletionIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	if m.completion.provider == nil {
		return nil
	}
	if prevValue == m.textInput.Value() && prevCursor == m.textInput.Position() {
		return nil
	}

	m.completion.reqSeq++
	reqID := m.completion.reqSeq
	return tea.Tick(m.completion.debounce, func(time.Time) tea.Msg {
		return completionDebounceMsg{RequestID: reqID}
	})
}

func (m *Model) scheduleDebouncedHelpBarIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	if m.helpBar.provider == nil {
		return nil
	}
	if prevValue == m.textInput.Value() && prevCursor == m.textInput.Position() {
		return nil
	}

	m.helpBar.reqSeq++
	reqID := m.helpBar.reqSeq
	return tea.Tick(m.helpBar.debounce, func(time.Time) tea.Msg {
		return helpBarDebounceMsg{RequestID: reqID}
	})
}

func (m *Model) scheduleDebouncedHelpDrawerIfNeeded(prevValue string, prevCursor int) tea.Cmd {
	if m.helpDrawer.provider == nil {
		return nil
	}
	if !m.helpDrawer.visible && !m.helpDrawer.prefetch {
		return nil
	}
	if m.helpDrawer.pinned {
		return nil
	}
	if prevValue == m.textInput.Value() && prevCursor == m.textInput.Position() {
		return nil
	}

	m.helpDrawer.reqSeq++
	reqID := m.helpDrawer.reqSeq
	return tea.Tick(m.helpDrawer.debounce, func(time.Time) tea.Msg {
		return helpDrawerDebounceMsg{RequestID: reqID}
	})
}

func (m *Model) handleDebouncedCompletion(msg completionDebounceMsg) tea.Cmd {
	if m.completion.provider == nil {
		return nil
	}
	if msg.RequestID != m.completion.reqSeq {
		return nil
	}

	req := CompletionRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     CompletionReasonDebounce,
		RequestID:  msg.RequestID,
	}
	m.completion.lastReqID = req.RequestID
	m.completion.lastReqKind = req.Reason
	return m.completionCmd(req)
}

func (m *Model) handleDebouncedHelpBar(msg helpBarDebounceMsg) tea.Cmd {
	if m.helpBar.provider == nil {
		return nil
	}
	if msg.RequestID != m.helpBar.reqSeq {
		return nil
	}

	req := HelpBarRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     HelpBarReasonDebounce,
		RequestID:  msg.RequestID,
	}
	m.helpBar.lastReqID = req.RequestID
	m.helpBar.lastReqKind = req.Reason
	return m.helpBarCmd(req)
}

func (m *Model) handleDebouncedHelpDrawer(msg helpDrawerDebounceMsg) tea.Cmd {
	if m.helpDrawer.provider == nil {
		return nil
	}
	if msg.RequestID != m.helpDrawer.reqSeq {
		return nil
	}
	if !m.helpDrawer.visible && !m.helpDrawer.prefetch {
		return nil
	}

	req := HelpDrawerRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		RequestID:  msg.RequestID,
		Trigger:    HelpDrawerTriggerTyping,
	}
	m.helpDrawer.loading = true
	return m.helpDrawerCmd(req)
}

func (m *Model) triggerCompletionFromShortcut(k tea.KeyMsg) tea.Cmd {
	if m.completion.provider == nil {
		return nil
	}
	if !key.Matches(k, m.keyMap.CompletionTrigger) {
		return nil
	}
	keyStr := k.String()

	m.completion.reqSeq++
	req := CompletionRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		Reason:     CompletionReasonShortcut,
		Shortcut:   keyStr,
		RequestID:  m.completion.reqSeq,
	}
	m.completion.lastReqID = req.RequestID
	m.completion.lastReqKind = req.Reason
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

			ctx, cancel := context.WithTimeout(m.appContext(), m.completion.reqTimeout)
			defer cancel()

			result, err = m.completion.provider.CompleteInput(ctx, req)
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

func (m *Model) helpBarCmd(req HelpBarRequest) tea.Cmd {
	return func() tea.Msg {
		var (
			payload   HelpBarPayload
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

			ctx, cancel := context.WithTimeout(m.appContext(), m.helpBar.reqTimeout)
			defer cancel()

			payload, err = m.helpBar.provider.GetHelpBar(ctx, req)
		}()

		if recovered != nil {
			log.Error().
				Interface("panic", recovered).
				Str("stack", stack).
				Uint64("request_id", req.RequestID).
				Msg("help bar provider panicked")
			return helpBarResultMsg{
				RequestID: req.RequestID,
				Err:       fmt.Errorf("help bar provider panic: %v", recovered),
			}
		}

		return helpBarResultMsg{
			RequestID: req.RequestID,
			Payload:   payload,
			Err:       err,
		}
	}
}

func (m *Model) helpDrawerCmd(req HelpDrawerRequest) tea.Cmd {
	return func() tea.Msg {
		var (
			doc       HelpDrawerDocument
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

			ctx, cancel := context.WithTimeout(m.appContext(), m.helpDrawer.reqTimeout)
			defer cancel()

			doc, err = m.helpDrawer.provider.GetHelpDrawer(ctx, req)
		}()

		if recovered != nil {
			log.Error().
				Interface("panic", recovered).
				Str("stack", stack).
				Uint64("request_id", req.RequestID).
				Msg("help drawer provider panicked")
			return helpDrawerResultMsg{
				RequestID: req.RequestID,
				Err:       fmt.Errorf("help drawer provider panic: %v", recovered),
			}
		}

		return helpDrawerResultMsg{
			RequestID: req.RequestID,
			Doc:       doc,
			Err:       err,
		}
	}
}

func (m *Model) handleCompletionResult(msg completionResultMsg) tea.Cmd {
	if msg.RequestID != m.completion.reqSeq {
		return nil
	}
	m.completion.lastReqID = msg.RequestID
	m.completion.lastResult = msg.Result
	m.completion.lastError = msg.Err
	if msg.Err != nil || !msg.Result.Show || len(msg.Result.Suggestions) == 0 {
		m.hideCompletionPopup()
		return nil
	}

	m.completion.selection = 0
	m.completion.visible = true
	m.completion.scrollTop = 0
	m.completion.visibleRows = 0
	m.completion.replaceFrom = clampInt(msg.Result.ReplaceFrom, 0, len(m.textInput.Value()))
	m.completion.replaceTo = clampInt(msg.Result.ReplaceTo, m.completion.replaceFrom, len(m.textInput.Value()))
	m.ensureCompletionSelectionVisible()
	return nil
}

func (m *Model) handleHelpBarResult(msg helpBarResultMsg) tea.Cmd {
	if msg.RequestID != m.helpBar.reqSeq {
		return nil
	}
	m.helpBar.lastReqID = msg.RequestID
	m.helpBar.lastErr = msg.Err
	if msg.Err != nil {
		m.helpBar.visible = false
		return nil
	}
	if !msg.Payload.Show || strings.TrimSpace(msg.Payload.Text) == "" {
		m.helpBar.visible = false
		return nil
	}

	m.helpBar.payload = msg.Payload
	m.helpBar.visible = true
	return nil
}

func (m *Model) handleHelpDrawerResult(msg helpDrawerResultMsg) tea.Cmd {
	if msg.RequestID != m.helpDrawer.reqSeq {
		return nil
	}
	m.helpDrawer.loading = false
	m.helpDrawer.err = msg.Err
	if msg.Err != nil {
		return nil
	}
	m.helpDrawer.doc = msg.Doc
	return nil
}

func (m *Model) renderHelpBar() string {
	if !m.helpBar.visible {
		return ""
	}
	return m.helpBarStyleForSeverity(m.helpBar.payload.Severity).Render(m.helpBar.payload.Text)
}

func (m *Model) helpBarStyleForSeverity(severity string) lipgloss.Style {
	switch severity {
	case "error":
		return m.styles.Error
	case "warning":
		return m.styles.HelpText
	default:
		return m.styles.Info
	}
}

func (m *Model) handleHelpDrawerShortcuts(k tea.KeyMsg) (bool, tea.Cmd) {
	if m.helpDrawer.provider == nil {
		return false, nil
	}

	switch {
	case key.Matches(k, m.keyMap.HelpDrawerToggle):
		return true, m.toggleHelpDrawer()
	case m.helpDrawer.visible && key.Matches(k, m.keyMap.HelpDrawerClose):
		if m.completion.visible && key.Matches(k, m.keyMap.CompletionCancel) {
			return false, nil
		}
		m.closeHelpDrawer()
		return true, nil
	case m.helpDrawer.visible && key.Matches(k, m.keyMap.HelpDrawerRefresh):
		return true, m.requestHelpDrawerNow(HelpDrawerTriggerManualRefresh)
	case m.helpDrawer.visible && key.Matches(k, m.keyMap.HelpDrawerPin):
		m.helpDrawer.pinned = !m.helpDrawer.pinned
		return true, nil
	}

	return false, nil
}

func (m *Model) toggleHelpDrawer() tea.Cmd {
	if m.helpDrawer.visible {
		m.closeHelpDrawer()
		return nil
	}

	m.helpDrawer.visible = true
	m.helpDrawer.err = nil
	return m.requestHelpDrawerNow(HelpDrawerTriggerToggleOpen)
}

func (m *Model) closeHelpDrawer() {
	m.helpDrawer.visible = false
	m.helpDrawer.loading = false
}

func (m *Model) requestHelpDrawerNow(trigger HelpDrawerTrigger) tea.Cmd {
	if m.helpDrawer.provider == nil {
		return nil
	}
	m.helpDrawer.loading = true
	m.helpDrawer.err = nil
	m.helpDrawer.reqSeq++
	req := HelpDrawerRequest{
		Input:      m.textInput.Value(),
		CursorByte: m.textInput.Position(),
		RequestID:  m.helpDrawer.reqSeq,
		Trigger:    trigger,
	}
	return m.helpDrawerCmd(req)
}

func (m *Model) handleCompletionNavigation(k tea.KeyMsg) (bool, tea.Cmd) {
	if !m.completion.visible {
		return false, nil
	}

	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		m.hideCompletionPopup()
		return false, nil
	}

	switch {
	case key.Matches(k, m.keyMap.CompletionCancel):
		m.hideCompletionPopup()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionPrev):
		if m.completion.selection > 0 {
			m.completion.selection--
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionNext):
		if m.completion.selection < len(suggestions)-1 {
			m.completion.selection++
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionPageUp):
		if m.completion.selection > 0 {
			m.completion.selection = max(0, m.completion.selection-m.completionPageStep())
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionPageDown):
		if m.completion.selection < len(suggestions)-1 {
			m.completion.selection = min(len(suggestions)-1, m.completion.selection+m.completionPageStep())
		}
		m.ensureCompletionSelectionVisible()
		return true, nil
	case key.Matches(k, m.keyMap.CompletionAccept):
		m.applySelectedCompletion()
		return true, nil
	}
	return false, nil
}

func (m *Model) applySelectedCompletion() {
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 || m.completion.selection >= len(suggestions) {
		m.hideCompletionPopup()
		return
	}

	selected := suggestions[m.completion.selection]
	input := m.textInput.Value()
	from := clampInt(m.completion.replaceFrom, 0, len(input))
	to := clampInt(m.completion.replaceTo, from, len(input))
	newInput := input[:from] + selected.Value + input[to:]

	m.textInput.SetValue(newInput)
	m.textInput.SetCursor(from + len(selected.Value))
	m.hideCompletionPopup()
}

func (m *Model) hideCompletionPopup() {
	m.completion.visible = false
	m.completion.selection = 0
	m.completion.replaceFrom = 0
	m.completion.replaceTo = 0
	m.completion.scrollTop = 0
	m.completion.visibleRows = 0
}

func (m *Model) computeCompletionOverlayLayout(header, timelineView string) (completionOverlayLayout, bool) {
	if !m.completion.visible || m.width <= 0 || m.height <= 0 {
		return completionOverlayLayout{}, false
	}
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		return completionOverlayLayout{}, false
	}

	inputY := lipgloss.Height(header) + 1 + lipgloss.Height(timelineView)
	popupStyle := m.completionPopupStyle()
	frameWidth := popupStyle.GetHorizontalFrameSize()
	frameHeight := popupStyle.GetVerticalFrameSize()

	contentWidth := 1
	for _, suggestion := range suggestions {
		w := runewidth.StringWidth("  " + suggestion.DisplayText)
		if w > contentWidth {
			contentWidth = w
		}
	}

	popupWidth := contentWidth + frameWidth
	if m.completion.minWidth > 0 {
		popupWidth = max(popupWidth, m.completion.minWidth)
	}
	if m.completion.maxWidth > 0 {
		popupWidth = min(popupWidth, m.completion.maxWidth)
	}
	popupWidth = min(popupWidth, m.width)
	contentWidth = max(1, popupWidth-frameWidth)

	desiredRows := len(suggestions)
	if m.completion.maxVisible > 0 {
		desiredRows = min(desiredRows, m.completion.maxVisible)
	}
	maxHeight := m.completion.maxHeight
	if maxHeight <= 0 {
		maxHeight = m.height
	}
	maxHeight = min(maxHeight, m.height)
	maxRowsByConfig := max(1, maxHeight-frameHeight)
	desiredRows = min(desiredRows, maxRowsByConfig)
	if desiredRows <= 0 {
		return completionOverlayLayout{}, false
	}

	margin := max(0, m.completion.margin)
	availableBelow := max(0, m.height-(inputY+1+margin))
	availableAbove := max(0, inputY-margin)
	belowRows := max(0, min(availableBelow, maxHeight)-frameHeight)
	aboveRows := max(0, min(availableAbove, maxHeight)-frameHeight)

	bottomRows := max(0, min(maxHeight, m.height-margin)-frameHeight)
	if belowRows == 0 && aboveRows == 0 && bottomRows == 0 {
		return completionOverlayLayout{}, false
	}

	visibleRows := desiredRows
	popupY := inputY + 1 + margin
	switch m.completion.placement {
	case CompletionOverlayPlacementAbove:
		visibleRows = min(visibleRows, aboveRows)
		popupY = inputY - margin - (visibleRows + frameHeight)
	case CompletionOverlayPlacementBelow:
		visibleRows = min(visibleRows, belowRows)
		popupY = inputY + 1 + margin
	case CompletionOverlayPlacementBottom:
		visibleRows = min(visibleRows, bottomRows)
		popupY = m.height - margin - (visibleRows + frameHeight)
	case CompletionOverlayPlacementAuto:
		placeBelow := belowRows >= desiredRows || belowRows >= aboveRows
		if placeBelow {
			visibleRows = min(visibleRows, belowRows)
		} else {
			visibleRows = min(visibleRows, aboveRows)
			popupY = inputY - margin - (visibleRows + frameHeight)
		}
	default:
		visibleRows = min(visibleRows, aboveRows)
		popupY = inputY - margin - (visibleRows + frameHeight)
	}
	if visibleRows <= 0 {
		return completionOverlayLayout{}, false
	}

	anchorX := m.completionAnchorColumn()
	popupX := anchorX
	if m.completion.horizontal == CompletionOverlayHorizontalGrowLeft {
		popupX -= popupWidth
	}
	popupX += m.completion.offsetX
	popupY += m.completion.offsetY
	popupX = clampInt(popupX, 0, max(0, m.width-popupWidth))
	popupY = clampInt(popupY, 0, max(0, m.height-1))

	return completionOverlayLayout{
		PopupX:       popupX,
		PopupY:       popupY,
		PopupWidth:   popupWidth,
		VisibleRows:  visibleRows,
		ContentWidth: contentWidth,
	}, true
}

func (m *Model) computeHelpDrawerOverlayLayout(header, timelineView string) (helpDrawerOverlayLayout, bool) {
	if !m.helpDrawer.visible || m.width <= 0 || m.height <= 0 {
		return helpDrawerOverlayLayout{}, false
	}

	widthPercent := clampInt(m.helpDrawer.widthPercent, 20, 90)
	heightPercent := clampInt(m.helpDrawer.heightPercent, 20, 90)
	panelWidth := max(32, m.width*widthPercent/100)
	panelHeight := max(8, m.height*heightPercent/100)
	panelWidth = min(panelWidth, max(20, m.width-2))
	panelHeight = min(panelHeight, max(6, m.height-2))

	panelStyle := m.helpDrawerPanelStyle()
	frameWidth := panelStyle.GetHorizontalFrameSize()
	frameHeight := panelStyle.GetVerticalFrameSize()
	contentWidth := max(1, panelWidth-frameWidth)
	contentHeight := max(1, panelHeight-frameHeight)

	margin := max(0, m.helpDrawer.margin)
	headerHeight := lipgloss.Height(header)
	inputY := headerHeight + 1 + lipgloss.Height(timelineView)

	panelX := 0
	panelY := 0
	switch m.helpDrawer.dock {
	case HelpDrawerDockRight:
		panelX = m.width - margin - panelWidth
		panelY = headerHeight + 1 + margin
	case HelpDrawerDockLeft:
		panelX = margin
		panelY = headerHeight + 1 + margin
	case HelpDrawerDockBottom:
		panelX = (m.width - panelWidth) / 2
		panelY = m.height - margin - panelHeight
	case HelpDrawerDockAboveRepl:
		fallthrough
	default:
		panelX = (m.width - panelWidth) / 2
		panelY = inputY - margin - panelHeight
	}
	panelX = clampInt(panelX, 0, max(0, m.width-panelWidth))
	panelY = clampInt(panelY, 0, max(0, m.height-panelHeight))

	return helpDrawerOverlayLayout{
		PanelX:        panelX,
		PanelY:        panelY,
		PanelWidth:    panelWidth,
		PanelHeight:   panelHeight,
		ContentWidth:  contentWidth,
		ContentHeight: contentHeight,
	}, true
}

func (m *Model) renderHelpDrawerPanel(layout helpDrawerOverlayLayout) string {
	if layout.ContentWidth <= 0 || layout.ContentHeight <= 0 {
		return ""
	}

	title := "Help Drawer"
	subtitle := "No contextual help provider content yet"
	bodyLines := []string{}
	doc := m.helpDrawer.doc
	hasDoc := strings.TrimSpace(doc.Title) != "" ||
		strings.TrimSpace(doc.Subtitle) != "" ||
		strings.TrimSpace(doc.Markdown) != "" ||
		len(doc.Diagnostics) > 0 ||
		strings.TrimSpace(doc.VersionTag) != ""
	if hasDoc {
		if strings.TrimSpace(doc.Title) != "" {
			title = doc.Title
		}
		if strings.TrimSpace(doc.Subtitle) != "" {
			subtitle = doc.Subtitle
		}
		if !doc.Show {
			subtitle = "No contextual help for current input"
		}
		if strings.TrimSpace(doc.Markdown) != "" {
			bodyLines = append(bodyLines, strings.TrimSpace(doc.Markdown))
		}
		if len(doc.Diagnostics) > 0 {
			bodyLines = append(bodyLines, "")
			bodyLines = append(bodyLines, "Diagnostics:")
			for _, d := range doc.Diagnostics {
				d = strings.TrimSpace(d)
				if d == "" {
					continue
				}
				bodyLines = append(bodyLines, "- "+d)
			}
		}
		if strings.TrimSpace(doc.VersionTag) != "" {
			bodyLines = append(bodyLines, "")
			bodyLines = append(bodyLines, "Version: "+doc.VersionTag)
		}
	}
	if m.helpDrawer.err != nil {
		subtitle = "Error"
		bodyLines = append(bodyLines, m.helpDrawer.err.Error())
	}
	if m.helpDrawer.loading {
		if hasDoc {
			subtitle = strings.TrimSpace(subtitle + " (refreshing)")
		} else {
			subtitle = "Loading..."
		}
	}
	if m.helpDrawer.pinned {
		subtitle = strings.TrimSpace(subtitle + " [pinned]")
	}

	toggleKey := bindingPrimaryKey(m.keyMap.HelpDrawerToggle, "ctrl+h")
	refreshKey := bindingPrimaryKey(m.keyMap.HelpDrawerRefresh, "ctrl+r")
	pinKey := bindingPrimaryKey(m.keyMap.HelpDrawerPin, "ctrl+g")
	footer := fmt.Sprintf("%s toggle • %s refresh • %s pin", toggleKey, refreshKey, pinKey)
	content := []string{
		m.helpDrawerTitleStyle().Render(title),
		m.helpDrawerSubtitleStyle().Render(subtitle),
	}
	if len(bodyLines) > 0 {
		content = append(content, strings.Join(bodyLines, "\n"))
	}
	content = append(content, m.styles.HelpText.Render(footer))

	rendered := m.helpDrawerPanelStyle().
		Width(layout.PanelWidth).
		Height(layout.PanelHeight).
		Render(strings.Join(content, "\n\n"))
	return rendered
}

func (m *Model) renderCompletionPopup(layout completionOverlayLayout) string {
	if layout.VisibleRows <= 0 || layout.ContentWidth <= 0 {
		return ""
	}
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		return ""
	}

	start := clampInt(m.completion.scrollTop, 0, max(0, len(suggestions)-1))
	end := min(len(suggestions), start+layout.VisibleRows)
	lines := make([]string, 0, layout.VisibleRows)
	for i := start; i < end; i++ {
		itemText := "  " + suggestions[i].DisplayText
		itemStyle := m.styles.CompletionItem
		if i == m.completion.selection {
			itemStyle = m.styles.CompletionSelected
			itemText = "› " + suggestions[i].DisplayText
		}
		itemText = runewidth.Truncate(itemText, layout.ContentWidth, "")
		if delta := layout.ContentWidth - runewidth.StringWidth(itemText); delta > 0 {
			itemText += strings.Repeat(" ", delta)
		}
		lines = append(lines, itemStyle.Render(itemText))
	}
	return m.completionPopupStyle().Width(layout.PopupWidth).Render(strings.Join(lines, "\n"))
}

func (m *Model) completionAnchorColumn() int {
	runes := []rune(m.textInput.Value())
	cursor := clampInt(m.textInput.Position(), 0, len(runes))
	prefix := string(runes[:cursor])
	return runewidth.StringWidth(m.textInput.Prompt + prefix)
}

func (m *Model) completionPopupStyle() lipgloss.Style {
	if !m.completion.noBorder {
		return m.styles.CompletionPopup
	}
	return m.styles.CompletionPopup.
		Border(lipgloss.HiddenBorder(), false, false, false, false).
		Padding(0, 0)
}

func (m *Model) helpDrawerPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
}

func (m *Model) helpDrawerTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33"))
}

func (m *Model) helpDrawerSubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))
}

func bindingPrimaryKey(b key.Binding, fallback string) string {
	if !b.Enabled() {
		return fallback
	}
	keyName := strings.TrimSpace(b.Help().Key)
	if keyName == "" {
		return fallback
	}
	return keyName
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
	if m.completion.visibleRows > 0 {
		return max(1, m.completion.visibleRows)
	}
	if m.completion.maxVisible > 0 {
		return m.completion.maxVisible
	}
	return 1
}

func (m *Model) completionPageStep() int {
	if m.completion.pageSize > 0 {
		return max(1, m.completion.pageSize)
	}
	return m.completionVisibleLimit()
}

func (m *Model) ensureCompletionSelectionVisible() {
	suggestions := m.completion.lastResult.Suggestions
	if len(suggestions) == 0 {
		m.completion.scrollTop = 0
		return
	}

	m.completion.selection = clampInt(m.completion.selection, 0, len(suggestions)-1)
	limit := m.completionVisibleLimit()
	maxTop := max(0, len(suggestions)-limit)
	m.completion.scrollTop = clampInt(m.completion.scrollTop, 0, maxTop)
	if m.completion.selection < m.completion.scrollTop {
		m.completion.scrollTop = m.completion.selection
	}
	visibleEnd := m.completion.scrollTop + limit - 1
	if m.completion.selection > visibleEnd {
		m.completion.scrollTop = m.completion.selection - limit + 1
	}
	m.completion.scrollTop = clampInt(m.completion.scrollTop, 0, maxTop)
}

func (m *Model) updateKeyBindings() { mode_keymap.EnableMode(&m.keyMap, m.focus) }

func normalizeHelpBarConfig(cfg HelpBarConfig) HelpBarConfig {
	if cfg.Debounce == 0 && cfg.RequestTimeout == 0 && !cfg.Enabled {
		return DefaultHelpBarConfig()
	}
	merged := DefaultHelpBarConfig()
	merged.Enabled = cfg.Enabled
	if cfg.Debounce > 0 {
		merged.Debounce = cfg.Debounce
	}
	if cfg.RequestTimeout > 0 {
		merged.RequestTimeout = cfg.RequestTimeout
	}
	return merged
}

func normalizeHelpDrawerConfig(cfg HelpDrawerConfig) HelpDrawerConfig {
	if cfg.Debounce == 0 &&
		cfg.RequestTimeout == 0 &&
		len(cfg.ToggleKeys) == 0 &&
		len(cfg.CloseKeys) == 0 &&
		len(cfg.RefreshShortcuts) == 0 &&
		len(cfg.PinShortcuts) == 0 &&
		cfg.Dock == "" &&
		cfg.WidthPercent == 0 &&
		cfg.HeightPercent == 0 &&
		cfg.Margin == 0 &&
		!cfg.PrefetchWhenHidden &&
		!cfg.Enabled {
		return DefaultHelpDrawerConfig()
	}

	merged := DefaultHelpDrawerConfig()
	merged.Enabled = cfg.Enabled
	if len(cfg.ToggleKeys) > 0 {
		merged.ToggleKeys = cfg.ToggleKeys
	}
	if len(cfg.CloseKeys) > 0 {
		merged.CloseKeys = cfg.CloseKeys
	}
	if len(cfg.RefreshShortcuts) > 0 {
		merged.RefreshShortcuts = cfg.RefreshShortcuts
	}
	if len(cfg.PinShortcuts) > 0 {
		merged.PinShortcuts = cfg.PinShortcuts
	}
	if cfg.Debounce > 0 {
		merged.Debounce = cfg.Debounce
	}
	if cfg.RequestTimeout > 0 {
		merged.RequestTimeout = cfg.RequestTimeout
	}
	if cfg.Dock != "" {
		merged.Dock = cfg.Dock
	}
	if cfg.WidthPercent > 0 {
		merged.WidthPercent = cfg.WidthPercent
	}
	if cfg.HeightPercent > 0 {
		merged.HeightPercent = cfg.HeightPercent
	}
	if cfg.Margin > 0 {
		merged.Margin = cfg.Margin
	}
	if cfg.PrefetchWhenHidden {
		merged.PrefetchWhenHidden = true
	}

	merged.Dock = normalizeHelpDrawerDock(merged.Dock)
	merged.WidthPercent = clampInt(merged.WidthPercent, 20, 90)
	merged.HeightPercent = clampInt(merged.HeightPercent, 20, 90)
	merged.Margin = max(0, merged.Margin)
	return merged
}

func normalizeHelpDrawerDock(v HelpDrawerDock) HelpDrawerDock {
	switch v {
	case HelpDrawerDockAboveRepl, HelpDrawerDockRight, HelpDrawerDockLeft, HelpDrawerDockBottom:
		return v
	default:
		return HelpDrawerDockAboveRepl
	}
}

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
		cfg.OverlayOffsetX == 0 &&
		cfg.OverlayOffsetY == 0 &&
		!cfg.OverlayNoBorder &&
		cfg.OverlayPlacement == "" &&
		cfg.OverlayHorizontalGrow == "" &&
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
	if cfg.OverlayOffsetX != 0 {
		merged.OverlayOffsetX = cfg.OverlayOffsetX
	}
	if cfg.OverlayOffsetY != 0 {
		merged.OverlayOffsetY = cfg.OverlayOffsetY
	}
	if cfg.OverlayNoBorder {
		merged.OverlayNoBorder = true
	}
	if cfg.OverlayPlacement != "" {
		merged.OverlayPlacement = cfg.OverlayPlacement
	}
	if cfg.OverlayHorizontalGrow != "" {
		merged.OverlayHorizontalGrow = cfg.OverlayHorizontalGrow
	}
	merged.OverlayPlacement = normalizeOverlayPlacement(merged.OverlayPlacement)
	merged.OverlayHorizontalGrow = normalizeOverlayHorizontalGrow(merged.OverlayHorizontalGrow)
	return merged
}

func normalizeOverlayPlacement(v CompletionOverlayPlacement) CompletionOverlayPlacement {
	switch v {
	case CompletionOverlayPlacementAuto,
		CompletionOverlayPlacementAbove,
		CompletionOverlayPlacementBelow,
		CompletionOverlayPlacementBottom:
		return v
	default:
		return CompletionOverlayPlacementAuto
	}
}

func normalizeOverlayHorizontalGrow(v CompletionOverlayHorizontalGrow) CompletionOverlayHorizontalGrow {
	switch v {
	case CompletionOverlayHorizontalGrowRight, CompletionOverlayHorizontalGrowLeft:
		return v
	default:
		return CompletionOverlayHorizontalGrowRight
	}
}

func (m *Model) ctrl() *timeline.Controller { return m.sh.Controller() }

func (m *Model) cancelAppContext() {
	if m.appStop != nil {
		m.appStop()
	}
}

func (m *Model) appContext() context.Context {
	if m.appCtx != nil {
		return m.appCtx
	}
	return context.Background()
}

func newTurnID(seq int) string {
	return timeNow().Format("20060102-150405.000000000") + ":" + fmt.Sprintf("%d", seq)
}

func timeNow() time.Time { return time.Now() }
