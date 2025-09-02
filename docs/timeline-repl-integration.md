---
Title: Build and Integrate a Timeline REPL (Bobatea)
Slug: timeline-repl-integration
Short: End-to-end guide to embed the timeline-first REPL in your Bubble Tea app or create a standalone REPL from scratch.
Topics:
- bobatea
- repl
- timeline
- bubbletea
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

## Build and Integrate a Timeline REPL

This tutorial shows how to add the new timeline-first REPL to an existing Bubble Tea application, and how to build a standalone REPL from scratch. The REPL emits semantic events, a transformer maps them to timeline UI entities, and a forwarder routes those entities into the UI. This decoupling gives you a durable, turn-centric transcript with rich renderers (text, markdown, logs, structured data) and streaming.

### What you’ll learn

- The evaluator contract for streaming events (timeline-first)
- Wiring an in-memory event bus, transformer, and UI forwarder
- Creating and running the REPL model with the timeline shell
- Available renderers and the props they expect

## Prerequisites

- Go 1.21+
- Familiarity with Bubble Tea
- A project set up similar to the Bobatea repository
- Optional: configure logging using zerolog via `bobatea/pkg/logutil`

## Core concepts (quick recap)

The REPL model renders a timeline transcript plus an input line. Evaluators stream semantic events; a transformer converts those into timeline lifecycle messages; the UI forwarder passes them into Bubble Tea, where renderer models produce the final view.

- Evaluator: implements `EvaluateStream(ctx, code, emit)` and metadata getters
- Transformer: subscribes to `repl.events`, publishes `timeline.created/updated/completed/deleted`
- UI forwarder: subscribes to `ui.entities` and `p.Send(...)` Bubble Tea messages
- Timeline shell: manages a `viewport` and selection/focus ergonomics

See also:
- `bobatea/docs/timeline.md` for controller, shell, and lifecycle details
- `bobatea/examples/js-repl/main.go` for a complete executable example

## Evaluator interface (timeline-first)

Your evaluator should stream events via the provided `emit` callback. Errors should generally be represented as events (stderr or result markdown) so they appear in the timeline.

```go
type Evaluator interface {
    EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error
    GetPrompt() string
    GetName() string
    SupportsMultiline() bool
    GetFileExtension() string
}
```

Common event kinds and props (handled by the built-in transformer):
- `repl.EventResultMarkdown`:
  - props: `{ "markdown": string }` (alias: `text`)
- `repl.EventStdout` / `repl.EventStderr`:
  - props: `{ "append": string }` or `{ "text": string }`, optional `is_error: true`
- `repl.EventLog`:
  - props: `{ "level": string, "message": string, "metadata"?: any, "fields"?: any }`
- `repl.EventStructuredLog`:
  - props: `{ "level": string, "message"?: string, "data"|"metadata"|"fields": any }`
- `repl.EventInspector`:
  - props: `{ "data": any }` or `{ "json": string }`

## From scratch: minimal standalone REPL

The example below mirrors `bobatea/examples/js-repl/main.go`, but simplified. You can either:

- Use the convenience helper `repl.RunTimelineRepl` (simplest), or
- Wire the bus/transformer/forwarder yourself (advanced embedding).

### Option A: use the helper (recommended)

```go
package main

import (
    "context"
    "github.com/go-go-golems/bobatea/pkg/repl"
)

type EchoEvaluator struct{}

func (e *EchoEvaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
    emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{
        "markdown": "You said: " + code,
    }})
    return nil
}
func (e *EchoEvaluator) GetPrompt() string        { return "echo> " }
func (e *EchoEvaluator) GetName() string          { return "Echo" }
func (e *EchoEvaluator) SupportsMultiline() bool  { return false }
func (e *EchoEvaluator) GetFileExtension() string { return ".txt" }

func main() {
    cfg := repl.DefaultConfig()
    cfg.Title = "Echo REPL"
    if err := repl.RunTimelineRepl(&EchoEvaluator{}, cfg); err != nil { panic(err) }
}
```

### Option B: wire bus + program manually

```go
package main

import (
    "context"
    "log"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/eventbus"
    "github.com/go-go-golems/bobatea/pkg/repl"
    "github.com/go-go-golems/bobatea/pkg/timeline"
)

// EchoEvaluator demonstrates EvaluateStream by echoing back markdown
type EchoEvaluator struct{}

func (e *EchoEvaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
    emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{
        "markdown": "You said: " + code,
    }})
    return nil
}

func (e *EchoEvaluator) GetPrompt() string          { return "echo> " }
func (e *EchoEvaluator) GetName() string            { return "Echo" }
func (e *EchoEvaluator) SupportsMultiline() bool    { return false }
func (e *EchoEvaluator) GetFileExtension() string   { return ".txt" }

func main() {
    // 1) Event bus
    bus, err := eventbus.NewInMemoryBus()
    if err != nil { log.Fatal(err) }

    // 2) Transformer: REPL events -> timeline UI entities
    repl.RegisterReplToTimelineTransformer(bus)

    // 3) Evaluator and REPL model
    evaluator := &EchoEvaluator{}
    cfg := repl.DefaultConfig()
    cfg.Title = "Echo REPL"
    model := repl.NewModel(evaluator, cfg, bus.Publisher)

    // 4) Bubble Tea program and UI forwarder
    p := tea.NewProgram(model, tea.WithAltScreen())
    timeline.RegisterUIForwarder(bus, p)

    // 5) Run bus and program together
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    errs := make(chan error, 2)
    go func() { errs <- bus.Run(ctx) }()
    go func() { _, e := p.Run(); cancel(); errs <- e }()
    if e := <-errs; e != nil { log.Fatal(e) }
}
```

