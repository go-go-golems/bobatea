# Tasks

## TODO


- [x] Create pkg/tui/inputhistory package and move History/Entry structures.
- [x] Preserve Add behavior including adjacent input dedupe and max size trimming.
- [x] Preserve NavigateUp/NavigateDown semantics including reset-to-empty behavior.
- [x] Preserve Clear and ResetNavigation behavior.
- [x] Preserve immutable getter behavior for entries and input history slices.
- [x] Port history-focused tests into new package with parity assertions.
- [x] Update repl model and tests to import and use pkg/tui/inputhistory.
- [x] Remove old repl/history.go or convert to thin type alias shim if needed during migration.
- [x] Run go test ./pkg/repl/... and verify history navigation behavior manually.
- [x] Add package README comments and examples for non-REPL usage.
