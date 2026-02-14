# Tasks

## TODO

- [x] Add runtime-aware identifier completion for JS REPL globals defined during prior evaluations
- [x] Add runtime-aware property completion for runtime objects (initially plain identifier base expressions)
- [x] Merge static jsparse candidates with runtime candidates with deterministic ranking and dedupe
- [x] Add evaluator regression tests for runtime-defined function and object property completion
- [x] Run manual probe script and `go test ./pkg/repl/evaluators/javascript -count=1` after implementation
