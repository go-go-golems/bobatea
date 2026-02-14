# Tasks

## DONE

- [x] Create detailed analysis and implementation guide for REPL command palette integration

## TODO

- [x] Update implementation guide to match BOBA-008 split architecture (`model.go` orchestration + feature files) and layer/z-order policy
- [x] Extend `repl.Config` with `CommandPaletteConfig` + slash policy enum + normalization defaults
- [x] Add REPL command descriptor/registry contracts and evaluator command provider hook
- [ ] Add command-palette state wiring in `repl.Model` and initialize `commandpalette.Model` in `NewModel`
- [ ] Implement keyboard open/close routing and command dispatch in input mode
- [ ] Implement conservative slash-open behavior with guard rails (default: empty-input policy)
- [ ] Render command palette as top lipgloss v2 layer over existing overlays
- [ ] Add tests for config normalization, key routing precedence, slash policy, and command execution behavior
- [ ] Run validation: `go test ./pkg/repl/... -count=1`
- [ ] Run validation: `golangci-lint run -v --max-same-issues=100 ./pkg/repl/...`
- [ ] Smoke test examples: `go run ./examples/repl/autocomplete-generic` and `go run ./examples/js-repl` (PTY-wrapped in non-TTY environments)
- [ ] Update BOBA-005 changelog and diary with task-by-task commits, failures, and validation output
