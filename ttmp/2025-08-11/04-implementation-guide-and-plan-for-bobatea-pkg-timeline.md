## Implementation Guide and Plan: `bobatea/pkg/timeline/`

Date: 2025-08-11

### 1) Purpose and scope

The `timeline` package provides an append-only, Turn-centric visualization layer for chat UIs in Bobatea. It consumes provider/middleware events translated by a TurnStore into UI entity lifecycle messages (Created/Updated/Completed/Deleted), maintains ordered in-memory state, and renders a timeline using a registry of pluggable renderers (e.g., `llm_text`, `tool_calls_panel`, `diff_view`). The package is provider-agnostic and decoupled from `conversation.Manager`.

Key goals:
- Append-only ordering, stable `EntityID`, timestamps (`StartedAt`, `UpdatedAt`).
- Pluggable renderers with width/theme-aware caching.
- Bubble Tea-friendly: expose `tea.Model` or composable view for embedding.
- Minimal surface for integration with existing programs (e.g., simple-chat-agent).

### 2) Architecture overview

- TimelineController
  - Owns `EntityStateStore` with append-only order and lookup by `EntityID`.
  - Applies lifecycle messages: Created, Updated, Completed, Deleted.
  - Maintains selection/focus and produces a rendered string via RendererRegistry.

- RendererRegistry
  - Registers `BlockRenderer`-like components by `RendererDescriptor.Key` or `Kind`.
  - Each renderer implements `Render(props, width, theme) (string, int)` and declares a `RelevantPropsHash(props)` for caching.

- Cache
  - Memoizes per-entity rendered output keyed by `(rendererKey, entityID, width, theme, propsHash)`.
  - Invalidated on `UIEntityUpdated` or presentation changes.

- Event Adapter (subscriber)
  - Subscribes to `ui.entities` (Watermill) or receives events via channels.
  - Decodes `UIEntity*` messages and forwards to `TimelineController`.

- Bubble Tea integration
  - TimelineModel implements `tea.Model` or exposes `Update(msg tea.Msg)` and `View()` for composition.
  - Supports `tea.WindowSizeMsg`, scrolling, and selection messages.

### 3) Public API proposal (pseudocode)

```go
package timeline

import (
    "time"
    tea "github.com/charmbracelet/bubbletea"
)

// IDs and descriptors
type EntityID struct {
    RunID   string `json:"run_id,omitempty"`
    TurnID  string `json:"turn_id,omitempty"`
    BlockID string `json:"block_id,omitempty"`
    LocalID string `json:"local_id,omitempty"`
    Kind    string `json:"kind"`
}

type RendererDescriptor struct {
    Key  string `json:"key"`  // e.g. renderer.llm_text.markdown.v1
    Kind string `json:"kind"` // e.g. llm_text, tool_calls_panel
}

// Lifecycle messages (already defined on the bus; re-declared as inputs)
type UIEntityCreated struct {
    ID        EntityID
    Renderer  RendererDescriptor
    Props     map[string]any
    StartedAt time.Time
    Labels    map[string]string
}

type UIEntityUpdated struct {
    ID        EntityID
    Patch     map[string]any
    Version   int64
    UpdatedAt time.Time
}

type UIEntityCompleted struct {
    ID     EntityID
    Result map[string]any
}

type UIEntityDeleted struct { ID EntityID }

// Renderer API
type Renderer interface {
    Key() string
    Kind() string
    // Render returns rendered string and computed height
    Render(props map[string]any, width int, theme string) (string, int, error)
    // RelevantPropsHash computes a stable hash for caching invalidation
    RelevantPropsHash(props map[string]any) string
}

type RendererRegistry interface {
    Register(r Renderer)
    GetByKey(key string) (Renderer, bool)
    GetByKind(kind string) (Renderer, bool)
}

// Timeline controller
type Controller interface {
    // Apply lifecycle changes
    OnCreated(evt UIEntityCreated)
    OnUpdated(evt UIEntityUpdated)
    OnCompleted(evt UIEntityCompleted)
    OnDeleted(evt UIEntityDeleted)

    // Presentation
    SetSize(width, height int)
    SetTheme(theme string)
    View() string

    // Selection
    SelectNext()
    SelectPrev()
}

// Bubble Tea integration
type Model struct { /* ... */ }
func NewModel(reg RendererRegistry, opts ...Option) *Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m *Model) View() string

// Event bridge examples
type WatermillSubscriber interface { Start() error; Close() error }
func NewWatermillSubscriber(bus WatermillBus, controller Controller) WatermillSubscriber
```

