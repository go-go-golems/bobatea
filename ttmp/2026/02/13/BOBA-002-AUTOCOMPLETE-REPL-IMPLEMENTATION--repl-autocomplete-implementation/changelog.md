# Changelog

## 2026-02-13

- Initial workspace created


## 2026-02-13

Overlay direction follow-up completed (commit `2b04556`): added explicit vertical placement controls (`auto`, `above`, `below`, `bottom`) and horizontal growth direction (`right`, `left`) so completion popup placement can be tuned without code edits.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/config.go — Added `OverlayPlacement` and `OverlayHorizontalGrow` config enums and defaults
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Added vertical placement strategy selection and left/right horizontal growth logic
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/autocomplete_model_test.go — Added bottom placement and left growth layout tests
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/repl_test.go — Added default assertions for placement/growth config


## 2026-02-13

Post-validation UX follow-up completed (commit `23095dc`): fixed debounce-time popup flicker by keeping the completion overlay visible while new debounced requests are pending, and added overlay positioning/appearance controls for minimal popup layouts.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Removed hide-on-input-change behavior, added popup style override path, and applied overlay X/Y offsets in layout computation
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/config.go — Added `OverlayOffsetX`, `OverlayOffsetY`, and `OverlayNoBorder` autocomplete config options
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/autocomplete_model_test.go — Added tests for no-flicker debounce behavior, offset placement, and borderless frame sizing
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/repl_test.go — Added default config assertions for new overlay controls


## 2026-02-13

Created a detailed autocomplete implementation guide focused on REPL-side debounce scheduling, completer-owned trigger decisions, stale-result handling, and optional keyboard shortcut trigger mode (including Tab) with key-conflict mitigation.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/design-doc/01-autocomplete-implementation-guide.md — Primary implementation guide document
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/index.md — Ticket overview updated to point at guide and constraints
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Execution checklist populated and template placeholder removed


## 2026-02-13

Expanded tasks into a phased execution checklist: generic mechanism first, minimal non-JS example, tmux+screenshot validation, JS jsparse integration, and tmux+screenshot validation; explicitly allows fresh-cutover rewrite of autocomplete widget without backward-compatibility constraints.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Detailed phased task plan for implementation and validation


## 2026-02-13

Task 2 completed: selected fresh-cutover autocomplete integration path (no backward compatibility constraints) and initialized detailed implementation diary for task-by-task execution.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/design-doc/01-autocomplete-implementation-guide.md — Added explicit implementation-path decision section
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Initialized detailed diary with Step 1
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Checked Task 2 complete


## 2026-02-13

Task 3 completed: added generic autocomplete contracts in pkg/repl (CompletionReason, CompletionRequest, CompletionResult, InputCompleter) and validated via go test ./pkg/repl/...

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/autocomplete_types.go — New generic contract definitions
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Step 2 execution notes and test log
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Task 3 checked


## 2026-02-13

Task 4 completed: implemented REPL-side debounced completion scheduling and stale-response filtering in repl.Model; validated with gofmt and go test ./pkg/repl/...

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Added completion debounce messages
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 3 execution notes
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Task 4 checked


## 2026-02-13

Task 5 completed: added explicit shortcut-trigger completion path in repl.Model (Reason=shortcut, Shortcut metadata) without REPL-side trigger heuristics; validated with gofmt and go test ./pkg/repl/...

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Added shortcut trigger handling and immediate request dispatch
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 4 execution notes
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Task 5 checked


## 2026-02-13

Task 6 completed: implemented completion popup rendering, keyboard navigation, and replace-range apply behavior in repl.Model; validated with gofmt and go test ./pkg/repl/...

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Added popup state/rendering/navigation/apply logic
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 5 execution notes
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Task 6 checked


## 2026-02-13

Task 7 completed: resolved tab/focus conflict by introducing configurable focus toggle key and model defaults (ctrl+t with completer, tab otherwise); validated with gofmt and go test ./pkg/repl/...

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/config.go — Added FocusToggleKey config field
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Switched focus routing to dynamic focusToggleKey
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 6 execution notes
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Task 7 checked


