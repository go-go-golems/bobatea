# Refactoring REPL to a Timeline‑Centric Design (No Backcompat)

This document specifies a detailed architecture, design, and implementation plan to refactor the REPL to a timeline‑first UI. We explicitly do not maintain backward compatibility; the REPL becomes a thin input + controller shell while the timeline renders the complete transcript via entity models.


## Goals

- Replace monolithic REPL transcript rendering with append‑only timeline entities.
- Make structured/streaming output first‑class via a new evaluator contract.
- Ship a minimal, testable slice first, then progressively add widgets.
- Keep ergonomics: simple input, history navigation, slash commands, external editor.


## High‑Level Architecture

- REPL Model (UI shell)
  - Owns a `timeline.Controller` and `timeline.Registry`.
  - Renders two regions:
    - Timeline viewport (transcript)
    - Input line (or multiline editor) with status/help
  - Handles focus switching (timeline selection vs input focus).

- Evaluator (streaming, structured)
  - New interface that emits typed events (stdout, stderr, markdown result, table, etc.).
  - REPL bridges events → timeline entity lifecycle.

- Renderers (widgets)
  - Reuse existing `llm_text`, `tool_calls_panel`, `plain`.
  - Add REPL‑specific minimal widgets and extend over time.


## Package Layout (Target)

- `pkg/repl/` (refactored in place; no backcompat)
  - `model.go` — New REPL shell model (timeline + input)
  - `evaluator.go` — New evaluator interfaces (streaming)
  - `bridge.go` — Maps evaluator events to timeline messages
  - `messages.go` — REPL messages (unchanged where possible)
  - `styles.go` — Styles for input/status; entity styles via timeline renderers
  - `history.go` — Kept, used only for input recall (no transcript rendering)
- `pkg/timeline/` — Unchanged; register additional factories here or in REPL init
- `pkg/timeline/renderers/`
  - `repl_input_model.go` — Minimal input code widget (new)
  - (later) `repl_stdout_model.go`, `repl_stderr_model.go`, `repl_progress_model.go`, etc.
- `cmd/repl-timeline-demo/` — Demo program exercising new REPL


## New Evaluator Contract

```go
// pkg/repl/evaluator.go

// EventKind enumerates structured output kinds.
type EventKind string
const (
  EventInput            EventKind = "repl_input"
  EventResultMarkdown   EventKind = "repl_result_markdown"
  EventStdout           EventKind = "repl_stdout"
  EventStderr           EventKind = "repl_stderr"
  EventToolCalls        EventKind = "repl_tool_calls"
  EventProgress         EventKind = "repl_progress"
  EventPerf             EventKind = "repl_perf"
  EventTable            EventKind = "repl_table"
  EventDiff             EventKind = "repl_diff"
  EventShellCmd         EventKind = "repl_shell_cmd"
  EventInspector        EventKind = "repl_inspector"
)

// Event carries a semantic payload for the UI.
type Event struct {
  Kind  EventKind
  Props map[string]any // renderer props patch
}

// Evaluator executes code and streams events.
type Evaluator interface {
  EvaluateStream(ctx context.Context, code string, emit func(Event)) error
  GetPrompt() string
  GetName() string
  SupportsMultiline() bool
  GetFileExtension() string
}
```

Notes:
- We eliminate the old `(string, error)` return. All output is events; errors are represented as events (stderr or result with error state) and the method’s terminal error.
- The REPL shell remains responsible for creating an input event and for turn ID management.


## Timeline Bridge

Responsibilities:
- Create a new TurnID per submit.
- Emit `UIEntityCreated` for the input entity.
- Lazily create stream entities (stdout/stderr) on first chunk; update with patches.
- Emit result entity/entities and mark completed at the end.

Sketch:

```go
// pkg/repl/bridge.go

type Bridge struct {
  ctrl *timeline.Controller
  reg  *timeline.Registry
  clock func() time.Time
}

func (b *Bridge) NewTurnID() string { /* counter or ULID */ }

func (b *Bridge) Emit(turnID, local, kind string, props map[string]any) timeline.EntityID {
  id := timeline.EntityID{TurnID: turnID, LocalID: local, Kind: kind}
  b.ctrl.OnCreated(timeline.UIEntityCreated{
    ID: id,
    Renderer: timeline.RendererDescriptor{Kind: kind},
    Props: props,
    StartedAt: b.clock(),
  })
  return id
}

func (b *Bridge) Patch(id timeline.EntityID, patch map[string]any) { /* OnUpdated */ }
func (b *Bridge) Complete(id timeline.EntityID, result map[string]any) { /* OnCompleted */ }
```

Mapping evaluator events → emit/patch/complete:
- `EventInput` → Emit once at start
- `EventStdout/Stderr` → Emit on first chunk; Patch thereafter
- `EventResultMarkdown/Table/Diff/Inspector/ToolCalls` → Emit immediately when available; Complete on finalize
- `EventProgress/Perf` → Emit once, Patch periodically, optional completion


## Minimal Widget Set (MVP)

Focus on smallest viable feature set to test end‑to‑end:
- `repl_input` (new): boxed monospace view of submitted code; copyable
- `repl_result_markdown` (reuse `llm_text`): supports markdown, code copy, metadata footer (duration/exit code)
- Fallback `plain` for anything else during MVP

Later increments:
- `repl_stdout`, `repl_stderr` (streaming text with distinct styles)
- `repl_progress` (single‑line progress/spinner)
- `repl_tool_calls` (extend `tool_calls_panel`)
- `repl_perf` (sparkline)
- `repl_table`, `repl_diff`, `repl_shell_cmd`, `repl_inspector`


## REPL Shell Model (New)

