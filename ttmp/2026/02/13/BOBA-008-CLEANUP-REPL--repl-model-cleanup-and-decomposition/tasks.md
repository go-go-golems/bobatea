# Tasks

## TODO

- [x] Write and publish BOBA-008 decomposition analysis document
- [x] Relate key REPL files to BOBA-008 analysis doc (`model.go`, `config.go`, `keymap.go`, help type contracts)

## Big-Bang Rewrite Scope

- [ ] Create the new internal file layout in one refactor branch (`model_messages.go`, `model_input.go`, `model_timeline_bus.go`, `completion_model.go`, `completion_overlay.go`, `helpbar_model.go`, `helpdrawer_model.go`, `helpdrawer_overlay.go`, `config_normalize.go`, `model_async_provider.go`)
- [ ] Introduce internal `completionModel`, `helpBarModel`, and `helpDrawerModel` structs and migrate feature state from root `Model`
- [ ] Refactor root `Model` to orchestration-only responsibilities (`Update` routing + `View` layer composition)
- [ ] Replace duplicated provider panic/timeout boilerplate with a shared async helper
- [ ] Remove old transitional paths and dead helpers from `model.go`

## Cutover Validation

- [ ] Update existing `pkg/repl` tests to reflect the new internal structure and keep behavior coverage intact
- [ ] Add regression tests for stale-drop ordering, debounce sequencing, and completion/drawer overlay bounds
- [ ] Run validation: `go test ./pkg/repl/... -count=1`
- [ ] Run validation: `golangci-lint run -v --max-same-issues=100 ./pkg/repl/...`
- [ ] Smoke test examples: `go run ./examples/repl/autocomplete-generic` and `go run ./examples/js-repl`

## Ticket Hygiene

- [ ] Update BOBA-008 changelog with rewrite summary and commit hashes
- [ ] Update BOBA-008 diary/analysis notes with migration details and known follow-ups
- [x] Add optional constructor/helper to inject external base context into repl.Model
- [x] Switch completion/helpBar/helpDrawer provider command contexts from Background to app context + timeout
- [x] Run validation (go test ./pkg/repl/... and golangci-lint ./pkg/repl/...) and commit task-by-task
- [x] Add model app context fields/cancel lifecycle and cancel on quit path
- [x] Add tests for app-context cancellation propagation to provider commands
