# Tasks

## DONE

- [x] Create detailed implementation guide for REPL autocomplete with debounce scheduling and optional shortcut trigger mode (`Tab` supported via config)

## TODO

- [ ] Implement completion request/result contracts and optional InputCompleter interface in pkg/repl
- [ ] Add AutocompleteConfig defaults (debounce, timeout, trigger keys, focus toggle key)
- [ ] Extend repl.Model with autocomplete state, request sequencing, and internal messages
- [ ] Wire debounced scheduling on input edits without REPL-side trigger heuristics
- [ ] Wire explicit shortcut trigger requests (including optional tab mode)
- [ ] Implement suggestion popup rendering, navigation, and apply behavior
- [ ] Resolve tab vs focus-toggle key conflict via config-driven keymap
- [ ] Add tests for debounce coalescing, stale-result drop, shortcut reason tagging, and key routing
