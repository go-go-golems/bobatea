# Refactoring the REPL Help System for Multi-Backend Titles and Markdown Attachments

Date: 2025-09-03
Author: manuel (@wesen)
Status: Proposal (Breaking Changes — no backward compatibility)

## 1) Purpose & Scope

We want to evolve the REPL help system to:
- Allow registering multiple help backends, each with its own title/identity, and present them nicely on the top-level help page and in lookups, rather than flattening via the current composite adapter.
- Allow users (and downstream apps) to attach pure Markdown help entries without needing Glazed frontmatter or a custom backend.

This document analyzes the current implementation and proposes a refactor with minimal disruption to existing code, focusing on `bobatea/pkg/repl/help` and its adapters. An example consumer exists in `oak/cmd/oak-repl/main.go`.

## 2) Current State (as of today)

Key files:
- `bobatea/pkg/repl/help/api.go` — Defines `Section`, `TopLevelPage`, and `Backend` interface (+ optional `RelatedBackend`).
- `bobatea/pkg/repl/help/handler.go` — `HandleHelpCommand(ctx, cfg, input)` orchestrates parsing, querying, and rendering.
- `bobatea/pkg/repl/help/render.go` — Default `Renderer` that outputs markdown for top-level, section, and query results.
- `bobatea/pkg/repl/help/parse.go` — Parser for `/help` inputs and DSL builder from flags.
- `bobatea/pkg/repl/help/adapters/glazed.go` — Adapter for Glazed `HelpSystem` to `Backend`.
- `bobatea/pkg/repl/help/adapters/slash.go` — Adapter to surface slash commands as help sections.
- `bobatea/pkg/repl/help/adapters/composite.go` — `MultiBackend` that flattens multiple backends.

Current usage examples:
- `oak/cmd/oak/commands/repl_help.go` builds a synthetic `/help` string and calls `HandleHelpCommand` with a single `GlazedBackend`.
- `oak/cmd/oak-repl/main.go` registers a `/help` slash command that constructs a `MultiBackend(GlazedBackend, SlashBackend)` and passes it into `HandleHelpCommand` for rendering.

Limitations today:
- `MultiBackend` flattens results; top-level page merges all sections without indicating their source. No way to label/group by backend identity.
- There is no first-class way to load arbitrary Markdown-only entries (without Glazed frontmatter) as help sections.

## 3) Design Goals

- Multiple backends can be registered with metadata (title, optional description, slug prefix) and be rendered in distinct groups on the top-level help page.
- Section lookup via slug should support disambiguation and optional namespacing (e.g., `glazed:<slug>`, `slash:<slug>`, `md:<slug>`). Defaults remain backward compatible when unambiguous.
- Querying can either broadcast to capable backends or allow scoping to a specific backend (e.g., flag `--backend=` or namespaced DSL fields like `backend:slash`).
- Provide an in-memory Markdown backend that supports adding plain markdown sections programmatically or by loading files.
- Maintain backward compatibility for existing callers using a single backend or the composite adapter. Offer a straightforward migration path to the new multi-backend manager.

## 4) Proposed API Changes

### 4.1 New BackendRegistration and Manager

Add a manager type that orchestrates multiple backends and exposes the existing `Backend` contract to `HandleHelpCommand`, but with richer rendering capabilities.

```go
// bobatea/pkg/repl/help/multi.go
package replhelp

type BackendRegistration struct {
    ID          string // stable id, e.g., "glazed", "slash", "md"
    Title       string // display title, e.g., "Project Docs (Glazed)", "Slash Commands"
    Description string // optional short description shown at top-level
    Backend     Backend
    SlugPrefix  string // optional, e.g., "slash-" for slash commands
}

type MultiManager struct {
    regs []BackendRegistration
}

func NewMultiManager(regs ...BackendRegistration) *MultiManager
```

Responsibilities:
- On `TopLevel`, fetch each backend’s page and render them as separate groups with the backend title/description.
- On `GetBySlug`, resolve slug in order:
  1) If input is `id:slug` (e.g., `slash:lang`), dispatch to that backend only.
  2) If input starts with a known `SlugPrefix`, normalize and dispatch to that backend.
  3) Otherwise, query backends in registration order and return the first match.
