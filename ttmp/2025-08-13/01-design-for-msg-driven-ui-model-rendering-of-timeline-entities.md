Title: Design for Msg-driven UI Model Rendering of Timeline Entities
Slug: timeline-msg-driven-entities
Short: Consolidated Bubble Tea models with a unified message set for creation, updates, completion, and clipboard actions
Topics:
- bobatea
- ui
- timeline
- bubbletea
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: Design

## Purpose and scope

This document describes a consolidation of the timeline UI so that all entity rendering goes through Bubble Tea models only, and proposes a full message set to drive entity lifecycle and interactions. It replaces stateless string renderers and the render cache with model factories and per-entity models. It also outlines how props and lifecycle changes flow through messages rather than direct method calls.

## What exists now (after consolidation)

- Controller manages append-only entities and selection/focus.
- Registry exposes only `EntityModelFactory` for `Key`/`Kind` to create models.
- Models live under `bobatea/pkg/timeline/renderers` and implement `EntityModel`.
- Demo wires factories for `llm_text`, `tool_calls_panel`, and `plain`.

## Unified message set for models

### Lifecycle messages (domain → UI)

- UIEntityCreated: existing struct, used by controller via `OnCreated`.
- UIEntityUpdated: existing struct, used by controller via `OnUpdated`.
- UIEntityCompleted: existing struct, used by controller via `OnCompleted`.
- UIEntityDeleted: existing struct, used by controller via `OnDeleted`.

Optional Bubble Tea equivalents (already added):
- EntityCreatedMsg { ID, Renderer, Props, StartedAt }
- EntityCompletedMsg { ID, Result }

These allow a parent Bubble Tea model to broadcast lifecycle without calling controller methods directly. The controller could expose an Update(msg) entry-point to consume these if desired.

### Prop and selection messages (controller → model)

- EntityPropsUpdatedMsg { ID, Patch }
- EntitySelectedMsg { ID }
- EntityUnselectedMsg { ID }

Controller emits these during `View()` and when updates arrive, ensuring models can react in `Update` instead of only via `OnProps`.

### Clipboard/copy actions

- EntityCopyTextMsg { ID }
- EntityCopyCodeMsg { ID }

Models respond by emitting:
- CopyTextRequestedMsg { Text string }
- CopyCodeRequestedMsg { Code string }

LLM text model: extracts first fenced code block for `EntityCopyCodeMsg`, falls back to text if none.
Plain/tool-calls models: return textual view for both.

## Handling prop updates via messages

In addition to calling `OnProps`, controller now also dispatches `EntityPropsUpdatedMsg` to the model so that purely message-driven models work. Future step: a pure message flow could be introduced where the controller simply routes `UIEntityUpdated` as `EntityPropsUpdatedMsg` without calling `OnProps`; models would own their internal state.

## Controller responsibilities (final shape)

- Maintain entity order and selected index.
- Instantiate models via registry on create.
- Route selection/focus hints and external Bubble Tea messages to the selected model.
- Render by calling each model's `View()`.

## Model guidance

Models should:
- Keep minimal internal state derived from props.
- Respect width/selection/focus hints via props and messages.
- Render efficiently; avoid heavy work in every `View()` call.
- Implement sensible clipboard behavior for both text and code requests.

## Future enhancements

- Add Enter/Exit focus messages for consistency instead of direct Focus/Blur calls.
- Adopt only message-driven lifecycle: `EntityCreatedMsg`, `EntityPropsUpdatedMsg`, `EntityCompletedMsg` handled by controller.Update().
- Add richer message set: expand/collapse, open-in-editor, jump-to-turn, etc.
- Introduce per-model command helpers for asynchronous work (e.g., background formatting).

## Summary of message API (Go signatures)

```go
// Selection & props
type EntitySelectedMsg struct { ID timeline.EntityID }
type EntityUnselectedMsg struct { ID timeline.EntityID }
type EntityPropsUpdatedMsg struct { ID timeline.EntityID; Patch map[string]any }

// Clipboard
type EntityCopyTextMsg struct { ID timeline.EntityID }
type EntityCopyCodeMsg struct { ID timeline.EntityID }
type CopyTextRequestedMsg struct { Text string }
type CopyCodeRequestedMsg struct { Code string }

// Optional lifecycle as Bubble Tea messages
type EntityCreatedMsg struct { ID timeline.EntityID; Renderer timeline.RendererDescriptor; Props map[string]any; StartedAt time.Time }
type EntityCompletedMsg struct { ID timeline.EntityID; Result map[string]any }
```

## Notes for implementers

- Keep using `RendererDescriptor{Key,Kind}` as the model selection hint; the name remains for compatibility, but it now refers to model factories.
- The `cache.go` and stateless `renderers.go` are obsolete; do not reintroduce them.
- In demos or apps, handle `CopyTextRequestedMsg` and `CopyCodeRequestedMsg` at the top-level program to integrate with the system clipboard.


