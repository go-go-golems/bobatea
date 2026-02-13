---
Title: Help Bar Analysis and Implementation Guide
Ticket: BOBA-003-HELP-BAR-REPL-IMPLEMENTATION
Status: active
Topics:
    - repl
    - help-bar
    - implementation
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: bobatea/pkg/autocomplete/autocomplete.go
      Note: Current autocomplete model used for interoperability analysis
    - Path: bobatea/pkg/repl/config.go
      Note: Help bar configuration defaults and feature switches
    - Path: bobatea/pkg/repl/evaluator.go
      Note: Optional provider capability extension point
    - Path: bobatea/pkg/repl/model.go
      Note: Primary integration surface for input-driven help bar scheduling and rendering
    - Path: bobatea/pkg/repl/styles.go
      Note: Style extensions for contextual help bar severity states
ExternalSources: []
Summary: Detailed analysis and implementation plan for a typing-triggered REPL help bar with completer/evaluator-backed symbol metadata.
LastUpdated: 2026-02-13T10:31:00-05:00
WhatFor: Build a fast contextual help bar that updates while typing and remains compatible with evolving autocomplete implementations.
WhenToUse: Use while implementing or reviewing REPL input-context UX features.
---


# Help Bar Analysis and Implementation Guide

## Executive Summary

This document proposes a contextual help bar for `bobatea/pkg/repl` that updates while the user types, similar to autocomplete cadence, but with separate rendering and behavior goals.

The help bar should:

- refresh on input edits through debounced scheduling,
- delegate semantic decisions to evaluator/completer-side providers,
- display concise context (symbol/type/signature/module hints),
- stay stable under asynchronous updates and slow providers.

> [!NOTE]
> Trigger policy remains provider-owned. The REPL schedules update opportunities; providers decide whether to show data.

## Problem Statement

The current REPL in `bobatea/pkg/repl/model.go` renders:

- title,
- timeline transcript,
- input line,
- static one-line keyboard help.

There is no live, typed-symbol feedback. Users cannot see type/signature context for the token at cursor position while typing.

Without a help bar:

- discoverability of evaluator APIs is poor,
- users rely on trial-and-error and runtime failures,
- high-frequency edit loops feel blind.

## Existing Architecture Context

Relevant surfaces:

- `bobatea/pkg/repl/model.go`
- input key handling and render composition.

- `bobatea/pkg/repl/config.go`
- natural place for help-bar feature flags and debounce tuning.

- `bobatea/pkg/repl/evaluator.go`
- current narrow evaluator interface; requires optional capability extension.

- `bobatea/pkg/autocomplete/autocomplete.go`
- existing completion UI model; currently not directly embedded in `repl.Model`.

> [!TIP]
> The help bar should not depend on the legacy autocomplete widget internals. This keeps room for replacing autocomplete with a new engine.

## UX Contract for Help Bar

### User-visible behavior

- As typing/cursor changes settle, help bar updates.
- If context is unknown or ambiguous, bar can hide or show neutral text.
- Update latency target: perceived near-real-time (typically <200ms after typing pause).

### Content scope

Single-line concise output, for example:

- function signature: `map(arr: []T, fn: (T) -> U) -> []U`
- inferred type: `obj.user.email: string`
- module hint: `require("database") exports: query, exec, transaction`

### Non-goals

- full documentation rendering (belongs to help drawer ticket),
- multi-line code examples in this bar,
- heavy parser diagnostics stream.

## Proposed Provider Contract

Introduce optional REPL capability interface (separate file recommended, e.g. `bobatea/pkg/repl/help_bar_types.go`):

