# Using Watermill for a Timeline‑Based REPL Design

This document analyzes Pinocchio’s simple‑chat‑agent (including its UI and backend), then proposes a Watermill‑based architecture for Bobatea’s REPL/timeline. The current channel‑heavy coupling in the REPL becomes difficult to reason about and scale; moving to a router/pub‑sub topology (Watermill) simplifies orchestration, decouples concerns, and provides a path to distributed operation (in‑memory or Redis Streams).


## 1) Reference: Pinocchio simple‑chat‑agent

Files reviewed:
- `pinocchio/cmd/agents/simple-chat-agent/main.go`
- `pinocchio/cmd/agents/simple-chat-agent/pkg/backend/tool_loop_backend.go`
- `pinocchio/cmd/agents/simple-chat-agent/pkg/ui/*`
- `pinocchio/pkg/redisstream/*`

### 1.1 Router + Sink + Topics
- Builds an `events.EventRouter` via `rediscfg.BuildRouter(settings, verbose)` that supports:
  - In‑memory (Watermill) when Redis disabled
  - Redis Streams publisher/subscriber when enabled (with group isolation support)
- Topics logically separated:
  - `chat` — engine/middleware/tool events
  - `ui` — UI forwarding (from router to Bubble Tea program)
  - `logs` — logging/persistence subscribers
- `middleware.NewWatermillSink(router.Publisher, "chat")` provides an event sink for engine and middlewares.

### 1.2 Backend: Event → Timeline Forwarder
- `ToolLoopBackend.MakeUIForwarder(p *tea.Program)` returns a Watermill handler that:
  - Parses Geppetto events (`EventPartialCompletion*`, `EventFinal`, `EventTool*`, `EventError`, …)
  - Translates to `timeline.UIEntity{Created,Updated,Completed}` and calls `p.Send(...)`
  - Emits dedicated renderers: `llm_text`, `tool_call`, `tool_call_result`, `log_event`, `agent_mode`
  - Uses stable local IDs and per‑event metadata to keep updates idempotent
- Backend runs the tool loop (`RunToolCallingLoop`) and sets `BackendFinishedMsg` when done; forwarder does not prematurely finish the backend.

### 1.3 UI Layer
- `pkg/ui/app.go` composes:
  - A REPL component for input/history
  - A viewport to show transcript (or an overlay for interactive forms)
  - A sidebar for mode/tool diagnostics, toggled at runtime
- Integration patterns:
  - Channel forwarder example: `xevents.AddUIForwarder(router, uiCh)` pushes events to a channel the model reads. This is optional — the primary path uses the backend forwarder to send timeline messages directly to `tea.Program`.
  - Tool forms integration via overlay model (generative UI pattern). Forms are fed via channel; replies return over channels.
- Logging/persistence handlers subscribe to `chat` and write to SQLite (and stdout pretty‑print if enabled).

### 1.4 Redis Streams Utilities
- `rediscfg.BuildRouter` returns a router with Redis publisher/subscriber or in‑memory fallback.
- `BuildGroupSubscriber` and `EnsureGroupAtTail` let you isolate handlers (e.g., UI vs logs) and avoid full backlog replay.


## 2) Why Watermill for Bobatea REPL

Problems with channel‑heavy approach:
- Tight coupling: producers and consumers depend on each other’s concrete channels and lifetimes.
- Scaling pain: when adding more features (copy/paste, structured output, tables, progress bars, file ops, logs, persistence), it’s hard to route and fan‑out safely.
- Backpressure and ordering: manual channel management risks deadlocks or dropped events without visibility.

Watermill advantages:
- Decoupling via topics: produce once, subscribe many (UI, logs, persistence, test injectors).
- Swappable transport: in‑memory by default; Redis Streams for durability and multi‑process.
- Handler composition: add or remove subscribers (UI forwarder, logger, metrics, recorder) with no code changes in producers.
- Isolation: consumer groups to isolate UI vs logging; replay control with `EnsureGroupAtTail`.


## 3) Proposed Bobatea Architecture (REPL + Timeline)

