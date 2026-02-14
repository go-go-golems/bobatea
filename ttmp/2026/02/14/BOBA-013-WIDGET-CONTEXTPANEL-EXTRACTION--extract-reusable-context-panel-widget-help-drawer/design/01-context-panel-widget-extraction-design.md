---
Title: Context Panel Widget Extraction Design
Ticket: BOBA-013-WIDGET-CONTEXTPANEL-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - help-drawer
    - implementation
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Design for extracting help drawer into a reusable contextual panel widget with dock and pin behavior."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Define reusable context panel API and host adaptation strategy."
WhenToUse: "Use during BOBA-013 implementation and review."
---

# Context Panel Widget Extraction Design

## Goal
Extract help drawer behavior into `pkg/tui/widgets/contextpanel` with reusable keyboard, debounce, pin, and overlay logic.

## Current Source
- `pkg/repl/helpdrawer_model.go`
- `pkg/repl/helpdrawer_overlay.go`
- `pkg/repl/help_drawer_types.go`

## Extraction Scope
In scope:
- toggle/open/close
- debounce while typing
- manual refresh trigger
- pinned mode
- request lifecycle + stale-result handling
- dock/size/margin based overlay layout
- panel rendering with title/subtitle/body/diagnostics/footer

Out of scope:
- REPL-specific completion cancel conflict logic
- REPL-specific key labels and command text

## Proposed Contracts
```go
type Provider interface {
    GetContextPanel(ctx context.Context, req Request) (Document, error)
}

type KeyMap struct {
    Toggle key.Binding
    Close key.Binding
    Refresh key.Binding
    Pin key.Binding
}

type Dock string // above, right, left, bottom
```

## REPL Adapter Rule
Keep one compatibility shim in `pkg/repl` mapping existing `HelpDrawer*` types to widget package types. Remove shim later once evaluator types are migrated.

## Risks
- dock math regressions (cutoff at right/bottom)
- footer key mismatch after extraction
- pin behavior regressions while typing

## Validation
- port `pkg/repl/help_drawer_model_test.go`
- manual run with drawer docking modes and pin toggle
