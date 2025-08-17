---
Title: How to Build an Entity Renderer and Wire It Up
Slug: how-to-build-entity-renderer
Short: Step-by-step guide to creating a Bubble Tea entity model, registering it, and sending UI lifecycle messages.
Topics:
- bobatea
- ui
- timeline
- renderers
- tutorial
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Tutorial
---

## What you’ll learn

This tutorial walks through building a custom timeline entity renderer in Bobatea, wiring it into the `timeline.Registry`, and emitting lifecycle messages so it appears in the TUI. You’ll create a Bubble Tea model that reacts to selection, size, and key events, then register it so the `timeline.Controller` can instantiate it on `UIEntityCreated`.

We use concrete package and file names from the repository so you can cross-reference and extend the design.

## Architecture recap

- Package `bobatea/pkg/timeline` provides:
  - Entity identity and lifecycle types: `types.go`
  - Registry and controller: `registry.go`, `controller.go`
  - A thin shell around a `viewport.Model`: `shell.go`
- Package `bobatea/pkg/timeline/renderers` hosts Bubble Tea models implementing `timeline.EntityModel`.
- A host (e.g., `bobatea/pkg/chat/model.go` or a backend forwarder) emits lifecycle messages:
  - `UIEntityCreated`, `UIEntityUpdated`, `UIEntityCompleted`, `UIEntityDeleted`.

Key design points:
- The controller is append-only and instantiates models via the registry.
- The controller forwards selection and focus hints to models and, when entering is active, routes key messages to the selected model. TAB/shift+TAB are special-cased to route even when not entering.

## 1) Create your renderer model

Create a file like `bobatea/pkg/timeline/renderers/my_widget_model.go`:

```go
package renderers

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/go-go-golems/bobatea/pkg/timeline"
)

type MyWidgetModel struct {
    title    string
    details  string
    width    int
    selected bool
    expanded bool
}

func (m *MyWidgetModel) Init() tea.Cmd { return nil }

func (m *MyWidgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case timeline.EntitySelectedMsg:
        m.selected = true
    case timeline.EntityUnselectedMsg:
        m.selected = false
        m.expanded = false // collapse when unselected
    case timeline.EntityPropsUpdatedMsg:
        if t, ok := v.Patch["title"].(string); ok { m.title = t }
        if d, ok := v.Patch["details"].(string); ok { m.details = d }
    case timeline.EntitySetSizeMsg:
        m.width = v.Width
        return m, nil
    case tea.KeyMsg:
        if m.selected && (v.String() == "tab" || v.String() == "shift+tab") {
            m.expanded = !m.expanded
            return m, nil
        }
    }
    return m, nil
}

func (m *MyWidgetModel) View() string {
    sty := lipgloss.NewStyle()
    if m.selected {
        sty = sty.Bold(true)
    }
    s := m.title
    if m.expanded && m.details != "" {
        s += "\n\n" + m.details
    }
    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(s)
}

type MyWidgetFactory struct{}

func (MyWidgetFactory) Key() string  { return "renderer.my_widget.v1" }
func (MyWidgetFactory) Kind() string { return "my_widget" }
func (MyWidgetFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &MyWidgetModel{}
    m.Update(timeline.EntityPropsUpdatedMsg{Patch: initialProps})
    return m
}
```

Notes:
- Use `EntitySelectedMsg`/`EntityUnselectedMsg` to track `selected` state and collapse on unselect.
- Use `EntitySetSizeMsg` to adapt to width.
- Use `tea.KeyMsg` for interactions; TAB/shift+TAB will arrive when selected even if not in entering mode (thanks to controller routing policy).

## 2) Register the renderer in a registry

In your host (e.g., `bobatea/pkg/chat/model.go`), register the factory before entities are created:

```go
reg := timeline.NewRegistry()
reg.RegisterModelFactory(renderers.MyWidgetFactory{})
ctrl := timeline.NewController(reg)
```

If you’re using the chat model’s shell, pass a hook:

```go
renderersHook := func(reg *timeline.Registry) {
    reg.RegisterModelFactory(renderers.MyWidgetFactory{})
}
ui := chat.InitialModel(backend, chat.WithTimelineRegister(renderersHook))
```

## 3) Emit lifecycle messages

From your backend or integration, create and update entities using the types in `bobatea/pkg/timeline/types.go`:

```go
id := timeline.EntityID{LocalID: "abc123", Kind: "my_widget"}
created := timeline.UIEntityCreated{
    ID: id,
    Renderer: timeline.RendererDescriptor{Kind: "my_widget"},
    Props: map[string]any{"title": "Hello", "details": "More info"},
}
ctrl.OnCreated(created)

// Later, update properties
ctrl.OnUpdated(timeline.UIEntityUpdated{ID: id, Patch: map[string]any{"details": "Updated"}})

// Complete when done
ctrl.OnCompleted(timeline.UIEntityCompleted{ID: id})
```

If you’re using the shell, call `shell.OnCreated/OnUpdated/OnCompleted` instead; it will refresh the viewport:

```go
shell := timeline.NewShell(reg)
shell.OnCreated(created)
shell.OnUpdated(timeline.UIEntityUpdated{ID: id, Patch: map[string]any{"details": "Updated"}})
shell.OnCompleted(timeline.UIEntityCompleted{ID: id})
```

## 4) Selection and key routing

- Selection helpers: `SelectNext`, `SelectPrev`, `SelectLast`, `SetSelectionVisible`, `EnterSelection`, `ExitSelection`.
- Routing policy (in `controller.go`):
  - Keys are routed to the selected model when entering is true.
  - TAB/shift+TAB are routed even when not entering.
- Models receive selection and focus hints via messages:
  - `EntitySelectedMsg`, `EntityUnselectedMsg`, `EntityFocusMsg`, `EntityBlurMsg`.

## 5) Copy-to-clipboard and side effects

- To request text copy from a model:
  - The model handles `timeline.EntityCopyTextMsg` and returns a command that emits `timeline.CopyTextRequestedMsg{Text: ...}`.
  - The host (e.g., `chat` model) listens for `CopyTextRequestedMsg` and performs the actual clipboard write.

## 6) Testing with the demo

- See `bobatea/cmd/timeline-demo/main.go` for a standalone example that generates entities and updates.
- Run:

```bash
cd bobatea && go run ./cmd/timeline-demo
```

Use the demo as a reference for how to hook up lifecycle events to the registry and controller.

## 7) Tips and best practices

- Keep `View()` fast; cache derived strings if needed.
- Use `EntitySetSizeMsg` for width-aware wrapping instead of computing width in constructors.
- Collapse large details when unselected; use TAB for quick toggles (users can navigate faster).
- Prefer routing keys via the controller; avoid duplicating key policy in hosts.

## References

- `bobatea/pkg/timeline/types.go`
- `bobatea/pkg/timeline/registry.go`
- `bobatea/pkg/timeline/controller.go`
- `bobatea/pkg/timeline/shell.go`
- `bobatea/pkg/timeline/renderers/*`
- `bobatea/pkg/chat/model.go`



