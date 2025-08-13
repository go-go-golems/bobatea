---
Title: Removing deprecated Stream* messages and conversation UI from Bobatea chat
Slug: remove-conversation-and-stream-msgs
Short: Analysis of eliminating bobatea/pkg/chat/conversation and Stream* messages in favor of timeline entities
Topics:
- chat
- timeline
- migration
- refactor
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Goal and scope

We want to determine whether the legacy `bobatea/pkg/chat/conversation` package (and its `Stream*` message types) can be removed and replaced by the newer timeline-based rendering and interaction model. The desired end-state is that backends emit timeline lifecycle events directly, and the chat UI no longer depends on conversation-specific message types or rendering styles.

## Current usage inventory

- Chat model imports conversation UI for styles and for stream message types; it converts `Stream*` messages into timeline events.

```386:446:bobatea/pkg/chat/model.go
case conversationui.StreamCompletionMsg,
    conversationui.StreamStartMsg,
    conversationui.StreamStatusMsg,
    conversationui.StreamDoneMsg,
    conversationui.StreamCompletionError:
    // … translate to timeline UIEntityCreated/Updated/Completed …
```

```85:96:bobatea/pkg/chat/model.go
style  *conversationui.Style
// …
style:               conversationui.DefaultStyles(),
```

- Fake backend emits `Stream*` messages and also (for tools) emits timeline events. This shows both patterns currently coexist.

```137:171:bobatea/cmd/chat/fake-backend.go
f.p.Send(conversationui.StreamStartMsg{ StreamMetadata: metadata })
// …
f.p.Send(conversationui.StreamCompletionMsg{ /* delta, completion */ })
// …
f.p.Send(conversationui.StreamDoneMsg{ StreamMetadata: metadata, Completion: msg })
```

```85:99:bobatea/cmd/chat/fake-backend.go
// Tool demo path emits timeline events directly
f.p.Send(
    timeline.UIEntityCreated{ /* key: renderer.tool.get_weather.v1 */ },
)
f.p.Send(timeline.UIEntityUpdated{ /* … */ })
f.p.Send(timeline.UIEntityCompleted{ /* … */ })
```

- CLI client (`cmd/chat-client`) provides endpoints for sending `Stream*` messages over HTTP for testing. If we remove the conversation UI, these helpers become obsolete or must be reworked to send timeline events instead.

```80:93:bobatea/cmd/chat-client/main.go
msg := conversationui.StreamStartMsg{ /* … */ }
sendRequest("start", msg)
```

## What the conversation package provides

- Message types: `StreamStartMsg`, `StreamStatusMsg`, `StreamCompletionMsg`, `StreamDoneMsg`, `StreamCompletionError`.
- A legacy `Model` for rendering a conversation tree and caching markdown, plus a `Style` type.

```484:512:bobatea/pkg/chat/conversation/model.go
type StreamStartMsg struct { StreamMetadata }
type StreamStatusMsg struct { StreamMetadata; Text string }
type StreamCompletionMsg struct { StreamMetadata; Delta, Completion string }
type StreamDoneMsg struct { StreamMetadata; Completion string }
type StreamCompletionError struct { StreamMetadata; Err error }
```

The chat UI no longer renders the conversation tree; it renders timeline entities. The only remaining roles for this package in the chat flow are: (1) the `Stream*` types, (2) the `Style` used around the input box.

## Feasibility of removal

Short answer: Yes, we can remove the conversation package from the chat path by:

- Emitting timeline lifecycle events directly from backends (for both LLM text and tools), and
- Switching the chat input panel styling to `pkg/timeline/chatstyle` (already present), and
- Removing the `Stream*` handling from the chat model.

Rationale:

- The chat model already converts `Stream*` messages into `timeline.UIEntity*` events. If backends emit `UIEntity*` directly, that adapter logic becomes unnecessary.
- We already have `chatstyle` for consistent styling in timeline components; reusing it in the chat model avoids the `conversationui.Style` dependency.

## Recommended end-state

- Backends:
  - Only send `timeline.UIEntityCreated/Updated/Completed` (plus `chat.BackendFinishedMsg`), never `conversationui.Stream*`.
  - For LLM text streaming, use a single entity of kind `llm_text` with progressive `Updated` patches to `text`, and finally `Completed`.

