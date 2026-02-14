# Tasks

## TODO

- [ ] Add runtime-aware identifier completion for JS REPL globals defined during prior evaluations
- [ ] Add runtime-aware property completion for runtime objects (initially plain identifier base expressions)
- [ ] Merge static jsparse candidates with runtime candidates with deterministic ranking and dedupe
- [ ] Add evaluator regression tests for runtime-defined function and object property completion
- [ ] Run manual probe script and `go test ./pkg/repl/evaluators/javascript -count=1` after implementation
