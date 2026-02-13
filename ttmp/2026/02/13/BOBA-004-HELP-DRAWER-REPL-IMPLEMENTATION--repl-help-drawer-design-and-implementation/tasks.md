# Tasks

## DONE

- [x] Create detailed analysis and implementation guide for keyboard-toggle REPL help drawer with adaptive typing updates

## TODO

- [ ] Add `HelpDrawerProvider` contracts and request/result types in `pkg/repl`
- [ ] Extend `repl.Config` with `HelpDrawerConfig` defaults and keymap
- [ ] Wire help drawer state/messages and toggle handling into `repl.Model`
- [ ] Add debounced adaptive updates while drawer is visible
- [ ] Implement drawer panel rendering (overlay-first, canvas-layer-ready)
- [ ] Add tests for toggle, adaptive updates, and stale response filtering
