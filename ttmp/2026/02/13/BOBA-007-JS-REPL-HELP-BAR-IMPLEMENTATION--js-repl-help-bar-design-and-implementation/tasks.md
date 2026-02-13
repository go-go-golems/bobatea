# Tasks

## DONE

- [x] Create implementation analysis and plan for JS REPL help bar

## TODO

### Phase 1: Provider Surface

- [ ] Implement `GetHelpBar(ctx, req)` in `pkg/repl/evaluators/javascript/evaluator.go`
- [ ] Add token/cursor extraction utility for help-bar lookup path
- [ ] Keep provider-owned show/hide policy (no trigger heuristics in REPL model)

### Phase 2: Metadata Resolution

- [ ] Reuse existing `jsparse` completion context and candidates in help-bar logic
- [ ] Add JS signature catalog for built-ins and known module aliases
- [ ] Reuse alias extraction (`requireAliasPattern`) for module-aware signatures

### Phase 3: Runtime Fallback

- [ ] Add side-effect-free runtime metadata fallback (class + arity/name only)
- [ ] Ensure no arbitrary expression evaluation during help lookup

### Phase 4: Validation

- [ ] Add evaluator tests for exact/prefix/module help outcomes
- [ ] Add integration checks in REPL package where appropriate
- [ ] Validate with `go test ./pkg/repl/evaluators/javascript -count=1`
- [ ] Validate with `go test ./pkg/repl/... -count=1`
- [ ] Validate with `golangci-lint run -v --max-same-issues=100`

### Phase 5: Example + Docs

- [ ] Update `examples/js-repl/main.go` copy and behavior expectations
- [ ] Add manual test playbook in ticket docs
- [ ] Update changelog and diary entries with implementation commits