Composition:
- Fields: `ctrl *timeline.Controller`, `reg *timeline.Registry`, `input textinput.Model`, `history *History`, `width,height int`, `focus enum{Input, Timeline}`, `bridge Bridge`, `turnSeq int`.
- Init: register renderers (`llm_text`, `tool_calls_panel`, `plain`, `repl_input`).
- Update:
  - `tea.WindowSizeMsg`: compute layout; `ctrl.SetSize(w, timelineHeight)` and adjust input width.
  - `tea.KeyMsg`:
    - Focus switching: `tab` toggles between Input and Timeline
    - When focus=Timeline: up/down move selection; enter toggles entering mode; `c/y` copy code/text
    - When focus=Input: up/down navigate history; enter submits
  - Submit flow:
    - Build `turnID := fmt.Sprintf("turn-%d", m.turnSeq++)`
    - Bridge emit input entity (`repl_input`) with prompt+code
    - Start evaluator goroutine; map events → bridge calls; on completion, emit result entity and complete
- View:
  - Header/title line
  - Timeline viewport (controller.View()) above
  - Input line and help/status below


## Keybindings (Initial)

- Global: `q`/`ctrl+c` quit
- Focus: `tab` toggle Input <-> Timeline, `esc` exit timeline entering mode
- Input focus:
  - `enter` submit; `ctrl+e` external editor; `up/down` history; `ctrl+j` add multiline line
- Timeline focus:
  - `up/down` selection; `enter` toggle entering; `c` copy code; `y` copy text


## Theming

- Maintain REPL input/status theming in `pkg/repl/styles.go`.
- Timeline entity theming via each renderer; allow `Controller.SetTheme("dark"|"light"|customKey)` to broadcast.
- Provide a mapping function from REPL theme → renderer props for consistency.


## Implementation Plan

Phase 0 — Preparation
- Create `cmd/repl-timeline-demo` skeleton app.
- Add basic logging (no external deps beyond zerolog if already in repo).

Phase 1 — New Evaluator + Bridge + Shell MVP
- Replace `pkg/repl/evaluator.go` with the new `Evaluator` interface and `Event`/`EventKind`.
- Add `pkg/repl/bridge.go` implementing timeline emission helpers.
- Refactor `pkg/repl/model.go`:
  - Remove transcript rendering; keep input/history/editor logic.
  - Add timeline controller + registry; register `llm_text`, `tool_calls_panel`, `plain`.
  - Add focus management and key routing to controller (`SelectNext/Prev`, `EnterSelection/ExitSelection`, `SendToSelected`).
  - On submit: create `turnID`, emit `repl_input`, run `EvaluateStream` and map events. For MVP, map final string‑like outcome to `EventResultMarkdown` (e.g., the evaluator can simply emit that event once).
- Add `pkg/timeline/renderers/repl_input_model.go` (minimal):
  - Props: `prompt`, `code`, `language?`, `cwd?`, optional `timestamp`
  - View: boxed code; implement `EntityCopyTextMsg` to copy `code`.

Phase 2 — Examples and Demo
- Add `cmd/repl-timeline-demo/main.go` creating a simple evaluator that emits just `EventResultMarkdown`.
- Update `examples/repl/*` to use the new REPL; remove examples using the old `(string,error)` API.

Phase 3 — Streaming + stdout/stderr
- Implement `repl_stdout_model.go` and `repl_stderr_model.go` with stream support (append text; handle width changes).
- Add a `ShellEvaluator` that emits `EventStdout`/`EventStderr` in real time and a final `EventResultMarkdown` with exit code/duration.

Phase 4 — Progressive Widgets
- Add `repl_progress` and integrate with long‑running evaluators.
- Extend `repl_tool_calls` with `tool_calls_panel` to mirror sub‑steps.
- Add `repl_perf` (sparkline) to visualize last N eval durations.
- Add `repl_table`, `repl_diff`, `repl_shell_cmd`, `repl_inspector` based on usage.

Phase 5 — Polishing and API Stabilization
- Theme propagation across timeline and input.
- Copy semantics and UX docs (key cheatsheet in footer).
- Benchmarks for large transcripts (thousands of entities).


## File‑Level Changes (Initial Commit Scope)

- Modify: `pkg/repl/evaluator.go` — new interfaces and types
- Modify: `pkg/repl/model.go` — new shell with timeline integration
- Add: `pkg/repl/bridge.go` — timeline bridge
- Add: `pkg/timeline/renderers/repl_input_model.go` — minimal renderer
- Modify: `pkg/repl/messages.go` — keep as is where possible; add any needed messages
- Keep: `pkg/repl/history.go`, `pkg/repl/styles.go` — used for input only
- Add: `cmd/repl-timeline-demo/main.go` — boot REPL shell and run


## Testing Strategy

- Manual demo: run `go run ./cmd/repl-timeline-demo`.
- Simulated evaluator: emit a fixed sequence of events (input → result_markdown) to validate entity creation, selection, copy, and resizing.
- Unit tests for bridge mapping (props patches, completion) and `repl_input_model` copy behavior.


## Risks and Mitigations

- Scope creep: enforce MVP (input + result markdown) before stdout/stderr.
- Event spam: coalesce frequent patches (batch stdout lines) to reduce churn.
- Width handling: always update models with `EntitySetSizeMsg`; renderers must wrap.
- API churn: document new evaluator contract and pin examples; no backcompat expected.


## Acceptance Criteria (MVP)

- Submitting input produces two entities: `repl_input` and `repl_result_markdown`.
- Timeline renders above input; selection and copy work (c/y).
- Window resize updates both timeline and input widths.
- Example evaluator emits result markdown and demo runs without panics.


## Follow‑ups

- Implement stdout/stderr and progress widgets.
- Theme harmonization across timeline renderers.
- Add navigable “turns” sidebar or filters (via labels) if needed.

