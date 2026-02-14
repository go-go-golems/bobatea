---
Title: Input History Utility Implementation Plan
Ticket: BOBA-015-WIDGET-INPUTHISTORY-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - history
    - implementation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Step-by-step migration guide for extracting input history utility from REPL."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Execution checklist for BOBA-015."
WhenToUse: "Use while implementing and reviewing input history extraction."
---

# Goal
Ship reusable `pkg/tui/inputhistory` and switch REPL to consume it.

# Context
History is already self-contained and should be extracted first among low-risk utilities.

# Quick Reference
## Steps
1. Create new package and move `History` implementation.
2. Add unit tests in new package.
3. Update REPL imports/usages.
4. Remove old `pkg/repl/history.go`.
5. Run full REPL tests.

## DoD
- zero behavior changes in history navigation
- all old tests passing
- no `pkg/repl` import in `pkg/tui/inputhistory`

# Usage Examples
```bash
go test ./pkg/tui/inputhistory/... -count=1
go test ./pkg/repl/... -count=1
```