- Chat model:
  - Remove the entire `case conversationui.Stream*` branch in `Update`.
  - Keep user-submit logic that creates a user `llm_text` entity immediately.
  - Replace `style *conversationui.Style` with `style *chatstyle.Style` and use `chatstyle.DefaultStyles()`.

- CLI client:
  - Either delete the `Stream*` routes or replace them with routes that send timeline entity lifecycle events for testing.

## Concrete changes (file-by-file)

- `bobatea/pkg/chat/model.go`:
  - Replace style type and constructor:
    - `style *conversationui.Style` → `style *chatstyle.Style`.
    - `conversationui.DefaultStyles()` → `chatstyle.DefaultStyles()`.
  - Remove imports of `conversationui`.
  - Delete `case conversationui.Stream*` handling; rely on `timeline.UIEntity*` and `chat.BackendFinishedMsg` paths.

- `bobatea/cmd/chat/fake-backend.go`:
  - Remove `conversationui` messages. For LLM text streaming, emit:
    - `timeline.UIEntityCreated{ ID: {LocalID, Kind:"llm_text"}, Renderer: {Kind:"llm_text"}, Props: {role:"assistant", text:""} }`
    - For each token/chunk: `timeline.UIEntityUpdated{ Patch: {text: <running concatenation>}, Version: n }`
    - At end: `timeline.UIEntityCompleted{ Result: {text: final} }` plus `chat.BackendFinishedMsg`.
  - Tool paths already use timeline events; keep as-is.

- `bobatea/cmd/chat-client/main.go`:
  - If kept, adjust dev routes to post timeline lifecycle messages instead of `Stream*`.
  - Or remove if not needed in the new flow.

## Example: backend streaming via timeline entities (pseudocode)

```go
// on Start(ctx, prompt)
id := newID()
send(UIEntityCreated{ID: {LocalID: id, Kind: "llm_text"}, Renderer: {Kind: "llm_text"}, Props: {role: "assistant", text: ""}, StartedAt: now})
acc := ""
for each chunk in stream {
    acc += chunk
    send(UIEntityUpdated{ID: {LocalID: id, Kind: "llm_text"}, Patch: {text: acc}, Version: ver++; UpdatedAt: now})
}
send(UIEntityCompleted{ID: {LocalID: id, Kind: "llm_text"}, Result: {text: acc}})
send(BackendFinishedMsg{})
```

## Risk analysis and mitigations

- Risk: removing `Stream*` breaks any external tooling (e.g., `chat-client`) expecting those endpoints.
  - Mitigation: provide a transition period where the server translates `Stream*` requests into timeline events server-side, deprecate docs, then remove.

- Risk: the chat model currently logs stream telemetry; removing the branch changes logs.
  - Mitigation: keep equivalent logs around `UIEntity*` handling.

- Risk: programmatic integrations relying on `conversation.Manager`.
  - Mitigation: the chat app no longer maintains a conversation tree; downstreams should rely on timeline entities or build their own storage.

## Decision

It is feasible and recommended to remove `bobatea/pkg/chat/conversation` from the chat application flow. The timeline controller and entity models fully supersede conversation rendering and `Stream*` messages. Concretely:

- Backends should emit only timeline events.
- The chat model should use `chatstyle` for styles and drop all `conversationui` references.
- Optional developer tooling like `cmd/chat-client` should be updated or retired.

## Pointers to current code (for migration)

- Conversation message handling in chat model:
```396:446:bobatea/pkg/chat/model.go
// case conversationui.Stream* … currently translates to timeline events
```

- Style initialization in chat model:
```137:146:bobatea/pkg/chat/model.go
style:               conversationui.DefaultStyles(),
```

- Fake backend stream and tool emission:
```69:110:bobatea/cmd/chat/fake-backend.go
// Tool demo emits timeline events; stream path still uses conversationui Stream*
```

## Suggested next steps

- [ ] Replace `conversationui.DefaultStyles()` with `chatstyle.DefaultStyles()` in `chat.model` and switch the style type.
- [ ] Remove `case conversationui.Stream*` from `chat.model.Update`.
- [ ] Update fake backend to emit timeline events for LLM text streaming.
- [ ] Update or remove `cmd/chat-client` routes that send `Stream*`.
- [ ] Delete `bobatea/pkg/chat/conversation` once no references remain.
- [ ] Run the app and verify streaming, selection, copy, and save-to-file flows work end-to-end.


