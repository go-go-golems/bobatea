---
Title: Runtime Symbol Autocomplete Bug Report and Fix Plan
Ticket: BOBA-009-JS-REPL-AUTOCOMPLETE-RUNTIME-SYMBOLS
Status: active
Topics:
    - repl
    - autocomplete
    - javascript
    - bug
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/js-repl/main.go
      Note: User-facing JS REPL entrypoint for manual validation
    - Path: pkg/repl/evaluators/javascript/evaluator.go
      Note: Autocomplete implementation path and root-cause location
    - Path: pkg/repl/evaluators/javascript/evaluator_test.go
      Note: Current coverage and missing runtime completion regression tests
    - Path: pkg/repl/model.go
      Note: Request lifecycle and completion command flow into evaluator
    - Path: ttmp/2026/02/13/BOBA-009-JS-REPL-AUTOCOMPLETE-RUNTIME-SYMBOLS--js-repl-autocomplete-missing-runtime-defined-functions/scripts/probe_runtime_defined_completion.go
      Note: Ticket-local reproduction probe
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-13T19:06:55.779973497-05:00
WhatFor: ""
WhenToUse: ""
---


# Runtime Symbol Autocomplete Bug Report and Fix Plan

## Executive Summary

The JS REPL currently executes user-defined symbols into the live Goja runtime, but autocomplete does not include those symbols in suggestion results. The core issue is architectural: completion is computed from the *current input snapshot only* (`jsparse.Analyze` + `ResolveCandidates`) and does not consult runtime state.

This is why:

- `function greetUser(name) { ... }` is callable immediately in the REPL.
- Help bar can show runtime info for `greetUser` via `runtimeHelpForIdentifier`.
- Autocomplete still returns no `greetUser` suggestion for `gre`.

The fix should be a hybrid completion pipeline:

- Keep static `jsparse` candidates (fast and good at local syntax context).
- Merge runtime-derived candidates for identifier/property contexts.
- Rank static matches first, then runtime matches, dedupe by label/value.

## Problem Statement

User-facing bug:

- After defining a function or variable in JS REPL, typing its prefix does not surface it in autocomplete.

Concrete repro (captured with ticket probe script):

```bash
go run ./ttmp/2026/02/13/BOBA-009-JS-REPL-AUTOCOMPLETE-RUNTIME-SYMBOLS--js-repl-autocomplete-missing-runtime-defined-functions/scripts/probe_runtime_defined_completion.go
```

Observed output:

```text
INPUT="gre" show=false suggestions=0
  has greetUser=false, has dataBucket=false
INPUT="dataB" show=false suggestions=0
  has greetUser=false, has dataBucket=false
INPUT="dataBucket." show=true suggestions=3
  has greetUser=false, has dataBucket=false
  first 3: hasOwnProperty, toString, valueOf
HELPBAR greetUser show=true kind="runtime" text="greetUser(...): function (arity 1)"
```

Expected behavior:

- `gre` should show `greetUser`.
- `dataB` should show `dataBucket`.
- `dataBucket.` should include object members from runtime value (at least own keys, optionally proto keys).

Impact:

- Primary REPL ergonomics regression for iterative workflows.
- Users perceive autocomplete as stale or broken after they define symbols.
- Help bar and autocomplete appear inconsistent, which is confusing.

## Proposed Solution

Implement **hybrid runtime-aware autocomplete** in JS evaluator:

1. Keep current static path:
   - Parse input via `tsParser.Parse`.
   - Analyze context via `jsparse.Analyze(...).CompletionContextAt(...)`.
   - Resolve static candidates via `jsparse.ResolveCandidates(...)`.
2. Add runtime augmentation:
   - `CompletionIdentifier`: inspect `runtime.GlobalObject()` names and filter by prefix.
   - `CompletionProperty`: if `ctx.BaseExpr` is a plain identifier, read runtime value and list keys from object.
3. Merge candidate sets:
   - Dedupe by `Label`.
   - Preserve source metadata (`detail`: `runtime global`, `runtime property`, etc.).
   - Stable sorting with static-first priority.
4. Keep existing debounce/shortcut trigger model unchanged.

> [!NOTE]
> Fundamental architecture model:
> Static parser knows current text, runtime knows historical execution state.
> The fix is not replacing static completion; it is composing both worlds.

### Current Flow (Today)

```text
Input changed -> REPL Model completionCmd -> Evaluator.CompleteInput(req)
    -> parse current req.Input
    -> jsparse.Analyze(current req.Input, nil)
    -> ResolveCandidates(...)
    -> return suggestions
```

### Desired Flow (After Fix)

```text
Input changed -> REPL Model completionCmd -> Evaluator.CompleteInput(req)
    -> staticCandidates := ResolveCandidates(...)
    -> runtimeCandidates := collectRuntimeCandidates(ctx, req.Input, req.CursorByte)
    -> merged := mergeCandidates(staticCandidates, runtimeCandidates)
    -> return suggestions
```

### Why this shape is best

- Minimal blast radius: only JS evaluator completion path changes.
- Preserves existing overlays/key handling/debounce in `pkg/repl/model.go`.
- Uses already-existing runtime (`e.runtime`) and existing runtime-help precedent.
- Supports future help drawer/help bar reuse of runtime symbol inventory.

