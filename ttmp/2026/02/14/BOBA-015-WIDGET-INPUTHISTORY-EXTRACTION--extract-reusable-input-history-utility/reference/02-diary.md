---
Title: Diary
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
Summary: "Implementation diary for input history extraction."
LastUpdated: 2026-02-14T22:12:00-05:00
WhatFor: "Capture migration details and validation evidence for BOBA-015."
WhenToUse: "Use when reviewing BOBA-015 implementation decisions and outcomes."
---

# Diary

## Goal
Extract REPL input history into a reusable package while preserving navigation behavior exactly.

## Step 1: Move history state machine and keep a REPL shim

The history implementation was moved from `pkg/repl/history.go` to `pkg/tui/inputhistory/history.go` with no behavioral changes to add/navigate/reset semantics.

A thin alias shim was added in `pkg/repl/history_alias.go` so current REPL code paths continue to use `NewHistory`, `History`, and `HistoryEntry` without large refactors.

### Prompt Context
**User prompt (verbatim):** "now implement them all, commit at each step, update your diary, test with tmux as you go too when opportune."

**Assistant interpretation:** Implement each extraction ticket incrementally with commits, diary updates, and practical runtime checks.

**Inferred user intent:** Reach reusable architecture with low-risk, auditable migrations and real-world behavior checks.

### What I did
- Added `pkg/tui/inputhistory/history.go`.
- Added `pkg/tui/inputhistory/doc.go`.
- Added `pkg/tui/inputhistory/history_test.go`.
- Removed old `pkg/repl/history.go` implementation.
- Added `pkg/repl/history_alias.go` compatibility shim.

### Why
- `History` is independent of REPL/timeline and belongs in reusable utility packages.
- Alias shim minimized migration churn and reduced regression risk.

### What worked
- `go test ./pkg/tui/inputhistory/... -count=1` passed.
- `go test ./pkg/repl/... -count=1` passed.
- tmux smoke test showed history recall still works:
  - entered `first`, `second`, pressed `Up`, input line became `second`.

### What didn't work
- tmux cleanup emitted a benign `can't find session` during teardown because the session had already exited after `Ctrl-C`.

### What I learned
- History extraction is low-risk when preserving API names and using type aliases in REPL.

### What was tricky to build
- Avoiding wide REPL test churn while still making the extraction real (not duplicated code).

### What warrants a second pair of eyes
- Verify alias shim approach is acceptable long-term versus immediate full import rewiring.

### What should be done in the future
- After remaining widget extractions, remove shim and import `pkg/tui/inputhistory` directly in REPL model/tests.

### Code review instructions
- Inspect `pkg/tui/inputhistory/history.go` and compare logic with previous implementation.
- Inspect `pkg/repl/history_alias.go` for compatibility surface.
- Re-run:
  - `go test ./pkg/tui/inputhistory/... -count=1`
  - `go test ./pkg/repl/... -count=1`

### Technical details
- tmux capture file during smoke test: `/tmp/boba015_tmux_capture.txt`.