### 4) Internal data structures (pseudocode)

```go
type entityRecord struct {
    ID        EntityID
    Renderer  RendererDescriptor
    Props     map[string]any
    StartedAt time.Time
    UpdatedAt time.Time
    Version   int64
    Completed bool
}

type entityStore struct {
    // append-only order of entity IDs
    order  []EntityID
    // index by ID
    byID   map[EntityID]*entityRecord
}

type cacheKey struct {
    RendererKey string
    EntityKey   string // e.g., stable JSON of EntityID
    Width       int
    Theme       string
    PropsHash   string
}

type renderCache struct { m map[cacheKey]string; h map[cacheKey]int }
```

### 5) Rendering flow (pseudocode)

```go
func (c *controller) render() string {
    b := strings.Builder{}
    for _, id := range c.store.order { // append-only order
        rec := c.store.byID[id]
        r := c.registryBy(rec)
        key := cacheKey{r.Key(), hashID(id), c.width, c.theme, r.RelevantPropsHash(rec.Props)}
        if s, ok := c.cache.m[key]; ok {
            b.WriteString(s)
            b.WriteByte('\n')
            continue
        }
        s, _, _ := r.Render(rec.Props, c.contentWidth(), c.theme)
        c.cache.m[key] = s
        b.WriteString(s)
        b.WriteByte('\n')
    }
    return b.String()
}
```

### 6) Selection and scrolling

- Maintain an index into `store.order` as the selected entity.
- Provide `SelectNext/Prev` and calculate offsets for viewport.
- Integration with Bubble Tea’s `viewport.Model` is recommended but optional; timeline can also just return a large string for parent models to set.

### 7) Reference renderers (first wave)

 - LLMTextRenderer
   - `Key: renderer.llm_text.simple.v1`, `Kind: llm_text`
   - Uses chatstyle boxed output (lipgloss) with width-aware wrapping.
   - Props: `{ text: string, role?: string, metadata?: {...} }`

- ToolCallsPanelRenderer
  - `Key: renderer.tools.panel.v1`, `Kind: tool_calls_panel`
  - Props: `{ calls: [{ id, name, args, status, result? }], summary?: {...} }`

- DiffViewRenderer
  - `Key: renderer.diff.unified.v1`, `Kind: diff_view`
  - Props: `{ before: string, after: string, mode: string }`

### 8) Append-only ordering and timestamps

- Append entities in the order `UIEntityCreated` messages arrive.
- Use `StartedAt` for tiebreaking if merging multiple sources (rare).
- Update `UpdatedAt` on `UIEntityUpdated`; leave `StartedAt` unchanged.

### 9) Integration guide (simple-chat-agent demo)

1. Produce entity events in TurnStore translation layer
   - For streaming text: create → update on partials → complete on final.
   - For tool calls/results: create a single `tool_calls_panel` entity and patch it for each call/result.
   - Optionally emit a `diff_view` entity after final if the agent produced a patch.

   Demo note: In the chat demo, tool entities can be triggered by the backend on slash commands. The fake backend sends `TriggerWeatherToolMsg` / `TriggerWebSearchToolMsg` to the UI when it sees `/weather ...` or `/search ...`, so the timeline shows real tool_call entities without needing keyboard shortcuts.

2. Wire router handler
   - Subscribe to `ui.entities` and hand decoded `UIEntity*` messages to `TimelineController`.

3. Embed TimelineModel in the UI
   - Replace or augment `conversationui` usage by placing the timeline view inside the main viewport.
   - Maintain scroll-to-bottom semantics when the selected entity is near the end.
   - Keyboard triggers are bound via Bubble Tea keymap and matched using `key.Matches(...)`, so Alt-based shortcuts (e.g., `alt+w`, `alt+s`) work where terminals support them.

Demo renderer keys: Weather `renderer.tool.get_weather.v1`, WebSearch `renderer.tool.web_search.v1` (both `Kind: tool_call`).

LLM text renderer key used by the demo: `renderer.llm_text.simple.v1`.

Logging defaults: the CLI sets the global log level to Debug to reduce noise from Trace logs. Override with `BOBATEA_LOG_LEVEL=trace|debug|info|...`.

