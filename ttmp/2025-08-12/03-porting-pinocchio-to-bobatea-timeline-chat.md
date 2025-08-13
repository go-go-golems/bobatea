---
Title: Porting Pinocchio to the Bobatea Timeline Chat Framework
Slug: porting-pinocchio-to-timeline-chat
Short: How to migrate Pinocchio’s chat stack (main, chatrunner, engine backend) to the new timeline-first chat UI
Topics:
- pinocchio
- chat
- timeline
- migration
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Purpose

Pinocchio currently wires a full chat experience around a conversation-centric UI and stream messages. This guide explains how to port it to the new Bobatea timeline chat framework that uses timeline entities and a simplified backend API Start(ctx, prompt string). The goal is to eliminate the legacy conversation coupling and leverage interactive per-entity models.

## Current integration overview (Pinocchio)

- Entry point initializes commands and help system, not the chat UI directly:

```60:76:pinocchio/cmd/pinocchio/main.go
func main() {
    // ... initRootCmd / initAllCommands ...
    err := rootCmd.Execute()
}
```

- The runtime chat orchestration lives in the chat runner. It creates an Engine, sets up an EventRouter and builds the Bobatea UI:

```100:126:pinocchio/pkg/chatrunner/chat_runner.go
eg.Go(func() error {
    // Engine and router are set up above
    backend := ui.NewEngineBackend(engine, uiSink)
    model := bobachat.InitialModel(cs.manager, backend, cs.uiOptions...)
    p := tea.NewProgram(model, cs.programOptions...)
    router.AddHandler("ui", "ui", ui.StepChatForwardFunc(p))
    // run handlers then p.Run()
})
```

- The backend bridges Engine events to the old chat widget by translating events.Event* to conversation Stream* messages:

```103:168:pinocchio/pkg/ui/backend.go
func StepChatForwardFunc(p *tea.Program) func(msg *message.Message) error {
    // parse events and send conversation Stream* messages to p
    switch e := e.(type) {
    case *events.EventPartialCompletionStart:
        p.Send(conversation2.StreamStartMsg{ /* ... */ })
    case *events.EventPartialCompletion:
        p.Send(conversation2.StreamCompletionMsg{ /* ... */ })
    case *events.EventFinal:
        p.Send(conversation2.StreamDoneMsg{ /* ... */ })
    // ...
    }
}
```

## Gaps vs. new framework

- Backend API changed to Start(ctx, prompt string) and timeline-only UI events.
- The chat UI no longer depends on conversation Stream* messages; it expects either:
  - Direct timeline.UIEntity* lifecycle messages from the backend, or
  - The chat model itself creating a user entity and the backend streaming the assistant entity via timeline updates.
- bobachat.InitialModel no longer takes a conversation manager.

## Migration plan (high level)

1) Replace conversation Stream* with timeline lifecycle in the UI bridge.
2) Update EngineBackend to implement the new Start(ctx, prompt string) and emit timeline entities (plus BackendFinishedMsg).
3) Update ChatRunner to construct the chat model with the new signature and drop the manager requirement for UI.
4) Keep the EventRouter pattern (good separation), but forward Engine events to timeline instead of conversation.
5) Optionally remove any conversation tree logic from runtime paths; keep only if needed for non-UI processing.

## Detailed changes

### 1) Chat runner: construct the new chat model

- Current code passes a manager into bobachat.InitialModel, which no longer exists. Switch to the new signature:

```118:122:pinocchio/pkg/chatrunner/chat_runner.go
-model := bobachat.InitialModel(cs.manager, backend, cs.uiOptions...)
+model := bobachat.InitialModel(backend, cs.uiOptions...)
```

- Keep program options and router logic as-is. The forwarding handler remains, but it should now emit timeline messages (see next section).

### 2) Backend Start signature and streaming

- Implement Start(ctx context.Context, prompt string) (tea.Cmd, error) in EngineBackend:

Pseudo-adapter (replace the Stream* send calls with timeline UI events):

```go
func (e *EngineBackend) Start(ctx context.Context, prompt string) (tea.Cmd, error) {
    if e.isRunning { return nil, errors.New("Engine is already running") }
    ctx, cancel := context.WithCancel(ctx)
    e.cancel = cancel
    e.isRunning = true
    return func() tea.Msg {
        // Kick off engine workflow associated with the prompt (seed turn construction as needed)
        // Engine will publish events → router → StepChatForwardFunc → p.Send(UIEntity*)
        return nil
    }, nil
}
```

