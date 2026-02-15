---
Title: JS REPL Help Bar Implementation Analysis and Plan
Ticket: BOBA-007-JS-REPL-HELP-BAR-IMPLEMENTATION
Status: active
Topics:
    - repl
    - javascript
    - help-bar
    - implementation
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-goja/pkg/jsparse/analyze.go
      Note: |-
        Analysis bundle and completion resolution entrypoints
        Analysis API used to derive completion/help context
    - Path: ../../../../../../../go-go-goja/pkg/jsparse/completion.go
      Note: |-
        Completion context/candidate model and detail metadata used by JS evaluator
        Source of completion context and candidate metadata
    - Path: examples/js-repl/main.go
      Note: |-
        Main JS REPL example that should demonstrate the help bar
        Interactive validation entrypoint
    - Path: pkg/repl/evaluators/javascript/evaluator.go
      Note: |-
        JS evaluator where GetHelpBar must be implemented
        Implementation target for JS help-bar provider
    - Path: pkg/repl/help_bar_types.go
      Note: |-
        Help bar provider contract consumed by evaluator implementations
        Help bar provider contract
    - Path: pkg/repl/model.go
      Note: |-
        Debounced help bar request scheduling and rendering lifecycle
        Help bar request lifecycle in REPL model
ExternalSources: []
Summary: Analysis and implementation plan for extending JS REPL with contextual help-bar support and pragmatic signature/type hints.
LastUpdated: 2026-02-13T16:12:00-05:00
WhatFor: Plan and guide implementation of JS-specific help bar behavior in the REPL.
WhenToUse: Use while implementing and reviewing JS evaluator help-bar functionality.
---


# JS REPL Help Bar Implementation Analysis and Plan

## Executive Summary

This ticket scopes JavaScript-specific help bar behavior for `examples/js-repl` and `pkg/repl/evaluators/javascript`.

The key finding is that JS completion is implemented and already carries lightweight metadata (`CompletionCandidate.Detail`), but there is currently no `GetHelpBar` implementation in the JS evaluator and no robust signature/type pipeline exposed to the help bar.

Recommended approach:

- implement `repl.HelpBarProvider` on the JS evaluator,
- start with static/index-driven hints (fast, deterministic, side-effect free),
- add safe runtime introspection as a fallback (for value class and function arity),
- keep outputs concise and single-line for the help bar.

## Problem Statement

`pkg/repl/model.go` now supports a generic help bar contract, but the JavaScript evaluator does not implement it. In practice:

- JS REPL users get completion but no contextual inline help text while typing.
- Function/type quality is ambiguous; the codebase does not currently provide rich signature inference.

A concrete design is needed to answer: should type/signature hints come from runtime lookup, static analysis, or both?

## Current Capability Audit

### What already exists

- Generic help bar support in REPL model:
  - debounced scheduling,
  - stale-result drop by request ID,
  - timeout-bounded provider calls,
  - panic recovery,
  - rendering pipeline.
- JS completion support:
  - `Evaluator.CompleteInput(...)` is implemented.
  - `jsparse` context + candidate resolution is wired.
- `jsparse` metadata surfaces:
  - completion context (`property`, `identifier`, etc.),
  - candidate kind and `Detail` fields,
  - lexical scope/binding resolution.

### What is missing

- No `GetHelpBar(ctx, req)` in `pkg/repl/evaluators/javascript/evaluator.go`.
- No dedicated JS signature registry for help-bar text.
- No end-to-end JS example behavior for help bar.

### Direct answer: runtime function types already implemented?

No, not fully.

- Existing metadata mostly gives coarse labels (`method`, `number`, binding kind), not robust signatures.
- Runtime can provide limited hints (function arity/name), but cannot by itself provide complete source-level type info.

## Proposed Solution

Implement `GetHelpBar(ctx, req)` in JS evaluator using a hybrid strategy:

1. Static/contextual first
- Reuse the existing parse/analyze/completion context flow.
- Resolve nearest exact/prefix symbol and candidate metadata from `jsparse`.

2. Signature catalog
- Add curated signature snippets for key APIs:
- built-ins (`console`, `JSON`, `Math`, common array/string methods),
- known node module aliases currently used by completion (`fs`, `path`, `url`).

3. Safe runtime fallback
- If static metadata is insufficient, inspect already-bound runtime globals safely:
- value class (`function`, `object`, etc.),
- function arity (`length`) and name where available,
- no invocation, no side-effectful property traversal.

4. Payload shaping
- Return concise `HelpBarPayload` with:
- `Show`, `Text`, `Kind`, `Severity`.
- Keep trigger/show policy provider-owned.

