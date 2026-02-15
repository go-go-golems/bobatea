# Tasks

## TODO


- [x] Create pkg/tui/widgets/suggest package skeleton and move shared completion types.
- [x] Define widget config, keymap, buffer snapshot, and buffer mutator interfaces.
- [x] Port debounce scheduling logic from repl completion model into widget state machine.
- [x] Port completion result handling including stale request drop and visibility transitions.
- [x] Port selection navigation and paging behavior (prev/next/page up/page down).
- [x] Port completion apply/replacement logic preserving replace range semantics.
- [x] Port overlay layout geometry (placement, horizontal growth, clamp, offsets, margins).
- [x] Port overlay rendering with style injection and border/no-border support.
- [x] Write unit tests for request coalescing, stale drop, selection, paging, replacement, and hide conditions.
- [x] Add overlay geometry tests for auto/above/below/bottom and left/right growth.
- [x] Build REPL adapter layer mapping existing keymap and textinput to suggest widget interfaces.
- [x] Replace direct completion fields/calls in repl.Model with suggest widget usage.
- [x] Run go test ./pkg/repl/... and examples/repl/autocomplete-generic manual parity checks.
- [x] Run examples/js-repl manual completion parity checks (tab trigger and accept).