### 3.1 Topics and Message Contracts
- Topics:
  - `ui.entities` — carries timeline lifecycle messages:
    - `UIEntityCreated{ID, Renderer, Props, StartedAt}`
    - `UIEntityUpdated{ID, Patch, Version, UpdatedAt}`
    - `UIEntityCompleted{ID, Result}`
    - `UIEntityDeleted{ID}`
  - `ui.logs` — optional log/info events (tool logs, info panels)
  - `repl.events` — optional structured REPL events (stdout/stderr/progress) if you keep a semantic layer above timeline
- ID conventions:
  - `EntityID{RunID, TurnID, LocalID, Kind}` — use Turn‑scoped LocalIDs and stable Kind for idempotence.
  - Version monotonicity for `Updated` to allow de‑dupe.

Encoding:
- JSON payloads with a `type` field, e.g.:
  - `{ "type": "timeline.created", ... }`
  - `{ "type": "timeline.updated", ... }`

### 3.2 REPL → Router → Timeline
- REPL submits code; an adapter publishes events:
  - Option A (direct): publish timeline lifecycle messages to `ui.entities` — entity renderers handle display.
  - Option B (semantic): publish `repl.events` (stdout/stderr/result/progress) and run a transformer subscriber to map them to `ui.entities`.
- UI subscriber:
  - In‑process `tea.Program` handler that reads `ui.entities` and `p.Send(...)` timeline messages (akin to Pinocchio’s `MakeUIForwarder`).
  - Alternatively, a channel forwarder that the REPL model consumes and translates to timeline calls.

### 3.3 Transport and Router
- Build router with in‑memory transport by default:
  - `events.NewEventRouter(WithVerbose(...))`
- Optional Redis Streams:
  - `rediscfg.BuildRouter(settings, verbose)` and `BuildGroupSubscriber()` for multiple independent subscribers.
- Subscribers:
  - `ui-forward` (program forwarder) on `ui.entities`
  - `event-logger` on `ui.entities` and/or `repl.events`
  - `event-persist` on `ui.entities` (e.g., SQLite transcripts)

### 3.4 Entity Renderers and Timeline Controller
- Reuse Bobatea’s timeline renderers (`text`, `markdown`, `structured_data`, etc.).
- Keep the REPL transcript as a timeline Shell (viewport) so that UI remains responsive.
- Ensure properties:
  - Coalesce frequent updates (stdout/stderr) at the publisher or transformer level to reduce UI churn.
  - Stable IDs and versions for idempotent updates.


## 4) Detailed Wiring (Pseudo‑code)

### 4.1 Setup
```go
// Build router (memory or Redis based on config)
router, err := buildRouterFromConfig(cfg)
uiTopic := "ui.entities"

// Program + UI forwarder
p := tea.NewProgram(appModel)
router.AddHandler("ui-forward", uiTopic, makeUIForwarder(p))

// Optional: logs and persistence
router.AddHandler("ui-logger", uiTopic, logHandler)
router.AddHandler("ui-persist", uiTopic, persistHandler)

// Start router and UI
ctx, cancel := context.WithCancel(context.Background())
eg, ctx := errgroup.WithContext(ctx)
eg.Go(func() error { return router.Run(ctx) })
eg.Go(func() error { _, err := p.Run(); cancel(); return err })
if err := eg.Wait(); err != nil { log.Error().Err(err).Msg("exit") }
```

### 4.2 REPL Publisher (Direct timeline messages)
```go
func publishTimelineCreated(pub message.Publisher, e timeline.UIEntityCreated) error {
    b, _ := json.Marshal(struct{ Type string; timeline.UIEntityCreated }{"timeline.created", e})
    return pub.Publish("ui.entities", message.NewMessage(watermill.NewUUID(), b))
}
```

### 4.3 UI Forwarder (Program handler)
```go
func makeUIForwarder(p *tea.Program) func(msg *message.Message) error {
    return func(msg *message.Message) error {
        defer msg.Ack()
        var env struct{ Type string; Payload json.RawMessage }
        if err := json.Unmarshal(msg.Payload, &env); err != nil { return err }
        switch env.Type {
        case "timeline.created":
            var e timeline.UIEntityCreated; _ = json.Unmarshal(env.Payload, &e)
            p.Send(e)
        case "timeline.updated":
            var e timeline.UIEntityUpdated; _ = json.Unmarshal(env.Payload, &e)
            p.Send(e)
        case "timeline.completed":
            var e timeline.UIEntityCompleted; _ = json.Unmarshal(env.Payload, &e)
            p.Send(e)
        case "timeline.deleted":
            var e timeline.UIEntityDeleted; _ = json.Unmarshal(env.Payload, &e)
            p.Send(e)
        }
        return nil
    }
}
```

