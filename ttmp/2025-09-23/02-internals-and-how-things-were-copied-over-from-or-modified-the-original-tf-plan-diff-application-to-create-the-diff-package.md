# Internals: Mapping tfplandiff to bobatea/pkg/diff

Date: 2025-09-23

## Overview

This document explains what we copied, adapted, or redesigned from the `tfplandiff` TUI to build the generic `pkg/diff` component. The goal is to preserve the proven UX (two-pane layout, visible search, filters, redaction) while removing Terraform-specific coupling.

## Mapping Table

- tfplandiff `model.go` → pkg/diff `model.go`
  - Focus, layout, resizing, search visibility, filter toggles
  - Differences: now generic `DataProvider` and `DiffItem` collection; no Terraform filtering flags

- tfplandiff `list.go` → pkg/diff `list.go`
  - List wrapper with item adapter; disabled quit, no internal filter
  - Differences: items are `DiffItem` with name + categories; description shows total changes

- tfplandiff `detail.go` → pkg/diff `detail.go`
  - Viewport wrapper with `SetSize` and `SetContent`
  - Kept minimal, generic

- tfplandiff `render.go` → pkg/diff `renderer.go`
  - Header with badges; filter strip; sections; line rendering; redaction
  - Differences: counts/badges now driven by `ChangeStatus`; search filters detail lines, not just resource-level

- tfplandiff `search.go` → pkg/diff `renderer.go` + `model.go`
  - Matching generalized: item names, change paths, before/after values
  - Visible search row; no leading slash prompt; ESC clears

- tfplandiff `styles.go` → pkg/diff `styles.go`
  - Borders, Title, Path, Added/Removed/Updated, SensitiveValue; new Badge/FilterOn/Off styles

- tfplandiff `values.go` → pkg/diff `renderer.go`
  - `valueToString` and redaction logic; can add more formatters later

## Key Changes

1) Interfaces
```
DataProvider → Title() string, Items() []DiffItem
DiffItem → ID(), Name(), Categories()
Category → Name(), Changes()
Change → Path(), Status(), Before(), After(), Sensitive()
```
This decouples the UI from Terraform-specific resource structures.

2) Search UX
- `/` shows a separate input line; prompt=""
- As you type, list narrows; detail renders only matching lines. Badges are based on filtered lines.

3) Filters & Badges
- `StatusFilter` toggles Added/Removed/Updated via keys 1/2/3
- Filter line renders ON/OFF state with distinct styles
- Badges show `+A -R ~U` counts; when search filters lines, counts reflect visible lines

4) Resizing
- Header height = title + search (if visible)
- Footer = help line
- Body = height - header - footer - safety line
- Panel inner sizes use `GetFrameSize()` from styles

5) Files and Responsibilities
- Moved logic into `list.go`, `detail.go`, `renderer.go`, `styles.go` to avoid a monolithic `model.go`
- `keymap.go` collects bindings similar to tfplandiff; help text remains in footer

## What We Didn’t Copy

- Terraform-specific filtering and action badges
- Pagination and advanced help UI
- Exporters and meta sections (deferred)

## Future Improvements

- Extract a dedicated `search.go` to encapsulate input model and helpers
- Add `values.go` with pluggable value formatters and redaction policies
- Virtualization for very large lists; streaming update support
- Feature flags in examples (`--no-search`, `--no-filters`)

## Notes & Caveats

- Rendering counts are recomputed per view on filtered content; for large datasets, we may want cached aggregates
- Redaction styles are intentionally subdued; ensure accessibility in dark/light terminals


