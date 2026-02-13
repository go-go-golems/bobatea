---
Title: REPL Model Decomposition Analysis and Split Plan
Ticket: BOBA-008-CLEANUP-REPL
Status: active
Topics:
    - repl
    - cleanup
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/repl/config.go
      Note: |-
        Config/default definitions that should own normalization helpers currently in model.go
        Config/default ownership used to justify moving normalization out of model.go
    - Path: pkg/repl/help_bar_types.go
      Note: |-
        Help bar provider contracts consumed from model.go
        Help bar provider contract considered in submodel extraction
    - Path: pkg/repl/help_drawer_types.go
      Note: |-
        Help drawer provider contracts consumed from model.go
        Help drawer provider contract considered in submodel extraction
    - Path: pkg/repl/keymap.go
      Note: |-
        Key binding surface used by input, completion, help bar, and help drawer slices
        Key binding surface impacted by input/help feature split
    - Path: pkg/repl/model.go
      Note: |-
        Current monolithic model that mixes orchestration, feature logic, overlays, async commands, and config normalization
        Monolithic REPL model analyzed for decomposition boundaries and extraction phases
ExternalSources: []
Summary: Big-bang rewrite plan for replacing pkg/repl/model.go with smaller feature models and a clean cutover.
LastUpdated: 2026-02-13T18:28:00-05:00
WhatFor: Plan a big-bang decomposition and cutover of the REPL model into smaller files and submodels.
WhenToUse: When implementing BOBA-008 refactors or reviewing REPL architecture boundaries.
---


# REPL Model Decomposition Analysis and Split Plan

## Executive Summary

`pkg/repl/model.go` has become a multi-domain file (1603 LOC, ~45 functions) that currently owns: root Bubble Tea update routing, timeline event forwarding, input/history behavior, autocomplete state machine, help bar state machine, help drawer state machine, overlay layout/rendering, async provider execution, and config normalization.

The best path is now a **single big-bang rewrite and cutover**:

1. Build the target architecture (`completion`, `helpbar`, `helpdrawer` submodels + thin root orchestrator) in one refactor branch.
2. Replace current `model.go` internals in one merge, updating tests and examples in the same change.
3. Remove transitional compatibility concerns and optimize for clean ownership boundaries.

Given the updated constraint, this is simpler than staged extraction and avoids carrying temporary architecture debt between phases.

> [!IMPORTANT]
> Backward compatibility is not a goal for BOBA-008. Prefer a clean cutover architecture and update call sites/examples in the same rewrite.

## Problem Statement

### 1) Domain mixing in one file

`pkg/repl/model.go` mixes unrelated concerns in one unit:

- Lifecycle/routing: `NewModel`, `Init`, `Update`, `updateInput`, `updateTimeline`, `View`
- Event bus publishing: `publishReplEvent`, `publishUIEntityCreated`
- Three independent async feature pipelines:
  - completion (`scheduleDebouncedCompletionIfNeeded`, `completionCmd`, `handleCompletionResult`)
  - help bar (`scheduleDebouncedHelpBarIfNeeded`, `helpBarCmd`, `handleHelpBarResult`)
  - help drawer (`scheduleDebouncedHelpDrawerIfNeeded`, `helpDrawerCmd`, `handleHelpDrawerResult`)
- Overlay layout/rendering for completion and drawer
- Config normalization (`normalizeHelpBarConfig`, `normalizeHelpDrawerConfig`, `normalizeAutocompleteConfig`)

This increases cognitive load and makes local changes harder to reason about.

### 2) Repeated async command boilerplate

`completionCmd`, `helpBarCmd`, and `helpDrawerCmd` each reimplement the same pattern:

- `context.WithTimeout`
- panic recovery + stack capture
- log + typed result message

Same shape appears three times with slightly different types.

### 3) Provider calls are not tied to app shutdown

Provider command contexts are currently derived from `context.Background()` inside command paths, so cancellation only happens on timeout. Exiting the UI should also cancel these calls so in-flight provider work can stop immediately.

### 4) Hidden coupling via shared mutable fields

`Model` currently owns large feature-specific field groups (`completion*`, `helpBar*`, `helpDrawer*`). `updateInput` triggers all three pipelines together. This creates implicit coupling and makes feature-local invariants less visible.

### 5) Rendering and state logic interleaving

`View()` currently performs:

- base section rendering
- completion overlay layout + render + selection-side effects (`ensureCompletionSelectionVisible`)
- help drawer overlay layout + render
- lipgloss v2 compositor setup

This combines “what to show” and “where to show it” across multiple features in one function.

## Current Architecture Map

### Surface size and hotspots

- File size: `pkg/repl/model.go` = 1603 LOC
- `*Model` methods: ~45
- Long functions:
  - `NewModel` (~108 lines)
  - `computeCompletionOverlayLayout` (~105 lines)
  - `renderHelpDrawerPanel` (~78 lines)
  - `computeHelpDrawerOverlayLayout` (~53 lines)