- On `Query`, if DSL contains `backend:<id>` or a new flag `--backend=<id>`, only query that backend; otherwise, fan out to all that support queries and merge results, tagging them for rendering.

`MultiManager` will implement `Backend` to remain compatible with `HandleHelpCommand` while allowing richer rendering with a new `Renderer` extension (see 4.2).

### 4.2 Renderer Extension for Grouped TopLevel and Source Tags

Extend `Renderer` with optional methods; keep the interface stable by adding a wrapper that detects support.

```go
// New optional extension; do NOT change the existing Renderer interface
// to preserve compatibility.
type GroupedRenderer interface {
    RenderTopLevelGrouped(groups []TopLevelGroup) string
    RenderQueryResultsGrouped(results []ScopedSection) string
}

type TopLevelGroup struct {
    BackendID    string
    BackendTitle string
    BackendDesc  string
    Page         *TopLevelPage
}

type ScopedSection struct {
    BackendID string
    Section   *Section
}
```

Behavior:
- If `Renderer` implements `GroupedRenderer`, `MultiManager` will provide grouped data; otherwise, it will transform grouped data back into a flattened `TopLevelPage` and `[]*Section` to feed the existing methods, preserving current rendering.

### 4.3 HelpArgs Enhancement (Optional)

Add an optional backend selector flag interpreted by `HandleHelpCommand` and passed via context to the manager:

- `--backend=<id>` restricts slug, top-level, and query operations to a specific backend.
- Keep existing input compatible: no change required for current callers.

Implementation detail: either extend `HelpArgs` with `BackendID string` and have `BuildDSL` inject `backend:<id>` or pass a separate hint down to `Backend.Query` via a lightweight context value.

### 4.4 Markdown Backend

Provide a small, dependency-free backend able to register arbitrary markdown sections.

```go
// bobatea/pkg/repl/help/adapters/markdown.go
package adapters

type MarkdownBackend struct {
    mu       sync.RWMutex
    sections map[string]*replhelp.Section // key: slug
}

func NewMarkdownBackend() *MarkdownBackend
func (m *MarkdownBackend) AddSection(s *replhelp.Section)
func (m *MarkdownBackend) AddFile(path string, withFrontmatter bool) error // optional helper
func (m *MarkdownBackend) TopLevel(ctx context.Context) (*replhelp.TopLevelPage, error)
func (m *MarkdownBackend) GetBySlug(ctx context.Context, slug string) (*replhelp.Section, error)
func (m *MarkdownBackend) Query(ctx context.Context, dsl string) ([]*replhelp.Section, bool, error)
```

Notes:
- Minimal text search for `Query` (quoted text or contains) similar to `SlashBackend`.
- If `withFrontmatter` is false, default fields come from filename and first heading; place the body in `Content` and set `Type` to `topic`.
- Top-level ordering by `Order`, then slug.

## 5) Behavior & UX

### 5.1 Top-Level Help
- Render a backend list with titles and descriptions, each showing its grouped sections (topics/examples/apps/tutorials).
- If only one backend is registered, behavior stays almost identical to current output.

### 5.2 Slug Resolution
- Prefer explicit `id:slug` when multiple backends may share slugs.
- Preserve existing prefixes (e.g., `slash-`) and allow custom `SlugPrefix` to maintain old links.

### 5.3 Querying
- Unscoped queries fan out; grouped renderer shows backend tags per result.
- `--backend=<id>` or `backend:<id>` in DSL scopes the query to a single backend.

## 6) Migration Strategy

- Existing users of a single backend: no change.
- Users of `adapters.NewMultiBackend(...)`: migrate to `replhelp.NewMultiManager(...)` with explicit registrations:

```go
mgr := replhelp.NewMultiManager(
    replhelp.BackendRegistration{ID: "glazed", Title: "Project Docs", Backend: &adapters.GlazedBackend{HS: helpSys}},
    replhelp.BackendRegistration{ID: "slash",  Title: "Slash Commands", Backend: adapters.NewSlashBackend(reg), SlugPrefix: "slash-"},
)
md := replhelp.HandleHelpCommand(ctx, replhelp.Config{Backend: mgr, ShowRelated: true, Renderer: replhelp.DefaultRenderer()}, "/help --all")
```

