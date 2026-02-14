---
Title: Reusable Widget Extraction Design for pkg/repl
Ticket: BOBA-010-REPL-WIDGET-EXTRACTION-CLEANUP
Status: active
Topics:
    - analysis
    - cleanup
    - repl
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: bobatea/pkg/repl/command_palette_model.go
      Note: Palette host adapter logic currently tied to REPL
    - Path: bobatea/pkg/repl/completion_model.go
      Note: Autocomplete state machine extraction candidate
    - Path: bobatea/pkg/repl/helpbar_model.go
      Note: Contextual help bar widget-grade logic
    - Path: bobatea/pkg/repl/helpdrawer_model.go
      Note: Help drawer interactions and debounce/pin behavior
    - Path: bobatea/pkg/repl/model.go
      Note: Current composition root that mixes app orchestration and widget internals
    - Path: bobatea/pkg/repl/model_async_provider.go
      Note: Generic provider timeout and panic recovery utility
    - Path: bobatea/pkg/repl/model_timeline_bus.go
      Note: REPL-specific timeline/eventbus glue that should remain app-level
ExternalSources: []
Summary: Detailed extraction plan for turning pkg/repl feature blocks into reusable Bubble Tea widgets without carrying REPL/timeline coupling.
LastUpdated: 2026-02-14T16:41:45.90419568-05:00
WhatFor: Guide pre-PR cleanup so we can split generic widgets from REPL-specific orchestration and reduce model coupling.
WhenToUse: Use when implementing or reviewing refactors that move autocomplete/help UI behavior out of pkg/repl into reusable TUI widget packages.
---


# Reusable Widget Extraction Design for `pkg/repl`

> [!IMPORTANT]
> **Goal:** identify which parts of `pkg/repl` are genuinely reusable widgets, which parts must remain REPL/timeline glue, and how to split them without reintroducing the regressions fixed in BOBA-002 through BOBA-009.

> [!TIP]
> **Context for new contributors:** `pkg/repl` currently behaves as both a shell application and a widget toolkit. This document separates those concerns so future features can be built once and reused in non-REPL apps.

## 1. Executive Summary

`pkg/repl` now contains several mature UI feature slices: autocomplete popup, help bar, help drawer, command palette host behavior, key help rendering, and async provider orchestration. Most of these are widget-grade and can be extracted. The core REPL model should remain as an integration host for evaluator execution, timeline/eventbus projection, and application-level focus routing.

**Extract now (high confidence):**
- `completion_model.go` + `completion_overlay.go` + `autocomplete_types.go` as a generic suggestion widget.
- `helpbar_model.go` + `help_bar_types.go` as a generic contextual status bar widget.
- `helpdrawer_model.go` + `helpdrawer_overlay.go` + `help_drawer_types.go` as a generic contextual panel widget.
- `model_async_provider.go` (`runProvider`) as shared async/debounce/provider runtime utilities.
- `history.go` as generic input history navigation state.

**Partially extract / keep thin REPL adapter:**
- `command_palette_model.go` + `command_palette_overlay.go` should become a reusable host adapter over `pkg/commandpalette`, but REPL-specific commands must remain in REPL.
- `model_help.go` should become reusable key-help layout logic; binding sets remain app-specific.

**Keep in REPL (application glue):**
- `model_timeline_bus.go`, `wm_transformer.go`, `bridge.go` (timeline/eventbus projection).
- `model_input.go` and key routing semantics that coordinate timeline focus vs input focus.
- evaluator contracts in `evaluator.go` and REPL event kinds.

## 2. Current Architecture Snapshot

### 2.1 Top-level model responsibilities today

The `Model` in `pkg/repl/model.go` currently owns all of these at once:
- app lifecycle/context cancellation.
- timeline shell orchestration.
- text input and history.
- feature widget state (`completion`, `helpBar`, `helpDrawer`, `palette`).
- keyboard dispatch + mode switching.
- lipgloss v2 layer composition.

This means one model is both:
- an application composition root, and
- a feature/widget runtime.

That dual role is what makes the package feel “clunky”: every feature improvement touches central orchestration code.

### 2.2 Feature coupling map

```text
pkg/repl/model.go
  ├─ updateInput() / updateTimeline() (routing)
  ├─ completion_model.go + completion_overlay.go
  ├─ helpbar_model.go
  ├─ helpdrawer_model.go + helpdrawer_overlay.go
  ├─ command_palette_model.go + command_palette_overlay.go
  ├─ model_help.go (short/full key help rendering)
  ├─ model_async_provider.go (timeouts, panic recovery)
  └─ model_timeline_bus.go + wm_transformer.go (REPL event -> timeline)
```

### 2.3 Existing reusable extraction precedent

