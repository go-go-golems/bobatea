## Design: A Turn-Centric Visualization Approach for the Chat UI

Date: 2025-08-11

### Purpose and scope

This document analyzes the current chat UI built around `conversation.Manager` and `conversationui.Model`, identifies gaps for the newer Turn/Block model, and proposes turn-centric rendering designs. The primary goal is to expose a “visual rendering” of a Turn’s output: sometimes just LLM text, often richer widgets (tool call panes, structured result views, metadata panels), while keeping Bubble Tea ergonomics and streaming performance in mind.

### Summary of findings

- The existing UI is conversation-centric: it renders a linear list of chat messages using `glamour` for markdown and caches rendered strings per `Message.ID`.
- Streaming is integrated via custom `Stream*Msg` types; the UI appends a placeholder assistant message on `StreamStartMsg` and replaces its content on `StreamCompletionMsg`/`StreamDoneMsg`.
- With the new Turns data model, engines append structured blocks (`llm_text`, `tool_call`, `tool_use`) to a Turn. Mapping turns to “conversation messages” is supported by helpers, but the UI cannot easily display tool calls/results as rich widgets.
- The rendering layer is tightly coupled to `glamour`-rendered strings and viewport content updates. This inhibits composing richer visual sections per Turn and per Block.

### Current architecture: where rendering and streaming happen

- The top-level UI model embeds a `conversationui.Model` that performs markdown rendering, caching and selection logic.

```startLine:124:endLine:147:bobatea/pkg/chat/model.go
ret := model{
    conversationManager: manager,
    conversation:        conversationui.NewModel(manager),
    filepicker:          fp,
    style:               conversationui.DefaultStyles(),
    keyMap:              DefaultKeyMap,
    backend:             backend,
    viewport:            viewport.New(0, 0),
    help:                help.New(),
    scrollToBottom:      true,
}
```

- In the embedded `conversationui.Model`, a `glamour.TermRenderer` is constructed and reused; it’s recreated on width change and the cache is invalidated. Rendering is cached per message ID with a `selected` flag.

```startLine:141:endLine:157:bobatea/pkg/chat/conversation/model.go
renderer, err := glamour.NewTermRenderer(
    glamour.WithStandardStyle(m.determinedStyle),
    glamour.WithWordWrap(m.getRendererContentWidth()),
)
...
// Invalidate cache as rendering depends on width
m.cache = make(map[conversation2.NodeID]cacheEntry)
```

- Message rendering uses `glamour` if available, falls back to a word wrapper otherwise.

```startLine:218:endLine:233:bobatea/pkg/chat/conversation/model.go
var v_ string
if m.renderer == nil {
    v_ = wrapWords(v, contentWidth)
} else {
    rendered, err := m.renderer.Render(v + "\n")
    if err != nil {
        v_ = wrapWords(v, contentWidth)
    } else {
        v_ = strings.TrimSpace(rendered)
    }
}
```

- Streaming lifecycle: on `StreamStartMsg`, a new assistant message is appended; duplicates are guarded. On completion/done, the content is updated and cached.

```startLine:327:endLine:406:bobatea/pkg/chat/conversation/model.go
case StreamStartMsg:
    // duplicate detection...
    msg_ := conversation2.NewChatMessage(
        conversation2.RoleAssistant, "",
        conversation2.WithID(msg.ID),
        conversation2.WithMetadata(metadata))
    if err := m.manager.AppendMessages(msg_); err != nil { /* ... */ }
    m.updateCache(msg_)
```

- The top-level `model.Update` forwards stream messages to `conversationui.Model.Update` and updates the viewport content, often forcing a `GotoBottom`.

```startLine:358:endLine:384:bobatea/pkg/chat/model.go
if m.scrollToBottom {
    v, _ := m.conversation.ViewAndSelectedPosition()
    m.viewport.SetContent(v)
    m.viewport.GotoBottom()
}
```

### The Turns model and helpers

The Turns data model defines an engine/middleware-friendly representation of interaction content as ordered Blocks, not just opaque strings. Engines append LLM text and tool-related blocks.

