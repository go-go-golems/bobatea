---
Title: Generic Example Playbook
Ticket: BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION
Status: active
Topics:
    - repl
    - autocomplete
    - implementation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/repl/autocomplete-generic/main.go
      Note: Minimal non-JS autocomplete demo
    - Path: pkg/repl/autocomplete_model_test.go
      Note: Automated test coverage corresponding to playbook checks
    - Path: pkg/repl/model.go
      Note: REPL integration behavior validated by playbook
ExternalSources: []
Summary: Runbook for the minimal non-JS autocomplete REPL example, including exact commands and pass/fail criteria.
LastUpdated: 2026-02-13T11:45:00-05:00
WhatFor: Provide a deterministic workflow to run and validate generic autocomplete behavior before JS integration.
WhenToUse: Use during BOBA-002 Phase 2 and Phase 3 preparation.
---


# Generic Example Playbook

## Goal

Provide exact commands and a validation checklist for the minimal generic (non-JS) autocomplete example so behavior can be tested consistently across local runs and tmux sessions.

## Context

The example program lives at `examples/repl/autocomplete-generic/main.go`.

It demonstrates both trigger paths required by BOBA-002:

- Debounce trigger (`CompletionReasonDebounce`) after typing input.
- Explicit shortcut trigger (`CompletionReasonShortcut`) via `tab`.

The completer intentionally keeps REPL trigger detection out of the model:

- Debounce only shows suggestions once token length is at least 2.
- Shortcut trigger can force a completion request immediately.

## Quick Reference

Build/compile check:

```bash
go test ./examples/repl/autocomplete-generic
```

Run the example:

```bash
go run ./examples/repl/autocomplete-generic
```

Expected key bindings in the example:

- `tab`: trigger completion explicitly and also accept when popup is open.
- `enter`: submit input or accept completion when popup is open.
- `ctrl+t`: toggle focus between input and timeline.
- `ctrl+?`: toggle full help.
- `ctrl+c` or `alt+q`: quit.

Success criteria (Tasks 11-14):

1. Popup appears after typing `co` and waiting briefly (debounce path).
2. Popup appears when pressing `tab` even if only a short prefix is typed (shortcut path).
3. `up`/`down` changes selected suggestion.
4. `enter` or `tab` inserts selected suggestion at cursor/range.
5. Input remains editable after completion apply.
6. Focus switching (`ctrl+t`) still works with autocomplete enabled.

## Usage Examples

Debounce trigger example:

1. Start program with `go run ./examples/repl/autocomplete-generic`.
2. Type `co`.
3. Wait ~120ms.
4. Confirm suggestions like `console`, `const`, `context` appear.

Shortcut trigger example:

1. Type `c`.
2. Press `tab`.
3. Confirm popup opens immediately and suggestions are shown.
4. Press `down` then `enter` to insert selected item.

Insertion behavior example:

1. Type `cons`.
2. Wait for popup and choose `console`.
3. Press `enter`.
4. Confirm input becomes `console`.

## Related

- `examples/repl/autocomplete-generic/main.go`
- `pkg/repl/model.go`
- `pkg/repl/autocomplete_model_test.go`
- `ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/design-doc/01-autocomplete-implementation-guide.md`