`pkg/commandpalette` is already a separate package and works. That is the strongest evidence that feature packages can be extracted safely if:
- they are given clean app-agnostic contracts,
- REPL-specific command registration stays outside the package.

## 3. Candidate-by-Candidate Extraction Analysis

## 3.1 Autocomplete popup widget

**Current files/symbols:**
- `pkg/repl/completion_model.go`
- `pkg/repl/completion_overlay.go`
- `pkg/repl/autocomplete_types.go`
- Key bindings consumed via `pkg/repl/keymap.go`

**Why reusable:**
- generic request shape (`input`, `cursor`, reason, shortcut).
- generic suggestion list with replacement span and selection.
- generic overlay placement logic (above/below/bottom, left/right growth, paging).

**Current REPL coupling to remove:**
- direct access to `m.textInput`.
- direct access to `m.keyMap` fields.
- direct use of `m.width`, `m.height`, and timeline-derived anchor Y assumptions.

**Proposed reusable contract:**

```go
// pkg/tui/widgets/suggest

type BufferSnapshot struct {
    Value      string
    CursorByte int
}

type BufferMutator interface {
    SetValue(v string)
    SetCursorByte(pos int)
}

type Provider interface {
    CompleteInput(ctx context.Context, req Request) (Result, error)
}

type Widget struct { ... }

func (w *Widget) OnBufferChanged(prev, cur BufferSnapshot) tea.Cmd
func (w *Widget) HandleKey(k tea.KeyMsg, km KeyMap, mut BufferMutator) (handled bool, cmd tea.Cmd)
func (w *Widget) OverlayLayout(anchor Anchor, viewport Size) (Layout, bool)
func (w *Widget) OverlayView(layout Layout, styles Styles) string
```

> [!NOTE]
> Anchor computation should be host-provided. REPL can anchor near input row; another app can anchor at cursor row in an editor panel.

## 3.2 Help bar widget

**Current files/symbols:**
- `pkg/repl/helpbar_model.go`
- `pkg/repl/help_bar_types.go`

**Why reusable:**
- It is already generic: debounce + provider call + payload visibility.
- no timeline dependency.

**Current REPL coupling to remove:**
- calls to `m.applyLayoutAndRefresh()`.
- style mapping bound to `m.styles`.

**Proposed reusable contract:**
- widget returns state-change flags (`visibilityChanged`) to host.
- host decides relayout strategy.
- widget exposes pure render using injected style struct.

## 3.3 Help drawer widget

**Current files/symbols:**
- `pkg/repl/helpdrawer_model.go`
- `pkg/repl/helpdrawer_overlay.go`
- `pkg/repl/help_drawer_types.go`

**Why reusable:**
- request/response types are content-agnostic.
- behavior (toggle, refresh, pin, debounce typing updates) is universal.
- docking/size model already parameterized.

**Current REPL coupling to remove:**
- references to completion key conflict behavior.
- use of REPL key map field names.
- panel style and footer binding names tied to REPL naming.

**Proposed reusable contract:**

```go
// pkg/tui/widgets/contextpanel

type Widget struct { ... }

type KeyMap struct {
    Toggle key.Binding
    Close  key.Binding
    Refresh key.Binding
    Pin key.Binding
}

func (w *Widget) HandleKey(k tea.KeyMsg, km KeyMap) (handled bool, cmd tea.Cmd)
func (w *Widget) OnBufferChanged(prev, cur BufferSnapshot) tea.Cmd
func (w *Widget) RequestNow(trigger Trigger) tea.Cmd
func (w *Widget) OverlayLayout(anchor Anchor, viewport Size) (Layout, bool)
func (w *Widget) OverlayView(layout Layout, styles Styles, footer FooterBindings) string
```

> [!IMPORTANT]
> Keep `doc.Show == false` semantics in the widget package. That behavior is subtle and currently used to distinguish “no relevant content” from provider errors.

## 3.4 Command palette host adapter

**Current files/symbols:**
- `pkg/repl/command_palette_model.go`
- `pkg/repl/command_palette_overlay.go`
- backing engine `pkg/commandpalette/model.go`

**What is already reusable:**
- `pkg/commandpalette` engine.

**What is still REPL-specific:**
- built-in command list manipulating REPL state (focus, quit, history, help drawer).
- slash-open policy provider typed to REPL context.

**Extraction plan:**
- Create `pkg/tui/widgets/palettehost` to manage open/close keys, slash policy, overlay placement, and command merge logic.
- Keep command definitions themselves in REPL.

## 3.5 Key-help adaptive full mode logic

**Current files/symbols:**
- `pkg/repl/model_help.go`
- `pkg/repl/keymap.go`

**Reusable part:**
- `computeHelpColumns`, `splitHelpColumns`, `fullHelpGroupsFit`.

**Keep app-specific:**
- the actual binding set and ordering (`ShortHelp`, `FullHelp`).

