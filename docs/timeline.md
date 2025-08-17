---
Title: Timeline UI (Turn-centric) for Bobatea
Slug: bobatea-timeline
Short: An append-only, renderer-driven timeline for streaming chat/agent UIs
Topics:
- bobatea
- ui
- timeline
- turns
- streaming
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

## Timeline UI (Turn-centric) for Bobatea

The Timeline package provides a self-contained, append-only visualization layer for chat/agent UIs. It decouples provider streaming and orchestration from rendering: engines and middlewares emit provider-agnostic events that are translated into UI entities; the timeline renders those entities using Bubble Tea entity models (no stateless renderers or cache). This yields a flexible UI capable of simple text, tool-call panels, diffs, and more.

### What you’ll build

- A standalone demo that shows streaming text and tool calls, with no LLM dependency
- A reusable `timeline` package with:
  - An append-only controller and store
  - A model factory registry (plug in text/panels/diffs)
  - No render cache (models render directly)
  - Bubble Tea-friendly surface (viewport-based)

## Core concepts

The timeline visualizes immutable creation order with incremental updates. An entity has a stable identity and lifecycle. Bubble Tea models convert entity props into strings that fit the current width/theme.

- Entity identity: `{ RunID?, TurnID?, BlockID?, LocalID, Kind }`
- Lifecycle messages:
  - Created: initial props and `StartedAt`
  - Updated: partial `Patch` with `UpdatedAt` and `Version`
  - Completed: optional final `Result` snapshot
  - Deleted: remove from timeline
- Append-only ordering: entities are appended in the order Created messages arrive
- Caching: `(rendererKey, entityID, width, theme, propsHash)`

## Package layout

- `bobatea/pkg/timeline/types.go` — entity IDs, descriptors, lifecycle messages
- `bobatea/pkg/timeline/registry.go` — entity model factory registry
- `bobatea/pkg/timeline/store.go` — append-only entity store
- `bobatea/pkg/timeline/cache.go` — removed (models render directly)
- `bobatea/pkg/timeline/controller.go` — applies lifecycle messages, renders via registry
- `bobatea/pkg/timeline/renderers` — Bubble Tea models (`llm_text`, `tool_calls_panel`, `plain`)
- Demo: `bobatea/cmd/timeline-demo/main.go`

## Minimal API (pseudocode)

```go
// Entity identity
type EntityID struct {
  RunID, TurnID, BlockID, LocalID string
  Kind string // e.g., "llm_text", "tool_calls_panel"
}

// Renderer selection
type RendererDescriptor struct { Key, Kind string }

// Lifecycle messages
type UIEntityCreated struct {
  ID EntityID
  Renderer RendererDescriptor
  Props map[string]any
  StartedAt time.Time
}
type UIEntityUpdated struct {
  ID EntityID
  Patch map[string]any
  Version int64
  UpdatedAt time.Time
}
type UIEntityCompleted struct { ID EntityID; Result map[string]any }
type UIEntityDeleted struct { ID EntityID }

// EntityModel interface
type EntityModel interface {
  tea.Model
  View() string
  OnProps(patch map[string]any)
  OnCompleted(result map[string]any)
  SetSize(width, height int)
  Focus()
  Blur()
}

// Controller (selected methods)
type Controller struct { /* store, cache, registry, size, theme */ }
func (c *Controller) OnCreated(e UIEntityCreated)
func (c *Controller) OnUpdated(e UIEntityUpdated)
func (c *Controller) OnCompleted(e UIEntityCompleted)
func (c *Controller) OnDeleted(e UIEntityDeleted)
func (c *Controller) SetSize(w, h int)
func (c *Controller) View() string
```

## Running the demo

The demo runs independently and generates fake streaming events.

```bash
cd bobatea
go run ./cmd/timeline-demo
```

Controls:
- t: start a streaming text entity (partial → final)
- o: start a tool-calls panel (multiple calls and results)
- q: quit

You should see a header immediately:

```
Timeline demo: press t = stream text, o = tool calls, q = quit
```

## Logging and debugging

The demo logs to `/tmp/timeline-demo.log`. It records WindowSizeMsg, key presses, controller apply/refresh cycles, renderer outputs, and cache hits/misses.

To inspect logs while running:

```bash
tail -f /tmp/timeline-demo.log
```

To script with tmux:

```bash
cd bobatea
: > /tmp/timeline-demo.log
tmux new-session -d -s tldemo 'go run ./cmd/timeline-demo'
sleep 1
tmux send-keys -t tldemo:0 t Enter
sleep 1
tmux send-keys -t tldemo:0 o Enter
sleep 2
tail -n 200 /tmp/timeline-demo.log
tmux send-keys -t tldemo:0 q Enter
tmux kill-session -t tldemo
```

## How it renders

Rendering is additive and width-aware:

- The viewport height reserves 2 lines for a fixed header (see `bobatea/cmd/timeline-demo/main.go`).
- For each entity (in append order):
  1) The controller instantiates/updates a Bubble Tea model via the registered factory
  2) The model receives selection/focus hints and renders itself with `View()`

Example sequence (simplified):

```text
Created llm_text → model.OnProps → model.View
Updated llm_text ("Hello") → model.OnProps → model.Update(EntityPropsUpdatedMsg)
Completed llm_text ("Hello, world!") → model.OnCompleted → model.View
```

## Extending with custom models

Add a new model by implementing the interface and registering a factory:

```go
type MyModelFactory struct{}
func (MyModelFactory) Key() string  { return "renderer.my_widget.v1" }
func (MyModelFactory) Kind() string { return "my_widget" }
func (MyModelFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel { /* return Bubble Tea model */ }

reg := timeline.NewRegistry()
reg.RegisterModelFactory(MyModelFactory{})
```

Then create/update entities with `RendererDescriptor{Kind: "my_widget"}` (or Key).

## Design choices

- Append-only ordering provides durable, predictable timelines for user navigation and debugging
- Bubble Tea models provide richer interactivity and composition
- TurnStore translation (outside this package) is the intended source of UIEntity messages; it maps provider/middleware events to entities without inspecting Turn state

## Troubleshooting

- Black screen or missing header: ensure WindowSizeMsg arrives; the demo reserves 2 header lines and logs viewport size. Resize the terminal to force a size event.
- No updates on keypress: verify logs show `KeyMsg`; if not, the terminal may not be focused.
- Performance issues: ensure models avoid excessive recomputation in `View()`.

## Next steps

- Add markdown renderer for `llm_text` (glamour-based) mirroring `conversation` UI styles
- Add richer tool-call list with expandable details
- Add a diff renderer and a metadata panel
- Provide an adapter from provider/middleware events to UIEntity messages (TurnStore translation)



## API updates (2025-08-17)

This section summarizes the latest changes to input routing and the role of the shell vs controller.

- Controller key routing:
  - `timeline.Controller.HandleMsg(tea.Msg)` now routes messages to the selected entity model when either:
    - entering mode is active (as before), or
    - the message is `tea.KeyMsg` for TAB or shift+TAB.
  - This allows compact toggles on selected entities without fully entering focus.
  - See `bobatea/pkg/timeline/controller.go` (method `HandleMsg`).

- Shell simplification:
  - `timeline.Shell` is a thin wrapper around `viewport.Model` plus selection helpers.
  - TAB/shift+TAB logic was removed from the shell so that routing policy is centralized in the controller.
  - The shell focuses on:
    - Viewport management (`SetSize`, `View`, `GotoBottom`, `ScrollUp/Down`)
    - Selection ergonomics (`SelectNext/Prev/Last`, `EnterSelection/ExitSelection`, `ScrollToSelected`)
    - Lightweight wrappers to apply lifecycle events and refresh the view.
  - See `bobatea/pkg/timeline/shell.go`.

- Interactive log event renderer example:
  - File: `bobatea/pkg/timeline/renderers/log_event_model.go`
  - Behavior: when selected, pressing TAB toggles YAML metadata visibility; when unselected, metadata auto-collapses. This is implemented by handling `tea.KeyMsg` in the model’s `Update` method and by the controller forwarding TAB even when not entering.

- Recommended renderer message handling (recap):
  - Selection and props: `EntitySelectedMsg`, `EntityUnselectedMsg`, `EntityPropsUpdatedMsg`
  - Sizing and focus hints: `EntitySetSizeMsg`, `EntityFocusMsg`, `EntityBlurMsg`
  - Copy actions: `EntityCopyTextMsg` / emit `CopyTextRequestedMsg`, `EntityCopyCodeMsg` / emit `CopyCodeRequestedMsg`