## 2026-02-13

Task 8 completed and idiomatic REPL integration added (commit `d2056ba`): migrated REPL key handling to `key.Binding`/`key.Matches`, integrated `bubbles/help` model with mode-aware bindings, switched View composition to lipgloss section joins, added dedicated completion popup styles, and excluded `ttmp/` from golangci-lint path scanning.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/keymap.go — New mode-tagged REPL key map with ShortHelp/FullHelp
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Key routing refactor, help model integration, lipgloss-composed view
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/styles.go — Completion popup and selected-row styles
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/config.go — Autocomplete config field docs
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/repl_test.go — Default autocomplete config assertions
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/.golangci.yml — Exclude `ttmp/` for lint
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/design-doc/01-autocomplete-implementation-guide.md — Added idiomatic-integration section
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 7 diary entry
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Checked Task 8 and recorded idiomatic integration task


## 2026-02-13

Task 9 completed (commit `650ea1e`): added focused unit tests for debounce coalescing, stale completion result drop, shortcut trigger reason tagging, and popup key-routing/apply behavior.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/autocomplete_model_test.go — New Task 9 unit test suite
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Task 9 checked
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 8 diary entry


## 2026-02-13

Task 10 completed (commit `49c6a9a`): added an integration-style autocomplete model test that drives `Update` from typing through debounce/result handling, popup navigation, and completion apply.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/autocomplete_model_test.go — Added end-to-end model flow test and command-drain helper
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Task 10 checked
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 9 diary entry


## 2026-02-13

Phase 2 Tasks 11-14 completed (code commit `d568e59`): added a minimal non-JS autocomplete REPL demo, ensured both debounce and shortcut trigger paths are represented, documented exact run commands, and defined explicit success criteria in a dedicated playbook.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/examples/repl/autocomplete-generic/main.go — Minimal generic autocomplete demo evaluator/completer
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/02-generic-example-playbook.md — Run commands and success criteria
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Tasks 11-14 checked
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 10 diary entry


## 2026-02-13

Phase 3 Tasks 15-19 completed (code commit `f59efb7`): ran generic example in tmux, executed manual validation checklist, captured required state snapshots into `various/`, and recorded pass/fail findings.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/examples/repl/autocomplete-generic/main.go — Added `BOBATEA_NO_ALT_SCREEN=1` mode for deterministic tmux pane capture
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/generic-01-idle.txt — Idle capture
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/generic-02-popup-open.txt — Popup open capture
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/generic-03-selection-moved.txt — Selection moved capture
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/generic-04-accepted.txt — Accepted completion capture
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/generic-05-focus-timeline.txt — Focus switched capture
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/generic-phase3-validation.md — Checklist findings and verdict
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Tasks 15-19 checked
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 11 diary entry


## 2026-02-13

Phase 4 Tasks 20-24 completed (commit `cabcadf`): implemented jsparse-backed JS completer on evaluator, added module-aware candidate enrichment, extended evaluator tests for representative contexts, and rewired `examples/js-repl` to use the package evaluator path.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/evaluators/javascript/evaluator.go — `CompleteInput` implementation and helpers
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/evaluators/javascript/evaluator_test.go — JS completer context tests
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/examples/js-repl/main.go — Updated JS example wiring for autocomplete
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Tasks 20-24 checked
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 12 diary entry


## 2026-02-13

Phase 5 Tasks 25-29 completed: ran JS example in tmux, validated property/module/shortcut/no-suggestion/focus behavior, and stored capture artifacts plus a validation report under `various/`.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/js-01-idle.txt — Idle JS state
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/js-02-property-popup.txt — Property completion popup
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/js-03-accept-result.txt — Accepted property completion
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/js-04-module-popup.txt — Module symbol completion popup
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/js-05-no-suggestion.txt — No suggestion case
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/js-06-focus-timeline.txt — Focus-toggle capture
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/various/js-phase5-validation.md — Validation checklist and verdict
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/tasks.md — Tasks 25-29 checked
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/reference/01-diary.md — Added Step 13 diary entry
