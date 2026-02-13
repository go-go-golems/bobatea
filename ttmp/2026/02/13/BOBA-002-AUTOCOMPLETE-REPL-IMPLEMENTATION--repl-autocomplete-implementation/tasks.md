# Tasks

## DONE

- [x] Create detailed implementation guide for REPL autocomplete with debounce scheduling and optional shortcut trigger mode (`Tab` supported via config)

## TODO

### Phase 1: Generic Autocomplete Mechanism (No JS)

- [x] Decide implementation path for autocomplete UI: refactor existing `pkg/autocomplete` or rewrite from scratch for fresh cutover (no backward compatibility constraints)
- [x] Define final generic contracts in `pkg/repl` for request/response (`CompletionRequest`, `CompletionResult`, `InputCompleter`)
- [x] Implement generic REPL-side debounce scheduling and request ID stale-result filtering
- [x] Implement generic shortcut-trigger path (`Tab` or configurable key) without REPL trigger heuristics
- [x] Implement generic suggestion popup rendering, selection navigation, and apply/replace-range behavior
- [x] Resolve `tab` conflict with timeline focus toggle by introducing explicit configurable focus key
- [x] Add config defaults and docs for debounce, timeout, trigger keys, accept keys, max suggestions, and focus toggle
- [x] Add unit tests for generic mechanism: debounce coalescing, stale drop, shortcut reason tagging, key routing, and apply behavior
- [x] Add integration-style REPL model test that exercises typing -> completion request -> popup -> selection apply end-to-end with a fake completer

### Phase 2: Minimal Generic Example (No JS)

- [ ] Add a minimal non-JS example program using a tiny in-memory generic completer (for example static symbols + prefix filtering)
- [ ] Ensure example exercises both debounce trigger and explicit shortcut trigger paths
- [ ] Document exact run command(s) in a playbook section inside this ticket
- [ ] Add success criteria for the example: popup appears, navigation works, accept inserts suggestion at correct cursor position

### Phase 3: Manual Validation in tmux (Generic Example)

- [ ] Run the minimal generic example in `tmux`
- [ ] Execute a manual validation script/checklist (typing flow, debounce behavior, shortcut behavior, accept/cancel behavior, focus switching)
- [ ] Take screenshots for each key state transition in tmux (idle input, popup open, selection moved, accepted completion)
- [ ] Store screenshots in ticket workspace under `various/` with clear names and short notes
- [ ] Record validation findings and pass/fail outcomes in the ticket changelog

### Phase 4: JavaScript Integration with jsparse

- [ ] Implement JS completer using `go-go-goja/pkg/jsparse` completion primitives (cursor context + candidates)
- [ ] Wire JS evaluator to expose/use the new generic `InputCompleter` contract
- [ ] Ensure JS completer honors REPL request reasons (`debounce`, `shortcut`) and decides when to show suggestions
- [ ] Add/extend JS-focused tests for representative cases (`obj.`, module symbols, partial identifiers, incomplete input)
- [ ] Update JS REPL example to use jsparse-backed autocomplete end-to-end

### Phase 5: Manual Validation in tmux (JS Example)

- [ ] Run the JS example in `tmux` with jsparse autocomplete enabled
- [ ] Validate completion behavior for representative JS contexts (property access, identifiers, module requires)
- [ ] Validate explicit shortcut trigger (`Tab`) and conflict-free focus behavior
- [ ] Take and store tmux screenshots for JS flows (before trigger, popup candidates, accept result, no-suggestion case)
- [ ] Record JS validation notes, edge cases, and final readiness status in ticket changelog
- [x] Adopt idiomatic bobatea REPL key bindings (key.Binding), bubbles help model rendering, and lipgloss-composed REPL layout
