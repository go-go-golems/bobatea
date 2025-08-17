---
Title: How to Build a Timeline Backend for LLM Chat Applications
Slug: how-to-build-timeline-backend
Short: Implement a backend that translates provider events into timeline entity messages (create, update, complete, delete).
Topics:
- backend
- timeline
- chat
- events
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Tutorial
---

## Overview

This tutorial explains how to implement a backend that converts LLM/provider events into `timeline` UI lifecycle messages. The backend publishes `UIEntityCreated`, `UIEntityUpdated`, `UIEntityCompleted`, and `UIEntityDeleted` to drive the append-only timeline UI in Bobatea.

We use the `pinocchio` simple chat agent’s backend as a concrete example, pointing at real files and functions.

## Key packages and files

- Timeline UI (Bobatea):
  - `bobatea/pkg/timeline/types.go`: UI lifecycle message types
  - `bobatea/pkg/timeline/controller.go`: append-only store, selection, message application
  - `bobatea/pkg/timeline/renderers/*`: Bubble Tea models
- Example backend:
  - `pinocchio/cmd/agents/simple-chat-agent/pkg/backend/tool_loop_backend.go`
  - Function `(*ToolLoopBackend).MakeUIForwarder(p *tea.Program)`

## Backend responsibilities

- For each provider/agent event, send one or more timeline messages to the program:
  - `UIEntityCreated` when an entity starts
  - `UIEntityUpdated` for streaming/incremental changes
  - `UIEntityCompleted` when done
  - `UIEntityDeleted` to remove an entity

The Bubble Tea program will deliver these to the chat model, which forwards them to the `timeline.Controller`.

## Wiring the forwarder

Inside your backend, construct a forwarder to map provider events to timeline messages. Example scaffold derived from `tool_loop_backend.go`:

```go
func (b *ToolLoopBackend) MakeUIForwarder(p *tea.Program) func(msg *message.Message) error {
    return func(msg *message.Message) error {
        msg.Ack()
        e, err := events.NewEventFromJson(msg.Payload)
        if err != nil { return err }
        md := e.Metadata()
        entityID := md.ID.String()

        switch e_ := e.(type) {
        case *events.EventPartialCompletionStart:
            p.Send(timeline.UIEntityCreated{ ID: timeline.EntityID{LocalID: entityID, Kind: "llm_text"}, Renderer: timeline.RendererDescriptor{Kind: "llm_text"}, Props: map[string]any{"role": "assistant", "text": "", "streaming": true}, StartedAt: time.Now() })
        case *events.EventPartialCompletion:
            p.Send(timeline.UIEntityUpdated{ ID: timeline.EntityID{LocalID: entityID, Kind: "llm_text"}, Patch: map[string]any{"text": e_.Completion, "streaming": true}, Version: time.Now().UnixNano(), UpdatedAt: time.Now() })
        case *events.EventFinal:
            p.Send(timeline.UIEntityCompleted{ ID: timeline.EntityID{LocalID: entityID, Kind: "llm_text"}, Result: map[string]any{"text": e_.Text} })
            p.Send(timeline.UIEntityUpdated{ ID: timeline.EntityID{LocalID: entityID, Kind: "llm_text"}, Patch: map[string]any{"streaming": false}, Version: time.Now().UnixNano(), UpdatedAt: time.Now() })
        case *events.EventLog:
            localID := fmt.Sprintf("log-%s-%d", md.TurnID, time.Now().UnixNano())
            p.Send(timeline.UIEntityCreated{ ID: timeline.EntityID{LocalID: localID, Kind: "log_event"}, Renderer: timeline.RendererDescriptor{Kind: "log_event"}, Props: map[string]any{"level": e_.Level, "message": e_.Message, "metadata": md} })
            p.Send(timeline.UIEntityCompleted{ ID: timeline.EntityID{LocalID: localID, Kind: "log_event"} })
        // ... other events like tools and agent modes
        }
        return nil
    }
}
```

## Deleting entities (backend-driven)

To remove an entity from the timeline (e.g., on cancel/retry/prune), emit `UIEntityDeleted`:

```go
delID := timeline.EntityID{LocalID: entityID, Kind: "llm_text"}
p.Send(timeline.UIEntityDeleted{ID: delID})
```

The controller will remove the entity and adjust selection. Use the same `LocalID`/`Kind` pair you used at creation.

## Best practices

- Ensure `LocalID` is unique per entity; using provider `message_id` or tool `id` is a good strategy. For ad-hoc items (logs), generate a unique ID with a timestamp or UUID.
- Set `StartedAt` on creation and increment `Version` on updates with `time.Now().UnixNano()` if you need strict ordering.
- Always send `UIEntityCompleted` to finalize long-lived entities, even when also updating.
- Delete only when your UX expects removal; otherwise prefer completion to preserve history.

## Testing your backend

1. Run your chat UI and start a request
2. Verify entities appear as you emit create/update/complete
3. Emit a delete and ensure the timeline removes the entity and selection remains valid

Combine with `timeline` logs to trace message flow.

## References

- `bobatea/pkg/timeline/types.go`
- `bobatea/pkg/timeline/controller.go`
- `pinocchio/cmd/agents/simple-chat-agent/pkg/backend/tool_loop_backend.go`
- `bobatea/docs/timeline.md` (API updates section)


