# Changelog

## 2026-02-14

- Initial workspace created


## 2026-02-14

Implemented async provider extraction: added pkg/tui/asyncprovider, migrated repl wrappers, and added dedicated tests.

### Related Files

- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model_async_provider.go — Repl now calls shared utility
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/tui/asyncprovider/run.go — New shared timeout/panic wrapper
- /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/tui/asyncprovider/run_test.go — Coverage for success

