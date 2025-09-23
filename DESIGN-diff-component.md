# Design Document: Reusable Diff Component for Bobatea

**Document Type:** Technical Design Document  
**Date:** 2025-09-23  
**Purpose:** Define a minimal, reusable structured data diff component (MVP) for bobatea.

## Executive Summary

We will ship a lean diff viewer for Bubble Tea applications focused on: two-pane list/detail, substring search, and value redaction. The design intentionally defers registries, exporters, multi-provider bundles, and advanced rendering to a roadmap document.

Advanced/deferred features are tracked in `ttmp/2025-09-23/01-advanced-diff-design-features-for-later.md`.

## Scope (MVP)

- Two-pane layout: list of items (left), detail pane (right)
- Substring search across item name, change paths, and rendered values
- Redaction toggle for sensitive values
- Default theme only (aligned with bobatea look and feel)
- Minimal interfaces and configuration

Out of scope for MVP: exporters, plugin/registry systems, multiple built-in providers, advanced renderers (code, schema, syntax highlighting), complex theming, virtualized lists. See ttmp roadmap.

## Minimal Interfaces

```go
// Data source
interface DataProvider {
    Title() string
    Items() []DiffItem
}

// Item with grouped changes
interface DiffItem {
    ID() string
    Name() string
    Categories() []Category
}

interface Category {
    Name() string
    Changes() []Change
}

interface Change {
    Path() string
    Status() ChangeStatus // "added"|"updated"|"removed"
    Before() any
    After() any
    Sensitive() bool
}

type ChangeStatus string
```

Notes:
- Keep interfaces focused on what the renderer needs. Derive badges/aggregates at render-time.
- Omit descriptions, metadata, icons, priorities, actions. Add only when required by a concrete use case.

## Minimal Configuration

```go
 type Config struct {
     Title           string
     RedactSensitive bool
     SplitPaneRatio  float64 // optional; default 0.35
 }
```

- Window size comes from Bubble Tea `WindowSizeMsg`; no explicit width/height config.
- Search is always enabled; no filters panel in MVP.

## Model and Behavior

- List: Bubble list for items; description shows simple counts (optional)
- Detail: viewport scrolling; render categories and change lines
- Search: lowercased substring match across name, paths, and values
- Redaction: masks `Before/After` values when enabled
- Keys: Up/Down, PageUp/PageDown (optional), Tab (switch pane), `/` (search), `r` (redact), `q` (quit)

## Rendering

- Default renderer only
- Change lines:
  - Header: category name
  - Lines: `- <before>` and `+ <after>` with status coloring
  - Redaction applies to values
- Keep styles minimal and consistent with bobatea components

## Package Layout (MVP)

```
pkg/diff/
  doc.go           # package docs
  model.go         # Bubble Tea model (list/detail/search/redaction)
  provider.go      # minimal interfaces
  renderer.go      # default renderer + RenderOptions
  config.go        # Config + defaults
  styles.go        # minimal styles/theme
examples/diff/basic-usage/main.go
```

## Public API (MVP)

```go
func NewModel(provider DataProvider, config Config) Model
func DefaultConfig() Config

// Bubble Tea
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string

// Options
func (m *Model) SetSize(width, height int)        // optional convenience
func (m *Model) SetRedactSensitive(enabled bool)
func (m *Model) SetSplitPaneRatio(ratio float64)
```

## Delivery Plan (1–1.5 weeks)

- Day 1–2: Interfaces, model skeleton, setSize/layout, styles
- Day 3: Search + redaction; wire list/detail update flow
- Day 4: JSON example provider + example app
- Day 5: Tests (basic), polish, README
- Day 6–7: Optional: status badges, dark theme variant

## Migration Note (tfplandiff)

- Implement a thin adapter in `tfplandiff` to implement `DataProvider` + `DiffItem` etc.
- Keep current tfplandiff behavior; adopt bobatea diff incrementally behind a flag.

## Risks (MVP)

- Domain mismatch: Keep interfaces minimal; adapt via provider
- Performance on large diffs: Accept baseline; profile before adding virtualization
- Scope creep: Advanced items are explicitly moved to the ttmp roadmap

## Roadmap

Deferred features and expanded APIs live in:
- `ttmp/2025-09-23/01-advanced-diff-design-features-for-later.md`
