---
Title: Async Provider Utility Implementation Plan
Ticket: BOBA-014-WIDGET-ASYNCPROVIDER-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - async
    - implementation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Implementation plan for shared async provider utility and call-site migration."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Execution plan for BOBA-014."
WhenToUse: "Use while extracting and integrating async provider utility."
---

# Goal
Extract and standardize async provider execution helper for widget packages.

# Context
Three REPL feature commands currently duplicate wrapper usage over `runProvider`; centralizing this utility reduces duplication and keeps panic/timeout behavior consistent.

# Quick Reference
## Implementation Steps
1. Create `pkg/tui/asyncprovider/run.go`.
2. Copy logic and keep log fields compatible.
3. Add unit tests (`success`, `panic`, `timeout`, `ctx canceled`).
4. Update `pkg/repl/model_async_provider.go` to call new utility.
5. Verify no behavior change through existing tests.

## DoD
- new package has direct tests
- REPL tests unchanged and green
- no direct `context.Background()` introduced at call sites

# Usage Examples
```bash
go test ./pkg/tui/asyncprovider/... -count=1
go test ./pkg/repl/... -count=1
```
