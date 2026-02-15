# Tasks

## TODO


- [x] Create pkg/tui/widgets/contextpanel package and move help drawer request/document types.
- [x] Implement widget state for visible/loading/error/pinned/request sequence and config.
- [x] Implement debounce-on-typing scheduling with pinned and prefetch rules preserved.
- [x] Implement keyboard handling API for toggle, close, refresh, and pin.
- [x] Implement request-now path for manual refresh and toggle-open triggers.
- [x] Implement stale result drop and document/error update handling.
- [x] Port overlay layout logic for above/right/left/bottom dock positions with clamp behavior.
- [x] Port panel rendering logic with title/subtitle/body/diagnostics/version/footer composition.
- [x] Add unit tests for toggle, close, pin, debounce coalescing, stale drop, and refresh behavior.
- [x] Add overlay tests for dock positions and no-cutoff guarantees.
- [x] Integrate widget via repl adapter preserving completion-cancel conflict behavior.
- [x] Run go test ./pkg/repl/... and manual checks for drawer docking/pinning in example apps.
