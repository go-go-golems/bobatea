--- 
Title: Bobatea Chat Architecture and Timeline Renderers
Slug: chat-architecture-and-renderers
Short: How the chat UI works, how messages flow through the backend, and how to register timeline renderers and interactive models
Topics:
- chat
- timeline
- renderers
- bubbletea
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Bobatea Chat Architecture and Timeline Renderers

The Bobatea chat UI composes a Bubble Tea model with a timeline engine for rendering and interacting with “entities” (messages, tool calls, panels). The chat model translates backend streaming events into timeline lifecycle events, while the timeline controller renders entities using registered renderers or, preferably, interactive Bubble Tea models. This design cleanly separates input handling, backend orchestration, and rich per-entity UI, enabling custom renderers, tool panels, and interactive widgets.

### Key concepts at a glance

- Timeline-first: messages and tools are timeline entities, each rendered by a renderer or an interactive entity model.
- Two renderer layers:
  - Stateless `Renderer` for simple text rendering.
  - Interactive `EntityModel` via `EntityModelFactory` for full Bubble Tea per-entity models.
- Chat model is responsible for:
  - Input box and app state.
  - Translating backend stream messages into timeline lifecycle events.
  - Delegating selection/focus to timeline controller and routing keys to selected entity while “entering”.
- Backends stream content via Bubble Tea messages and can also emit entity lifecycle messages directly (e.g., agent tool calls).

## Architecture Overview

The chat UI is a Bubble Tea model with:
- A `viewport` for the timeline output
- A `textArea` for user input
- A `timeline.Controller` for rendering entities and handling selection/focus
- A `timeline.Registry` for renderers and interactive model factories
- A `Backend` abstraction for streaming completions and tool-originated UI events

Core initialization:

```126:170:bobatea/pkg/chat/model.go
// Initialize timeline components
ret.timelineReg = timeline.NewRegistry()
// Register interactive entity model factories
ret.timelineReg.RegisterModelFactory(renderers.NewLLMTextFactory())
ret.timelineReg.RegisterModelFactory(renderers.ToolCallsPanelFactory{})
ret.timelineReg.RegisterModelFactory(renderers.PlainFactory{})
if ret.timelineRegHook != nil {
	ret.timelineRegHook(ret.timelineReg)
}
ret.timelineCtrl = timeline.NewController(ret.timelineReg)
ret.entityVers = map[string]int64{}
```

Program setup (logging, flags, backend, and registration hook):

```84:106:bobatea/cmd/chat/main.go
func runChatWithOptions(backendFactory func() chat.Backend, tlHook func(*timeline.Registry)) {
	status := &chat.Status{}
	backend := backendFactory()
	options := []tea.ProgramOption{tea.WithMouseCellMotion(), tea.WithAltScreen()}
	model := chat.InitialModel(backend, chat.WithStatus(status), chat.WithTimelineRegister(tlHook))
	p := tea.NewProgram(model, options...)
	if setterBackend, ok := backend.(interface{ SetProgram(*tea.Program) }); ok {
		setterBackend.SetProgram(p)
	}
	if _, err := p.Run(); err != nil { /* ... */ }
}
```

## Message Flow and State Machine

The chat model runs a small state machine to coordinate input, streaming, selection, and file-saving. It translates backend stream messages into timeline lifecycle events:

```412:446:bobatea/pkg/chat/model.go
// Translate stream messages to timeline entity lifecycle
case conversationui.StreamStartMsg:
    id := v.ID.String()
    m.entityVers[id] = 0
    m.timelineCtrl.OnCreated(timeline.UIEntityCreated{
        ID:        timeline.EntityID{LocalID: id, Kind: "llm_text"},
        Renderer:  timeline.RendererDescriptor{Kind: "llm_text"},
        Props:     map[string]any{"role": "assistant", "text": ""},
        StartedAt: time.Now(),
    })
case conversationui.StreamCompletionMsg:
    id := v.ID.String()
    m.entityVers[id] = m.entityVers[id] + 1
    m.timelineCtrl.OnUpdated(timeline.UIEntityUpdated{
        ID:        timeline.EntityID{LocalID: id, Kind: "llm_text"},
        Patch:     map[string]any{"text": v.Completion},
        Version:   m.entityVers[id],
        UpdatedAt: time.Now(),
    })
case conversationui.StreamDoneMsg:
    id := v.ID.String()
    m.timelineCtrl.OnCompleted(timeline.UIEntityCompleted{
        ID:     timeline.EntityID{LocalID: id, Kind: "llm_text"},
        Result: map[string]any{"text": v.Completion},
    })
```