### Control-flow sketch (today)

```text
Update(msg)
  -> root switch
     -> input key path (updateInput)
        -> drawer shortcuts
        -> completion nav
        -> completion shortcut trigger
        -> history / submit / textinput update
        -> schedule debounce for completion + helpbar + helpdrawer
     -> timeline key path (updateTimeline)
     -> timeline UI entity events
     -> debounce msgs (3)
     -> result msgs (3)

View()
  -> header + timeline + input + helpbar + key help
  -> completion overlay layout/render
  -> drawer overlay layout/render
  -> lipgloss v2 layer compose
```

### Feature ownership clusters in `model.go`

- Root orchestration: `pkg/repl/model.go:114`, `pkg/repl/model.go:222`, `pkg/repl/model.go:228`, `pkg/repl/model.go:409`
- Input/timeline routing: `pkg/repl/model.go:310`, `pkg/repl/model.go:378`
- Bus publish path: `pkg/repl/model.go:489`, `pkg/repl/model.go:504`, `pkg/repl/model.go:514`
- Completion pipeline: `pkg/repl/model.go:595`, `pkg/repl/model.go:646`, `pkg/repl/model.go:727`, `pkg/repl/model.go:856`, `pkg/repl/model.go:985`, `pkg/repl/model.go:1058`, `pkg/repl/model.go:1294`
- Help bar pipeline: `pkg/repl/model.go:610`, `pkg/repl/model.go:665`, `pkg/repl/model.go:770`, `pkg/repl/model.go:878`, `pkg/repl/model.go:911`
- Help drawer pipeline: `pkg/repl/model.go:625`, `pkg/repl/model.go:684`, `pkg/repl/model.go:813`, `pkg/repl/model.go:898`, `pkg/repl/model.go:929`, `pkg/repl/model.go:1163`, `pkg/repl/model.go:1216`
- Normalization helpers: `pkg/repl/model.go:1417` onward

## Proposed Solution

### Design goal

Split REPL internals into **small, feature-oriented files and submodels** while keeping `repl.Model` as the public root object.

### Target internal layout

```text
pkg/repl/
  model.go                         # root type + NewModel + Init + Update + View (thin)
  model_messages.go                # internal tea message structs
  model_input.go                   # updateInput + updateTimeline + focus switching
  model_timeline_bus.go            # submit + publishReplEvent + publishUIEntityCreated + refresh scheduler

  completion_model.go              # completion state + handlers + nav + apply/hide
  completion_overlay.go            # completion layout + popup rendering

  helpbar_model.go                 # help bar state + scheduling + command + result handling + render

  helpdrawer_model.go              # drawer state + shortcuts + toggle/close/pin + request/result
  helpdrawer_overlay.go            # drawer layout + panel rendering

  model_async_provider.go          # shared panic/timeout wrapper for provider calls
  config_normalize.go              # normalize* helpers moved out of model.go
  model_layout_primitives.go       # clamp/binding helpers that are UI-independent
```

### Submodel decomposition (target in big-bang cutover)

Implement real internal submodels directly as part of the rewrite.

```go
// unexported, owned by root Model
type completionModel struct {
    provider InputCompleter
    state    completionState
    cfg      completionConfig
}

type helpBarModel struct {
    provider HelpBarProvider
    state    helpBarState
    cfg      helpBarConfig
}

type helpDrawerModel struct {
    provider HelpDrawerProvider
    state    helpDrawerState
    cfg      helpDrawerConfig
}
```

Root `Model` remains orchestrator:

```go
type Model struct {
    // stable public shell identity
    evaluator Evaluator
    config    Config
    styles    Styles

    // shared runtime context
    textInput textinput.Model
    focus     string
    keyMap    KeyMap
    sh        *timeline.Shell

    // feature engines
    completion completionModel
    helpBar    helpBarModel
    helpDrawer helpDrawerModel
}
```

### Shared input context (explicit, not implicit)

Introduce a small immutable context passed into feature schedulers:

```go
type InputSnapshot struct {
    Value      string
    CursorByte int
    Focus      string
    Width      int
    Height     int
}
```

This removes ad-hoc reads from root mutable fields and clarifies what each feature depends on.

## Design Decisions

### Decision 1: Keep one exported root model

**Decision:** Keep `repl.Model` as the public Bubble Tea model. Do not expose feature submodels publicly.

**Why:** Preserves compatibility across examples and call sites while allowing aggressive internal cleanup.

### Decision 2: Big-bang rewrite with direct cutover

**Decision:** Skip staged intermediate states. Rewrite to target architecture in one integrated refactor and switch over immediately.

**Why:** The team is explicitly fine with big-bang; this reduces total churn, avoids interim glue code, and gets to the clean architecture faster.

### Decision 3: Deduplicate provider command pattern

**Decision:** Introduce a shared async helper in `model_async_provider.go`.

Pseudocode:

```go
func runProvider[T any](baseCtx context.Context, timeout time.Duration, reqID uint64, label string, f func(ctx context.Context) (T, error)) (out T, err error) {
    defer recoverToError(&err, label, reqID)
    ctx, cancel := context.WithTimeout(baseCtx, timeout)
    defer cancel()
    return f(ctx)
}
```

**Why:** Reduces boilerplate, keeps panic/timeout semantics consistent across features, and allows app-level cancellation to interrupt provider calls.

### Decision 4: Tie provider execution to app context lifecycle

**Decision:** Add a model-owned app context (`appCtx`) and cancel function; derive provider call contexts from `appCtx` (plus timeout), and cancel on UI shutdown.

**Why:** In-flight provider requests should terminate when the UI exits, not continue until timeout.

### Decision 5: Move normalization out of `model.go`

**Decision:** Move `normalize*` functions to `config_normalize.go`.

**Why:** They are config concerns, not runtime update/view logic.

### Decision 6: Keep overlay composition centralized, but not feature logic

**Decision:** Keep final lipgloss v2 layer composition in root `View`, but move per-feature layout/render into feature files/submodels.

**Why:** One place for layer z-order policy, multiple places for feature-specific geometry/content.

## Alternatives Considered

### A) Mechanical file split only (same `*Model`, no submodels)

- Pros: lower immediate risk
- Cons: retains coupling and requires follow-on rewrites anyway

**Verdict:** rejected for BOBA-008 since we are intentionally choosing a full cutover.

### B) Full rewrite into independent internal feature models now

- Pros: maximal separation, single migration window, no temporary architecture debt
- Cons: higher one-time integration risk and larger review payload

**Verdict:** chosen approach for BOBA-008.

### C) Create a new package `pkg/repl/internal/model` immediately

- Pros: stronger compile-time boundaries
- Cons: larger scope; can obscure core rewrite signal with package churn

**Verdict:** optional follow-up after big-bang cutover if still needed.

## Implementation Plan

### Big-Bang Rewrite Plan

1. Introduce new internal architecture files in one branch:
   - `model_messages.go`
   - `model_input.go`
   - `model_timeline_bus.go`
   - `completion_model.go`
   - `completion_overlay.go`
   - `helpbar_model.go`
   - `helpdrawer_model.go`
   - `helpdrawer_overlay.go`
   - `config_normalize.go`
   - `model_async_provider.go`
2. Define and wire internal feature submodels (`completionModel`, `helpBarModel`, `helpDrawerModel`) and move feature fields out of root `Model`.
3. Refactor root `Model` to orchestration only (`Update` dispatch, `View` layer composition, shared input/timeline context).
4. Replace duplicated provider call wrappers with one generic panic/timeout helper.
5. Add app-context plumbing so provider commands use `context.WithTimeout(appCtx, ...)` and cancel on quit.
6. Update tests and examples in the same changeset for clean cutover.
7. Remove dead transitional code and ensure no dual-path logic remains.

Validation gates:

- `go test ./pkg/repl/... -count=1`
- `golangci-lint run -v --max-same-issues=100 ./pkg/repl/...`
- run at least one REPL example (`go run ./examples/repl/autocomplete-generic` and `go run ./examples/js-repl`) for smoke verification

## Migration Risk and Mitigations

### Risk: Behavior drift in debounced request sequencing (single large diff)

- Mitigation: preserve current request-id invariants; add regression tests for stale-drop across all three feature channels.

### Risk: Overlay placement regressions

- Mitigation: snapshot-style tests for completion/drawer layout edge cases (small terminal, offsets, margins, dock variants).

### Risk: Key binding regressions during cutover

- Mitigation: keep key handling order exactly as today in `updateInput` and assert with focused key-path tests.

### Risk: Context-cancellation regressions in providers

- Mitigation: add tests that cancel app context and assert provider commands return `context.Canceled` quickly (without waiting full timeout).

## Concrete Big-Bang Scope

The BOBA-008 merge should include the full internal model rewrite in one integrated changeset:

1. Root model reduced to orchestration.
2. Feature logic moved to dedicated files and internal submodels.
3. Shared async provider execution helper in place.
4. Provider contexts chained from app context (no raw `context.Background()` in provider command paths).
5. Config normalization removed from `model.go`.
6. Tests/examples updated to the new behavior shape in the same PR.

## Open Questions

1. Should multiline/external-editor flow be included in BOBA-008 if currently dormant in `Model` fields, or handled in a follow-up?
2. Do we want a generalized “typing-context feature bus” now, or keep direct calls from `updateInput` for simplicity?
3. Should overlay z-order policy remain hardcoded in root `View`, or moved to a small `overlayComposer` helper after phase 2?

## References

- `pkg/repl/model.go`
- `pkg/repl/config.go`
- `pkg/repl/keymap.go`
- `pkg/repl/help_bar_types.go`
- `pkg/repl/help_drawer_types.go`
