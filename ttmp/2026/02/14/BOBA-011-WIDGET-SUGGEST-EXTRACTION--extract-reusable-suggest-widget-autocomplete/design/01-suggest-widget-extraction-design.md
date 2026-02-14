---
Title: Suggest Widget Extraction Design
Ticket: BOBA-011-WIDGET-SUGGEST-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - autocomplete
    - implementation
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Design for extracting autocomplete state, key handling, and overlay rendering into a reusable suggest widget package."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Define architecture and boundaries for suggest widget extraction without REPL/timeline coupling."
WhenToUse: "Use when implementing BOBA-011 extraction work and reviewing API boundaries for completion UI reuse."
---

# Suggest Widget Extraction Design

## Goal
Extract `pkg/repl` completion behavior into a reusable widget package that can be used by REPL and non-REPL Bubble Tea applications.

## Current Source of Truth
- `pkg/repl/completion_model.go`
- `pkg/repl/completion_overlay.go`
- `pkg/repl/autocomplete_types.go`

## Extraction Scope
In scope:
- debounce request scheduling
- async completion request dispatch
- completion result lifecycle (stale-drop, show/hide, selection, paging)
- replacement application logic
- overlay geometry and rendering

Out of scope:
- REPL-specific keymap field names
- evaluator-specific completion semantics
- timeline/input layout policy owned by REPL host

## Target Package
`pkg/tui/widgets/suggest`

## Proposed Interfaces
```go
type BufferSnapshot struct {
    Value string
    CursorByte int
}

type BufferMutator interface {
    SetValue(string)
    SetCursorByte(int)
}

type Provider interface {
    CompleteInput(ctx context.Context, req Request) (Result, error)
}

type KeyMap struct {
    Trigger key.Binding
    Accept key.Binding
    Cancel key.Binding
    Prev key.Binding
    Next key.Binding
    PageUp key.Binding
    PageDown key.Binding
}

type Anchor struct { X int; Y int }
type Viewport struct { Width int; Height int }
```

## State Machine Rules
- Only latest request id is authoritative.
- Debounce request is skipped when input snapshot is unchanged.
- Accept applies replacement range `[ReplaceFrom:ReplaceTo]`.
- Popup visibility transitions are deterministic:
  - hidden on provider error
  - hidden when `Show=false` or no suggestions
  - shown with selection reset when valid result arrives

## Overlay Rules
- keep current behavior parity:
  - vertical placement: `auto|above|below|bottom`
  - horizontal growth: `left|right`
  - clamp to viewport bounds
  - support max/min width, max height, margins, offsets, no-border mode

## REPL Adapter Strategy
`pkg/repl` keeps a thin adapter layer:
- maps REPL keymap to suggest widget keymap
- provides anchor and viewport from REPL layout
- bridges `textinput.Model` through `BufferMutator`

## Risks
- key precedence regressions (drawer/palette/completion conflict)
- overlay Y anchor drift in REPL-specific layout
- scroll/selection regressions for page navigation

## Validation
- port and pass existing completion tests from `pkg/repl/autocomplete_model_test.go`
- run manual parity in:
  - `examples/repl/autocomplete-generic`
  - `examples/js-repl`