External lifecycle events (e.g., tools) are accepted directly and rendered:

```468:507:bobatea/pkg/chat/model.go
case timeline.UIEntityCreated:
    m.timelineCtrl.OnCreated(msg_)
    if m.scrollToBottom { m.viewport.SetContent(m.timelineCtrl.View()); m.viewport.GotoBottom() }
    return m, nil
case timeline.UIEntityUpdated:
    m.timelineCtrl.OnUpdated(msg_)
    if m.scrollToBottom { m.viewport.SetContent(m.timelineCtrl.View()); m.viewport.GotoBottom() }
    return m, nil
case timeline.UIEntityCompleted:
    m.timelineCtrl.OnCompleted(msg_)
    if m.scrollToBottom { m.viewport.SetContent(m.timelineCtrl.View()); m.viewport.GotoBottom() }
    return m, nil
case timeline.UIEntityDeleted:
    m.timelineCtrl.OnDeleted(msg_)
    if m.scrollToBottom { m.viewport.SetContent(m.timelineCtrl.View()); m.viewport.GotoBottom() }
    return m, nil
```

States:
- StateUserInput: typing in the input box
- StateStreamCompletion: backend streaming; input is disabled
- StateMovingAround: selection mode; arrow keys and page up/down navigate the timeline; “enter” toggles entering mode; keys are routed to the selected entity
- StateSavingToFile: shows file picker
- StateError: displays error view

Selection/entering routing:

```313:350:bobatea/pkg/chat/model.go
if m.state == StateMovingAround {
    switch msg_.String() {
    case "enter":
        m.timelineCtrl.EnterSelection()
        m.viewport.SetContent(m.timelineCtrl.View())
        return m, nil
    case "esc":
        if m.timelineCtrl.IsEntering() {
            m.timelineCtrl.ExitSelection()
            m.viewport.SetContent(m.timelineCtrl.View())
            return m, nil
        }
        m.state = StateUserInput
        m.textArea.Focus()
        m.updateKeyBindings()
        m.timelineCtrl.SetSelectionVisible(false)
        m.timelineCtrl.Unselect()
        m.viewport.SetContent(m.timelineCtrl.View())
        return m, nil
    }
    if m.timelineCtrl.IsEntering() {
        cmd := m.timelineCtrl.HandleMsg(msg_)
        m.viewport.SetContent(m.timelineCtrl.View())
        return m, cmd
    }
}
```

## Backend Interface and Streaming

Backends abstract streaming completions and tool-originated UI events. The chat UI currently uses `SubmitPrompt` and relies on backends to emit the stream messages and, optionally, timeline events for tools.

```33:56:bobatea/pkg/chat/backend.go
type Backend interface {
	Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error)
	SubmitPrompt(ctx context.Context, prompt string) (tea.Cmd, error)
	Interrupt()
	Kill()
	IsFinished() bool
}
```

Submit path in the chat model:
- Adds the user message entity to the timeline.
- Calls `backend.SubmitPrompt`, which should return a `tea.Cmd` that streams messages back into the program.

```909:940:bobatea/pkg/chat/model.go
// Add entity to timeline (user message)
m.timelineCtrl.OnCreated(timeline.UIEntityCreated{ ID: timeline.EntityID{LocalID: id, Kind: "llm_text"}, Renderer: timeline.RendererDescriptor{Kind: "llm_text"}, Props: map[string]any{"role": "user", "text": userMessage}})
m.timelineCtrl.OnCompleted(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: id, Kind: "llm_text"}})
// Start backend for prompt
backendCmd := func() tea.Msg {
    ctx := context2.Background()
    cmd, err := m.backend.SubmitPrompt(ctx, userMessage)
    if err != nil { return ErrorMsg(err) }
    return cmd()
}
```

