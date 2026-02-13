# Tasks

## DONE

- [x] Create detailed analysis and implementation guide for keyboard-toggle REPL help drawer with adaptive typing updates

## TODO

- [x] Add `HelpDrawerProvider` contracts and request/result types in `pkg/repl`
- [x] Extend `repl.Config` with `HelpDrawerConfig` defaults and keymap
- [x] Wire help drawer state/messages and toggle handling into `repl.Model`
- [x] Add debounced adaptive updates while drawer is visible
- [x] Implement drawer panel rendering (overlay-first, canvas-layer-ready)
- [x] Add tests for toggle, adaptive updates, and stale response filtering