Run it, type text, and watch the markdown entity appear in the transcript.

## Integrate into an existing Bubble Tea app

If you already have a Bubble Tea model, instantiate and delegate to the REPL model alongside your own components.

```go
type AppModel struct {
    replModel *repl.Model
}

func NewAppModel(bus *eventbus.Bus) *AppModel {
    evaluator := &MyEvaluator{}
    cfg := repl.DefaultConfig()
    cfg.Title = "My App REPL"
    m := repl.NewModel(evaluator, cfg, bus.Publisher)
    return &AppModel{replModel: m}
}

func (m *AppModel) Init() tea.Cmd { return m.replModel.Init() }

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    *m.replModel, cmd = m.replModel.Update(msg)
    return m, cmd
}

func (m *AppModel) View() string { return m.replModel.View() }
```

In your `main`, wire the bus, transformer, UI forwarder, and program as shown in the standalone example.

### Emitting events from other parts of your app (Watermill)

When embedding the REPL into a broader application, any subsystem can enrich the transcript by publishing semantic REPL events to the bus. The REPL transformer (`RegisterReplToTimelineTransformer`) listens on `repl.events` and produces timeline entities.

```go
// Envelope expected by RegisterReplToTimelineTransformer
payload, _ := json.Marshal(struct {
    TurnID string     `json:"turn_id"`
    Event  repl.Event `json:"event"`
    Time   time.Time  `json:"time"`
}{
    TurnID: "ext-1",
    Event:  repl.Event{Kind: repl.EventLog, Props: map[string]any{"level": "info", "message": "external note"}},
    Time:   time.Now(),
})
_ = bus.Publisher.Publish(eventbus.TopicReplEvents, message.NewMessage(watermill.NewUUID(), payload))
```

Guidelines:
- Choose a stable `TurnID` per logical operation to group entities
- Prefer `EventStdout`/`EventStderr` with `append` for streaming text
- Use `EventStructuredLog` with `data` for rich inspection; it renders with YAML/JSON renderers

## Logging configuration

When running TUIs, logs should not interfere with stdout. Use the provided helpers to send logs to a file or discard output while preserving levels:

```go
import (
    "flag"
    "github.com/go-go-golems/bobatea/pkg/logutil"
    "github.com/rs/zerolog"
)

func parseLevel(s string) zerolog.Level { /* map: trace/debug/info/warn/error */ }

func main() {
    ll := flag.String("log-level", "error", "log level: trace, debug, info, warn, error")
    lf := flag.String("log-file", "", "log file path (optional)")
    flag.Parse()

    level := parseLevel(*ll)
    if *lf != "" {
        logutil.InitTUILoggingToFile(level, *lf)
    } else {
        logutil.InitTUILoggingToDiscard(level)
    }

    // ... then run the REPL (helper or manual wiring)
}
```

These helpers configure global zerolog state to avoid polluting the terminal UI while still capturing diagnostics when needed.

## Renderers and expected props

Registered by default in the REPL model’s registry:

- Text (`kind: "text"`):
  - props: `text` (replace), `append` (streaming), `is_error` (style), `streaming` (hint)
- Markdown (`kind: "markdown"`):
  - props: `markdown` (preferred) or `text` (alias)
  - Copy actions: raw markdown or all code blocks (`c`)
- Log event (`kind: "log_event"`):
  - props: `level`, `message`, optional `metadata`, `fields`
  - Selected: press `Tab` to toggle YAML metadata visibility
- Structured log event (`kind: "structured_log_event"`):
  - props: `level`, `message?`, `yaml` or any of `data/metadata/fields` (auto-marshalled)
- Structured data (`kind: "structured_data"`):
  - props: `data` (any Go value) or `json` (string) — pretty-printed as JSON

You can register additional factories with `timeline.Registry.RegisterModelFactory(...)` before running the program.

## Keyboard shortcuts

- Tab: switch focus between input and timeline; selected entities may also handle Tab
- Enter: submit input (or toggle enter/exit selection inside timeline)
- Up/Down: history (input) or selection (timeline)
- c: copy code (renderer-specific)
- y: copy text
- Ctrl+C: quit

## Troubleshooting

- No transcript updates: ensure you registered both the transformer and UI forwarder, and that the bus is running
- Black screen: resize the terminal to force a `WindowSizeMsg`; the shell refreshes on size changes
- Markdown not rendering: renderer falls back to plain text if Glamour fails or no TTY is detected

## Next steps

- Build richer renderers (diffs, panels) by implementing `timeline.EntityModel` and registering a factory
- Integrate with your agent or tool backend by emitting REPL events from your evaluator logic