Example backend (fake) that emits stream and tool entities:

```69:110:bobatea/cmd/chat/fake-backend.go
// Recognize tool commands and emit lifecycle messages directly
if strings.HasPrefix(content, "/weather ") {
	f.p.Send(timeline.UIEntityCreated{
		ID:       timeline.EntityID{LocalID: localID, Kind: "tool_call"},
		Renderer: timeline.RendererDescriptor{Key: "renderer.tool.get_weather.v1", Kind: "tool_call"},
		Props:    map[string]any{"location": "Paris", "units": "celsius"},
		StartedAt: time.Now(),
	})
	f.p.Send(timeline.UIEntityUpdated{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}, Patch: map[string]any{"result": "22C, Sunny"}, Version: 1, UpdatedAt: time.Now()})
	f.p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}})
	return
}
```

## Timeline Registry and Controller

The registry maps renderer `Key`/`Kind` to either stateless `Renderer` or interactive `EntityModelFactory`. Prefer `EntityModelFactory` for interactive experiences.

```1:35:bobatea/pkg/timeline/registry.go
type Renderer interface {
	Key() string
	Kind() string
	Render(props map[string]any, width int, theme string) (string, int, error)
	RelevantPropsHash(props map[string]any) string
}
type EntityModel interface {
	tea.Model
	View() string
	OnProps(patch map[string]any)
	OnCompleted(result map[string]any)
	SetSize(width, height int)
	Focus()
	Blur()
}
type EntityModelFactory interface {
	Key() string
	Kind() string
	NewEntityModel(initialProps map[string]any) EntityModel
}
```

Controller instantiates models when entities are created, updates props on patch, and renders either interactive models or stateless renderers. It also routes messages to the selected model in entering mode.

```53:76:bobatea/pkg/timeline/controller.go
func (c *Controller) OnCreated(e UIEntityCreated) {
	rec := &entityRecord{ID: e.ID, Renderer: e.Renderer, Props: cloneMap(e.Props), StartedAt: e.StartedAt.UnixNano()}
	// Prefer factory by key
	if e.Renderer.Key != "" {
		if f, ok := c.reg.GetModelFactoryByKey(e.Renderer.Key); ok {
			rec.model = f.NewEntityModel(rec.Props)
			rec.model.SetSize(c.width, c.height)
			if c.theme != "" { rec.model.OnProps(map[string]any{"theme": c.theme}) }
		}
	}
	// Fallback to factory by kind
	if rec.model == nil && e.Renderer.Kind != "" {
		if f, ok := c.reg.GetModelFactoryByKind(e.Renderer.Kind); ok {
			rec.model = f.NewEntityModel(rec.Props)
			rec.model.SetSize(c.width, c.height)
			if c.theme != "" { rec.model.OnProps(map[string]any{"theme": c.theme}) }
		}
	}
	c.store.add(rec)
	if c.selected < 0 { c.selected = 0 }
}
```

Rendering and routing:

```165:178:bobatea/pkg/timeline/controller.go
sel := c.selectionVisible && c.selected >= 0 && keyID(id) == keyID(c.store.order[c.selected])
rec.model.OnProps(map[string]any{"selected": sel})
if sel { rec.model.Update(EntitySelectedMsg{ID: rec.ID}) } else { rec.model.Update(EntityUnselectedMsg{ID: rec.ID}) }
if sel && c.entering { rec.model.Focus() } else { rec.model.Blur() }
s := rec.model.View()
```

```293:313:bobatea/pkg/timeline/controller.go
// Route a Bubble Tea message to selected model when entering
func (c *Controller) HandleMsg(msg tea.Msg) tea.Cmd {
	if !c.entering { return nil }
	if c.selected < 0 || c.selected >= len(c.store.order) { return nil }
	id := c.store.order[c.selected]
	rec, ok := c.store.get(id)
	if !ok || rec.model == nil { return nil }
	m2, cmd := rec.model.Update(msg)
	if mm, ok := m2.(EntityModel); ok { rec.model = mm }
	return cmd
}
```

## Registering Renderers and Interactive Models

