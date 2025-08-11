## Analysis: Async streaming flow from Turn-based engines to Bobatea chat UI

Date: 2025-08-11

### Scope

End-to-end trace of streaming events from engine RunInference (Turn-based) through Watermill to the Bubble Tea UI, and how those events mutate the current conversation in Bobatea. Includes identifiers used (message_id, run_id, turn_id), and how they map into UI message IDs.

### Components and responsibilities

- Engines (OpenAI/Claude) emit events while processing a `*turns.Turn`:
  - Start, partial, final, error, interrupt, tool-call, tool-result.
  - Event metadata carries `message_id` (NodeID), optional `run_id`, `turn_id`, engine info, usage.
- Event router (Watermill) transports JSON events on topic `ui`.
- UI forwarder `StepChatForwardFunc(p)` deserializes events and sends Bubble Tea messages (`StreamStartMsg`, `StreamCompletionMsg`, `StreamDoneMsg`, `StreamCompletionError`) to the program.
- Conversation UI (`conversationui.Model`) updates the `conversation.Manager` and its render cache accordingly.

### Event types and metadata

Event constructors and metadata structure:

```startLine:85:endLine:101:geppetto/pkg/events/chat-events.go
type EventPartialCompletionStart struct { EventImpl }
func NewStartEvent(metadata EventMetadata) *EventPartialCompletionStart
```

```startLine:120:endLine:137:geppetto/pkg/events/chat-events.go
type EventFinal struct { EventImpl; Text string }
func NewFinalEvent(metadata EventMetadata, text string) *EventFinal
```

```startLine:139:endLine:153:geppetto/pkg/events/chat-events.go
type EventError struct { EventImpl; ErrorString string }
func NewErrorEvent(metadata EventMetadata, err error) *EventError
```

```startLine:184:endLine:201:geppetto/pkg/events/chat-events.go
type EventToolCall struct { EventImpl; ToolCall ToolCall }
func NewToolCallEvent(metadata EventMetadata, toolCall ToolCall) *EventToolCall
```

```startLine:208:endLine:224:geppetto/pkg/events/chat-events.go
type EventToolResult struct { EventImpl; ToolResult ToolResult }
func NewToolResultEvent(metadata EventMetadata, toolResult ToolResult) *EventToolResult
```

```startLine:266:endLine:289:geppetto/pkg/events/chat-events.go
type EventPartialCompletion struct { EventImpl; Delta, Completion string }
func NewPartialCompletionEvent(metadata EventMetadata, delta, completion string) *EventPartialCompletion
```

EventMetadata and identifiers:

```startLine:295:endLine:336:geppetto/pkg/events/chat-events.go
type EventMetadata struct {
    conversation.LLMMessageMetadata
    ID     conversation.NodeID `json:"message_id"`
    RunID  string              `json:"run_id,omitempty"`
    TurnID string              `json:"turn_id,omitempty"`
    Extra  map[string]any      `json:"extra,omitempty"`
}
```

Notes:
- `ID` is the per-stream message identifier used by the UI as `StreamMetadata.ID` to correlate `start/partial/final` to the same assistant message.
- `RunID`/`TurnID` correlate across a broader workflow but are not consumed directly by the current UI.

### Engine emission of events (examples)

OpenAI engine emits Start → Partial → Final, and ToolCall events during streaming:

```startLine:161:endLine:170:geppetto/pkg/steps/ai/openai/engine_openai.go
startEvent := events.NewStartEvent(metadata)
e.publishEvent(ctx, startEvent)
... e.publishEvent(ctx, events.NewErrorEvent(metadata, err))
```

```startLine:257:endLine:305:geppetto/pkg/steps/ai/openai/engine_openai.go
partialEvent := events.NewPartialCompletionEvent(metadata, delta, message)
e.publishEvent(ctx, partialEvent)
... toolCallEvent := events.NewToolCallEvent(metadata, toolCall)
e.publishEvent(ctx, toolCallEvent)
finalEvent := events.NewFinalEvent(metadata, message)
e.publishEvent(ctx, finalEvent)
```