```startLine:16:endLine:37:geppetto/pkg/doc/topics/08-turns.md
## Turns and Blocks in Geppetto
... The Turn data model provides ... ordered Blocks.
### Helpers
- turns.BuildConversationFromTurn
- turns.BlocksFromConversationDelta
```

Block/Conversation mapping helpers exist to translate back and forth:

```startLine:9:endLine:29:geppetto/pkg/turns/conv_conversation.go
func BuildConversationFromTurn(t *Turn) conversation.Conversation {
    for _, b := range t.Blocks {
        switch b.Kind {
        case BlockKindUser:
            text, _ := getString(b.Payload, PayloadKeyText)
            msgs = append(msgs, conversation.NewChatMessage(conversation.RoleUser, text))
        case BlockKindLLMText:
            text, _ := getString(b.Payload, PayloadKeyText)
            msgs = append(msgs, conversation.NewChatMessage(conversation.RoleAssistant, text))
        case BlockKindSystem:
            text, _ := getString(b.Payload, PayloadKeyText)
            if text != "" {
                msgs = append(msgs, conversation.NewChatMessage(conversation.RoleSystem, text))
            }
        case BlockKindToolCall:
            // map to ToolUseContent request from assistant
            /* ... */
        }
    }
}
```

Engines now append `llm_text` and `tool_call` blocks directly to the Turn (streaming path shown here):

```startLine:291:endLine:301:geppetto/pkg/steps/ai/openai/engine_openai.go
if len(message) > 0 {
    turns.AppendBlock(t, turns.NewAssistantTextBlock(message))
}
for _, tc := range mergedToolCalls {
    turns.AppendBlock(t, turns.NewToolCallBlock(tc.ID, tc.Function.Name, args))
}
```

```startLine:206:endLine:217:geppetto/pkg/steps/ai/claude/engine_claude.go
for _, c := range response.Content {
    switch v := c.(type) {
    case api.TextContent:
        if s := v.Text; s != "" {
            turns.AppendBlock(t, turns.NewAssistantTextBlock(s))
        }
    case api.ToolUseContent:
        turns.AppendBlock(t, turns.NewToolCallBlock(v.ID, v.Name, args))
    }
}
```

### Problem analysis

- Conversation-centric rendering hides structure. Tool calls and tool results are just text or opaque markers. We cannot promote them to proper UI widgets without parsing and conventions.
- Rendering and state are tied to message strings; there’s no notion of a “Turn” with blocks that can be rendered by type.
- Width changes invalidate the entire message cache; recomputation churn is high for long conversations.
- The current streaming flow appends/edits a single assistant message. With Turns, multiple blocks may be appended (text + tool_call), and we may want to show a multi-section UI for a single Turn.

### Goals

- Expose a “visual rendering” of a Turn output:
  - Default: LLM text area, markdown-rendered.
  - Rich: panels for tool calls, tool results, citations, metadata, errors.
- Decouple rendering from `conversation.Message` and instead render from `turns.Turn` and `turns.Block`.
- Provide a cache keyed on `(turnID, width, theme, selection state)` to avoid re-render work.
- Fit seamlessly in Bubble Tea: produce string output for viewport composition; allow future per-section widgets.
- Preserve existing streaming UX: scroll-to-bottom, status updates, selection.

### Brainstorming design options

1) New TurnUI model with renderer registry (recommended)
   - Introduce `turnui.Model` that holds a `TurnStore` and a registry of `BlockRenderer`s and/or a `TurnRenderer`.
   - Rendering outputs a list of “segments” per Turn, each a wrapped string (or widget later). Cache at the segment and turn level.
   - Map streaming events to Turn mutations via a small adapter; stop writing placeholder `assistant` messages to the conversation manager for UI purposes.
   - Pros: clean separation, unlocks widgets, aligns with Turns architecture.
   - Cons: higher initial integration effort; requires a TurnStore and event adapter.

2) Adapter inside current `conversationui.Model`
   - Keep current model, but when rendering assistant messages, reconstruct a Turn using `BlocksFromConversationDelta` and pass to a `TurnRenderer` to render richer sections.
   - Pros: minimal disruption; reuses manager and cache.
   - Cons: structural mismatch; hard to show tool-use sequences as separate sections; caching is message-centric, not turn-centric.