You can register both stateless renderers and interactive models.

- In the `chat` command, register demo tool renderers and a test interactive widget via the `WithTimelineRegister` hook:

```75:81:bobatea/cmd/chat/main.go
runChatWithOptions(func() chat.Backend { return NewFakeBackend() }, func(reg *timeline.Registry) {
	reg.Register(&ToolWeatherRenderer{})
	reg.Register(&ToolWebSearchRenderer{})
	reg.RegisterModelFactory(CheckboxFactory{})
})
```

- Stateless tool renderer example (`Key` identifies a unique renderer, `Kind` groups similar entities):

```13:51:bobatea/cmd/chat/tool_renderers.go
type ToolWeatherRenderer struct{}
func (r *ToolWeatherRenderer) Key() string  { return "renderer.tool.get_weather.v1" }
func (r *ToolWeatherRenderer) Kind() string { return "tool_call" }
func (r *ToolWeatherRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    // choose style based on selection; render lines into a styled panel
    // return content and height
}
```

- Interactive entity example: a checkbox model with a factory keyed as a tool renderer:

```186:193:bobatea/cmd/chat/tool_renderers.go
type CheckboxFactory struct{}
func (CheckboxFactory) Key() string  { return "renderer.test.checkbox.v1" }
func (CheckboxFactory) Kind() string { return "" }
func (CheckboxFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &CheckboxModel{}
    m.OnProps(initialProps)
    return m
}
```

- Built-in interactive LLM text model factory:

```114:127:bobatea/pkg/timeline/renderers/llm_text_model.go
type LLMTextFactory struct{}
func (f *LLMTextFactory) Key() string  { return "renderer.llm_text.simple.v1" }
func (f *LLMTextFactory) Kind() string { return "llm_text" }
func (f *LLMTextFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &LLMTextModel{ renderer: f.renderer }
    m.OnProps(initialProps)
    return m
}
```

### Choosing Key vs. Kind

- Use `Key` for a specific implementation/version (e.g., renderer.tool.get_weather.v1).
- Use `Kind` to handle generic entity types (e.g., llm_text, tool_call).
- The controller prefers a `Key`-matched `EntityModelFactory`. If no key factory is found, it tries `Kind`. If no model exists, it uses the stateless `Renderer`.

## Extending the Chat with Custom Entities

To add a new tool panel or widget:

1) Implement an interactive `EntityModel` and `EntityModelFactory`:
- Implement `Init`, `Update`, `View`, `OnProps`, `OnCompleted`, `SetSize`, `Focus`, `Blur`.
- Provide `Key` or `Kind` in the factory.

Pseudocode:

```go
type MyPanelModel struct { width int; selected bool; /* ... */ }
func (m *MyPanelModel) Init() tea.Cmd { return nil }
func (m *MyPanelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle selection, keys, props */ return m, nil }
func (m *MyPanelModel) View() string { /* render */ return "..."}
func (m *MyPanelModel) OnProps(patch map[string]any) { /* apply props */ }
func (m *MyPanelModel) OnCompleted(result map[string]any) {}
func (m *MyPanelModel) SetSize(w, h int) { m.width = w }
func (m *MyPanelModel) Focus() {}
func (m *MyPanelModel) Blur()  {}

type MyPanelFactory struct{}
func (MyPanelFactory) Key() string  { return "renderer.tool.my_panel.v1" }
func (MyPanelFactory) Kind() string { return "tool_call" }
func (MyPanelFactory) NewEntityModel(p map[string]any) timeline.EntityModel { m := &MyPanelModel{}; m.OnProps(p); return m }
```

2) Register it via the hook:

```go
chat.InitialModel(backend, chat.WithTimelineRegister(func(reg *timeline.Registry){
    reg.RegisterModelFactory(MyPanelFactory{})
}))
```

3) Emit lifecycle events either from the backend or the UI:
- Backend (preferred for agent-produced tools): send `UIEntityCreated` -> `UIEntityUpdated` patches as state evolves -> `UIEntityCompleted`.
- UI (for quick demos): call `timelineCtrl.OnCreated/OnUpdated/OnCompleted`.

## Keyboard and Interaction Model

