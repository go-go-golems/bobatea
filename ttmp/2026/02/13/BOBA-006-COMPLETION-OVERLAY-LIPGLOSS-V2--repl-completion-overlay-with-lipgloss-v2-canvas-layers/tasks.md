# Tasks

## TODO

- [x] Add overlay config controls to `AutocompleteConfig` (max width/height, min width, margin, paging)
- [x] Add completion paging key bindings (page up/down) and help integration
- [x] Add REPL completion viewport state (scroll offset / visible rows) and selection-visibility helpers
- [x] Add lipgloss v2 dependency and implement overlay renderer in `Model.View` (non-inline popup)
- [x] Implement placement/sizing logic with terminal clamping and above/below fallback
- [x] Implement scrolling/paging behavior for long suggestion lists
- [x] Add/adjust tests for viewport and overlay behavior
- [x] Run focused validation (`go test`, lint on changed packages, manual `examples/js-repl` run)

## Done (Research)

- [x] Create BOBA-006 ticket workspace and scaffold docs
- [x] Analyze current `pkg/repl` inline completion rendering flow
- [x] Review grail-js lipgloss v2 layer/canvas usage and related ticket docs
- [x] Validate lipgloss v2 layer/compositor behavior with targeted probes
- [x] Write detailed BOBA-006 analysis/design document
- [x] Write detailed BOBA-006 research diary
- [x] Upload BOBA-006 bundle to reMarkable and verify cloud listing