**Package target:** `pkg/tui/keyhelp` with pure helpers:
- `ComputeColumns(bindings []key.Binding, model help.Model) [][]key.Binding`
- `RenderAdaptiveFullHelp(...) string`

This avoids duplicating the same tricky width-fit behavior in other tools.

## 3.6 Async provider execution utility

**Current files/symbols:**
- `pkg/repl/model_async_provider.go` (`runProvider` + 3 wrappers)

**Why reusable:**
- generic timeout + panic recovery + typed return.
- no REPL semantics except naming strings.

**Package target:** `pkg/tui/asyncprovider` or `pkg/asyncutil`.

Potential API:

```go
func Run[T any](
    baseCtx context.Context,
    requestID uint64,
    timeout time.Duration,
    providerName string,
    panicPrefix string,
    fn func(context.Context) (T, error),
) (T, error)
```

## 3.7 History state

**Current file/symbols:**
- `pkg/repl/history.go`

**Why reusable:**
- no REPL dependencies.
- generic input history navigation.

**Package target:** `pkg/tui/inputhistory`.

## 3.8 Timeline/eventbus glue (non-extract)

**Keep in REPL package:**
- `pkg/repl/model_timeline_bus.go`
- `pkg/repl/wm_transformer.go`
- `pkg/repl/bridge.go`

These are not widget logic; they are app transport/persistence semantics for REPL event projection.

## 4. Proposed Target Package Layout

```text
pkg/
  repl/
    model.go                  # composition root
    model_input.go            # host routing
    model_timeline_bus.go     # app glue (stay)
    wm_transformer.go         # app glue (stay)
    keymap.go                 # app key bindings (stay)

  tui/
    widgets/
      suggest/                # from completion_* + autocomplete_types
      contextbar/             # from helpbar_* + help_bar_types
      contextpanel/           # from helpdrawer_* + help_drawer_types
      palettehost/            # from command_palette_model/overlay glue
    keyhelp/
      renderer.go             # from model_help helpers
    inputhistory/
      history.go              # from repl/history.go
    asyncprovider/
      run.go                  # from runProvider
```

> [!TIP]
> Keep extraction under `pkg/tui/...` (not `pkg/repl/widgets/...`) to make independence obvious and prevent accidental import cycles back into REPL.

## 5. Dependency and Ownership Rules

To keep these packages reusable, define strict ownership:

- Widgets may depend on:
  - Bubble Tea/Bubbles/Lipgloss types.
  - standard library.
- Widgets may not depend on:
  - `pkg/repl` model.
  - `pkg/timeline`.
  - `pkg/eventbus`.

- REPL host owns:
  - evaluator adaptation.
  - timeline event mapping.
  - app-level focus mode semantics.
  - feature wiring and config default values.

## 6. Integration Design (Host + Widgets)

### 6.1 Host responsibilities after extraction

`pkg/repl.Model` becomes a coordinator that:
- snapshots buffer/focus/viewport state.
- forwards key/input events to widgets in precedence order.
- asks widgets for overlay layers.
- composes base timeline + widget overlays in lipgloss v2 canvas.

### 6.2 Suggested input routing order

Keep this order because it reflects current UX correctness:

1. palette (if open).
2. help drawer shortcuts.
3. completion navigation/accept.
4. completion trigger shortcut.
5. host-level submit/history/focus.
6. text input update.
7. post-update debounce scheduling for widgets.

## 7. Phased Refactor Plan (Pre-PR Cleanup)

## Phase A: Extract low-risk pure utilities
- move `history.go` -> `pkg/tui/inputhistory`.
- move `runProvider` -> `pkg/tui/asyncprovider`.
- move key-help column-fit helpers -> `pkg/tui/keyhelp`.

**Why first:** no visible behavior change, fast confidence.

## Phase B: Extract contextbar + contextpanel widgets
- move state machines and types.
- keep REPL adapters small (`toWidgetRequest`, `fromWidgetState`).
- preserve all tests from `help_bar_model_test.go`, `help_drawer_model_test.go` with package rename/adaptation.

## Phase C: Extract suggest widget
- migrate completion state machine + overlay layout.
- generalize anchor provider and buffer mutator.
- preserve no-flicker/selection/paging behavior.

## Phase D: Extract palette host adapter
- keep `pkg/commandpalette` as engine.
- move only reusable host mechanics (open/close/slash/layout).
- keep REPL built-ins in `pkg/repl`.

## Phase E: Final REPL slim pass
- `model.go` becomes orchestration-only.
- remove feature internals from `pkg/repl`.
- enforce import boundaries with lints/checklist.

## 8. Acceptance Criteria Before PR

A cleanup PR should not merge until these hold:

- behavior parity for:
  - completion trigger/accept/cancel/page + overlay placement options.
  - help bar debounce and visibility transitions.
  - help drawer dock/pin/manual refresh/typing refresh behavior.
  - command palette keyboard semantics.
  - full help adaptive columns and short help default behavior.

- no new coupling:
  - `pkg/tui/widgets/*` must not import `pkg/repl` or `pkg/timeline`.

- test coverage:
  - existing `pkg/repl/*_test.go` scenarios preserved or ported.
  - new widget package tests added for pure behavior.

- manual checks:
  - `examples/repl/autocomplete-generic`
  - `examples/js-repl`

## 9. Detailed Cut List (File-Level)

**Move candidates:**
- `pkg/repl/completion_model.go`
- `pkg/repl/completion_overlay.go`
- `pkg/repl/autocomplete_types.go`
- `pkg/repl/helpbar_model.go`
- `pkg/repl/help_bar_types.go`
- `pkg/repl/helpdrawer_model.go`
- `pkg/repl/helpdrawer_overlay.go`
- `pkg/repl/help_drawer_types.go`
- `pkg/repl/model_async_provider.go` (utility part)
- `pkg/repl/history.go`
- `pkg/repl/model_help.go` (helper part)

**Stay in REPL:**
- `pkg/repl/model.go`
- `pkg/repl/model_input.go`
- `pkg/repl/model_layout.go`
- `pkg/repl/model_timeline_bus.go`
- `pkg/repl/wm_transformer.go`
- `pkg/repl/bridge.go`
- `pkg/repl/evaluator.go`
- `pkg/repl/keymap.go` (binding policy)
- `pkg/repl/config.go` (host defaults; widget config structs can mirror/alias)

## 10. Risk Register

- **Risk:** keybinding precedence regressions.
  - **Mitigation:** keep routing order fixed and covered with tests.

- **Risk:** overlay position drift across widgets.
  - **Mitigation:** shared overlay geometry helpers in one package; golden tests for placements.

- **Risk:** hidden coupling via direct model field access.
  - **Mitigation:** explicit host interfaces (`BufferSnapshot`, `Viewport`, `Anchor`).

- **Risk:** context cancellation regressions.
  - **Mitigation:** keep `appContext()` ownership in host and pass it into widget async calls.

## 11. Pseudocode: Slim REPL Host After Extraction

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case tea.WindowSizeMsg:
        m.viewport = Size{v.Width, v.Height}
        return m, m.relayout()

    case tea.KeyMsg:
        if handled, cmd := m.palette.HandleKey(v, m.hostCtx()); handled { return m, cmd }
        if handled, cmd := m.drawer.HandleKey(v); handled { return m, cmd }
        if handled, cmd := m.suggest.HandleKey(v, m.bufferMutator()); handled { return m, cmd }
        return m.handleHostKeysOrInput(v)

    case suggest.ResultMsg:
        return m, m.suggest.ApplyResult(v)
    case contextbar.ResultMsg:
        return m, m.bar.ApplyResult(v)
    case contextpanel.ResultMsg:
        return m, m.drawer.ApplyResult(v)
    }
    return m, nil
}
```

```go
func (m *Model) View() string {
    base := m.renderTimelineAndInput()
    layers := []Layer{ Layer(base, 0,0,0) }

    if l, ok := m.drawer.Layer(m.anchor(), m.viewport); ok { layers = append(layers, l) }
    if l, ok := m.suggest.Layer(m.anchor(), m.viewport); ok { layers = append(layers, l) }
    if l, ok := m.palette.Layer(m.viewport); ok { layers = append(layers, l) }

    return Compose(layers, m.viewport)
}
```

## 12. Research Notes (What informed this design)

Reviewed in detail:
- `pkg/repl/model.go` for orchestration boundaries.
- `pkg/repl/completion_model.go` and `pkg/repl/completion_overlay.go` for widget-grade state and overlay logic.
- `pkg/repl/helpbar_model.go`, `pkg/repl/helpdrawer_model.go`, and `pkg/repl/helpdrawer_overlay.go` for context widget behavior.
- `pkg/repl/model_async_provider.go` for reusable timeout/panic handling.
- `pkg/repl/model_help.go` for key-help fitting behavior.
- `pkg/repl/command_palette_model.go` + `pkg/commandpalette/model.go` for existing package boundary precedent.
- `pkg/repl/model_timeline_bus.go`, `pkg/repl/wm_transformer.go`, `pkg/repl/bridge.go` to identify non-widget app glue.

## 13. Recommendation

Proceed with extraction in Phases A-E above, starting with pure utility moves, then context widgets, then completion, then palette host.

This gives us a cleaner pre-PR story:
- REPL becomes an app-level orchestrator.
- shared widgets become reusable for other Bubble Tea apps.
- regression risk stays manageable because behavior is preserved behind adapters.