- Input mode (default): type prompt; Enter submits; Esc switches to selection.
- Selection mode: navigation keys move selection; entities are highlighted.
- Entering mode: hitting Enter while in selection focuses the entity model; subsequent keys are routed to it; Esc exits entering; Esc again exits selection to input.
- Copy flows:
  - `alt+c` in selection requests copy from selected entity.
  - Copy last assistant text when not in selection is supported via messages to timeline.

Relevant controller routing:

```332:341:bobatea/pkg/timeline/controller.go
func (c *Controller) EnterSelection()  { c.entering = true }
func (c *Controller) ExitSelection()   { c.entering = false }
func (c *Controller) IsEntering() bool { return c.entering }
```

## Running the Chat

- Commands:
  - `chat fake` or `chat chat` (alias) runs with the demo backend and registers demo renderers.
- Logging:
  - Writes to `/tmp/fake-chat.log`; set `BOBATEA_LOG_LEVEL` to adjust verbosity.

```44:68:bobatea/cmd/chat/main.go
logFile, _ := os.OpenFile("/tmp/fake-chat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
// zerolog configuration...
```

## Integration Hooks and Embedding

- `WithTimelineRegister(func(*timeline.Registry))` allows you to register custom renderers or factories from the caller.
- `WithStatus(*chat.Status)` shares UI state out-of-band (state, input text, selection index, message count, error).
- `WithAutoStartBackend(bool)` can auto-trigger initial backend start message.
- `GetUIState()` exposes `Status` as a map usable for embedding.

```99:125:bobatea/pkg/chat/model.go
type ModelOption func(*model)
func WithTimelineRegister(hook func(*timeline.Registry)) ModelOption
func WithStatus(status *Status) ModelOption
func WithAutoStartBackend(autoStartBackend bool) ModelOption
```

## Troubleshooting and Best Practices

- Prefer interactive models via `EntityModelFactory` for anything with focus/keyboard interactions or changing state.
- Use `Key` to version your renderers (e.g., `.v1`, `.v2`).
- Backends should emit stream messages for LLM output and can emit tool entity lifecycle events directly.
- Respect selection/entering: models should respond to `EntitySelectedMsg`, `EntityUnselectedMsg`, and patch updates via `EntityPropsUpdatedMsg`.

For writing style and documentation structure, see `glaze help how-to-write-good-documentation-pages` [[memory:5699956]].

### Minimal end-to-end example (pseudocode)

```go
// 1) Register custom factory
model := chat.InitialModel(backend, chat.WithTimelineRegister(func(reg *timeline.Registry) {
    reg.RegisterModelFactory(MyPanelFactory{})
}))

// 2) When your backend detects a tool call or needs UI, emit lifecycle
p.Send(timeline.UIEntityCreated{
    ID: timeline.EntityID{LocalID: id, Kind: "tool_call"},
    Renderer: timeline.RendererDescriptor{Key: "renderer.tool.my_panel.v1", Kind: "tool_call"},
    Props: map[string]any{"title": "My Panel", "status": "starting"},
    StartedAt: time.Now(),
})
// Later: updates
p.Send(timeline.UIEntityUpdated{ID: timeline.EntityID{LocalID: id, Kind: "tool_call"}, Patch: map[string]any{"status": "done"}, Version: 1, UpdatedAt: time.Now()})
p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: id, Kind: "tool_call"}})
```

- The controller will instantiate `MyPanelModel` from `MyPanelFactory`, route selection/focus into it, and render it as part of the timeline.
- The chat model will keep the viewport in sync and manage state transitions across input, selection, and streaming.

---

- Added a complete documentation page describing the chat architecture, backend flow, and renderer/model registration with code citations to `bobatea/pkg/chat/model.go`, `bobatea/cmd/chat/main.go`, `bobatea/pkg/chat/backend.go`, `bobatea/pkg/timeline/*`, and demo renderers in `bobatea/cmd/chat/tool_renderers.go`.
- Included YAML front matter, topic-focused section intros, and short runnable-style pseudocode per the `glaze` guide.
- Covered lifecycle messages mapping, keyboard/selection routing, and how to add custom interactive panels.