```go
type HelpBarReason string

const (
    HelpBarReasonDebounce HelpBarReason = "debounce"
    HelpBarReasonShortcut HelpBarReason = "shortcut"
    HelpBarReasonManual   HelpBarReason = "manual"
)

type HelpBarRequest struct {
    Input      string
    CursorByte int
    RequestID  uint64
    Reason     HelpBarReason
}

type HelpBarPayload struct {
    Show      bool
    Text      string
    Kind      string // signature|type|module|error|info
    Severity  string // info|warning|error
    Ephemeral bool
}

type HelpBarProvider interface {
    GetHelpBar(ctx context.Context, req HelpBarRequest) (HelpBarPayload, error)
}
```

Provider-owned decisions:

- whether bar should be shown (`Show`),
- what content appears,
- semantic classification (`Kind`).

## Integration Strategy in `repl.Model`

Add local state:

```go
helpBarProvider   HelpBarProvider
helpBarVisible    bool
helpBarPayload    HelpBarPayload
helpBarReqSeq     uint64
helpBarDebounce   time.Duration
helpBarTimeout    time.Duration
helpBarLastErr    error
```

Add messages:

```go
type helpBarTickMsg struct{ RequestID uint64 }
type helpBarResultMsg struct {
    RequestID uint64
    Payload   HelpBarPayload
    Err       error
}
```

### Scheduling rule

On input mutation:

- increment `helpBarReqSeq`,
- schedule `helpBarTickMsg` after debounce,
- on tick, drop stale IDs and request provider asynchronously.

### Stale-drop rule

If `msg.RequestID != helpBarReqSeq`, ignore.

This matches safe async patterns proposed for autocomplete.

## Dataflow Diagram

```mermaid
flowchart TD
  A[Input edited or cursor moved] --> B[updateInput records new text state]
  B --> C[schedule helpBarTickMsg debounce]
  C --> D{tick id latest?}
  D -- no --> E[drop]
  D -- yes --> F[HelpBarProvider.GetHelpBar]
  F --> G[helpBarResultMsg]
  G --> H{result id latest?}
  H -- no --> E
  H -- yes --> I{payload.Show}
  I -- no --> J[hide help bar]
  I -- yes --> K[render concise contextual line]
```

## Rendering Design

Current `View()` order in `model.go` is transcript -> input -> static help line.

Recommended layout:

1. timeline view
2. input
3. help bar (dynamic)
4. static key help

Rendering helper:

```go
func (m *Model) renderHelpBar() string {
    if !m.helpBarVisible {
        return ""
    }
    style := m.styles.Info
    switch m.helpBarPayload.Severity {
    case "warning":
        style = m.styles.HelpText
    case "error":
        style = m.styles.Error
    }
    return style.Render(m.helpBarPayload.Text)
}
```

## Key Design Decisions

### Decision 1: Keep help bar separate from autocomplete list

Rationale:

- different purpose and density,
- bar should still work if suggestion popup is hidden,
- supports replacing autocomplete widget later.

### Decision 2: Reuse scheduling pattern but keep independent provider interface

Rationale:

- shared mental model for stale handling,
- independent iteration speed and testing.

### Decision 3: Provider decides whether context is display-worthy

Rationale:

- keeps REPL policy-simple,
- language-specific heuristics live near evaluator/parser.

## Alternatives Considered

### A) Help bar derived from selected autocomplete item only

Pros:

- easy if list already exists,
- no new provider.

Cons:

- no help without visible completion list,
- cannot show context when no candidate list is relevant.

Status: rejected.

### B) Render help bar as timeline entity

Pros:

- reuses timeline renderers.

Cons:

- semantically wrong (help bar is transient input state, not transcript event),
- pollutes transcript history.

Status: rejected.

### C) Single unified `InputContextProvider` for autocomplete+help bar+drawer

Pros:

- one parser call per debounce cycle,
- shared cursor context and cache,
- cleaner cross-feature consistency.

Cons:

- larger initial refactor.

Status: recommended medium-term architecture.

## Implementation Plan

### Phase 1: Contracts and Config

- Add `HelpBarProvider` types and request/result structs.
- Extend `Config` with:

```go
type HelpBarConfig struct {
    Enabled        bool
    Debounce       time.Duration
    RequestTimeout time.Duration
    ShowLoading    bool
}
```

### Phase 2: Model Wiring

- Add help bar state and async messages in `model.go`.
- Discover optional provider via type assertion in `NewModel`.

### Phase 3: Input-driven Scheduling

- Detect text/cursor changes in `updateInput`.
- Schedule debounce tick.
- Dispatch provider call on latest tick only.

### Phase 4: Rendering and Styling

- Add `renderHelpBar()` helper.
- Extend `Styles` in `bobatea/pkg/repl/styles.go` if needed for `HelpBarInfo`, `HelpBarWarn`, `HelpBarError`.

### Phase 5: Tests

Add `repl` tests for:

- debounce coalescing,
- stale response drop,
- hide/show behavior based on `Show` flag,
- view composition order.

## Pseudocode

```go
func (m *Model) onInputChanged() tea.Cmd {
    if !m.config.HelpBar.Enabled || m.helpBarProvider == nil {
        m.helpBarVisible = false
        return nil
    }
    m.helpBarReqSeq++
    id := m.helpBarReqSeq
    return tea.Tick(m.config.HelpBar.Debounce, func(time.Time) tea.Msg {
        return helpBarTickMsg{RequestID: id}
    })
}

func (m *Model) handleHelpBarTick(msg helpBarTickMsg) tea.Cmd {
    if msg.RequestID != m.helpBarReqSeq {
        return nil
    }
    req := HelpBarRequest{
        Input: m.textInput.Value(),
        CursorByte: m.textInput.Position(),
        RequestID: msg.RequestID,
        Reason: HelpBarReasonDebounce,
    }
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), m.config.HelpBar.RequestTimeout)
        defer cancel()
        payload, err := m.helpBarProvider.GetHelpBar(ctx, req)
        return helpBarResultMsg{RequestID: msg.RequestID, Payload: payload, Err: err}
    }
}
```

## Performance and Reliability Considerations

- Keep request timeouts short (200-500ms typical).
- Use request IDs to avoid stale flicker.
- Optional lightweight cache in provider keyed by `(input,cursor)`.
- Avoid expensive markdown rendering in help bar path.

## Interop with Autocomplete Replacement

> [!IMPORTANT]
> This design intentionally does not depend on `pkg/autocomplete.Model` internals.

If autocomplete is replaced (or rewritten headless), help bar still works because it only depends on:

- input snapshot,
- cursor position,
- provider result contract.

Recommended future direction:

- `InputContextProvider` returning bundle:
  - completion candidates,
  - help bar payload,
  - help drawer payload seed.

## Risks and Mitigations

Risk: noisy frequent updates while cursor moves.

- Mitigation: debounce + same-request dedupe.

Risk: provider latency degrades typing feel.

- Mitigation: async non-blocking requests, timeout, stale drop.

Risk: inconsistent semantics across languages.

- Mitigation: per-evaluator providers behind shared contract.

## Acceptance Criteria

- Help bar updates from typing cadence (debounced).
- REPL has no symbol-trigger heuristics for bar display.
- Provider can suppress display with `Show=false`.
- Stale async responses are ignored.
- Feature is safe when autocomplete widget is replaced or rewritten.

## Checklist

- [ ] Add provider interface/types.
- [ ] Add `HelpBarConfig` defaults.
- [ ] Wire model state and messages.
- [ ] Implement scheduling and async handling.
- [ ] Implement dynamic render line and styles.
- [ ] Add tests for debounce/stale/view behavior.

## References

- `bobatea/pkg/repl/model.go`
- `bobatea/pkg/repl/config.go`
- `bobatea/pkg/repl/evaluator.go`
- `bobatea/pkg/repl/styles.go`
- `bobatea/pkg/autocomplete/autocomplete.go`
- `bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/design-doc/01-autocomplete-implementation-guide.md`
