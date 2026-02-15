---
Title: Context Bar Widget Extraction Design
Ticket: BOBA-012-WIDGET-CONTEXTBAR-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - help-bar
    - implementation
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Design for extracting help bar debounce/provider/render lifecycle into a reusable context bar widget."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Define context bar widget boundaries and host integration contract."
WhenToUse: "Use during BOBA-012 implementation and review."
---

# Context Bar Widget Extraction Design

## Goal
Move help bar mechanics out of REPL model into `pkg/tui/widgets/contextbar`.

## Current Source
- `pkg/repl/helpbar_model.go`
- `pkg/repl/help_bar_types.go`

## Extraction Scope
In scope:
- debounce scheduling
- async request/response lifecycle
- stale result rejection
- visibility and payload state transitions
- style-severity mapping hooks

Out of scope:
- host relayout calls
- host-level key routing or focus policy

## Target API
```go
type Provider interface {
    GetHelpBar(ctx context.Context, req Request) (Payload, error)
}

type Widget struct { ... }

func (w *Widget) OnBufferChanged(prev, cur BufferSnapshot) tea.Cmd
func (w *Widget) HandleResult(msg ResultMsg) (visibilityChanged bool)
func (w *Widget) View(styles Styles) string
```

## Host Responsibilities
- provide input snapshot + cursor
- run returned commands through Bubble Tea
- reflow layout when `visibilityChanged=true`

## Risks
- accidental layout churn due to over-refresh
- stale result handling regressions

## Validation
- port `pkg/repl/help_bar_model_test.go`
- verify JS help bar integration test remains green
