package repl

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    renderers "github.com/go-go-golems/bobatea/pkg/timeline/renderers"
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

    // streaming
    events chan tea.Msg // EvalEventMsg or EvalDoneMsg
    turnSeq int
    streams map[string]struct{ Stdout, Stderr bool }

    // refresh scheduling
    refreshPending   bool
    refreshScheduled bool
}

// NewModel constructs a new REPL shell with timeline transcript.
func NewModel(evaluator Evaluator, config Config) *Model {
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
    reg.RegisterModelFactory(renderers.MarkdownFactory{})
    reg.RegisterModelFactory(renderers.StructuredDataFactory{})

    sh := timeline.NewShell(reg)

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
        events:    make(chan tea.Msg, 128),
        streams:   map[string]struct{ Stdout, Stderr bool }{},
    }
}

// Messages
type EvalEventMsg struct {
    TurnID string
    Event  Event
}
type EvalDoneMsg struct {
    TurnID string
    Err    error
}

func waitForEvents(ch <-chan tea.Msg) tea.Cmd {
    return func() tea.Msg { return <-ch }
}

// Init subscribes to evaluator events.
func (m *Model) Init() tea.Cmd {
    return tea.Batch(textinput.Blink, waitForEvents(m.events), m.sh.Init())
}

// Update handles TUI events.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case tea.WindowSizeMsg:
        m.width, m.height = v.Width, v.Height
        m.textInput.Width = max(10, v.Width-10)
        // give most of the space to timeline shell viewport
        tlHeight := max(0, v.Height-4)
        m.sh.SetSize(v.Width, tlHeight)
        // initial refresh to fit new size
        m.sh.RefreshView(false)
        return m, nil

    case tea.KeyMsg:
        switch m.focus {
        case "input":
            return m.updateInput(v)
        case "timeline":
            return m.updateTimeline(v)
        }

    case EvalEventMsg:
        m.applyEvent(v.TurnID, v.Event)
        // schedule refresh to coalesce bursts
        return m, tea.Batch(waitForEvents(m.events), m.scheduleRefresh())

    case EvalDoneMsg:
        // For now, do nothing; evaluators should have emitted final result entity
        return m, waitForEvents(m.events)
    case timelineRefreshMsg:
        m.refreshScheduled = false
        if m.refreshPending {
            m.sh.RefreshView(true)
            m.refreshPending = false
        }
        return m, nil
    }

    // default: update input component
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)
    return m, cmd
}

func (m *Model) updateInput(k tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch k.Type {
    case tea.KeyCtrlC:
        return m, tea.Quit
    }
    switch k.String() {
    case "tab":
        m.focus = "timeline"
        return m, nil
    case "enter":
        input := m.textInput.Value()
        if strings.TrimSpace(input) == "" {
            return m, nil
        }
        m.textInput.Reset()
        if m.config.EnableHistory { m.history.Add(input, "", false); m.history.ResetNavigation() }
        return m, m.submit(input)
    case "up":
        if m.config.EnableHistory {
            if entry := m.history.NavigateUp(); entry != "" { m.textInput.SetValue(entry) }
        }
        return m, nil
    case "down":
        if m.config.EnableHistory {
            entry := m.history.NavigateDown(); m.textInput.SetValue(entry)
        }
        return m, nil
    }
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(k)
    return m, cmd
}

