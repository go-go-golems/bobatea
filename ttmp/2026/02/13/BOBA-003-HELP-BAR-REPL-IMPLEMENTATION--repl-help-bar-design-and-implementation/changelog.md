# Changelog

## 2026-02-13

- Initial workspace created


## 2026-02-13

Implemented BOBA-003 help bar end-to-end in REPL code: provider contracts, config defaults, model scheduling/state, async request handling with stale-drop and timeout, rendering, and tests. Also added `ttmp/go.mod` to isolate ticket scripts from root lint/typecheck traversal.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/help_bar_types.go — New help bar provider/request/payload contracts
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/config.go — Added `HelpBarConfig` and defaults
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Help bar model integration, scheduling, async handling, and rendering
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/help_bar_model_test.go — Dedicated help bar behavior tests
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/repl_test.go — Config default assertions for help bar
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/go.mod — Nested module marker to prevent root typecheck/lint collisions from ticket scripts

### Validation

- `go test ./pkg/repl/... -count=1`
- `golangci-lint run -v --max-same-issues=100`

### Commit

- `76feb91` — `repl: implement BOBA-003 typing-triggered help bar`


## 2026-02-13

Prepared BOBA-003 for implementation after BOBA-002 closeout: added a dedicated diary document, expanded tasks into phased execution (contracts/config, model integration, rendering, tests, ticket hygiene), and updated design guidance with BOBA-002 lessons on stale-drop, debounce flicker avoidance, and config normalization.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/reference/01-diary.md — New implementation diary scaffold
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/tasks.md — Phased task plan updated
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/design-doc/01-help-bar-analysis-and-implementation-guide.md — Added implementation lessons and constraints from BOBA-002
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/index.md — Linked diary in ticket overview


## 2026-02-13

Added a detailed analysis and implementation guide for a typing-triggered REPL help bar with debounced scheduling, provider-owned trigger semantics, async stale-result protection, and compatibility with replacing the autocomplete widget.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/design-doc/01-help-bar-analysis-and-implementation-guide.md — Primary analysis and implementation document
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/index.md — Ticket overview updated with guide linkage
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/tasks.md — Execution checklist seeded