Claude engine (via content block merger) emits Start, ToolCall, Final as needed:

```startLine:156:endLine:217:geppetto/pkg/steps/ai/claude/content-block-merger.go
return []events.Event{events.NewStartEvent(cbm.metadata)}, nil
... return []events.Event{events.NewToolCallEvent(cbm.metadata, events.ToolCall{ ... })}, nil
... return []events.Event{events.NewFinalEvent(cbm.metadata, cbm.response.FullText())}, nil
```

### Forwarding events to the UI

The Watermill handler translates events to Bubble Tea messages:

```startLine:103:endLine:168:pinocchio/pkg/ui/backend.go
func StepChatForwardFunc(p *tea.Program) func(msg *message.Message) error {
    e, _ := events.NewEventFromJson(msg.Payload)
    meta := e.Metadata()
    metadata := conversation2.StreamMetadata{ ID: meta.ID, EventMetadata: &meta }
    switch e := e.(type) {
    case *events.EventError:
        p.Send(conversation2.StreamCompletionError{ StreamMetadata: metadata, Err: errors.New(e.ErrorString) })
    case *events.EventPartialCompletion:
        p.Send(conversation2.StreamCompletionMsg{ StreamMetadata: metadata, Delta: e.Delta, Completion: e.Completion })
    case *events.EventFinal, *events.EventInterrupt:
        // Final/Interrupt become StreamDoneMsg with full text
        p.Send(conversation2.StreamDoneMsg{ StreamMetadata: metadata, Completion: e.Text })
    case *events.EventToolCall:
        p.Send(conversation2.StreamDoneMsg{ StreamMetadata: metadata, Completion: fmt.Sprintf("%s(%s)", e.ToolCall.Name, e.ToolCall.Input) })
    case *events.EventToolResult:
        p.Send(conversation2.StreamDoneMsg{ StreamMetadata: metadata, Completion: fmt.Sprintf("Result: %s", e.ToolResult.Result) })
    case *events.EventPartialCompletionStart:
        p.Send(conversation2.StreamStartMsg{ StreamMetadata: metadata })
    }
}
```

Identifier mapping:
- Watermill message (JSON) → `events.Event` → `EventMetadata.ID` preserved.
- `StreamMetadata.ID` is set to the same `NodeID`, used by `conversationui.Model` to correlate the UI message.

### UI handling and conversation updates

The top-level `chat` model forwards stream messages to the embedded `conversationui.Model` and updates viewport content.

```startLine:292:endLine:351:bobatea/pkg/chat/model.go
case conversationui.StreamCompletionMsg, conversationui.StreamStartMsg, conversationui.StreamStatusMsg, conversationui.StreamDoneMsg, conversationui.StreamCompletionError:
    old := m.conversation.SelectedIdx()
    m.conversation, cmd = m.conversation.Update(msg)
    if m.scrollToBottom { v, _ := m.conversation.ViewAndSelectedPosition(); m.viewport.SetContent(v); m.viewport.GotoBottom() }
```

Inside `conversationui.Model` the messages mutate the conversation via `conversation.Manager` and update the render cache:

```startLine:327:endLine:406:bobatea/pkg/chat/conversation/model.go
case StreamStartMsg:
    // create new assistant message with ID = StreamMetadata.ID, append to manager
    msg_ := conversation2.NewChatMessage(conversation2.RoleAssistant, "", conversation2.WithID(msg.ID), conversation2.WithMetadata(metadata))
    _ = m.manager.AppendMessages(msg_)
    m.updateCache(msg_)
```

```startLine:263:endLine:305:bobatea/pkg/chat/conversation/model.go
case StreamCompletionMsg, StreamDoneMsg:
    msg_, ok := m.manager.GetMessage(msg.ID)
    textMsg := msg_.Content.(*conversation2.ChatMessageContent)
    textMsg.Text = msg.Completion
    msg_.LastUpdate = time.Now()
    msg_.Metadata["event_metadata"] = msg.EventMetadata
    if msg.EventMetadata != nil { msg_.LLMMessageMetadata = &msg.EventMetadata.LLMMessageMetadata }
    m.updateCache(msg_)
```

