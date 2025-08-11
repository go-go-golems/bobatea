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

The Timeline package provides a self-contained, append-only visualization layer for chat/agent UIs. It decouples provider streaming and orchestration from rendering: engines and middlewares emit provider-agnostic events that are translated into UI entities; the timeline renders those entities using pluggable renderers, with width/theme-aware caching. This yields a flexible UI capable of simple text, tool-call panels, diffs, and more.

### What you’ll build

- A standalone demo that shows streaming text and tool calls, with no LLM dependency
- A reusable `timeline` package with:
  - An append-only controller and store
  - A renderer registry (plug in text/panels/diffs)
  - A per-entity render cache
  - Bubble Tea-friendly surface (viewport-based)

## Core concepts

The timeline visualizes immutable creation order with incremental updates. An entity has a stable identity and lifecycle. Renderers convert entity props into strings (for now) that fit the current width/theme. Caching ensures we only re-render pieces that changed.

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
- `bobatea/pkg/timeline/registry.go` — renderer registry
- `bobatea/pkg/timeline/store.go` — append-only entity store
- `bobatea/pkg/timeline/cache.go` — per-entity render cache
- `bobatea/pkg/timeline/controller.go` — applies lifecycle messages, renders via registry
- `bobatea/pkg/timeline/renderers.go` — minimal renderers (`llm_text`, `tool_calls_panel`, fallback)
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

// Renderer interface
type Renderer interface {
  Key() string
  Kind() string
  Render(props map[string]any, width int, theme string) (string, int, error)
  RelevantPropsHash(props map[string]any) string
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
  1) The controller asks the registry for a renderer, by key then by kind
  2) A cache lookup decides whether to reuse a previous render (same width, theme, props hash)
  3) On a miss, the renderer produces a string; we cache and append it

Example sequence (simplified):

```text
Created llm_text → render "" (miss, store)
Updated llm_text ("Hello") → invalidate → render (miss, store)
Completed llm_text ("Hello, world!") → invalidate → render (miss, store)
```

## Extending with custom renderers

Add a new renderer by implementing the interface and registering it:

```go
type MyRenderer struct{}
func (r *MyRenderer) Key() string  { return "renderer.my_widget.v1" }
func (r *MyRenderer) Kind() string { return "my_widget" }
func (r *MyRenderer) RelevantPropsHash(p map[string]any) string { return hashOf(p) }
func (r *MyRenderer) Render(p map[string]any, width int, theme string) (string, int, error) {
  // return a wrapped string for now
  return fmt.Sprintf("[my_widget] %v", p["title"]), 1, nil
}

reg := timeline.NewRegistry()
reg.Register(&MyRenderer{})
```

Then create/update entities with `RendererDescriptor{Kind: "my_widget"}` (or Key).

## Design choices

- Append-only ordering provides durable, predictable timelines for user navigation and debugging
- Minimal string-based renderers keep the first iteration simple and composable; future work can adopt richer widget components
- TurnStore translation (outside this package) is the intended source of UIEntity messages; it maps provider/middleware events to entities without inspecting Turn state

## Troubleshooting

- Black screen or missing header: ensure WindowSizeMsg arrives; the demo reserves 2 header lines and logs viewport size. Resize the terminal to force a size event.
- No updates on keypress: verify logs show `KeyMsg`; if not, the terminal may not be focused.
- Performance/caching issues: inspect cache hit/miss logs; confirm props hashes are stable when content is unchanged.

## Next steps

- Add markdown renderer for `llm_text` (glamour-based) mirroring `conversation` UI styles
- Add richer tool-call list with expandable details
- Add a diff renderer and a metadata panel
- Provide an adapter from provider/middleware events to UIEntity messages (TurnStore translation)


