---
Title: Context Panel Widget Implementation Plan
Ticket: BOBA-013-WIDGET-CONTEXTPANEL-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - help-drawer
    - implementation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Implementation checklist for context panel extraction and REPL cutover."
LastUpdated: 2026-02-14T21:50:00-05:00
WhatFor: "Execution guide for BOBA-013."
WhenToUse: "Use during BOBA-013 coding and PR review."
---

# Goal
Ship reusable context panel widget and replace REPL help drawer internals with adapter wiring.

# Context
Help drawer currently mixes reusable behavior with small REPL-specific key conflict logic. Plan preserves behavior while peeling reusable layers.

# Quick Reference
## Sequence
1. Create widget package and migrate types/state.
2. Migrate overlay layout and rendering helpers.
3. Add host-facing keymap + style injection.
4. Port tests and add overlay edge tests.
5. Integrate with REPL adapter.

## Required Parity Checks
- `alt+h` toggle still works
- `esc` close behavior unchanged
- `ctrl+r` refresh and `ctrl+g` pin unchanged
- no right cutoff for right dock
- no stale flicker while loading updates

## DoD
- all drawer tests passing in new package
- REPL tests and examples pass
- no API dependency from widget package to `pkg/repl`

# Usage Examples
```bash
go test ./pkg/tui/widgets/contextpanel/... -count=1
go test ./pkg/repl/... -count=1
```
