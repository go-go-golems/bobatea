---
Title: Autocomplete Overlay with Lipgloss v2 Canvas Layers — Analysis and Design
Ticket: BOBA-006-COMPLETION-OVERLAY-LIPGLOSS-V2
Status: active
Topics:
    - analysis
    - repl
    - autocomplete
    - lipgloss
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/2026-02-10--grail-js/internal/grailui/view.go
      Note: |-
        Reference lipgloss v2 canvas/layer composition in production-style code
        Lipgloss v2 layer/canvas composition reference
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/2026-02-10--grail-js/pkg/tealayout/chrome.go
      Note: |-
        Reusable layer helper patterns
        Layer helper patterns for overlay construction
    - Path: pkg/overlay/overlay.go
      Note: |-
        Existing v1 string-level overlay utility and limitations
        Legacy v1 string overlay baseline
    - Path: pkg/repl/config.go
      Note: |-
        Existing autocomplete config shape and defaults
        Autocomplete config extension target
    - Path: pkg/repl/keymap.go
      Note: |-
        Existing completion key bindings and help model integration
        Completion navigation and help bindings
    - Path: pkg/repl/model.go
      Note: |-
        Current inline completion rendering, scheduling, and key handling
        Current inline completion rendering and update flow
    - Path: pkg/repl/styles.go
      Note: |-
        Existing popup styles to be adapted for overlay container styling
        Popup style definitions
ExternalSources:
    - https://github.com/charmbracelet/lipgloss/discussions/506
Summary: REPL autocomplete is currently inline and must be redesigned as a size-constrained lipgloss v2 overlay with paging/scrolling; this document analyzes options and specifies a recommended implementation plan.
LastUpdated: 2026-02-13T11:42:07.294063568-05:00
WhatFor: Design and plan the REPL completion popup migration from inline rendering to a lipgloss v2 overlay model with proper anchoring and viewport behavior.
WhenToUse: Use when implementing or reviewing overlay-based completion UI in bobatea REPL.
---


# Autocomplete Overlay with Lipgloss v2 Canvas Layers

## Executive Summary

`pkg/repl` currently renders completion suggestions inline (directly below input) via `lipgloss.JoinVertical` in `Model.View()` (`pkg/repl/model.go`). This causes structural layout shifts and prevents true popup behavior.

The recommended direction is to introduce a **dedicated overlay rendering path** powered by **lipgloss v2 layers/canvas/compositor**, while keeping current autocomplete trigger policy unchanged (debounce + optional shortcut).

The overlay system should:

- anchor near the input cursor (or replace range),
- enforce max width/height constraints from config,
- support list viewport scrolling/paging when content exceeds available space,
- avoid pushing timeline/input/help sections down,
- keep key binding/help integration idiomatic with existing bobatea `KeyMap` + `help.Model`.

> [!NOTE]
> Trigger detection remains in the completer (as requested in earlier ticket work). The REPL overlay work is presentation + navigation + sizing, not semantic trigger policy.

## Why It Is Inline Today

Current flow in `pkg/repl/model.go`:

- `Model.View()` builds `header`, `timelineView`, `inputView`, `helpView`.
- `renderCompletionPopup()` returns a styled block string.
- If non-empty, code does:

```go
inputBlock := inputView
if popup := m.renderCompletionPopup(); popup != "" {
    inputBlock = lipgloss.JoinVertical(lipgloss.Left, inputView, popup)
}
```

This guarantees the popup is part of normal document flow. It cannot float above other sections.

`renderCompletionPopup()` currently only limits by item count (`completionMaxVisible`), not by terminal geometry or box height/width limits, and there is no scroll offset.

## Relevant Locations

### In bobatea

- `pkg/repl/model.go`
  - `View()`: inline layout path today
  - `renderCompletionPopup()`: popup string construction
  - `handleCompletionNavigation()`: selection movement/apply/cancel
  - `applySelectedCompletion()`: replace range application
