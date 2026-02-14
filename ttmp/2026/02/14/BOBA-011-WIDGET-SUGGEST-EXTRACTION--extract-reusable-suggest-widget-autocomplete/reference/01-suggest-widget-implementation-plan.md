---
Title: Suggest Widget Implementation Plan
Ticket: BOBA-011-WIDGET-SUGGEST-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - autocomplete
    - implementation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Step-by-step build plan for extracting and integrating the reusable suggest widget."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Execution checklist for BOBA-011 implementation."
WhenToUse: "Use while implementing and reviewing suggest widget extraction PRs."
---

# Goal
Ship `pkg/tui/widgets/suggest` and switch REPL to consume it with zero behavior regressions.

# Context
Current completion logic in `pkg/repl` is stable but tightly coupled to `Model` fields. This plan moves logic into a host-agnostic widget and leaves a thin REPL adapter.

# Quick Reference
## Target Files
- New:
  - `pkg/tui/widgets/suggest/*.go`
- Update:
  - `pkg/repl/model.go`
  - `pkg/repl/model_input.go`
  - `pkg/repl/keymap.go`
  - `pkg/repl/model_layout.go` (if needed for anchor changes)

## Sequence
1. Create widget package with copied state machine + types.
2. Add widget tests ported from existing completion tests.
3. Add REPL adapter layer.
4. Replace direct completion fields/calls in REPL.
5. Validate examples and full test suite.

## DoD
- all completion tests pass
- no flicker regression when typing with open overlay
- tab/enter/esc/up/down/pgup/pgdown behavior unchanged
- overlay placement/growth config still works

# Usage Examples
```bash
go test ./pkg/tui/widgets/suggest/... -count=1
go test ./pkg/repl/... -count=1
go run ./examples/repl/autocomplete-generic
```