- For Markdown: add

```go
mdBackend := adapters.NewMarkdownBackend()
mdBackend.AddFile("docs/repl/intro.md", false)
mdBackend.AddSection(&replhelp.Section{Slug: "quick-start", Title: "Quick Start", Content: "# Quick Start\n...", Type: replhelp.TypeTopic})

mgr := replhelp.NewMultiManager(
    replhelp.BackendRegistration{ID: "glazed", Title: "Project Docs", Backend: &adapters.GlazedBackend{HS: helpSys}},
    replhelp.BackendRegistration{ID: "slash",  Title: "Slash Commands", Backend: adapters.NewSlashBackend(reg), SlugPrefix: "slash-"},
    replhelp.BackendRegistration{ID: "md",     Title: "Local Notes",  Backend: mdBackend},
)
```

## 7) Concrete Changes

- New: `bobatea/pkg/repl/help/multi.go` (manager, registration, grouping structs, context helpers).
- New: `bobatea/pkg/repl/help/adapters/markdown.go` (in-memory markdown backend, optional file loader).
- Update: `bobatea/pkg/repl/help/render.go`
  - Add `GroupedRenderer` optional extension (new interface) and a small adapter to flatten when not implemented.
- Update: `bobatea/pkg/repl/help/parse.go`
  - Optional: support `--backend=<id>`; plumb into context or DSL as `backend:<id>`.
- Keep: existing `MultiBackend` in `adapters/composite.go` (deprecate in docs, but keep for compatibility).

## 8) Backward Compatibility

This refactor intentionally drops backward compatibility:

- The `Renderer` interface changes to grouped rendering methods (see 4.2); old renderers will not compile.
- `HandleHelpCommand` now requires a multi-backend manager in `Config` instead of a single backend; old callers must migrate.
- The previous `adapters.MultiBackend` is superseded by the new manager and is no longer used.
- Slug resolution prefers explicit `id:slug` and configured `SlugPrefix`; existing slugs like `slash-...` should continue to work if `SlugPrefix` is set accordingly, but ambiguous lookups now require `id:` prefixes.

## 9) Example Wiring in oak main.go

Today:
```go
backend := helpadapters.NewMultiBackend(
    &helpadapters.GlazedBackend{HS: helpSys},
    helpadapters.NewSlashBackend(reg),
)
md := replhelp.HandleHelpCommand(ctx, replhelp.Config{Backend: backend, ShowRelated: true, Renderer: replhelp.DefaultRenderer()}, s)
```

With the proposal:
```go
mgr := replhelp.NewMultiManager(
    replhelp.BackendRegistration{ID: "glazed", Title: "Project Docs", Backend: &helpadapters.GlazedBackend{HS: helpSys}},
    replhelp.BackendRegistration{ID: "slash",  Title: "Slash Commands", Backend: helpadapters.NewSlashBackend(reg), SlugPrefix: "slash-"},
)
// Optionally add markdown
mdBackend := helpadapters.NewMarkdownBackend()
mdBackend.AddFile("./docs/repl/intro.md", false)
mgr = replhelp.NewMultiManager(mgr.Registrations()..., replhelp.BackendRegistration{ID: "md", Title: "Local Notes", Backend: mdBackend})

md := replhelp.HandleHelpCommand(ctx, replhelp.Config{Backend: mgr, ShowRelated: true, Renderer: replhelp.DefaultRenderer()}, s)
```

## 10) Open Questions / Future Work

- Do we want cross-backend related-content suggestions? For now, related is delegated per backend (if it implements `RelatedBackend`). A future extension could compute cross-source relationships by tag/topic overlap.
- Consider a simple caching layer for top-level aggregation and slug lookup maps when backends are large.
- Consider rendering small backend badges next to items to indicate source where useful (opt-in in `GroupedRenderer`).

## 11) Rollout Plan

- Implement `MultiManager` and `MarkdownBackend`.
- Add a minimal `GroupedRenderer` implementation for the default renderer (or keep flattened output until an improved renderer is added).
- Update docs and examples; keep `MultiBackend` but document `MultiManager` as the recommended approach.
- Update `oak` example to use `MultiManager` in a small follow-up change.

