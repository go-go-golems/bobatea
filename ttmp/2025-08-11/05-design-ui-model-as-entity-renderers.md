---
Title: UI Models as Entity Renderers (Interactive Timeline Blocks)
Slug: ui-model-as-entity-renderers
Short: Make timeline entity renderers optionally be Bubble Tea models to enable sub-widgets, timers, and input handling; specify deletion semantics
Topics:
- bobatea
- ui
- timeline
- bubbletea
- architecture
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Design
---

## 1) Purpose and scope

The current `timeline` renderers are stateless functions that return strings. This is ideal for simple blocks (assistant/user text), but limiting for interactive sub-widgets that need timers, incremental updates, and local input handling (e.g., tool panels with progress, results, expand/collapse, copy buttons).

This design proposes an optional path where a renderer can be a full Bubble Tea model, managed per-entity by the `timeline.Controller`. We also formalize the use of `UIEntityDeleted` to asynchronously remove entities from the timeline.

Goals:
- Keep simple renderers simple (string-based), with zero behavior change.
- Allow advanced renderers to be stateful UI models with their own `Init/Update/View`.
- Route relevant messages (ticks, keys, mouse) and size/theme to child models.
- Cleanly handle deletion via `UIEntityDeleted` and optional soft-removal UX.

## 2) Terminology

- Stateless renderer: implements today’s `timeline.Renderer` (Render(props) -> string).
- Interactive entity model: a per-entity Bubble Tea model (`tea.Model`) created by a factory; it owns local state and returns a `View()` string.

## 3) Public API additions (pseudocode)

```go
package timeline

import tea "github.com/charmbracelet/bubbletea"

// Existing
type Renderer interface {
    Key() string
    Kind() string
    Render(props map[string]any, width int, theme string) (string, int, error)
    RelevantPropsHash(props map[string]any) string
}

// New: an interactive renderer instance per entity
type EntityModel interface {
    tea.Model
    // Lifecycle hooks from bus
    OnProps(p map[string]any)
    OnCompleted(result map[string]any)
    OnDeleted()
    // Presentation hooks
    SetSize(width, height int)
    SetTheme(theme string)
}

// Factory registered in the registry for a given Key/Kind
type EntityModelFactory interface {
    Key() string
    Kind() string
    NewEntityModel(initialProps map[string]any) EntityModel
}

// Registry supports both kinds
type Registry interface {
    // Unified API: one Register that introspects the type
    Register(any)
    // Get resolves to an interactive model factory if present, else a stateless renderer
    GetByKey(key string) (renderer Renderer, modelFactory EntityModelFactory, ok bool)
    GetByKind(kind string) (renderer Renderer, modelFactory EntityModelFactory, ok bool)
}

// Controller additions
type Controller struct { /* ... */ }

// For parent UIs that want to bubble messages into child models
func (c *Controller) Update(msg tea.Msg) (tea.Cmd, bool) // returns handled flag
func (c *Controller) Init() tea.Cmd                      // init child models
```

Notes:
- Existing `Renderer` stays unchanged for backwards compatibility.
- Callers keep using `reg.Register(&MyRenderer{})` or `reg.Register(&MyFactory{})`. The registry uses a type switch to store either/both. `Get*` returns both, and the controller prefers the factory if available.
- A Key/Kind may register both; the factory takes precedence for rendering.

## 4) Controller internals

Store additions (per entity):
- `model EntityModel` (optional; nil for stateless)
- `height int` cache for the last rendered height

Lifecycle handling:
- `OnCreated`: if a model factory exists, instantiate `model := f.NewEntityModel(Props)`; store model; call `SetSize/SetTheme`, enqueue `model.Init()`.
- `OnUpdated`: if model present, call `model.OnProps(patch)`; else, apply to props as today and invalidate cache.
- `OnCompleted`: call `model.OnCompleted(result)` if present; else, patch props and invalidate cache.
- `OnDeleted`: call `model.OnDeleted()` if present; remove from store/order; invalidate caches.

Rendering:
- If `rec.model != nil`, call `view := rec.model.View()`; do not use the string cache for interactive models (view already local-state-driven). Width is already pushed via `SetSize`.
- If stateless, keep today’s cache keyed by `(rendererKey, entityID, width, theme, propsHash)`.