### 3) Event forwarding: conversation → timeline

- Rewrite StepChatForwardFunc to send timeline.UIEntity* for LLM text, not conversation Stream* messages.
- Strategy: maintain an accumulator per event stream ID and map engine events to a single llm_text entity lifecycle.

Event mapping (per ID):
- On first partial start: send UIEntityCreated{Kind:"llm_text", Props:{role:"assistant", text:""}}.
- On partial deltas: update accumulated text and send UIEntityUpdated{Patch:{text: acc}, Version: n}.
- On final or interrupt: UIEntityCompleted{Result:{text: acc}} and optionally send boba_chat.BackendFinishedMsg{} at the end of a run.
- On error: mark completed with an error text and finish.

### 4) Remove conversation dependencies from UI path

- Eliminate imports of bobatea/pkg/chat/conversation from both chatrunner and ui/backend.go once the forwarding is timeline-based.
- The chat model already renders messages via the timeline controller; no conversation style or cache is needed.

### 5) Managing Turns and Blocks in the Backend (UI no longer owns conversation)

With the UI decoupled from conversation management, the backend is responsible for building and maintaining the Turn/Block state. Use Geppetto’s Turn-based model to seed user input, run inference, stream events, and (optionally) persist history. See also: geppetto docs on Turns and Engines.

Packages to import:

```go
import (
    "context"
    "time"
    // Geppetto core
    "github.com/go-go-golems/geppetto/pkg/turns"
    "github.com/go-go-golems/geppetto/pkg/inference/engine"
    "github.com/go-go-golems/geppetto/pkg/inference/engine/factory"
    "github.com/go-go-golems/geppetto/pkg/inference/middleware"
    "github.com/go-go-golems/geppetto/pkg/events"
    // Bobatea timeline messages
    "github.com/go-go-golems/bobatea/pkg/timeline"
)
```

Backend responsibilities:

- Create/maintain a Turn per run. Append a user-text Block on submit.
- Initialize an Engine (via factory) and attach an event sink for streaming.
- Translate Engine events to timeline entities for the UI, while also updating Turn Blocks.
- Optionally persist Turns in memory/disk to reconstruct history or export later.

Minimal backend run skeleton:

```go
type RunState struct {
    Turn       *turns.Turn
    AccText    string
    Version    int64
    EntityID   string // for timeline llm_text mapping
}

type BackendImpl struct {
    // keyed by stream/run id
    runs map[string]*RunState
    // engine factory and router/sink setup
}

func (b *BackendImpl) Start(ctx context.Context, prompt string) (tea.Cmd, error) {
    id := newID() // stable id for this streaming run
    // Seed a Turn with the user message
    t := &turns.Turn{}
    turns.AppendBlock(t, turns.NewUserTextBlock(prompt))

    // Prepare engine with streaming sink (Watermill or other)
    sink := middleware.NewWatermillSink(b.router.Publisher, "ui")
    e, err := factory.NewEngineFromParsedLayers(b.parsedLayers, engine.WithSink(sink))
    if err != nil { return nil, err }

    b.runs[id] = &RunState{Turn: t, AccText: "", Version: 0, EntityID: id}

    // Emit entity created to UI immediately (assistant placeholder)
    b.program.Send(timeline.UIEntityCreated{
        ID:       timeline.EntityID{LocalID: id, Kind: "llm_text"},
        Renderer: timeline.RendererDescriptor{Kind: "llm_text"},
        Props:    map[string]any{"role": "assistant", "text": ""},
        StartedAt: time.Now(),
    })

    // Kick off the engine; events will be forwarded (see StepChatForwardFunc below)
    return func() tea.Msg {
        _, _ = e.RunInference(ctx, t)
        return nil
    }, nil
}

// In the event forwarder (router handler), update Turn + send timeline events
func (b *BackendImpl) onEvent(ev events.Event) {
    md := ev.Metadata(); id := md.ID.String()
    st, ok := b.runs[id]
    if !ok { return }
    switch e := ev.(type) {
    case *events.EventPartialCompletionStart:
        // No turn change yet; assistant message begins
    case *events.EventPartialCompletion:
        st.AccText += e.Delta
        st.Version++
        // Reflect in Turn (append or update an assistant text block)
        turns.AppendBlock(st.Turn, turns.NewAssistantTextBlock(st.AccText))
        // Update UI entity
        b.program.Send(timeline.UIEntityUpdated{
            ID:      timeline.EntityID{LocalID: st.EntityID, Kind: "llm_text"},
            Patch:   map[string]any{"text": st.AccText},
            Version: st.Version,
            UpdatedAt: time.Now(),
        })
    case *events.EventFinal, *events.EventInterrupt:
        // Finalize Turn (ensure last assistant text is present)
        // Complete UI entity
        b.program.Send(timeline.UIEntityCompleted{
            ID:     timeline.EntityID{LocalID: st.EntityID, Kind: "llm_text"},
            Result: map[string]any{"text": st.AccText},
        })
        b.program.Send(bobatea.BackendFinishedMsg{})
    case *events.EventError:
        // Mark error in Turn if desired, and complete UI as error text
        b.program.Send(timeline.UIEntityCompleted{
            ID:     timeline.EntityID{LocalID: st.EntityID, Kind: "llm_text"},
            Result: map[string]any{"text": "**Error**\n\n" + e.ErrorString},
        })
        b.program.Send(bobatea.BackendFinishedMsg{})
    }
}
```