3) Hybrid “section renderers” for assistant messages
   - Continue appending assistant messages, but parse `EventMetadata` or embedded markers to extract block-like data and render sections.
   - Pros: drop-in for today’s event stream; good for incremental improvements.
   - Cons: brittle parsing; diverges from Turns; doesn’t leverage engine-produced blocks.

4) Block timeline UI (Turn/Block explorer)
   - Render a timeline grouped by Turns; each turn shows blocks in order with per-kind renderers.
   - Pros: maximal fidelity to Turns; great for debugging tool flows.
   - Cons: larger UI change; may be noisier for pure-chat scenarios.

### Recommended architecture: TurnUI model with renderer registry

Core ideas
- Introduce a store that owns the current Turn (or a small window of recent Turns) and is updated by events or by engine returns.
- Provide a renderer registry by BlockKind (and optionally by a discriminator in payload/name) and a Turn-level renderer for composition.
- Produce a fully wrapped string view per Turn for the Bubble Tea viewport; segment-level caches are keyed by (block identity, width, theme).
- Maintain selection state at the Turn or Block level.

Key interfaces (pseudocode)

```go
// geppetto/pkg/turns based types are inputs
type BlockRenderer interface {
    CanRender(b turns.Block) bool
    Render(b turns.Block, width int, theme Theme) (string, int) // string + height
}

type TurnRenderer interface {
    RenderTurn(t *turns.Turn, width int, theme Theme) (string, TurnPosition)
}

type TurnPosition struct { Offset, Height int }

type RendererRegistry interface {
    Register(br BlockRenderer)
    RenderBlocks(blocks []turns.Block, width int, theme Theme) (string, []int)
}

type Cache interface {
    Get(key string) (string, bool)
    Put(key, value string)
}
```

TurnUI model skeleton (pseudocode)

```go
type TurnUIModel struct {
    store       *TurnStore            // owns current Turn(s)
    registry    RendererRegistry
    cache       Cache                 // width/theme-sensitive
    viewport    viewport.Model
    width, height int
    theme       Theme
    selectedTurn int
}

func (m *TurnUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case StreamStartMsg, StreamCompletionMsg, StreamDoneMsg:
        // translate to Turn mutations instead of appending conversation messages
        m.store.ApplyStreamMsg(v)
        m.refreshViewport()
    case tea.WindowSizeMsg:
        m.width = v.Width; m.height = v.Height
        m.invalidateCacheOnWidthChange()
        m.refreshViewport()
    }
    return m, nil
}

func (m *TurnUIModel) refreshViewport() {
    t := m.store.CurrentTurn()
    view := m.renderTurn(t)
    m.viewport.SetContent(view)
    m.viewport.GotoBottom()
}

func (m *TurnUIModel) renderTurn(t *turns.Turn) string {
    // key by (turnID, width, theme, block-hash)
    if s, ok := m.cache.Get(cacheKey(t, m.width, m.theme)); ok { return s }
    s, _ := m.registry.RenderBlocks(t.Blocks, m.contentWidth(), m.theme)
    m.cache.Put(cacheKey(t, m.width, m.theme), s)
    return s
}
```

Renderers
- LLMTextRenderer: markdown via `glamour`, with the same style selection already in `conversationui.Model`.
- ToolCallRenderer: compact panel showing tool name, arguments (formatted as YAML/JSON), and status while awaiting tool_use.
- ToolUseRenderer: expandable panel or preformatted block for results/errors.
- SystemRenderer, OtherRenderer: minimal wrappers.

Caching
- Width/theme-sensitive cache keys.
- Per-block caching is possible: each block rendering can use `(block.ID or hash, width, theme)`; the turn string is joined from cached block strings.

Streaming integration
- Keep today’s `Stream*Msg` envelope, but update a `TurnStore` instead of appending placeholder `assistant` messages. The backend adapter translates stream deltas and metadata into Turn block mutations.
- For engines that already emit final blocks, we can update the Turn wholesale when the engine returns a Turn.

### Migration path and integration points

