## Migration Plan: Move Chat View to the Timeline Framework

Date: 2025-08-11

### Purpose

Migrate the existing conversation-centric chat UI to the new turn-centric Timeline framework, while preserving current UX (header, boxed messages, scrolling, selection, keybindings) and enabling richer entity-based rendering.

### Current state (key files)

- `bobatea/pkg/chat/model.go`: Top-level Bubble Tea model for chat UI. Owns `viewport`, `textarea`, selection/scrolling, and an embedded `conversationui.Model`.
- `bobatea/pkg/chat/conversation/model.go`: Renders conversation messages (markdown + cache) and reacts to streaming messages (`StreamStartMsg`, `StreamCompletionMsg`, `StreamDoneMsg`, `StreamCompletionError`).
- `bobatea/pkg/chat/conversation/style.go`: Defines `DefaultStyles()` and the boxed look using `lipgloss` borders.
- `bobatea/pkg/chat/backend.go`: Declares `Backend` interface and the message types used to drive the UI.

New framework components:
- `bobatea/pkg/timeline/types.go`: `EntityID`, `RendererDescriptor`, `UIEntityCreated/Updated/Completed/Deleted`.
- `bobatea/pkg/timeline/registry.go`: Renderer registry.
- `bobatea/pkg/timeline/store.go`: Append-only entity store.
- `bobatea/pkg/timeline/cache.go`: Per-entity render cache.
- `bobatea/pkg/timeline/controller.go`: Applies lifecycle messages, renders via registry, manages size.
- `bobatea/pkg/timeline/renderers.go`: Minimal `llm_text` and `tool_calls_panel` renderers. `LLMTextRenderer` now uses chat-style boxes.
- `bobatea/cmd/timeline-demo/pkg/chatstyle/`: Extracted chat styling (`DefaultStyles` + `RenderBox`).

### Goals

- Replace `conversationui.Model` view with Timeline rendering for message content.
- Map streaming messages to `UIEntity*` lifecycle events.
- Preserve boxed styling (system/user/assistant) and metadata placement.
- Maintain scroll-to-bottom behavior, selection navigation, and keymap.

### Migration approach

Replace `conversationui.Model` rendering with a `timeline.Controller` directly inside `bobatea/pkg/chat/model.go` (no flag/backwards compatibility). Reuse `viewport`, `textarea`, keymap, and status updates; translate streaming messages to `UIEntity*` events in-place (inside `chat.model.Update`).

### Detailed plan (phased)

Phase A — Replace Conversation UI with Timeline (no flag)
1. Remove the `conversationui.Model` usage in `bobatea/pkg/chat/model.go`.
2. Instantiate a `timeline.Controller` and `timeline.Registry` in `InitialModel(...)`.
   - Register renderers: `LLMTextRenderer`, `ToolCallsPanelRenderer`.
   - Keep existing `viewport.Model`, `textarea.Model`, keymap, and status updates.
3. In `tea.WindowSizeMsg`, call `ctrl.SetSize(m.width, m.height)` and update `viewport` content with `ctrl.View()`.
4. In `View()`, compose header + `viewport.View()`; the viewport content is always `ctrl.View()`.

Phase B — Translate streaming messages to Timeline lifecycle
1. In `bobatea/pkg/chat/model.go`, inside `Update(...)` (stream section):
   - For `conversationui.StreamStartMsg`: emit `timeline.UIEntityCreated` with `EntityID{LocalID: msg.ID.String(), Kind: "llm_text"}` and `Props={"role":"assistant","text":""}`.
   - For `conversationui.StreamCompletionMsg`: emit `timeline.UIEntityUpdated{ID:..., Patch={"text": msg.Completion}, Version:inc, UpdatedAt: time.Now()}`.
   - For `conversationui.StreamDoneMsg`: emit `timeline.UIEntityCompleted{ID:..., Result={"text": msg.Completion}}`.
   - For `conversationui.StreamCompletionError`: emit `timeline.UIEntityCompleted` with `Result={"text": fmt.Sprintf("**Error**\\n\\n%s", msg.Err)}` and optionally a dedicated `error` entity.
2. Call `ctrl.OnCreated/OnUpdated/OnCompleted` accordingly and refresh the viewport content.
3. Stop appending/updating the conversation manager for assistant messages when `useTimeline` is on (or gate it with a flag) to avoid double content.
4. For user-submitted text (in `submit()`): also emit a `UIEntityCreated+Completed` with `Kind: "llm_text"` and `Props={"role":"user","text": userMessage}` so the user box is shown by the timeline.
5. For existing system prompt: at init time, emit a single `UIEntityCreated+Completed` with `role: "system"` and initial system message (from the conversation manager seed).