Error case updates the same message text with a markdown-formatted error:

```startLine:306:endLine:326:bobatea/pkg/chat/conversation/model.go
case StreamCompletionError:
    msg_, _ := m.manager.GetMessage(msg.ID)
    textMsg := msg_.Content.(*conversation2.ChatMessageContent)
    textMsg.Text = "**Error**\n\n" + msg.Err.Error()
    m.updateCache(msg_)
```

### End-to-end sequence (Mermaid)

```mermaid
sequenceDiagram
    participant Engine as Engine (RunInference on Turn)
    participant Router as Watermill Router (topic: ui)
    participant UIForwarder as StepChatForwardFunc(p)
    participant Chat as bobatea/pkg/chat.model
    participant ConvUI as bobatea/pkg/chat/conversation.Model
    participant Manager as geppetto/pkg/conversation.Manager

    Note over Engine: Build EventMetadata {
      message_id: NodeID, run_id?, turn_id?, engine, usage
    }

    Engine->>Router: NewStartEvent(meta)
    Router->>UIForwarder: JSON(EventType=start, meta.ID=message_id)
    UIForwarder->>Chat: StreamStartMsg{ ID=meta.ID, EventMetadata=meta }
    Chat->>ConvUI: forward StreamStartMsg
    ConvUI->>Manager: AppendMessages(NewChatMessage(role=assistant, id=ID))
    ConvUI->>ConvUI: updateCache(newMessage)

    loop streaming
      Engine->>Router: NewPartialCompletionEvent(meta, delta, completion)
      Router->>UIForwarder: JSON(EventType=partial, meta.ID=message_id)
      UIForwarder->>Chat: StreamCompletionMsg{ ID=meta.ID, Delta, Completion }
      Chat->>ConvUI: forward StreamCompletionMsg
      ConvUI->>Manager: GetMessage(ID) and update text=Completion
      ConvUI->>ConvUI: updateCache(updatedMessage)
    end

    alt final
      Engine->>Router: NewFinalEvent(meta, text)
      Router->>UIForwarder: JSON(EventType=final, meta.ID=message_id)
      UIForwarder->>Chat: StreamDoneMsg{ ID=meta.ID, Completion=text }
      Chat->>ConvUI: forward StreamDoneMsg
      ConvUI->>Manager: GetMessage(ID) and set final text
      ConvUI->>ConvUI: updateCache(updatedMessage)
    else error/interrupt
      Engine->>Router: NewErrorEvent / NewInterruptEvent
      UIForwarder->>Chat: StreamCompletionError / StreamDoneMsg
      Chat->>ConvUI: forward
      ConvUI->>Manager: GetMessage(ID), set error text / final text
    end
```

### IDs in play (per event and UI message)

- EventMetadata.ID (message_id):
  - Type: `conversation.NodeID`
  - Purpose: Correlate start/partial/final across a single assistant response
  - Propagation: Engine → Event → Router → UIForwarder → `StreamMetadata.ID` → `conversation.Message.ID`

- EventMetadata.RunID, TurnID:
  - Optional correlation identifiers for runs/turns; currently attached to `StreamMetadata.EventMetadata` and stored on the `conversation.Message.Metadata["event_metadata"]` but not interpreted by the UI layout.

- Block IDs (Turn blocks):
  - Engines append `turns.Block` to the Turn (with IDs for tool calls/results) but the current UI flow does not use block IDs.
  - Forwarded tool_call/tool_result events are currently converted to `StreamDoneMsg` strings; block identity is not linked to conversation messages.

### Observations and improvement opportunities

- The UI uses `message_id` to manage one assistant message per streaming response—works well for text-only flows.
- Tool events are flattened to text; block IDs and structure are lost in the conversation UI path.
- Rich Turn visualizations would benefit from consuming block IDs and kinds to render structured panes for tool calls/results (see the companion design doc `01-design-for-a-new-chat-UI-turn-visualization-approach.md`).
