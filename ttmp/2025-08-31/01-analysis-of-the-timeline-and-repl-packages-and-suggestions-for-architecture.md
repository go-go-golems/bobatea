# Timeline + REPL: Analysis and Architecture Proposals

This document analyzes the timeline and REPL components in this repository and proposes an architecture to combine them for a richer REPL experience. The goal is to adopt the append-only, entity-driven timeline architecture to render REPL interactions as first-class entities with interactive, customizable widgets.


## 1) What Exists Today

### Timeline (Turn-centric UI)
- Append-only visualization: Entities are created, updated, completed, deleted. Ordering follows creation time.
- Decoupled rendering: Renderers are Bubble Tea models that receive props/patches and render directly (no stateless renderers or caches).
- Controller + Registry:
  - `Controller` manages entity lifecycle, selection/focus, routing messages to entity models, and rendering a viewport-friendly view.
  - `Registry` maps `RendererDescriptor{Key, Kind}` to factories producing interactive entity models.
- Entity identity and lifecycle:
  - `EntityID{RunID, TurnID, BlockID, LocalID, Kind}` + `RendererDescriptor{Key, Kind}`
  - Messages: `UIEntityCreated`, `UIEntityUpdated`, `UIEntityCompleted`, `UIEntityDeleted`
  - Controller also routes selection/focus and copy requests (`EntityCopyTextMsg`, `EntityCopyCodeMsg`).
- Renderers present:
  - `llm_text` (markdown, role, streaming animation, metadata footer, code copy)
  - `tool_calls_panel` (compact panel of tool calls)
  - `plain` (debugging: prints props)
- Demo (`cmd/timeline-demo`): keyboard-driven example that streams text and a tools panel; shows controller APIs, selection, and copy flows.

### REPL
- Embeddable Bubble Tea REPL with a pluggable `Evaluator` interface:
  - `Evaluate(ctx, code) (string, error)` + getters for prompt/name/multiline/extension
  - History, themes/styles, multiline, external editor, slash commands, async via Tea commands
  - View prints a traditional linear transcript of input/output.
- Examples include basic usage, custom evaluators, custom themes, multipane embedding.


## 2) Observations From Examples

- Timeline demo shows how entities can stream, be selected/focused, and expose secondary actions (copy text/code). It is well-suited to visualize turns or steps.
- REPL examples show classic REPL behavior (input → evaluation → output), with room for richer outputs (tables, panels, progress, tool-call summaries), but current rendering is monolithic.
- The sparkline package exists and could power REPL performance widgets.


## 3) Proposed Vision: A Timeline-First REPL

Move the REPL from a single flat transcript to a timeline of entities. Each user input and its outputs become a “turn” comprised of multiple entities:
- Input entity (the code or command the user submitted)
- Output entities (stdout, stderr, result, logs, tool-call summaries, structured artifacts)
- Status entities (progress, timing, resource usage, metadata)

The REPL view becomes:
- Top: timeline viewport (append-only, selectable, interactive entities)
- Bottom: REPL input line (or multiline editor) controlling new turns
- Controller mediates selection/focus; entities can handle their own key interactions when focused (expand/collapse, copy, rerun, etc.)

Benefits:
- Rich per-result widgets (tables, diffs, inspectors, plots, markdown) instead of plain text
- Clear separation of concerns: evaluators emit events, timeline renders them
- Easy extension: registering new renderers adds capabilities without rewriting REPL core


## 4) Entity Taxonomy for REPL

Suggested entity kinds and their expected props. Each should have a corresponding renderer factory.

- `repl_input`
  - Props: `prompt`, `code`, `language`, `cwd`, `timestamp`
  - Widget: monospace boxed code block, copyable, optional syntax highlighting
- `repl_stdout`
  - Props: `text`, `streaming`, `chunks` (append), `timestamp`
  - Widget: stream-aware text area, optional copy-as-text
- `repl_stderr`
  - Props: `text`, `streaming`, `severity`, `timestamp`
  - Widget: emphasized/error style, stream-aware
- `repl_result_markdown`
  - Props: `markdown`, `metadata` (duration, exit-code), `streaming`
  - Widget: based on `llm_text` renderer; supports code copy
- `repl_tool_calls`
  - Props: `calls: []{id,name,status,result?}`, `summary`
  - Widget: reuse/extend `tool_calls_panel`
- `repl_progress`
  - Props: `label`, `current`, `total`, `phase`, `spinner`
  - Widget: progress bar + spinner line
- `repl_perf`
  - Props: `durations_ms: []int`, `last_ms`, `success_rate`
  - Widget: sparkline + summary footer (use `pkg/sparkline`)
- `repl_table`
  - Props: `headers: []string`, `rows: [][]string`, `max_rows`, `truncated`
  - Widget: compact table, scroll or expand when focused
- `repl_diff`
  - Props: `a_label`, `b_label`, `diff_text` (unified), `stats`
  - Widget: colored unified diff with expand/collapse
- `repl_fs_event`
  - Props: `path`, `op` (created/updated/deleted), `size`, `preview?`
  - Widget: compact line, expandable details
