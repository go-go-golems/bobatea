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

## Checklist (migration tasks)

- [ ] Update pinocchio/pkg/chatrunner/chat_runner.go to use bobachat.InitialModel(backend, ...).
- [ ] Change pinocchio/pkg/ui/backend.go EngineBackend to implement Start(ctx, prompt string).
- [ ] Rewrite StepChatForwardFunc to emit timeline.UIEntity* and boba_chat.BackendFinishedMsg, remove conversation sends.
- [ ] Remove bobatea/pkg/chat/conversation imports from Pinocchio code paths.
- [ ] Verify interactive behaviors (selection, focus, copy) work with the timeline models.
- [ ] Delete or refactor any dev tooling that depended on Stream* messages.

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