Tool calling (optional):

- Attach a Turn-scoped tool registry/config in `Turn.Data` before calling `RunInference`.
- Let middleware handle `tool_call`/`tool_use` and update `Turn.Blocks` accordingly.
- Optionally mirror tool blocks to the UI as timeline `tool_call` entities for a richer experience.

Persistence options:

- In-memory: keep a slice/log of `*turns.Turn` per session.
- Export: on demand, convert to a conversation (`turns.BuildConversationFromTurn`) or serialize your Turn and Blocks directly.

## Example diffs (pseudocode)

- ChatRunner UI construction:

```118:122:pinocchio/pkg/chatrunner/chat_runner.go
-model := bobachat.InitialModel(cs.manager, backend, cs.uiOptions...)
+model := bobachat.InitialModel(backend, cs.uiOptions...)
```

- StepChatForwardFunc timeline mapping (sketch):

```go
func StepChatForwardFunc(p *tea.Program) func(msg *message.Message) error {
    return func(msg *message.Message) error {
        e, _ := events.NewEventFromJson(msg.Payload)
        md := e.Metadata(); id := md.ID
        switch e := e.(type) {
        case *events.EventPartialCompletionStart:
            p.Send(timeline.UIEntityCreated{ID: timeline.EntityID{LocalID: id.String(), Kind: "llm_text"},
                Renderer: timeline.RendererDescriptor{Kind: "llm_text"}, Props: map[string]any{"role":"assistant","text":""}, StartedAt: time.Now()})
        case *events.EventPartialCompletion:
            // acc[id] += e.Delta
            p.Send(timeline.UIEntityUpdated{ID: timeline.EntityID{LocalID: id.String(), Kind: "llm_text"}, Patch: map[string]any{"text": /* acc */}, Version: /* inc */, UpdatedAt: time.Now()})
        case *events.EventFinal, *events.EventInterrupt:
            p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: id.String(), Kind: "llm_text"}, Result: map[string]any{"text": /* acc */}})
            p.Send(boba_chat.BackendFinishedMsg{})
        case *events.EventError:
            p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: id.String(), Kind: "llm_text"}, Result: map[string]any{"text": "**Error**\n\n" + e.ErrorString}})
            p.Send(boba_chat.BackendFinishedMsg{})
        }
        return nil
    }
}
```

## Considerations

- Router lifecycle remains unchanged; only message payloads to the UI change.
- If Pinocchio still needs a stored conversation for non-UI flows (e.g., exporting), it can maintain it independently of the chat UI.
- Ensure the backend respects the “single stream at a time” invariant used by the chat model (ignore submits while streaming). The chat widget already guards this.

## Carrying conversation context across Turns (critical)

Turns are a normalized, provider-agnostic representation of interaction state. To continue a cohesive conversation from one step to the next, you must carry forward all prior Blocks when constructing the next Turn. Practically, this means:

- On each new user action (submit), build the next Turn by starting from the previous Turn’s Blocks, then append the new user Block.
- Engines read this full Turn and produce additional Blocks (e.g., `llm_text`, `tool_call`, `tool_use`).
- After the engine completes, persist the updated Turn (or merge the new Blocks back into your long-lived Turn state) so the next iteration includes the full context again.

Two common patterns:

1) Rolling Turn (single growing Turn)

```go
// Keep one Turn that accumulates Blocks over time
type Session struct{ Current *turns.Turn }

// On each prompt:
turn := clone(session.Current)              // shallow copy slice and metadata as needed
turns.AppendBlock(turn, turns.NewUserTextBlock(userPrompt))
updated, _ := engine.RunInference(ctx, turn)
session.Current = updated                   // now contains prior + newly appended Blocks
```

2) Turn log + reduce (append immutable Turns, then derive current Blocks)

```go
type Session struct{ History []*turns.Turn }

func reduce(history []*turns.Turn) *turns.Turn {
    out := &turns.Turn{}
    for _, t := range history { turns.AppendBlocks(out, t.Blocks...) }
    return out
}

// On prompt:
seed := reduce(session.History)
turns.AppendBlock(seed, turns.NewUserTextBlock(userPrompt))
updated, _ := engine.RunInference(ctx, seed)
session.History = append(session.History, updated)
```

UI implications:

- The backend should maintain and update this Turn (or the reduced Turn) so timeline entities reflect the whole conversation as it evolves.
- When starting the chat UI mid-session, seed the backend with the fully reduced Turn so the timeline contains prior assistant/user/tool Blocks.

Pitfalls and tips:

- Do not drop prior Blocks, or the provider will lose context and replies will drift.
- Avoid duplicating Blocks. If you capture streaming partials as a growing assistant text, either patch a single `llm_text` entity (and update the corresponding Block’s text) or replace the last assistant text Block; don’t append a fresh assistant text Block per token.
- Tool workflows: carry both `tool_call` and `tool_use` Blocks forward between iterations so the provider can see prior tool results.

## Checklist (migration tasks)

- [x] Update `pinocchio/pkg/chatrunner/chat_runner.go` to use `bobachat.InitialModel(backend, ...)`.
- [x] Change `pinocchio/pkg/ui/backend.go` EngineBackend to implement `Start(ctx, prompt string)`.
- [x] Rewrite `StepChatForwardFunc` to emit `timeline.UIEntity*` and `boba_chat.BackendFinishedMsg`, remove conversation sends.
- [ ] Remove `bobatea/pkg/chat/conversation` imports from Pinocchio code paths (partially done; still present in some files for legacy compatibility).
- [ ] Verify interactive behaviors (selection, focus, copy) work with the timeline models.
- [ ] Delete or refactor any dev tooling that depended on Stream* messages.
- [x] Turn-first flows: build initial Turn from `system` + pre-seeded `messages` + `prompt` (`buildInitialTurn`) and seed backend before auto-submit.
- [x] Implement log + reduce strategy in backend (`history []*turns.Turn` + `reduceHistory()`) so each run carries full prior Blocks.

## What we changed in this codebase (so far)

- Backend API and event mapping
  - Implemented `Start(ctx, prompt string)` in the Pinocchio backend; engines stream to UI via timeline UIEntity*.
  - Rewrote `StepChatForwardFunc` to emit `UIEntityCreated/Updated/Completed` for `llm_text` and send `BackendFinishedMsg` when done.

- Chat runner and CLI wiring
  - `chatrunner`: `bobachat.InitialModel(backend, ...)` with router-based readiness. Seeds backend from prior conversation or rendered Turn so timeline shows history.
  - CLI chat mode (`cmds/cmd.go`): waits for `<-router.Running()` before auto-submitting; seeds backend from a rendered Turn built out of `system`, pre-seeded `messages` (now strings mapped to user blocks), and the runtime prompt.
  - When seeding, backend now emits UI entities for prior blocks (user/assistant only) to avoid missing history in chat; system blocks are intentionally not emitted to prevent duplication with other previews. Block IDs are used to deduplicate emissions.

