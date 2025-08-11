---
Title: Using Bubble Tea Models Consistently for Timeline Entities and Aligning Chat Navigation
Slug: bubbletea-models-consistent-entities
Short: Migrate all timeline renderers to Bubble Tea models, restore original navigation behavior, and standardize entity-level events (copy, etc.)
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

## 1) Context and goals

- The chat view originally provided robust navigation and copy actions via `conversation.Model`.
- We migrated to a timeline entity system with a mix of stateless renderers and a growing number of interactive entity models.
- To unify behavior and simplify integration with Bubble Tea’s Update/View model, we will migrate all renderers to Bubble Tea `tea.Model`s and standardize event messages to handle common actions (copy text/source, focus/blur, selection changes).

Goals:
- Consistent entity rendering through Bubble Tea models.
- Restore intuitive navigation behavior (Shift+Up/Down selection, PgUp/PgDn/Up/Down viewport scroll) consistent with the original chat.
- Define entity-level events to support actions like copy, selection, focus, and props updates.

## 2) Current architecture overview

- Data flow (runtime):
  - Backend (fake/http/agent) emits UI lifecycle messages: `UIEntityCreated/Updated/Completed/Deleted`.
  - `chat.model.Update` receives these and forwards to `timeline.Controller.On*`.
  - `timeline.Controller` applies changes to an append-only `entityStore` and invalidates the per-entity `renderCache`.
  - `timeline.Controller.View()` renders the timeline by iterating entities in order and delegating to either:
    - Stateless path (legacy): `Renderer.Render(props, width, theme)` with memoization.
    - Interactive path (new): `EntityModel.View()` for Bubble Tea models created by a `Registry` model factory.

- Registry:
  - `Registry.RegisterModelFactory(factory)` registers per-entity factories keyed by `RendererDescriptor.Key` (and by Kind as a fallback).
  - On `OnCreated`, the controller instantiates a model when a matching factory exists and sets its size/theme via hooks.

- Controller responsibilities:
  - Selection: maintains a `selected` index and `selectionVisible` flag; provides `SelectPrev/Next/Last`, `Unselect()`.
  - Entering/focus: toggles an `entering` flag. When true, `HandleMsg(tea.Msg)` routes messages to the selected entity model and calls `Focus()/Blur()` as needed.
  - Events to models: sends `EntitySelectedMsg/EntityUnselectedMsg` on selection changes; `EntityPropsUpdatedMsg` on `UIEntityUpdated` to let models react in `Update`.
  - Scrolling support: `ViewAndSelectedPosition()` returns full view plus `(offset, height)` of the selected entity so the parent can implement scroll-to-selected behavior.

- Chat model composition:
  - Owns `viewport`, `textarea`, `help`, and the `timeline.Controller`.
  - UI states: `user-input`, `moving-around` (selection), `stream-completion`, `saving-to-file`, `error`.
  - Key handling:
    - ESC enters selection (select last, show highlight). Enter toggles entering; ESC exits entering; ESC again returns to input.
    - Shift+Up/Down selects prev/next and calls scroll-to-selected; Up/Down and PgUp/PgDn scroll the viewport.
    - When entering, keys are routed to the selected entity model via `controller.HandleMsg`.

- Rendering pipeline:
  - Parent computes header/help/textarea; timeline view is placed inside the main `viewport`.
  - Renderers/models use `timeline/chatstyle` boxed styles; selected entities render with the pink border; focused entities render with the focus border.

- Logging:
  - Controller logs create/update/complete/delete applications, selection changes, cache hits/misses.
  - Chat logs transitions between input/selection/entering and routing of keys to entity models.

## 2) Current gaps identified

- Arrow/selection behavior diverged from original; Up/Down were mixed into selection keys. We should use explicit selection keys (Shift+Up/Down) and reserve Up/Down for viewport scroll like before.
- Stateless renderers do not receive events and cannot interact like models; hybrid behavior is confusing.
- Copy actions depend on conversation; should be driven by entity models.

## 3) Migration plan (phased)

### Phase A: Interfaces and controller support
1. Finalize `timeline.EntityModel` = `tea.Model` with hooks:
   - `OnProps(map[string]any)`, `OnCompleted(map[string]any)`, `SetSize(w,h)`, `Focus()`, `Blur()`.
2. Controller responsibilities:
   - Instantiate models via `Registry.RegisterModelFactory`.
   - Route selection messages: `EntitySelectedMsg`, `EntityUnselectedMsg`.
   - Route update messages: `EntityPropsUpdatedMsg` upon `UIEntityUpdated`.
   - Provide `HandleMsg(tea.Msg)` to forward keys/mouse to focused model.
   - Expose `ViewAndSelectedPosition()` for scroll-to-selected logic.