func (m *Model) updateTimeline(k tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch k.String() {
    case "tab":
        m.focus = "input"
        return m, nil
    case "up":
        m.sh.SelectPrev(); return m, nil
    case "down":
        m.sh.SelectNext(); return m, nil
    case "enter":
        if m.sh.IsEntering() { m.sh.ExitSelection() } else { m.sh.EnterSelection() }
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
    if title == "" { title = fmt.Sprintf("%s REPL", m.evaluator.GetName()) }
    b.WriteString(m.styles.Title.Render(" "+title+" "))
    b.WriteString("\n\n")
    // timeline view (viewport-wrapped)
    b.WriteString(m.sh.View())
    b.WriteString("\n")
    // input
    b.WriteString(m.textInput.View())
    b.WriteString("\n")
    // help
    help := "TAB: switch focus | Enter: submit | Up/Down: history/selection | c: copy code | y: copy text | Ctrl+C: quit"
    b.WriteString(m.styles.HelpText.Render(help))
    b.WriteString("\n")
    return b.String()
}

// submit runs evaluation and streams events to m.events
func (m *Model) submit(code string) tea.Cmd {
    turnID := newTurnID(m.turnSeq)
    m.turnSeq++
    // Emit input as a markdown code fence so it looks nice
    go func() {
        m.events <- EvalEventMsg{TurnID: turnID, Event: Event{Kind: EventInput, Props: map[string]any{"markdown": "```\n"+code+"\n```"}}}
        // stream from evaluator
        _ = m.evaluator.EvaluateStream(context.Background(), code, func(e Event) {
            m.events <- EvalEventMsg{TurnID: turnID, Event: e}
        })
        m.events <- EvalDoneMsg{TurnID: turnID, Err: nil}
    }()
    return nil
}

func (m *Model) applyEvent(turnID string, e Event) {
    switch e.Kind {
    case EventInput:
        id := timeline.EntityID{TurnID: turnID, LocalID: "input", Kind: "markdown"}
        m.ctrl().OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "markdown"}, Props: e.Props, StartedAt: timeNow()})
    case EventStdout:
        id := timeline.EntityID{TurnID: turnID, LocalID: "stdout", Kind: "text"}
        if st, ok := m.streams[turnID]; !ok || !st.Stdout {
            m.ctrl().OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "text"}, Props: map[string]any{"text": "", "streaming": true}, StartedAt: timeNow()})
            st.Stdout = true
            if !ok { st.Stderr = false }
            m.streams[turnID] = st
        }
        m.ctrl().OnUpdated(timeline.UIEntityUpdated{ID: id, Patch: ensureAppendPatch(e.Props), Version: timeNow().UnixNano(), UpdatedAt: timeNow()})
    case EventStderr:
        id := timeline.EntityID{TurnID: turnID, LocalID: "stderr", Kind: "text"}
        if st, ok := m.streams[turnID]; !ok || !st.Stderr {
            m.ctrl().OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "text"}, Props: map[string]any{"text": "", "streaming": true, "is_error": true}, StartedAt: timeNow()})
            st.Stderr = true
            if !ok { st.Stdout = false }
            m.streams[turnID] = st
        }
        p := ensureAppendPatch(e.Props)
        p["is_error"] = true
        m.ctrl().OnUpdated(timeline.UIEntityUpdated{ID: id, Patch: p, Version: timeNow().UnixNano(), UpdatedAt: timeNow()})
    case EventResultMarkdown:
        // Each result gets its own local id sequence
        local := fmt.Sprintf("result-%d", m.turnSeq)
        id := timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "markdown"}
        m.ctrl().OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "markdown"}, Props: e.Props, StartedAt: timeNow()})
        // mark complete immediately
        m.ctrl().OnCompleted(timeline.UIEntityCompleted{ID: id, Result: nil})
    case EventInspector:
        local := fmt.Sprintf("inspect-%d", m.turnSeq)
        id := timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "structured_data"}
        m.ctrl().OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "structured_data"}, Props: e.Props, StartedAt: timeNow()})
        m.ctrl().OnCompleted(timeline.UIEntityCompleted{ID: id, Result: nil})
    default:
        // Fallback to text
        local := fmt.Sprintf("event-%d", m.turnSeq)
        id := timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "text"}
        m.ctrl().OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "text"}, Props: e.Props, StartedAt: timeNow()})
        m.ctrl().OnCompleted(timeline.UIEntityCompleted{ID: id, Result: nil})
    }
    m.refreshPending = true
}

func max(a, b int) int { if a > b { return a }; return b }

func ensureAppendPatch(props map[string]any) map[string]any {
    if props == nil { return map[string]any{} }
    if _, ok := props["append"]; ok { return props }
    if s, ok := props["text"].(string); ok { return map[string]any{"append": s} }
    return props
}

// internal helpers
type timelineRefreshMsg struct{}

func (m *Model) scheduleRefresh() tea.Cmd {
    if m.refreshScheduled { return nil }
    m.refreshScheduled = true
    return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg { return timelineRefreshMsg{} })
}

func (m *Model) ctrl() *timeline.Controller { return m.sh.Controller() }

func newTurnID(seq int) string {
    return timeNow().Format("20060102-150405.000000000") + ":" + fmt.Sprintf("%d", seq)
}

func timeNow() time.Time { return time.Now() }