### 10) Testing strategy

- Unit tests
  - Store: creation/update/complete/delete; append-only order; timestamps.
  - Cache: invalidation on updates and on width/theme changes.
  - Renderers: verify output shape and width wrapping.

- Integration tests
  - Simulate sequences of UIEntity* events and verify the rendered view string.
  - Streaming scenarios: partial updates interleaved with tool events.

### 11) Performance considerations

- Avoid global re-render by caching per-entity output and only rebuilding changed segments.
- Debounce bursts of updates (optional) before recomputing the full view string.
- Keep props minimal; large payloads can carry a compact summary plus an expand action.

### 12) Migration plan

Phase A (parallel path):
- Add `timeline` alongside existing `conversationui`.
- TurnStore starts publishing `ui.entities` in addition to today’s `ui` stream.
- A small demo in `simple-chat-agent` toggles between conversation view and timeline view.

Phase B (feature parity):
- Implement tool panel and text renderers; wire selection and copy actions.
- Validate scroll and resize behavior.

Phase C (switch-over):
- Make timeline the default in demo; keep compatibility flag for legacy view.

### 13) Work plan (TODO)

1. Package scaffolding
   - [ ] Create `bobatea/pkg/timeline/` package skeleton (`controller.go`, `model.go`, `store.go`, `registry.go`, `cache.go`)
   - [ ] Define public types: `EntityID`, `RendererDescriptor`, lifecycle structs

2. Store and controller
   - [ ] Implement `entityStore` with append-only `order` and `byID` map
   - [ ] Implement `Controller` with OnCreated/OnUpdated/OnCompleted/OnDeleted
   - [ ] Add selection and size/theme handling

3. Registry and renderers
   - [ ] Implement `RendererRegistry`
   - [ ] Implement `LLMTextRenderer` (glamour-based)
   - [ ] Implement `ToolCallsPanelRenderer`
   - [ ] Implement `DiffViewRenderer`

4. Caching
   - [ ] Implement `renderCache` with key `(rendererKey, entityID, width, theme, propsHash)`
   - [ ] Wire cache invalidation on updates and on size/theme change

5. Bubble Tea integration
   - [ ] Implement `Model` (tea.Model) with `Update` and `View`
   - [ ] Optional: integrate `viewport.Model` with selection-aware scroll

6. Event bridge
   - [ ] Implement Watermill subscriber for `ui.entities` → controller
   - [ ] Add example code to `simple-chat-agent` to start subscriber and embed model

7. Tests and examples
   - [ ] Unit tests for store/controller/registry/cache
   - [ ] Integration test driving entity events and asserting the rendered output
   - [ ] Example program under `bobatea/cmd/timeline-demo`

### 14) Pseudocode: applying events

```go
func (c *controller) OnCreated(e UIEntityCreated) {
    if _, ok := c.store.byID[e.ID]; ok { return }
    rec := &entityRecord{ID: e.ID, Renderer: e.Renderer, Props: clone(e.Props), StartedAt: e.StartedAt}
    c.store.byID[e.ID] = rec
    c.store.order = append(c.store.order, e.ID)
}

func (c *controller) OnUpdated(e UIEntityUpdated) {
    if rec, ok := c.store.byID[e.ID]; ok {
        applyPatch(rec.Props, e.Patch)
        rec.Version = max(rec.Version, e.Version)
        rec.UpdatedAt = e.UpdatedAt
        c.cache.invalidateByID(e.ID)
    }
}

func (c *controller) OnCompleted(e UIEntityCompleted) {
    if rec, ok := c.store.byID[e.ID]; ok {
        if len(e.Result) > 0 { applyPatch(rec.Props, e.Result) }
        rec.Completed = true
        c.cache.invalidateByID(e.ID)
    }
}

func (c *controller) OnDeleted(e UIEntityDeleted) {
    if _, ok := c.store.byID[e.ID]; !ok { return }
    delete(c.store.byID, e.ID)
    c.store.order = filterOut(c.store.order, e.ID)
    c.cache.invalidateByID(e.ID)
}
```

### 15) References

- Turn-centric entity messaging and translation: `01-` and `03-` design docs in this folder
- Provider/middleware events: `geppetto/pkg/doc/topics/08-turns.md`, `geppetto/pkg/doc/topics/09-middlewares.md`
- Existing UI patterns: `bobatea/pkg/chat/conversation/model.go`


