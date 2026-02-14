# Tasks

## TODO


- [x] Create pkg/tui/asyncprovider package and implement generic Run[T] utility.
- [x] Preserve timeout context wrapping semantics and request-id-aware panic logging.
- [x] Preserve panic recovery behavior returning typed zero value and wrapped error.
- [x] Write tests for success path with returned value and nil error.
- [x] Write tests for provider panic path including non-nil error.
- [x] Write tests for timeout path and context cancellation propagation.
- [x] Update repl model async command wrappers to call shared asyncprovider.Run.
- [x] Ensure provider names and panic prefixes are unchanged at call sites.
- [x] Run go test ./pkg/repl/... and verify no behavior regressions.
- [x] Add package-level docs clarifying intended widget-host usage.