> [!IMPORTANT]
> Help-bar lookup must remain side-effect free. No evaluation of arbitrary expressions just to infer types.

## Design Decisions

### Decision 1: Hybrid metadata model

Rationale:

- static data is deterministic and fast,
- runtime fallback improves usefulness for dynamic/global symbols,
- avoids immediate heavy dependency on full TS language service.

### Decision 2: Single-line concise help text

Rationale:

- this is an inline signal channel,
- detailed docs and longer explanations belong to help drawer (BOBA-004).

### Decision 3: Provider-owned display policy

Rationale:

- keeps REPL model generic,
- JS evaluator decides confidence thresholds for show/hide.

### Decision 4: Reuse existing completion context

Rationale:

- minimizes duplicate parser logic,
- keeps autocomplete and help-bar behavior consistent.

## Alternatives Considered

### A) Runtime-only introspection

Pros:

- reflects live runtime state.

Cons:

- weak signature fidelity,
- risky unless heavily sandboxed from side effects.

Status: rejected as primary strategy.

### B) Static-only using existing candidate details

Pros:

- fastest to ship.

Cons:

- often too generic to be truly helpful.

Status: acceptable baseline, insufficient alone for target UX.

### C) Full TypeScript language-service integration

Pros:

- best signature/type fidelity.

Cons:

- complexity and dependency overhead too high for this ticket.

Status: future work, out of scope.

## Implementation Plan

### Phase 1: Baseline JS help-bar provider

- Implement `GetHelpBar(ctx, req)` in `pkg/repl/evaluators/javascript/evaluator.go`.
- Add token-at-cursor helper (or reuse existing token extraction utility).
- Return basic symbol hints for exact matches from static maps.

Deliverable:

- JS REPL shows help-bar text for common typed symbols.

### Phase 2: Completion-context reuse

- Reuse `TSParser` + `jsparse.Analyze` context and candidates in help-bar path.
- Match by exact symbol first, prefix second.
- Reuse `requireAliasPattern` logic for module aliases.

Deliverable:

- Help bar aligns with completion candidate universe.

### Phase 3: Signature dictionary

- Add curated signature text map keyed by symbol (including module methods).
- Prefer signatures over generic candidate detail where available.

Deliverable:

- Higher-quality inline hints (e.g. `console.log(...args): void`).

### Phase 4: Safe runtime fallback

- Add optional runtime metadata helper for globals:
- value class and function arity/name only,
- no code execution.

Deliverable:

- improved hints for dynamic symbols introduced in session.

### Phase 5: Tests + example validation

- Add unit tests in JS evaluator package for:
- exact symbol help,
- prefix behavior,
- alias/module mapping,
- hide policy.
- Add integration-style REPL tests where useful.
- Update `examples/js-repl/main.go` placeholder/help copy.

Deliverable:

- reproducible quality checks and clear manual test flow.

## Testing Plan

### Automated

- `go test ./pkg/repl/evaluators/javascript -count=1`
- `go test ./pkg/repl/... -count=1`
- `golangci-lint run -v --max-same-issues=100`

### Manual

- `go run ./examples/js-repl`
- Type `con`, `console.lo`, `Math.ma`, and module flows (`const fs = require('fs'); fs.re`).
- Confirm help bar shows stable and relevant text under input.

## Dataflow Sketch

```mermaid
flowchart TD
  A[Typing in JS REPL] --> B[Debounced help request from repl.Model]
  B --> C[Evaluator.GetHelpBar(req)]
  C --> D[Token and cursor extraction]
  D --> E[jsparse context + candidate resolution]
  E --> F{Signature catalog hit?}
  F -- yes --> G[Return signature payload]
  F -- no --> H[Safe runtime metadata fallback]
  H --> I{Useful hint?}
  I -- yes --> G
  I -- no --> J[Show=false]
  G --> K[Help bar rendered in REPL view]
```

## Open Questions

1. Should signature metadata for native modules live in `bobatea` evaluator code, or be exported from go-go-goja module metadata?
2. Do we want call-argument-position hints in help bar now, or reserve for help drawer work?
3. Should help-bar analysis cache by `(input, cursor)` to reduce repeated parse costs under rapid typing?

## References

- `pkg/repl/help_bar_types.go`
- `pkg/repl/model.go`
- `pkg/repl/evaluators/javascript/evaluator.go`
- `examples/js-repl/main.go`
- `/home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/go-go-goja/pkg/jsparse/completion.go`
- `/home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/go-go-goja/pkg/jsparse/analyze.go`
