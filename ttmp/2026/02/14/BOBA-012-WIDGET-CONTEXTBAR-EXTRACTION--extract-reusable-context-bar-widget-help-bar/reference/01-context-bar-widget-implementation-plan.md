---
Title: Context Bar Widget Implementation Plan
Ticket: BOBA-012-WIDGET-CONTEXTBAR-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - help-bar
    - implementation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Implementation checklist and test plan for context bar extraction."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Execution guide for BOBA-012."
WhenToUse: "Use during coding and review of context bar extraction PR."
---

# Goal
Introduce `pkg/tui/widgets/contextbar` and migrate REPL help bar logic to it.

# Context
Help bar behavior is already mostly reusable; extraction should be low risk if relayout triggers are preserved in host.

# Quick Reference
## Build Steps
1. Copy types + state machine into new package.
2. Add typed result/debounce messages in widget package.
3. Port existing tests.
4. Integrate widget into REPL model through adapter methods.
5. Remove old helpbar internals from REPL package.

## Test Matrix
- debounce coalescing
- stale response drop
- hide on error
- hide on empty payload
- remain visible while new debounce request pending

## DoD
- `go test ./pkg/repl/...` and `go test ./pkg/tui/widgets/contextbar/...` pass
- no visual regression in `examples/js-repl`

# Usage Examples
```bash
go test ./pkg/tui/widgets/contextbar/... -count=1
go test ./pkg/repl/... -count=1
go run ./examples/js-repl
```