Near-term integration
- Add a new `turnui` package next to `conversationui` and a new top-level `chat` model constructor to opt into it (e.g., `InitialTurnModel`).
- In `pinocchio/pkg/cmds/cmd.go`, select the new model behind a feature flag or profile.
- Reuse selection and scroll behaviors from the current `chat` model.

Bridging from today’s flow
- The current UI depends on `conversation.Manager` and calls `ViewAndSelectedPosition()` to set viewport content. The turn UI can implement the same Bubble Tea model surface so the parent model code remains similar.
- Continue using existing `Stream*Msg` types; only the implementation of how the message affects UI state differs (conversation vs turn store).

Longer-term
- Promote blocks and turns as first-class in the UI: add a Turn timeline mode for debugging, grouping blocks per Turn with navigation.
- Offer plug-in renderers registered by BlockKind and by `payload.name` for tools.

### Trade-offs

- Separate model vs adapter-in-place
  - Separate model is cleaner and unlocks widgets; adapter is simpler short-term but fights the abstraction.
- String rendering vs widget composition
  - Starting with string rendering maintains compatibility and simplicity. We can evolve to per-section Bubble Tea components later.
- Cache coherence
  - Turn-based cache aligns with the new data model and reduces whole-conversation invalidations on width change.

### Concrete change points (code references)

- `bobatea/pkg/chat/model.go` forwards stream messages and recomputes the viewport on changes.

```startLine:627:endLine:669:bobatea/pkg/chat/model.go
m.conversation.SetWidth(m.width)
m.viewport.Width = m.width
m.viewport.Height = newHeight
v, _ := m.conversation.ViewAndSelectedPosition()
m.viewport.SetContent(v)
m.viewport.GotoBottom()
```

- `bobatea/pkg/chat/conversation/model.go` is the right reference for width-aware markdown rendering and caching; we will replicate/abstract these in the LLMText block renderer.

```startLine:159:endLine:173:bobatea/pkg/chat/conversation/model.go
func (m Model) getRendererContentWidth() int {
    style := m.style.UnselectedMessage
    w, _ := style.GetFrameSize()
    padding := style.GetHorizontalPadding()
    contentWidth := m.width - w - padding
    return contentWidth
}
```

### Proposed package layout and APIs

- `bobatea/pkg/turnui/`
  - `model.go`: Bubble Tea model with viewport, width handling, selection.
  - `store.go`: `TurnStore` that holds `*turns.Turn` and applies stream/engine updates.
  - `render/`
    - `registry.go`: registry for block renderers.
    - `renderers_llm_text.go`: glamour-based text renderer (adapts from `conversationui`).
    - `renderers_tool_call.go`, `renderers_tool_use.go`, `renderers_system.go`.
    - `cache.go`: width/theme-sensitive cache.

Minimal constructor (pseudocode)

```go
func NewTurnUIModel(store *TurnStore, options ...Option) TurnUIModel
```

### Next steps

- [ ] Implement `bobatea/pkg/turnui` with the skeleton model, store, and registry.
- [ ] Port the markdown/width/caching logic from `conversationui` into `renderers_llm_text.go`.
- [ ] Implement tool call/result renderers with compact panels.
- [ ] Add a feature flag to choose between `conversationui` and `turnui` in `pinocchio/pkg/cmds/cmd.go`.
- [ ] Add tests for width changes and streaming updates to ensure cache behavior and scroll-to-bottom.

### Appendix: Key references

- Streaming message handling in `conversationui`:

```startLine:263:endLine:305:bobatea/pkg/chat/conversation/model.go
case StreamCompletionMsg: /* updates text, metadata, cache */
case StreamDoneMsg:       /* final text, metadata, cache */
```

- Turn helpers and docs:

```startLine:60:endLine:74:geppetto/pkg/turns/conv_conversation.go
func BlocksFromConversationDelta(updated conversation.Conversation, startIdx int) []Block { /* ... */ }
```

```startLine:38:endLine:51:geppetto/pkg/doc/topics/08-turns.md
### Engine mapping and Tool workflow with Turns
1. Engine appends tool_call
2. Middleware executes tools and appends tool_use
3. Engine invoked again with updated Turn
```


