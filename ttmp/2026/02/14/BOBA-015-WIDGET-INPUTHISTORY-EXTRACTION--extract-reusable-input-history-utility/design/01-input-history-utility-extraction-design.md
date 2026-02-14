---
Title: Input History Utility Extraction Design
Ticket: BOBA-015-WIDGET-INPUTHISTORY-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - history
    - implementation
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Design for moving REPL history state into reusable input history utility package."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Define extraction boundaries and API for reusable input history manager."
WhenToUse: "Use during BOBA-015 implementation and code review."
---

# Input History Utility Extraction Design

## Goal
Move history navigation/data structure from `pkg/repl/history.go` into `pkg/tui/inputhistory`.

## Current Source
- `pkg/repl/history.go`

## Scope
In scope:
- append entry
- dedupe adjacent repeated inputs
- bounded max size
- `NavigateUp`/`NavigateDown`
- reset/clear
- immutable getters

Out of scope:
- REPL submit semantics
- UI-level key handling

## API Proposal
```go
type Entry struct {
    Input string
    Output string
    IsErr bool
}

type History struct { ... }

func New(maxSize int) *History
func (h *History) Add(input, output string, isErr bool)
func (h *History) NavigateUp() string
func (h *History) NavigateDown() string
func (h *History) ResetNavigation()
func (h *History) Clear()
func (h *History) Entries() []Entry
func (h *History) Inputs() []string
```

## Risks
- off-by-one regressions in navigation boundaries
- accidental behavior change in dedupe rule

## Validation
- port existing `history` tests in `pkg/repl/repl_test.go`
- verify input navigation parity in REPL examples