### Phase B: Convert all renderers to models
1. LLMTextRenderer → LLMTextModelFactory:
   - Implement `tea.Model` that wraps the old `conversation.renderMessage` logic (boxed style, metadata line, width-aware wrapping). (see pkg/chat/conversation/model.go)
   - Props: `{ role, text, metadata }`. Selection/focus toggles selected/focused styles.
2. ToolCallsPanelRenderer → ToolCallsPanelModelFactory:
   - Model with state `{ calls, summary }`, updates via `EntityPropsUpdatedMsg`.
3. Plain/fallback renderer → PlainModelFactory:
   - Simple model for debugging that prints key/value props.

### Phase C: Chat integration and navigation
1. Keep navigation keys:
   - Selection: Shift+Up/Down (mapped to SelectPrev/SelectNext).
   - Viewport scroll: Up/Down line, PgUp/PgDn page.
   - ESC enters selection mode from text; Enter toggles entering to route keys to model; ESC exits entering; ESC again returns to text.
2. Replace remaining `conversation` clipboard helpers with entity-driven actions:
   - When moving-around, Copy should invoke model-level Copy events (see events in Section 4).
3. Reuse `conversation` styles where appropriate by migrating styling to `timeline/chatstyle` and importing into entity models.

### Phase D: Remove stateless path
1. Remove `Renderer`-only registration once all built-ins are models.
2. Keep a debug adapter that can wrap a stateless function into a temporary model for experiments.

## 4) Entity-level event set (to be sent via controller → model Update)

- Selection and focus:
  - `EntitySelectedMsg{ID}`
  - `EntityUnselectedMsg{ID}`
  - `EntityFocusedMsg{ID}` (optional if we don’t rely solely on `Focus()` hook)
  - `EntityBlurredMsg{ID}` (optional counterpart)

- Props/data updates:
  - `EntityPropsUpdatedMsg{ID, Patch}`
  - `EntityCompletedMsg{ID, Result}` (optional if `OnCompleted` covers it)

- Clipboard and actions:
  - `EntityCopyTextMsg{ID}`: ask model to expose canonical text; model returns via a callback or emits `tea.Cmd` to parent.
  - `EntityCopyCodeMsg{ID}`: ask model to extract code blocks.
  - `EntityCopyMetadataMsg{ID}`: optional.
  - `EntityToggleExpandMsg{ID}`: expand/collapse if supported.
  - `EntityOpenLinkMsg{ID, URL}`: request to open a link (parent handles OS integration).

- Navigation within entity (if needed):
  - `EntityScrollLineUp/Down`, `EntityScrollPageUp/Down` for entities with internal scrollable content.

Notes:
- Models respond to these via `Update(msg tea.Msg) (tea.Model, tea.Cmd)`.
- Parent Chat model listens for returned `Cmd`s and executes side-effects (copy to clipboard, open URL, HTTP calls).

## 5) Detailed steps to convert LLM text rendering

1. Create `LLMTextModel` with fields `{ role, text, metadata, width, selected, focused }`.
2. Implement `View()` by porting the code from `conversation/model.go::renderMessage` (boxed style, metadata line, width-aware wrap).
3. Handle `EntityPropsUpdatedMsg` to update text/metadata; handle `EntitySelectedMsg`/`EntityUnselectedMsg` to toggle styles; `SetSize()` adjusts wrapping width.
4. Register with `Registry.RegisterModelFactory(LLMTextFactory{})` and update chat initialization to use this instead of `LLMTextRenderer`.
5. Remove `timeline/renderers.go::LLMTextRenderer` when the model is fully wired.

## 6) Restoring original navigation feel

Implement keys as:
- Shift+Up: SelectPrev; Shift+Down: SelectNext; call `scrollToSelected()`.
- Up/Down: `viewport.LineUp/LineDown(1)`.
- PgUp/PgDn: `viewport.PageUp/PageDown(1)`.
- ESC: text → selection (select last, show highlight); selection → entering (Enter); entering → selection (ESC); selection → text (ESC).

## 7) Testing and validation

- Unit tests for controller selection and `ViewAndSelectedPosition()`.
- Integration tests: simulate sequences of entity events and navigation keystrokes; assert viewport YOffset behaviors and selected styling.
- Visual validation for LLM text model parity with original `conversation.Model` output.

## 8) Rollout and cleanup

- Convert built-in renderers (LLMText, ToolCallsPanel, Plain) to models.
- Update examples and docs.
- Remove stateless renderer path and registry methods once no longer used.


