# Tasks

## TODO


- [x] Create pkg/tui/widgets/contextbar package and move help bar request/payload types.
- [x] Implement context bar widget state (req sequence, debounce, timeout, visibility, payload).
- [x] Implement buffer-change debounce scheduling with unchanged snapshot short-circuit.
- [x] Implement result handling for stale drop, hide-on-error, hide-on-empty, show-on-valid payload.
- [x] Expose host-facing visibility-change signal for relayout decisions.
- [x] Expose style-driven render function for severity-specific rendering.
- [x] Port and adapt help bar tests from repl package to widget package.
- [x] Add tests for pending debounce behavior while bar remains visible.
- [x] Integrate contextbar widget into repl.Model through adapter methods.
- [x] Remove old direct help bar internals from repl model where replaced.
- [x] Run go test ./pkg/repl/... and ensure js help bar integration tests remain green.
- [x] Run examples/js-repl manually and verify inline help bar behavior is unchanged.
