---
Title: Manual Test Playbook
Ticket: BOBA-004-HELP-DRAWER-REPL-IMPLEMENTATION
Status: active
Topics:
    - repl
    - help-drawer
    - implementation
    - analysis
DocType: playbooks
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/repl/autocomplete-generic/main.go
      Note: Primary manual demo entrypoint
    - Path: pkg/repl/model.go
      Note: Drawer behavior under manual validation
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-13T17:59:48.887429922-05:00
WhatFor: ""
WhenToUse: ""
---


# Manual Test Playbook

## Goal

Validate BOBA-004 help-drawer behavior in automated and manual flows.

## Context

Help drawer is keyboard-toggle UI in `pkg/repl` backed by optional `HelpDrawerProvider`. This playbook checks toggle behavior, adaptive debounce refresh while typing, and stale response safety.

## Quick Reference

### Automated Checks

```bash
go test ./pkg/repl -count=1
go test ./pkg/repl/... -count=1
golangci-lint run -v --max-same-issues=100
```

### Manual Demo Run

```bash
go run ./examples/repl/autocomplete-generic
```

### Manual Scenarios

1. Press `ctrl+h`.
Expected: help drawer opens with contextual panel.

2. Type `co`.
Expected: drawer content updates after debounce (typing trigger), plus help bar/completion updates still work.

3. Press `ctrl+r`.
Expected: drawer refreshes immediately (manual-refresh trigger).

4. Press `esc` while drawer is open and completion popup is closed.
Expected: drawer closes.

5. Open drawer and type quickly (`c`, `co`, `con`).
Expected: no stale content overwrite from older async requests.

## Usage Examples

Use after editing:

- `pkg/repl/help_drawer_types.go`
- `pkg/repl/model.go`
- `pkg/repl/config.go`
- `pkg/repl/keymap.go`
- `examples/repl/autocomplete-generic/main.go`

## Related

- `design-doc/01-help-drawer-analysis-and-implementation-guide.md`
- `reference/01-diary.md`
