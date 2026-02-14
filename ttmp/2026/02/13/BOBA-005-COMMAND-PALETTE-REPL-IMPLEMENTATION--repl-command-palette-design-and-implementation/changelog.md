# Changelog

## 2026-02-13

- Initial workspace created


## 2026-02-13

Added a detailed analysis and implementation guide for REPL command palette integration covering keyboard and slash-trigger entry, command registry design, key-routing precedence, and compatibility with replacing autocomplete internals.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/design-doc/01-command-palette-analysis-and-implementation-guide.md — Primary analysis and implementation document
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/index.md — Ticket overview updated with guide linkage
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/tasks.md — Execution checklist seeded


## 2026-02-13

Updated BOBA-005 task plan and implementation guide to align with post-BOBA-008 REPL split architecture, including explicit file ownership, routing precedence, and lipgloss v2 overlay layering policy for command palette integration.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/design-doc/01-command-palette-analysis-and-implementation-guide.md — Updated architecture and implementation steps for BOBA-008-aligned design
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/reference/01-diary.md — Added diary step documenting planning update before coding
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/tasks.md — Refined execution checklist to task-by-task split-model implementation sequence


## 2026-02-13

Completed task 3: extended REPL config with command palette defaults, slash-policy enum, and normalization logic; added config tests for defaults and sanitization. (commit 98c9f37)

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/config.go — Added CommandPaletteConfig and slash policy defaults
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/config_normalize.go — Added normalizeCommandPaletteConfig and slash policy sanitization
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/repl_test.go — Added command palette config default and normalization tests


## 2026-02-13

Completed task 4: added command palette contracts and evaluator extension hook interfaces (PaletteCommand, PaletteCommandProvider, PaletteCommandRegistry). (commit a9d3f24)

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/command_palette_types.go — Introduced command descriptor and evaluator-provider contracts for command palette integration

## 2026-02-13

Completed task 5: added command palette feature state in repl.Model and initialized commandpalette.Model in NewModel with normalized config; wired palette sizing on window resize. (commit ddae48a)

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/command_palette_model.go — Added commandPaletteModel internal state container
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Integrated command palette state initialization and size updates


## 2026-02-13

Completed task 6: implemented keyboard open/close routing for command palette and action dispatch through selected palette commands, including built-in command set and evaluator-provided command merge. (commit a944fdc)

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/commandpalette/model.go — Added SetCommands for replacing command list on palette open
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/command_palette_model.go — Implemented input routing
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/keymap.go — Added command palette open/close key bindings and help entries
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model_input.go — Inserted command palette handling precedence before other input features


## 2026-02-13

Completed task 7: implemented guarded slash-open behavior for command palette with policies empty-input (default), column-zero, and provider delegation; slash key is consumed only when policy allows opening. (commit 4640706)

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/command_palette_model.go — Added slash key detection and policy guard rails
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/command_palette_types.go — Added slash policy provider request/provider contracts


## 2026-02-13

Completed task: moved command palette overlay rendering from `pkg/repl/model.go` into `pkg/repl/command_palette_overlay.go` while keeping lipgloss v2 layer composition in `View()` (`command-palette-overlay` at z=30). Ran `go test ./pkg/repl/... -count=1` successfully. (commit 7be8bcc)

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/command_palette_overlay.go — New command palette overlay renderer and centering placement helper
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go — Root view now calls `renderCommandPaletteOverlay()` and keeps compositor layering only
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/tasks.md — Marked overlay extraction task complete


## 2026-02-13

Completed task: added focused command palette tests for config normalization bounds, input routing precedence, slash policy variants, and command execution/close semantics. (commit 77faeed)

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/command_palette_model_test.go — New focused command palette test suite
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/ttmp/2026/02/13/BOBA-005-COMMAND-PALETTE-REPL-IMPLEMENTATION--repl-command-palette-design-and-implementation/tasks.md — Marked focused test task complete
