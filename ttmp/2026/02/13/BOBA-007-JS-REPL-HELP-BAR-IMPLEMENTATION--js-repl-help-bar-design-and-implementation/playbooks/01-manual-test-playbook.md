---
Title: Manual Test Playbook
Ticket: BOBA-007-JS-REPL-HELP-BAR-IMPLEMENTATION
Status: active
Topics:
    - repl
    - javascript
    - help-bar
    - implementation
    - analysis
DocType: playbooks
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/js-repl/main.go
      Note: Manual validation entrypoint
    - Path: pkg/repl/evaluators/javascript/evaluator.go
      Note: Help-bar behavior under test
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-13T17:04:18.30090388-05:00
WhatFor: ""
WhenToUse: ""
---


# Manual Test Playbook

## Goal

Verify JS REPL help-bar behavior end to end after BOBA-007 implementation.

## Context

This playbook validates `GetHelpBar` behavior in the JavaScript evaluator plus REPL help-bar rendering in `examples/js-repl`.

## Quick Reference

### Automated Checks

```bash
go test ./pkg/repl/evaluators/javascript -count=1
go test ./pkg/repl/... -count=1
golangci-lint run -v --max-same-issues=100
```

### Manual Run

```bash
go run ./examples/js-repl
```

### Scenarios

1. Type `cons` and pause for debounce.
Expected: help bar appears with `console` signature/object summary.

2. Type `console.log`.
Expected: help bar shows `console.log(...args): void`.

3. Type:
```text
const fs = require("fs");
fs.re
```
Expected: help bar shows a module-aware `fs.*` signature (for example `fs.readFile(...)` or another matching candidate).

4. Type:
```text
function localFn(a, b, c) { return a + b + c; }
localFn
```
Expected: help bar falls back to runtime metadata and shows function arity/name (`arity 3`).

5. Type a single character `c` and pause.
Expected: no help bar text for debounce one-character identifier input.

## Usage Examples

Use this when reviewing regressions after changing:

- `pkg/repl/evaluators/javascript/evaluator.go`
- `pkg/repl/model.go`
- `examples/js-repl/main.go`

## Related

- `design-doc/01-js-repl-help-bar-implementation-analysis-and-plan.md`
- `reference/01-diary.md`