## Design Decisions

1. **Hybrid static + runtime candidate model**
   - Rationale: static-only misses executed symbols; runtime-only misses syntactic context and local lexical hints.

2. **Runtime augmentation only in identifier/property contexts**
   - Rationale: these are the contexts users expect symbol lookup from prior REPL state.
   - Skip argument-only/none contexts unless needed later.

3. **Safe runtime introspection**
   - Use `Runtime.GlobalObject()` and object key access with panic protection (`Runtime.Try` or `recover` around helper) because `Object.Keys()` can throw.

4. **No trigger-detection logic changes**
   - Rationale: matches current BOBA direction (trigger source is input completer wiring; evaluator focuses on resolution).

5. **Prefer plain identifier base expressions first**
   - For property completion, start with `ctx.BaseExpr` that maps directly to runtime symbol names.
   - Complex expressions (`foo().bar`) can remain static-only initially to reduce risk.

## Alternatives Considered

1. **Prepend REPL history into `req.Input` before static analysis**
   - Pros: maybe gives parser more symbols.
   - Cons: brittle with syntax errors/multiline history, does not reflect runtime mutations from native modules, memory growth, cursor offset complexity.
   - Decision: rejected.

2. **Maintain explicit symbol table during `Evaluate`**
   - Pros: deterministic and cheap lookup.
   - Cons: requires JS AST parse of evaluated snippets and misses dynamic writes (`globalThis[x] = ...`), aliasing, and module-driven mutations.
   - Decision: not first step; runtime inspection is higher fidelity.

3. **Use full LSP/TS language service**
   - Pros: rich typing and completions.
   - Cons: heavy integration, out of scope for this bug fix, non-trivial runtime bridging.
   - Decision: rejected for this ticket.

4. **Runtime-only completion**
   - Pros: includes dynamic state.
   - Cons: loses local syntax-sensitive candidate quality.
   - Decision: rejected.

## Implementation Plan

### Files and symbols to update

- `pkg/repl/evaluators/javascript/evaluator.go`
  - `CompleteInput(...)`
  - new helpers:
    - `runtimeIdentifierCandidates(partial string) []jsparse.CompletionCandidate`
    - `runtimePropertyCandidates(baseExpr, partial string) []jsparse.CompletionCandidate`
    - `mergeCompletionCandidates(static, runtime []jsparse.CompletionCandidate) []jsparse.CompletionCandidate`
    - optional safety helper for key extraction.
- `pkg/repl/evaluators/javascript/evaluator_test.go`
  - add regression tests for runtime-defined function/object visibility in completion.
- `ttmp/.../BOBA-009.../scripts/probe_runtime_defined_completion.go`
  - keep as manual verification probe.

### Pseudocode

```go
func (e *Evaluator) CompleteInput(ctx context.Context, req repl.CompletionRequest) (repl.CompletionResult, error) {
    // existing parse/analyze/context code...
    static := jsparse.ResolveCandidates(cctx, analysis.Index, root)
    if cctx.Kind == jsparse.CompletionProperty {
        static = append(static, moduleCandidatesFromRequireAliases(...)...)
    }

    runtime := []jsparse.CompletionCandidate{}
    switch cctx.Kind {
    case jsparse.CompletionIdentifier:
        runtime = e.runtimeIdentifierCandidates(cctx.PartialText)
    case jsparse.CompletionProperty:
        runtime = e.runtimePropertyCandidates(cctx.BaseExpr, cctx.PartialText)
    }

    candidates := mergeCompletionCandidates(static, runtime)
    // existing mapping to autocomplete.Suggestion...
}
```

### Acceptance criteria

- Defining `function greetUser(){}` then typing `gre` offers `greetUser`.
- Defining `const dataBucket = {count:1}` then typing `dataB` offers `dataBucket`.
- Typing `dataBucket.` includes runtime-derived keys (exact set documented by test).
- Existing completions (e.g., `console.lo`, `require("fs")` alias completions) remain intact.
- No panic regressions in concurrent completion test path.

### Test additions

- `TestEvaluator_CompleteInput_RuntimeDefinedIdentifier`
- `TestEvaluator_CompleteInput_RuntimeDefinedObjectProperty`
- `TestEvaluator_CompleteInput_MergesStaticAndRuntimeCandidates`
- Keep/extend concurrent completion stress test.

## Open Questions

1. Should runtime property completion include only own enumerable keys (`Keys`) or all own properties (`GetOwnPropertyNames`)?
2. Should prototype keys be included by default or only after no own-key matches?
3. How aggressively should we filter builtins from global object suggestions?
4. Do we want ranking boosts for recently-defined symbols?

## References

- `pkg/repl/evaluators/javascript/evaluator.go`
- `pkg/repl/evaluators/javascript/evaluator_test.go`
- `pkg/repl/model.go`
- `pkg/repl/autocomplete_types.go`
- `examples/js-repl/main.go`
- `ttmp/2026/02/13/BOBA-009-JS-REPL-AUTOCOMPLETE-RUNTIME-SYMBOLS--js-repl-autocomplete-missing-runtime-defined-functions/scripts/probe_runtime_defined_completion.go`