Message routing:
- Add `Controller.Update(msg tea.Msg)` that routes to the focused/selected entity model first. If unhandled or no focus, route to all models that might need ticks (e.g., via a small broadcast for `TickMsg`). Aggregate returned commands via `tea.Batch`.
- Add `Controller.Init()` to return a batch of child `Init()` commands at program start or after creation.

Selection/focus:
- Keep `Controller.SelectNext/Prev`. When selected entity changes, call `Focus/Blur` on models if they implement optional interfaces (or encode focus inside `OnProps`/`SetTheme`).

## 5) Parent integration (chat model)

- On `tea.WindowSizeMsg`, continue to call `timelineCtrl.SetSize(w,h)`; controller will fan out to child models.
- In `Update`, forward certain messages (e.g., `tea.TickMsg`, key/mouse when timeline is focused) into `timelineCtrl.Update(msg)`. Use the returned `tea.Cmd` in the parent’s command batch.
- For program startup and when entities are created, ensure `timelineCtrl.Init()` (or entity-specific init) commands are included.

This keeps the chat model as the host application while allowing embedded entity UIs.

Selection/focus behavior (ESC to select entities):
- When the user presses ESC in chat, the chat model switches to a "moving-around" mode and delegates selection to the timeline controller.
- Controller maintains `selected` index; `SelectNext/Prev` move the selection; `View()` renders the selected entity in the highlighted style.
- When focus moves to timeline, chat textbox is blurred and keybindings update; returning focus to input restores previous behavior.
- Interactive models receive focus changes implicitly (optional): controller may call `OnProps({"focused": true/false})` or provide optional `Focus()/Blur()` hooks.

## 6) UIEntityDeleted design

`types.go` already declares:

```go
type UIEntityDeleted struct { ID EntityID }
```

Semantics:
- Emitted by the backend (or orchestrator) to remove an entity from the timeline asynchronously.
- Controller behavior: `store.remove(id)`, clear any model instance, invalidate caches. Selection index is clamped to the new range.

Optional UX:
- Soft delete with animation: introduce a label/prop (e.g., `{"transient": true}`) and a policy in the controller to delay physical removal until after an animation period. This remains a presentation concern; the API continues to use `UIEntityDeleted` as the removal trigger.

Selection on delete:
- If the deleted entity was selected, selection moves to the previous entity (or clamped to end if last). The parent chat model should recompute the viewport and keep scroll-to-bottom semantics if enabled.

## 7) Example: interactive Web Search entity

Renderer factory:

```go
type WebSearchModel struct {
    vpWidth int
    theme   string
    query   string
    results []string
    spinIdx int
}

func (m *WebSearchModel) Init() tea.Cmd { return tea.Tick(time.Second/8, func(t time.Time) tea.Msg { return tick{} }) }
func (m *WebSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case tick:
        m.spinIdx++
        return m, tea.Tick(time.Second/8, func(time.Time) tea.Msg { return tick{} })
    }
    return m, nil
}
func (m *WebSearchModel) View() string { /* render boxed title + vertical results + spinner */ }
func (m *WebSearchModel) OnProps(p map[string]any) { /* append to results, update spinIdx */ }
func (m *WebSearchModel) OnCompleted(res map[string]any) { /* stop spinner */ }
func (m *WebSearchModel) OnDeleted() {}
func (m *WebSearchModel) SetSize(w, h int) { m.vpWidth = w }
func (m *WebSearchModel) SetTheme(theme string) { m.theme = theme }

type WebSearchFactory struct{}
func (WebSearchFactory) Key() string  { return "renderer.tool.web_search.v1" }
func (WebSearchFactory) Kind() string { return "tool_call" }
func (WebSearchFactory) NewEntityModel(props map[string]any) EntityModel {
    return &WebSearchModel{query: getString(props, "query")}
}
```

Backend continues to emit `UIEntityCreated` (with initial props), multiple `UIEntityUpdated` (accumulating `results`) and finally `UIEntityCompleted`.

## 8) Backwards compatibility

- Existing stateless renderers and registry APIs continue to work unmodified.
- Only renderers registered via `RegisterModelFactory` are treated as interactive.
- Caching remains for stateless renderers; interactive models are not cached beyond their own internal state.

## 9) Testing strategy