- `pkg/repl/config.go`
  - `AutocompleteConfig`: currently has debounce/timeout/keys/max suggestions
- `pkg/repl/keymap.go`
  - completion navigation/accept/cancel bindings and help entries
- `pkg/repl/styles.go`
  - popup/item/selected styles
- `pkg/overlay/overlay.go`
  - legacy v1 string compositing helper (not layer model)

### In grail-js (reference implementation style)

- `/home/manuel/code/wesen/corporate-headquarters/2026-02-10--grail-js/internal/grailui/view.go`
  - collects `[]*lipgloss.Layer`, composes via `lipgloss.NewCompositor`, draws to `lipgloss.NewCanvas(width,height)`
- `/home/manuel/code/wesen/corporate-headquarters/2026-02-10--grail-js/pkg/tealayout/chrome.go`
  - reusable helper functions producing layers at fixed coordinates and z-order

## Lipgloss v2 Findings That Matter Here

From local API/source inspection (same v2 family used in grail-js):

- `lipgloss.NewLayer(content).X(x).Y(y).Z(z).ID(id)` gives absolute placement and stacking.
- `lipgloss.NewCompositor(layers...)` flattens and z-sorts layers.
- `lipgloss.NewCanvas(width,height)` + `canvas.Compose(compositor)` renders clipped to canvas size.
- `Compositor.Hit(x,y)` exists (useful future mouse interactions).

Observed behavior from probes:

- Oversized layer content is clipped by canvas bounds.
- Off-screen/right-shifted layer content is partially visible, clipped correctly.
- Higher `Z` wins for overlapping `Hit` regions.

> [!WARNING]
> There is currently an import-path/version split in the ecosystem:
> - grail-js uses `charm.land/lipgloss/v2` pseudo-version with `Layer/Canvas/Compositor` APIs,
> - tagged `github.com/charmbracelet/lipgloss/v2` betas can differ in API shape.
>
> The implementation ticket should explicitly pin/verify the v2 variant that provides the required layer APIs.

## Requirements for This Ticket

Functional requirements:

- Completion list must render as overlay, not inline flow block.
- Overlay must obey configurable `max width` and `max height`.
- If suggestions exceed visible area, support scrolling/paging.
- Keep debounce and shortcut triggers unchanged.
- Keep REPL key help idiomatic (`help.Model` with key bindings).

Usability requirements:

- No layout jump when popup appears/disappears.
- Selected suggestion always visible.
- Placement flips above input when insufficient space below.

Non-goals:

- No completer semantic trigger changes.
- No full Bubble Tea v2 migration in this ticket.

## Design Options

### Option A: Keep inline and only truncate

Pros:

- very low code churn.

Cons:

- still not overlay;
- still shifts layout;
- still visually wrong for REPL UX.

Verdict: reject.

### Option B: Use existing `pkg/overlay` (v1 string compositing)

Pros:

- minimal dependency changes;
- can produce floating effect in Bubble Tea v1 quickly.

Cons:

- string-level ANSI compositing is fragile and harder to extend;
- no layer graph, no built-in hit model, limited long-term path;
- diverges from grail-js direction and lipgloss v2 layer model.

Verdict: acceptable emergency fallback, not preferred.

### Option C: Add lipgloss v2 overlay rendering path in REPL (recommended)

Pros:

- aligned with requested direction and grail-js references;
- robust positioning/z-order model;
- straightforward extension to future help drawer/bar/palette overlays.

Cons:

- mixed v1+v2 lipgloss use in same repo/package;
- need careful dependency pinning due v2 API variants.

Verdict: recommended.

## Recommended Architecture

### New Overlay State in `repl.Model`

Add state fields (names illustrative):