### 4.4 Transformer (Optional: REPL events → timeline)
```go
// Subscribes to repl.events; publishes mapped timeline messages to ui.entities
router.AddHandler("repl-to-ui", "repl.events", func(msg *message.Message) error {
    defer msg.Ack()
    ev := parseReplEvent(msg.Payload)
    out := mapReplEventToTimeline(ev)
    return pub.Publish("ui.entities", out...)
})
```


## 5) Operational Considerations

- Backpressure:
  - Use buffered topics and coalescing; drop non‑critical updates (e.g., logs) when the UI is saturated.
  - Watermill provides handler concurrency; keep UI forwarder single‑threaded to preserve ordering.
- Idempotence:
  - Stable IDs and monotonic versions for `Updated` messages; subscribers may reprocess after failures.
- Consumer groups (Redis):
  - UI, logging, and persistence each with their own group to prevent interference.
  - Use `EnsureGroupAtTail` for first‑run to avoid replay storms.
- Error handling:
  - Always Ack messages; log parse/mapping issues separately.
- Security/sandboxing:
  - If evaluators run untrusted code, isolate publishers so UI can’t be blocked by evaluator issues.


## 6) Incremental Plan for Bobatea

Phase 1 — Foundations
- Add a small `pkg/eventbus` wrapper:
  - Build in‑memory router by default; enable Redis via config.
  - Provide `Publisher()` and helpers for timeline lifecycle messages with `{Type,Payload}` envelopes.
- Add a `ui-forward` handler to send timeline messages to the Bubble Tea program.

Phase 2 — REPL Integration
- Replace channel coupling with a REPL publisher:
  - Directly publish `UIEntity{Created,Updated,Completed,Deleted}` to `ui.entities` during evaluation.
  - Or publish `repl.events` and include a transformer subscriber.
- Keep the timeline Shell as the transcript viewport. Ensure refresh is debounced and coalesced for heavy output.

Phase 3 — Observability
- Add `ui-logger` and `ui-persist` handlers.
- Optional: pretty console sink (like pinocchio’s `xevents.AddPrettyHandlers`) for CLI debugging.

Phase 4 — Redis Streams (Optional)
- Wire `rediscfg.BuildRouter` and per‑handler `BuildGroupSubscriber`.
- Use `EnsureGroupAtTail` for `ui.entities` and `repl.events` streams.


## 7) Mapping Examples

- stdout chunk → `timeline.updated{ id: {turn, local: "stdout", kind: "text"}, patch: {append: "..."}}`
- stderr → same as stdout with `is_error: true`
- final result markdown → `timeline.created{ id: {local:"result-<n>", kind:"markdown"}, props:{markdown:"..."}}` followed by `timeline.completed`.
- structured JSON result → `structured_data` renderer with `data` or `json` prop.


## 8) Why this fixes our scaling issues

- Replaces bespoke channels with explicit topics; adding features is additive (new handlers).
- Decouples evaluator lifecycles from UI rendering; retry and replay semantics defined by the router.
- Makes multi‑process distribution straightforward: evaluators can run elsewhere, UI stays local.
- Preserves rich timeline rendering with minimal changes to the REPL core.


## 9) Action Items

- Implement `pkg/eventbus` with Watermill router setup (memory + Redis).
- Implement `pkg/ui/forwarder` for `ui.entities` → `tea.Program.Send(...)` (mirroring Pinocchio’s backend forwarder).
- Update the REPL to publish timeline messages instead of calling controller directly.
- Add a demo wiring (like `cmd/repl-timeline-bus-demo`) showing:
  - REPL input → publishes events
  - UI forwarder → applies to timeline
  - Optional logger/persist subscribers

---

Pinocchio’s simple‑chat‑agent demonstrates a clean, scalable pattern: engines + middlewares publish to a router; specialized handlers forward to the timeline UI, log, and persist — all without entangling components. Adopting Watermill for Bobatea’s REPL will simplify our architecture, improve resilience, and make advanced features (tools, inspectors, progress, logs) easier to integrate and maintain.

