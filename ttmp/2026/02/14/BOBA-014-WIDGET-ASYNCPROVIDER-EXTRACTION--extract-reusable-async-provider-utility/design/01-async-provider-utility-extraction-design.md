---
Title: Async Provider Utility Extraction Design
Ticket: BOBA-014-WIDGET-ASYNCPROVIDER-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - async
    - implementation
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Design for extracting runProvider timeout/panic handling utility from REPL into reusable package."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Define reusable async provider runner API and migration plan."
WhenToUse: "Use during BOBA-014 implementation and review."
---

# Async Provider Utility Extraction Design

## Goal
Move `runProvider` out of `pkg/repl/model_async_provider.go` into a shared package used by all widgets.

## Current Source
- `pkg/repl/model_async_provider.go`

## Target Package
`pkg/tui/asyncprovider`

## API Proposal
```go
func Run[T any](
    baseCtx context.Context,
    requestID uint64,
    timeout time.Duration,
    providerName string,
    panicPrefix string,
    fn func(context.Context) (T, error),
) (T, error)
```

## Behavioral Guarantees
- wraps call in `context.WithTimeout`
- recovers provider panic
- logs panic with stack + provider name + request id
- returns typed zero value + wrapped error on panic

## Risks
- losing log detail fields during extraction
- accidentally changing timeout semantics

## Validation
- add focused tests for timeout, cancellation, panic recovery, success path
- ensure REPL command wrappers (`completionCmd`, `helpBarCmd`, `helpDrawerCmd`) still match previous behavior