Phase C — Preserve chat styling and metadata
1. Use `cmd/timeline-demo/pkg/chatstyle.DefaultStyles()` inside `LLMTextRenderer` (already in place) to render boxed messages in timeline.
2. Add metadata support in renderer (optional): if `props["metadata"]` exists, append a right-aligned metadata line using `MetadataStyle` (similar to `conversation/model.go`).
3. Ensure width computations mirror chat: call controller `SetSize(...)` from the existing `recomputeSize()` to keep layout consistent.

Phase D — Scrolling, selection, and keymap alignment
1. Reuse existing scrolling logic around `viewport.GotoBottom()` during streaming when `m.scrollToBottom` is true.
2. Selection: initially, selection can remain managed by chat model (message-level selection). Later, add per-entity selection by tracking `store.order` index inside controller and exposing it for navigation.
3. Keymap: no changes required initially. Later, map select next/prev to controller selection and use a highlighting style via `chatstyle`.

Phase E — Clean-up
1. Delete conversation-based rendering code paths; the timeline is now the sole renderer.
2. Keep `conversation.Manager` for persistence and user message capture if needed, or progressively migrate persistence to Turn/Entities.

### Mapping summary

- StreamStartMsg → UIEntityCreated{Kind: llm_text, role: assistant, text: ""}
- StreamCompletionMsg(delta/complete) → UIEntityUpdated{Patch: {text: completion}}
- StreamDoneMsg(final) → UIEntityCompleted{Result: {text: final}}
- StreamCompletionError → UIEntityCompleted{Result: {text: formatted error}}
- User submission → UIEntityCreated+Completed{Kind: llm_text, role: user, text}
- Initial system → UIEntityCreated+Completed{Kind: llm_text, role: system, text}

### Code change points (by file)

- `bobatea/pkg/chat/model.go`
  - Add `useTimeline` flag + functional option `WithTimeline(true)`.
  - Add fields: `timelineCtrl *timeline.Controller`, `timelineReg *timeline.Registry`.
  - Init: when using timeline, register renderers and seed system/user entities.
  - Update (stream messages): emit `UIEntity*` and call `timelineCtrl.On...`.
  - WindowSizeMsg: call `timelineCtrl.SetSize(m.width, m.height)` and set `viewport` content to `timelineCtrl.View()`.
  - View(): prefix header + `viewport.View()`.

- `bobatea/pkg/timeline/renderers.go`
  - Extend `LLMTextRenderer` to support metadata props (optional) with `MetadataStyle`.

- `bobatea/pkg/timeline/chatstyle/`
  - Use `DefaultStyles()` and `RenderBox()` (moved from the demo) for consistent chat-like boxes.

### Testing strategy

- Unit test a thin adapter that maps `Stream*Msg` to `UIEntity*` (table tests for roles and text content).
- Run the app with timeline on and fake backend:
  - Verify immediate header render.
  - Press submit; ensure user message appears as boxed entity.
  - Trigger fake backend streaming; ensure assistant entity updates incrementally.
- Inspect `/tmp/fake-chat.log` (existing) and `/tmp/timeline-demo.log` (if reused) to confirm render cycles, cache hit/miss, widths.

### Rollout plan

1. Ship behind a hidden flag `WithTimeline(true)`.
2. Enable for developers by default in demo binaries.
3. Gather feedback, fix regressions (scrolling, selection edge cases, metadata lines).
4. Remove conversation-based rendering.

### Risks and mitigations

- Double rendering or mismatched sources (conversation vs timeline):
  - Gate conversation updates (assistant messages) behind the timeline flag.
- Width discrepancies causing wrapping issues:
  - Align `SetSize(...)` calls; ensure header line reservation is consistent.
- Performance regressions:
  - Rely on per-entity cache; add debounce for frequent partial updates if needed.

### TODO checklist

- [ ] Add `WithTimeline(bool)` option and wire `timeline.Controller` in `pkg/chat/model.go`
- [ ] Translate `Stream*Msg` to `UIEntity*` inside chat model
- [ ] Emit system/user seed entities
- [ ] Update `WindowSizeMsg` and `View()` to use timeline output
- [ ] Validate styling parity with `chatstyle.RenderBox`
- [ ] Add metadata rendering in `LLMTextRenderer` (optional)
- [ ] Tests for mapping and rendering
- [ ] Enable flag in demo binary
- [ ] Remove conversation view path after parity