```go
// geometry + viewport
completionOverlayVisible    bool
completionOverlayX          int
completionOverlayY          int
completionOverlayWidth      int
completionOverlayHeight     int
completionScrollTop         int   // first visible suggestion index
completionVisibleRows       int   // rows available for items (excluding header/footer)

// anchor bookkeeping
completionAnchorByte        int   // usually cursor or replace start
completionAnchorScreenX     int
completionAnchorScreenY     int
```

Keep existing selection fields; add scroll window management around them.

### Config Extensions

Extend `AutocompleteConfig` in `pkg/repl/config.go`:

```go
type AutocompleteConfig struct {
    // existing fields...
    OverlayEnabled   bool // default true when v2 available
    OverlayMaxWidth  int  // absolute cap, e.g. 56
    OverlayMaxHeight int  // absolute cap, e.g. 12
    OverlayMinWidth  int  // keep readable, e.g. 24
    OverlayPadding   int  // gap from anchor, e.g. 1
    PageSize         int  // optional explicit paging size; default = visible rows
}
```

Defaults should remain conservative and terminal-safe.

### Keymap Extensions (Idiomatic Help Model Integration)

In `pkg/repl/keymap.go`, add optional bindings for viewport navigation:

- `CompletionPageDown` (`pgdown`, `ctrl+f`)
- `CompletionPageUp` (`pgup`, `ctrl+b`)
- optional `CompletionTop` (`home`)
- optional `CompletionBottom` (`end`)

Include these in `ShortHelp`/`FullHelp` groups with mode tag `input`.

### Overlay Placement Algorithm

Inputs:

- terminal size: `m.width`, `m.height`
- base layout heights: header/timeline/input/help
- anchor point near cursor in input row
- suggestion content measurements

Algorithm:

1. Compute `desiredWidth` from longest rendered item + padding + border.
2. Clamp width to `[OverlayMinWidth, OverlayMaxWidth]` and available space.
3. Compute `desiredHeight` from rows needed (`items + chrome`), clamp to max and available vertical space.
4. Prefer below-input placement if space available; otherwise place above.
5. Clamp `x/y` into screen bounds.

Pseudo:

```go
spaceBelow := m.height - (inputY + 1)
spaceAbove := inputY

placeBelow := spaceBelow >= minPopupRows || spaceBelow >= spaceAbove
if placeBelow {
    y = inputY + 1 + cfg.OverlayPadding
    maxH = spaceBelow - cfg.OverlayPadding
} else {
    maxH = spaceAbove - cfg.OverlayPadding
    y = inputY - popupH - cfg.OverlayPadding
}
```

### Viewport / Paging Behavior

Define:

- `visibleRows`: rows available for suggestions inside popup
- `scrollTop`: first visible row index
- `selected`: current suggestion index

Invariants:

- `0 <= selected < len(suggestions)`
- `0 <= scrollTop <= max(0, len(suggestions)-visibleRows)`
- `selected` always within `[scrollTop, scrollTop+visibleRows-1]`

On `up/down`:

- move `selected` by ±1
- adjust `scrollTop` to keep selection visible.

On `page down/up`:

- jump by `pageSize` (or `visibleRows`).

Render footer hints when clipped, for example:

- `12/73  ↑↓ move  PgUp/PgDn page  Enter accept  Esc close`

### Render Pipeline (Overlay Path)

High-level:

1. Render the existing REPL base view string (without inline completion block).
2. Build popup content string for visible suggestion window.
3. Compose layers:
   - base layer: z=0, full REPL view
   - popup layer: z=20
4. Draw on canvas sized to current terminal and return `canvas.Render()`.

Pseudo:

```go
func (m *Model) View() string {
    base := m.renderBaseViewWithoutInlineCompletion()
    if !m.completionOverlayVisible {
        return base
    }

    popup := m.renderCompletionOverlayContent()
    layers := []*lgv2.Layer{
        lgv2.NewLayer(base).X(0).Y(0).Z(0).ID("repl-base"),
        lgv2.NewLayer(popup).
            X(m.completionOverlayX).
            Y(m.completionOverlayY).
            Z(20).
            ID("completion-overlay"),
    }
    comp := lgv2.NewCompositor(layers...)
    canvas := lgv2.NewCanvas(m.width, m.height)
    canvas.Compose(comp)
    return canvas.Render()
}
```