- Unit tests for controller:
  - Creation of interactive model; propagation of size/theme; deletion clearing state.
  - Routing of `TickMsg` and key messages to the selected model.
  - Mixed environments: some entities interactive, some stateless.

- Integration tests:
  - Simulated web search with multiple `Updated` events; verify visual output grows vertically and spinner stops on completed.
  - Deletion tests to ensure selection index and view update correctly.

## 10) Migration plan

1. Extend `Registry` and `Controller` with the new interfaces; keep existing methods.
2. Convert demo `web_search` tool to an interactive model (keep weather stateless or make it interactive as well).
3. Wire message routing in chat model: forward `TickMsg` and key events to `timeline.Controller.Update` when focus is in the timeline.
4. Add examples and documentation.

---

Appendix A: Labels for presentation policies

- Use `Labels` in `UIEntityCreated` to hint presentation (e.g., `{"transient":"true"}`, `{"priority":"low"}`) — the controller may ignore them, but advanced UIs can use them to adjust transitions or pruning.

Appendix B: Conversation dependencies to remove or adapt

Even with timeline entities in place, the chat model still references conversation structures for selection and clipboard helpers. Key areas to refactor:
- `bobatea/pkg/chat/model.go`:
  - Seeding from `conversationManager.GetConversation()` at init (switch to seeding from persisted entity stream or TurnStore snapshot).
  - Status metrics `len(m.conversationManager.GetConversation())` (replace with entity counts from controller).
  - `startBackend()` passes `GetConversation()` to backends (define a new adapter that reads from the entity stream or turns, or build the minimal prompt context from entities).
  - Selection paths using `m.conversation.SelectedIdx()` and `m.conversation.Conversation()` in copy helpers (switch to timeline selection and entity text extraction from props).

Appendix C: Test entity – Checkbox selector

Purpose: demonstrate a minimal interactive entity that holds local state and can signal choices to the backend.

Pseudocode:

```go
type CheckboxMsg struct{ Checked bool }

type CheckboxModel struct {
    label   string
    checked bool
    width   int
}

func (m *CheckboxModel) Init() tea.Cmd { return nil }
func (m *CheckboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch k := msg.(type) {
    case tea.KeyMsg:
        if k.String() == "space" { m.checked = !m.checked; return m, func() tea.Msg { return CheckboxMsg{Checked: m.checked} } }
    }
    return m, nil
}
func (m *CheckboxModel) View() string { box := "[ ]"; if m.checked { box = "[x]" }; return fmt.Sprintf("%s %s", box, m.label) }
func (m *CheckboxModel) OnProps(p map[string]any) { if v, ok := p["checked"].(bool); ok { m.checked = v } }
func (m *CheckboxModel) OnCompleted(_ map[string]any) {}
func (m *CheckboxModel) OnDeleted() {}
func (m *CheckboxModel) SetSize(w, _ int) { m.width = w }
func (m *CheckboxModel) SetTheme(_ string) {}

type CheckboxFactory struct{}
func (CheckboxFactory) Key() string  { return "renderer.test.checkbox.v1" }
func (CheckboxFactory) Kind() string { return "test" }
func (CheckboxFactory) NewEntityModel(p map[string]any) EntityModel { return &CheckboxModel{ label: getString(p, "label") } }
```

Backend communication ideas for interactive entities:
- Message bubbling to parent:
  - Entity models return a `tea.Cmd` that emits a strongly-typed UI event (e.g., `CheckboxMsg`) which the parent chat model (or a thin adapter) translates into a backend API call or a `timeline.UIEntityUpdated` reflecting the new state.
- Event bus:
  - Provide a shared event sink in `Controller` (callback or channel) so entity models can `emit(UIEvent{EntityID, Type, Payload})`. A dedicated goroutine forwards these to the backend (HTTP, WS, or Watermill topics).
- User backend HTTP endpoints:
  - Models call a function exposed by the parent (injected via context) that performs HTTP calls to the `/user` backend. Useful when entities need to fetch or commit state server-side.
- Timeline patching:
  - For purely UI-local state, models send `UIEntityUpdated` to themselves (via controller) to persist state in props. The backend may observe the same bus and react.

These approaches can co-exist: small UI actions produce a local `Updated` patch; significant actions are also forwarded to the backend via a sink.


