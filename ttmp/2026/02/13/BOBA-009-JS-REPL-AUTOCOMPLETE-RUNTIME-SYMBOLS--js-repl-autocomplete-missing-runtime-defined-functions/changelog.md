# Changelog

## 2026-02-13

- Initial workspace created
- Added detailed bug report and fix plan in `design-doc/01-runtime-symbol-autocomplete-bug-report-and-fix-plan.md`.
- Added manual reproduction probe script `scripts/probe_runtime_defined_completion.go` and captured current behavior (runtime symbols missing from autocomplete while help bar sees them).
- Expanded actionable task list for implementation and validation.

## 2026-02-14

- Implemented BOBA-009 runtime-aware JS autocomplete end-to-end.
- Moved JS-specific completion lookup and merge logic into `go-go-goja/pkg/jsparse`:
  - require alias extraction
  - known module candidates
  - runtime identifier/property lookup
  - static/module/runtime candidate merge with deterministic ranking and dedupe
  - reusable top-level binding extraction for REPL lexical declarations.
- Updated `bobatea` JS evaluator to use the new `jsparse` APIs and maintain runtime declaration hints across evaluations.
- Added regression tests covering runtime-defined function, runtime-defined const identifier completion, and runtime object property completion.
- Validated with:
  - `go test ./pkg/jsparse/... -count=1` (go-go-goja)
  - `go test ./pkg/repl/evaluators/javascript -count=1`
  - `go test ./pkg/repl/... -count=1`
  - `golangci-lint run -v --max-same-issues=100 ./pkg/repl/...`
  - manual probe script `scripts/probe_runtime_defined_completion.go` (now shows `greetUser`, `dataBucket`, and `dataBucket.` keys).