### Architecture Diagram

```text
Key/Input Events
      |
      v
Model.Update()
  |- debounce/shortcut completion request
  |- completion results + selection
  |- viewport state (selected, scrollTop)
      |
      v
Model.View()
  |- base sections: header/timeline/input/help
  |- overlay metrics: anchor + size + placement
  |- popup content window (paged/scrolling)
  |- lipgloss v2 layers -> compositor -> canvas
      |
      v
Rendered terminal frame (no layout shift)
```

## Refactor Boundaries

Keep unchanged:

- `CompletionRequest` / `CompletionResult` contracts
- completer invocation, debounce behavior, panic safety
- submission/history/timeline mechanics

Refactor:

- split base view rendering and completion overlay rendering
- replace inline `renderCompletionPopup()` usage with overlay rendering path
- extend navigation handlers for page movement + scroll bookkeeping

## Implementation Plan (Phased)

Phase 1: groundwork

- Add config fields and defaults.
- Add new key bindings + help entries.
- Add overlay state fields + helpers (`clamp`, `ensureSelectionVisible`, `pageUp/pageDown`).

Phase 2: rendering migration

- Extract `renderBaseViewWithoutInlineCompletion()`.
- Implement overlay metrics computation + popup viewport rendering.
- Add lipgloss v2 composition path behind `OverlayEnabled`.

Phase 3: behavior hardening

- Add tests for placement and viewport invariants.
- Add tests for no-layout-shift behavior (string snapshot style).
- Add manual scenario checklist in example REPL (`examples/js-repl`).

Phase 4: stabilization

- Validate in small terminals (80x24) and large terminals.
- Validate long suggestion lists and long labels.

## Test Strategy

Unit tests:

- geometry/placement:
  - below vs above placement
  - x/y clamping at edges
- viewport:
  - selection movement with scroll updates
  - page up/down with bounds
- rendering:
  - overlay hidden when no suggestions
  - overlay visible with clipped list + footer

Integration/manual:

- `go run ./examples/js-repl`
- type patterns that return large suggestion sets
- verify:
  - popup overlays without moving timeline/help,
  - `up/down/page` navigation remains consistent,
  - accept/cancel behavior unchanged.

## Risk Matrix

- v2 API mismatch/import-path churn:
  - mitigate by pinning known-good version and documenting exact import path.
- mixed v1/v2 style APIs in same package:
  - mitigate by isolating v2 usage in overlay helper file(s).
- cursor anchor drift with styled input text:
  - mitigate by anchoring to replace range and validating width with printable width helpers.

## Open Questions

1. Should we anchor popup to cursor column or replace-range start?
2. Should page keys be enabled by default or only in full-help mode?
3. Should overlay include semantic detail columns (kind/type) now, or only label/value?

## Recommendation

Proceed with **Option C**: add a dedicated lipgloss v2 overlay rendering path in `pkg/repl`, with explicit max-size and viewport logic. Keep completer semantics untouched, and keep a guarded fallback path only if v2 dependency constraints block immediate integration.

## References

- Lipgloss v2 discussion and API direction:
  - https://github.com/charmbracelet/lipgloss/discussions/506
- Grail-js layer composition examples:
  - `/home/manuel/code/wesen/corporate-headquarters/2026-02-10--grail-js/internal/grailui/view.go`
  - `/home/manuel/code/wesen/corporate-headquarters/2026-02-10--grail-js/pkg/tealayout/chrome.go`
- Bobatea current REPL:
  - `pkg/repl/model.go`
  - `pkg/repl/config.go`
  - `pkg/repl/keymap.go`
  - `pkg/repl/styles.go`