- `repl_shell_cmd`
  - Props: `command`, `cwd`, `exit_code`, `stdout`, `stderr`, `duration_ms`
  - Widget: command header line, foldable stdout/stderr sections
- `repl_inspector`
  - Props: `object_kind`, `summary`, `details` (JSON), `expandable`
  - Widget: pretty-printed JSON or tree explorer

The base `plain` renderer is a good fallback for any entity kind during development.


## 5) Event Flow and Mapping

Introduce a bridge layer (adapter) that maps REPL lifecycle and evaluator signals to timeline lifecycle messages.

- On user submit:
  - Create a new TurnID (e.g., timestamp or incrementing counter)
  - Emit `UIEntityCreated{ID: {TurnID, LocalID: "input", Kind: "repl_input"}, Props: {prompt, code, language, ...}}`
- During evaluation:
  - For streaming stdout/stderr from evaluator, emit `UIEntityCreated` for each stream entity upon first chunk, then `UIEntityUpdated` with `Patch: {chunks: append(...), text: concat}` on subsequent chunks
  - For progress, `repl_progress` updates via `UIEntityUpdated` with `current/total`
- On result:
  - Emit one or more result entities: `repl_result_markdown`, `repl_table`, `repl_diff`, `repl_inspector`, etc.
  - Emit `UIEntityCompleted` for all in-flight entities to freeze their state
- On failure:
  - Emit/complete a `repl_stderr` entity (or result with error styling) with error message and metadata (exit code, duration)

IDs and labels:
- Use `EntityID{TurnID: <eval-id>, LocalID: <sequence>, Kind: <...>}`
- Optionally include `Labels` (e.g., language, evaluator name) to aid filtering.


## 6) REPL Model Integration Patterns

Two viable integration strategies:

1) Non-invasive Bridge (Adapter) around current REPL
- Keep the existing REPL UI intact for input. Replace/augment its transcript printing by emitting timeline events to a co-located timeline viewport above it.
- Pros: minimal changes; preserves existing API; quick path to value.
- Cons: the history list inside REPL becomes redundant; better to phase it out in favor of timeline entities.

2) Timeline-First REPL (Refactor)
- Make the timeline the source of truth for the transcript. The REPL’s `View()` renders only the input line and status bar; all previous outputs come from the timeline controller’s `View()`.
- Pros: single rendering path, richer widgets, selection/focus support out of the box.
- Cons: larger refactor; requires an adapter for evaluators to emit structured/streaming events.

Recommendation: Start with (1), then evolve into (2).


## 7) Evaluator Extensions for Rich Output

The current `Evaluator` returns a single `(string, error)`. To enable structured, streaming outputs:

- Add an optional interface:
  ```go
  type StreamingEvaluator interface {
    EvaluateStream(ctx context.Context, code string, emit func(Event)) error
  }

  // Example event union
  type Event struct {
    Kind string // stdout, stderr, result_markdown, table, diff, progress, tool_call
    Props map[string]any
  }
  ```
- The REPL bridge detects `StreamingEvaluator` and, if present, forwards events to the timeline controller by mapping to `UIEntityCreated/Updated/Completed`.
- For non-streaming evaluators, the bridge wraps the output into a single `repl_result_markdown` entity.

Note: This maintains backward compatibility while enabling structured outputs.


## 8) Renderer/Widget Catalog and Reuse

- Reuse `llm_text` renderer for `repl_result_markdown` immediately (it already supports markdown, code copy, streaming indicator, metadata footer). Set `role` to `"result"` and use `metadata` for timing/exit-code.
- Reuse/extend `tool_calls_panel` for `repl_tool_calls` to summarize steps a REPL/evaluator takes (e.g., shell commands, network calls, sub-invocations).
- Build new renderers incrementally:
  - `repl_stdout` and `repl_stderr`: simple text models with distinct styles and stream awareness.
  - `repl_progress`: single-line progress with spinner and right-aligned metadata (ETA, current/total).
  - `repl_perf`: small sparkline using `pkg/sparkline` to visualize recent eval durations.
  - `repl_table`: compact table; when selected/focused, allow horizontal scrolling or expand-in-place.
  - `repl_diff`: unified diff; expose key to toggle context.
  - `repl_shell_cmd`: header + foldable stdout/stderr sub-sections; keys to copy stdout/stderr separately.

All renderers should implement copy semantics via `EntityCopyTextMsg`/`EntityCopyCodeMsg` where applicable.


## 9) Keyboard UX and Focus Model

Leverage the timeline controller’s selection and entering/focus model:
- Global keys:
  - Up/Down: move selection across entities
  - Enter: toggle “entering” mode; while entering, route keys to the selected entity model
  - Tab/Shift+Tab: allow compact toggles even outside entering
  - c: copy code (controller sends `EntityCopyCodeMsg` to selected entity; renderer chooses best content)
  - y: copy text (send `EntityCopyTextMsg`)
- Entity-level keys (handled inside models when focused):
  - Expand/collapse, switch tabs (stdout/stderr), page through tables, toggle diff context, etc.

The input field remains always accessible; consider a key (e.g., Esc) to exit entity focus and return to input.


## 10) Minimal Implementation Plan

