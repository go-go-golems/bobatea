# Tasks

## DONE

- [x] Create detailed analysis and implementation guide for REPL command palette integration

## IMPLEMENTED (BOBA-008-aligned core)

- [x] Update implementation guide to match BOBA-008 split architecture (`model.go` orchestration + feature files) and layer/z-order policy
- [x] Extend `repl.Config` with `CommandPaletteConfig` + slash policy enum + normalization defaults
- [x] Add REPL command descriptor/registry contracts and evaluator command provider hook
- [x] Add command-palette state wiring in `repl.Model` and initialize `commandpalette.Model` in `NewModel`
- [x] Implement keyboard open/close routing and command dispatch in input mode
- [x] Implement conservative slash-open behavior with guard rails (default: empty-input policy)
- [x] Render command palette as top lipgloss v2 layer over existing overlays

## REMAINING (new design completion)

- [x] Move command palette overlay geometry/rendering out of `pkg/repl/model.go` into `pkg/repl/command_palette_overlay.go` so feature ownership matches BOBA-008 decomposition style
- [x] Add focused tests for command palette behavior:
  - config normalization and defaults
  - key routing precedence (palette open > drawer/completion navigation)
  - slash policy behavior (`empty-input`, `column-zero`, `provider`)
  - command execution dispatch and close semantics
- [x] Run validation: `go test ./pkg/repl/... -count=1`
- [x] Run validation: `golangci-lint run -v --max-same-issues=100 ./pkg/repl/...`
- [x] Smoke test examples with PTY wrapping:
  - `script -q -c "timeout 7s go run ./examples/repl/autocomplete-generic" /dev/null`
  - `script -q -c "timeout 7s go run ./examples/js-repl" /dev/null`
- [x] Update BOBA-005 changelog and diary with task-by-task commits, failures, and validation output