- Prompt templating (render templates before model invocation)
  - Added `Variables map[string]interface{}` to `run.RunContext` and `run.WithVariables(...)` to pass parsed layer variables.
  - Implemented `PinocchioCommand.buildInitialTurn(vars)` in `cmd.go` to render `systemPrompt`, each block text (converted from YAML messages), and the `prompt` using `glazed/pkg/helpers/templating` prior to building a `turns.Turn`.
  - Used rendered Turn in blocking mode, for chat seeding, and for chat auto-start submission (so the UI sends the rendered prompt text, not the raw template).
  - Example wiring when invoking the run:

  ```246:255:pinocchio/pkg/cmds/cmd.go
  messages, err := g.RunWithOptions(ctx,
      run.WithStepSettings(stepSettings),
      run.WithWriter(w),
      run.WithRunMode(runMode),
      run.WithUISettings(uiSettings),
      run.WithConversationManager(manager),
      run.WithRouter(router),
      run.WithVariables(parsedLayers.GetDefaultParameterLayer().Parameters.ToMap()),
  )
  ```

  Rendering helper and usage:

  ```88:101:pinocchio/pkg/cmds/cmd.go
  func buildInitialTurnRendered(systemPrompt string, msgs []*conversation.Message, userPrompt string, vars map[string]interface{}) (*turns.Turn, error) {
      sp, err := renderTemplateString("system-prompt", systemPrompt, vars)
      if err != nil { return nil, err }
      renderedMsgs, err := renderMessages(msgs, vars)
      if err != nil { return nil, err }
      up, err := renderTemplateString("prompt", userPrompt, vars)
      if err != nil { return nil, err }
      return buildInitialTurn(sp, renderedMsgs, up), nil
  }
  ```

- Turn-first flows
  - YAML `messages` switched to `[]string`. Loader converts strings to `turns.NewUserTextBlock` and stores them in `PinocchioCommand.Blocks`.
  - `PinocchioCommand.buildInitialTurn(vars)` used in blocking and chat flows. Blocking mode runs inference on a Turn and converts back to conversation only for output.
  - Backend gained `SetSeedTurn` (and can still accept `SetSeedFromConversation` for legacy), and chat flow now seeds from the rendered Turn before auto-submit.

- Logging and readiness
  - Avoid arbitrary sleeps. Use the Watermill router’s `Running()` channel before sending auto-submit messages.
  - Consolidated logging via standard zerolog; removed ad-hoc local loggers.

## What we learned

- UI auto-start must be aligned with router readiness (use `<-Running()`); otherwise messages can be dropped and the UI appears stuck.
- The chat widget is timeline-first; conversation trees are not needed for rendering and should not be pushed into the UI layer.
- Turns must carry the full set of Blocks across iterations to maintain provider context. Treat intermediate streaming as patching a single logical assistant message, not a series of disjoint messages.
- Seeding is important: when transitioning into chat (or resuming), reduce prior state into a Turn and seed the backend, then continue the conversation with a new appended user Block.

## Next steps

1) Complete removal of `conversation.Manager`
- Replace `chatrunner` seeding from Manager with `buildInitialTurn` (system/messages/prompt) and drop direct `conversation` APIs.
- Update `run/context.go`, agents, and examples to produce Turns directly. Use `turns.BuildConversationFromTurn` only at output boundaries.

---

Later: 
2) Tool visualization and workflows
- Map `tool_call`/`tool_use` Blocks into timeline entities with dedicated interactive models (params/result viewers, copy actions).
- Ensure tool Blocks are carried across iterations (log+reduce) so context is visible to the provider and the UI.

3) Persistence and export
- Introduce a session storage for `history []*turns.Turn` with save/load and export to conversation or JSON for reproducibility.
- Add CLI flags to write/read history files.

4) Robustness and UX
- Add tests for `buildInitialTurn`, `reduceHistory`, and auto-start gating on `<-Running()`.
- Handle engine errors and interrupts with clear UI entities and statuses.
- Provide configuration for auto-start (on/off), prompt sources, and logging verbosity.

5) Documentation and examples
- Update examples/agents to Turn-first patterns.
- Expand docs with a mini end-to-end sample (build Turn, run engine, map events to timeline, log+reduce follow-up Turn).

## References

- Bobatea chat submit path creating user entity and calling backend Start:

```909:941:bobatea/pkg/chat/model.go
// create user message as llm_text and call backend.Start(ctx, userMessage)
```

- Backend interface (new signature):

```33:53:bobatea/pkg/chat/backend.go
type Backend interface {
    Start(ctx context.Context, prompt string) (tea.Cmd, error)
    Interrupt(); Kill(); IsFinished() bool
}
```