Phase 1: Bridge and Basic Widgets
- Add a new package `pkg/repltimeline` (or integrate into `pkg/repl`) that:
  - Creates a `timeline.Controller` + `timeline.Registry`
  - Registers `llm_text`, `tool_calls_panel`, and `plain`
  - Exposes helpers like `EmitInput(turnID, code)`, `EmitStdout(turnID, chunk)`, `EmitResult(turnID, md, meta)`
- Update the REPL example app to render the timeline viewport above the input line and route keys to both (viewport + input).
- Map evaluation completion to a `repl_result_markdown` using `llm_text` now.

Phase 2: Structured/Streaming Evaluations
- Introduce `StreamingEvaluator` interface and adapt the JS, Shell examples to stream stdout/stderr chunks and progress.
- Implement `repl_stdout`/`repl_stderr` renderers.

Phase 3: Enrich with Widgets
- Add `repl_shell_cmd`, `repl_progress`, `repl_perf` (sparkline), and `repl_table`.
- Provide reusable helpers to convert common evaluator outputs (e.g., CSV/JSON) into tables or inspectors.

Phase 4: Make Timeline the Primary Transcript
- Remove the REPL’s inline history rendering; rely solely on timeline entities.
- Add commands to interact with prior turns (e.g., “rerun this input”, “open in editor”, “copy result”), leveraging selection.


## 11) Data Contracts and Props

Standardize props for cross-renderer consistency:
- Every entity supports `selected` boolean (controller already injects it) and uses a common style system (`chatstyle.DefaultStyles()` is a good baseline).
- Timestamps: prefer `StartedAt` from `UIEntityCreated`. Add `duration_ms` to result/progress entities.
- Copy behavior: renderers implement copy-message handling and choose the best payload (e.g., fenced code blocks for markdown).


## 12) Example: Shell Evaluator → Timeline

Pseudocode for a streaming shell evaluator emitting stdout/stderr and a final result:

```go
func (e *ShellStreamingEvaluator) EvaluateStream(ctx context.Context, code string, emit func(Event)) error {
  turnID := newTurnID()
  emit(Event{Kind: "repl_input", Props: map[string]any{"prompt": e.GetPrompt(), "code": code, "cwd": e.workingDir}})

  // Create streams lazily on first chunk
  var outCreated, errCreated bool
  var start = time.Now()

  cmd := exec.CommandContext(ctx, "bash", "-c", code)
  // wire stdout/stderr readers...
  go readLines(cmd.Stdout, func(line string) {
    if !outCreated { emit(Event{Kind: "repl_stdout", Props: map[string]any{"text": "", "streaming": true}}); outCreated = true }
    emit(Event{Kind: "repl_stdout", Props: map[string]any{"append": line + "\n"}})
  })
  go readLines(cmd.Stderr, func(line string) {
    if !errCreated { emit(Event{Kind: "repl_stderr", Props: map[string]any{"text": "", "streaming": true}}); errCreated = true }
    emit(Event{Kind: "repl_stderr", Props: map[string]any{"append": line + "\n"}})
  })

  err := cmd.Run()
  dur := time.Since(start)
  exitCode := 0
  if err != nil { exitCode = extractExitCode(err) }

  emit(Event{Kind: "repl_result_markdown", Props: map[string]any{
    "markdown": fmt.Sprintf("Exit: %d\nDuration: %s", exitCode, dur),
    "metadata": map[string]any{"exit_code": exitCode, "duration_ms": dur.Milliseconds()},
  }})
  return err
}
```

The bridge converts Events to timeline messages, maintaining a consistent TurnID for grouping.


## 13) Theming and Cohesion

- Use `chatstyle` as a base for entities to ensure consistent borders, padding, and selection/focus states.
- Expose a REPL theme that maps to entity styles (e.g., `stdout` vs `stderr` colors).
- Propagate global theme via `Controller.SetTheme()` and have models react by updating their styles.


## 14) Testing and Debugging

- Leverage the timeline demo’s logging strategy to inspect message flow.
- Use the `plain` renderer during development to validate props and lifecycle.
- Add a debug command `/dump` to emit a `plain` entity with the current evaluator context.


## 15) Risks and Mitigations

- Complexity creep: Start with the bridge, reuse existing renderers, and grow incrementally.
- Streaming backpressure: Use buffered channels and coalesce updates (e.g., batch stdout lines).
- Width/height handling: Controller already dispatches `EntitySetSizeMsg`; ensure renderers wrap content.
- Copy behavior variance: Standardize copy semantics across renderers and document keybindings.


## 16) Summary

- The timeline architecture is a natural fit for a modern REPL: append-only turns, interactive entities, and rich widgets.
- Begin with a bridge that surfaces REPL interactions as timeline entities using existing `llm_text` and `tool_calls_panel` renderers.
- Introduce streaming and structured outputs via an optional `StreamingEvaluator` interface.
- Gradually adopt timeline as the primary transcript, enabling selection, focus, copy, and advanced widgets (progress, tables, diffs, shell panels, inspectors, perf sparklines).

This approach keeps today’s REPL usable while unlocking a clear path to a significantly richer, extensible, and composable user experience.

