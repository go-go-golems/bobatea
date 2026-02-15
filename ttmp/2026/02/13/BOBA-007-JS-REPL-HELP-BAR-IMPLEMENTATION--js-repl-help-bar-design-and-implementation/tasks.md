# Tasks

## DONE

- [x] Create implementation analysis and plan for JS REPL help bar

## TODO

### Phase 1: Provider Surface

- [x] Implement `GetHelpBar(ctx, req)` in `pkg/repl/evaluators/javascript/evaluator.go`
- [x] Add token/cursor extraction utility for help-bar lookup path
- [x] Keep provider-owned show/hide policy (no trigger heuristics in REPL model)

### Phase 2: Metadata Resolution

- [x] Reuse existing `jsparse` completion context and candidates in help-bar logic
- [x] Add JS signature catalog for built-ins and known module aliases
- [x] Reuse alias extraction (`requireAliasPattern`) for module-aware signatures

### Phase 3: Runtime Fallback

- [x] Add side-effect-free runtime metadata fallback (class + arity/name only)
- [x] Ensure no arbitrary expression evaluation during help lookup

### Phase 4: Validation

- [x] Add evaluator tests for exact/prefix/module help outcomes
- [x] Add integration checks in REPL package where appropriate
- [x] Validate with `go test ./pkg/repl/evaluators/javascript -count=1`
- [x] Validate with `go test ./pkg/repl/... -count=1`
- [x] Validate with `golangci-lint run -v --max-same-issues=100`

### Phase 5: Example + Docs

- [x] Update `examples/js-repl/main.go` copy and behavior expectations
- [x] Add manual test playbook in ticket docs
- [x] Update changelog and diary entries with implementation commits
