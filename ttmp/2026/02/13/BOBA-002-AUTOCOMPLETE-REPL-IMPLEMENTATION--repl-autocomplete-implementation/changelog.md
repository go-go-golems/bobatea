# Changelog

## 2026-02-13

- Initial workspace created


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
