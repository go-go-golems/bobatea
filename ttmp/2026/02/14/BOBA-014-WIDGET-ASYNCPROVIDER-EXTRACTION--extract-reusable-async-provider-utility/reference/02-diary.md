---
Title: Diary
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
Summary: "Implementation diary for async provider utility extraction."
LastUpdated: 2026-02-14T22:05:00-05:00
WhatFor: "Capture what changed, why, and validation outcomes for BOBA-014."
WhenToUse: "Use when reviewing BOBA-014 implementation history and test evidence."
---

# Diary

## Goal
Extract `runProvider` from `pkg/repl` into a reusable shared utility and keep behavior identical.

## Step 1: Extract utility and cut REPL call sites

This step moved provider timeout/panic handling into `pkg/tui/asyncprovider` and switched REPL completion/help/help-drawer wrappers to call the new package.

The work was intentionally done before larger widget extraction because async behavior is a common dependency across all widgets and the change is low-risk when tested directly.

### Prompt Context
**User prompt (verbatim):** "now implement them all, commit at each step, update your diary, test with tmux as you go too when opportune."

**Assistant interpretation:** Implement each high-confidence extraction ticket incrementally with commits and documented progress.

**Inferred user intent:** Reduce `pkg/repl` coupling through real code changes with auditable progress and verification evidence.

### What I did
- Added `pkg/tui/asyncprovider/run.go` with generic `Run[T]`.
- Added `pkg/tui/asyncprovider/run_test.go` covering success, panic recovery, timeout, canceled context.
- Added `pkg/tui/asyncprovider/doc.go` package docs.
- Updated `pkg/repl/model_async_provider.go` to call `asyncprovider.Run` for:
  - `completionCmd`
  - `helpBarCmd`
  - `helpDrawerCmd`
- Removed in-file `runProvider` from REPL package.

### Why
- Keep panic/timeout logic centralized and reusable for extracted widgets.
- Avoid duplicated resilience logic across feature packages.

### What worked
- Behavior parity preserved in REPL tests.
- New async utility tests pass and cover all critical paths.

### What didn't work
- Initial compile failure due to stale `time` import in `pkg/repl/model_async_provider.go`.
  - Command: `go test ./pkg/repl/... -count=1`
  - Error: `"time" imported and not used`
- Fixed by removing unused import.

### What I learned
- This extraction is mechanically simple but should happen early because every widget provider path depends on it.

### What was tricky to build
- Preserving request metadata logging fields exactly (`request_id`, `provider`) to avoid losing diagnostics.

### What warrants a second pair of eyes
- Panic/error formatting string parity (`"%s panic: %v"`) relied upon by tests or logs.
- Any external tooling parsing provider log fields.

### What should be done in the future
- Reuse `asyncprovider.Run` in extracted widget packages directly and avoid re-wrapping in REPL.

### Code review instructions
- Start at `pkg/tui/asyncprovider/run.go`.
- Check REPL call-site changes in `pkg/repl/model_async_provider.go`.
- Validate with:
  - `go test ./pkg/tui/asyncprovider/... -count=1`
  - `go test ./pkg/repl/... -count=1`

### Technical details
- API:
```go
Run[T any](baseCtx context.Context, requestID uint64, timeout time.Duration, providerName, panicPrefix string, fn func(context.Context) (T, error)) (T, error)
```
