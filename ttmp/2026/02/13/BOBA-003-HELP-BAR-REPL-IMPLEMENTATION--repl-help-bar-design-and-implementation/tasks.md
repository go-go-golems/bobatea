# Tasks

## DONE

- [x] Create detailed analysis and implementation guide for typing-triggered REPL help bar

## TODO

### Phase 1: Contracts and Config (BOBA-002 lessons applied)

- [ ] Add `HelpBarProvider` contracts and request/result types in `pkg/repl` (`help_bar_types.go`)
- [ ] Keep provider trigger policy provider-owned; REPL only schedules requests (`reason`, `request_id`, `cursor`)
- [ ] Extend `repl.Config` with `HelpBarConfig` defaults (enabled, debounce, timeout)
- [ ] Add config normalization for help bar settings in `model.go` (same pattern as autocomplete)

### Phase 2: Model Integration and Scheduling

- [ ] Add help bar provider discovery in `NewModel` via optional capability interface
- [ ] Wire help bar state/messages into `repl.Model` (`visible`, `payload`, `error`, request sequence)
- [ ] Add debounced help bar scheduling on input/cursor changes
- [ ] Add stale-result filtering by request ID and timeout-bounded provider calls
- [ ] Ensure debounce scheduling does not introduce visual flicker in help-bar visibility transitions

### Phase 3: Rendering and UX

- [ ] Render help bar line in `View()` between input row and static key help
- [ ] Apply severity-aware style mapping (`info|warning|error`)
- [ ] Keep rendering compatible with existing completion overlay layering and focus help model behavior
- [ ] Preserve behavior when no provider is present (feature inert, no regression)

### Phase 4: Tests and Validation

- [ ] Add unit tests for request scheduling, stale-drop, timeout, and visibility transitions
- [ ] Add model tests for input mutation -> debounced request -> result render flow
- [ ] Add tests for severity style selection and no-provider behavior
- [ ] Run focused validation: `go test ./pkg/repl/... -count=1` and `golangci-lint run -v --max-same-issues=100 ./pkg/repl/...`

### Phase 5: Ticket Hygiene

- [ ] Maintain detailed diary entries for each implementation step
- [ ] Update changelog with commit hashes and related file links
- [ ] Check tasks as each phase/task is completed
