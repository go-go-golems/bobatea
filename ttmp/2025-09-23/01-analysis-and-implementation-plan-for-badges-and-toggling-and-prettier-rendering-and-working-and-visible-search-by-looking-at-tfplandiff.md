## Diff TUI: Analysis and Implementation Plan for Badges, Filters, Prettier Rendering, and a Working Visible Search (by studying tfplandiff)

Date: 2025-09-23

### 1) Purpose and Scope

This document analyzes the tfplandiff TUI implementation and derives a concrete, step-by-step plan to bring the same UX polish to the new, generic `pkg/diff` component in bobatea. The goal is to copy/adapt the proven patterns for:

- A visible search widget that actually resizes content and filters items
- Status filter toggles (added/removed/updated), on-by-default, with quick hotkeys
- Prettier rendering in the detail panel (headers, badges, groups/sections, consistent styles)
- Clean keymap conventions and help/footer hints
- A maintainable composition of smaller models/files

The end result should be an MVP that matches tfplandiff’s usability while preserving our generic interfaces and keeping the implementation modular.

### 2) Prior Art Overview (tfplandiff)

Files of interest (provided):
- `model.go` — owns overall model/state, search wiring, focus management, resizing
- `list.go` — wraps Bubble list with custom item and styles
- `detail.go` — contains a `viewport.Model` and setters for size/content
- `filters.go` — encapsulates detail filters and rendering of filter line
- `render.go` — composes the detail view: header, badges, filters line, sections
- `search.go` — matching strategy and filtering
- `options.go` — key bindings
- `styles.go` — lipgloss styles for list/detail and visual states
- `values.go` — value formatting and redaction

Key takeaways:
- The model maintains a single source of truth for state (focus, redaction, filters, search).
- Search UI is a `textinput.Model`; when visible, total layout height is recomputed; list/detail sizes are derived from body height minus frames.
- Filtering happens at two levels: list filtering via search; detail filtering via status toggles.
- Rendering is composed top→down (header, badges, filter-line, sections), and sections are composed left→right (e.g., prefix markers, values, and path labels).
- Key bindings mirror common TUI patterns (↑/↓, j/k, tab, /, r, q) and avoid quitting from Bubble list internals.

### 3) Target Architecture (bobatea/pkg/diff)

We already have:
- Minimal interfaces: `DataProvider`, `DiffItem`, `Category`, `Change` and `ChangeStatus`
- A working list/detail search MVP with redaction and frame-size-aware layout
- Config options for search and status filters (`WithSearch`, `WithStatusFilters`) and an initial `StatusFilter`

We will extend by adding modular files mirroring tfplandiff’s layering:
- `pkg/diff/model.go` — keep high-level orchestration only
- `pkg/diff/list.go` — list wrapper and item adapter/delegate styles
- `pkg/diff/detail.go` — viewport wrapper and size/content setters
- `pkg/diff/filters.go` — status filter state + renderer (badges/line)
- `pkg/diff/search.go` — search UI model and filtering functions
- `pkg/diff/render.go` — header, badges, filter-line, section renderers
- `pkg/diff/styles.go` — unified styles (list/detail, headers, badges, dimming)
- `pkg/diff/values.go` — value formatting, redaction helpers
- `pkg/diff/keymap.go` — clean key binding construction

Notes:
- We should minimize export surface area; keep helpers file-local unless required by examples.
- Preserve generic semantics in renderer: no domain coupling (Terraform). Use `ChangeStatus` and `Category` names.

### 4) Detailed Behavior Spec

4.1 Visible search widget
- Press `/` toggles visible `textinput.Model` at the top (header region). ESC hides.
- While visible, body height shrinks. When hidden, body height expands back. Recompute layout after both.
- Search matches against: `DiffItem.Name()`, `Change.Path()` and stringified `Before/After` values.
- Filtering affects left list (visible items). Detail panel renders the selected visible item.
- Optional: show “N of M matched” inline next to the search input placeholder.

4.2 Status filter toggles (added/removed/updated)
- Three toggles mapped to keys `1`, `2`, `3` respectively.
- They affect which change lines are shown in the detail viewport.
- The current state is displayed in a one-line filter strip under the header: e.g., `[+] Added ON   [−] Removed ON   [~] Updated ON`.
- Default: all ON. Configurable via `InitialFilter`.

4.3 Badges and header
- Show item name as the main title in detail.
- Optional badges to summarize counts: `+A −R ~U` (added, removed, updated) placed to the right of the title in a subtle line.
- Consider dimming when search is active but the item has zero matching change lines (still selectable but visually de-emphasized).

4.4 Rendering of change lines
- Use consistent prefix markers and color styles:
  - Removed: `- <before>` (red)
  - Added: `+ <after>` (green)
  - Updated: two lines (removed then added)
- Right-align or gray the path label; ensure values wrap reasonably; avoid truncating the value text.
- Redaction: when a `Change.Sensitive()` is true, render `[redacted]` (style muted) instead of the actual value.

4.5 List model
- Bubbles list with a custom delegate similar to tfplandiff (selected/dimmed styles).
- No built-in filter; all filtering is our own (search narrows our backing slice and `SetItems`).
- Disable quit bindings in list.

4.6 Keymap
- Up/Down or j/k: navigate list
- Tab: switch focus between list and detail
- `/`: toggle search (visible/hidden); focuses input
- ESC: while in search, hide and clear search
- `r`: toggle redaction
- `1`/`2`/`3`: toggle Added/Removed/Updated
- `q`: quit

### 5) File-by-File Implementation Guide

5.1 `pkg/diff/search.go`
- Define a `searchModel` that holds a `textinput.Model`, `visible` bool, and `query` string.
- Expose `Init() tea.Cmd`, `Update(msg)`, `View()` returning a single-line string.
- Provide `MatchesItem(DiffItem, lowerQuery string) bool` and `FilterItems([]DiffItem, query string)`; reuse/extend current logic.

Pseudocode:
```
type searchModel struct {
    input textinput.Model
    visible bool
    query string
}

func (s *searchModel) Show() { s.visible = true; s.input.Focus() }
func (s *searchModel) Hide() { s.visible = false; s.query = ""; s.input.Blur(); s.input.SetValue("") }
func (s *searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
    var cmd tea.Cmd
    s.input, cmd = s.input.Update(msg)
    s.query = strings.TrimSpace(s.input.Value())
    return *s, cmd
}
```

5.2 `pkg/diff/filters.go`
- `StatusFilter { ShowAdded, ShowRemoved, ShowUpdated }` (already defined in config). Add small renderer:
```
func renderFilterLine(f StatusFilter, styles Styles) string {
    badge := func(label string, on bool) string { /* style: on/off */ }
    return lipgloss.JoinHorizontal(lipgloss.Left,
        badge("+", f.ShowAdded),
        " ", badge("-", f.ShowRemoved),
        " ", badge("~", f.ShowUpdated),
    )
}
```
- Handle key toggles in `model.Update` (already partially implemented), then re-render detail.

5.3 `pkg/diff/list.go`
- Encapsulate `list.Model` construction and a custom delegate similar to tfplandiff (selected/dimmed styles).
- Provide `newItemList(items []DiffItem) list.Model` and `setItems(l *list.Model, items []DiffItem)`.

5.4 `pkg/diff/detail.go`
- Keep a wrapper with `viewport.Model` plus `SetSize(w,h)` and `SetContent(string)` that does `GotoTop()`.

5.5 `pkg/diff/render.go`
- Compose detail view:
  1) Title line (item name) + badges summarizing counts
  2) Filter line (status toggles) when enabled
  3) Category sections (grouped `Change` lists)
- Implement a small change counter: iterate item categories and accumulate counts by status.
- Move current line rendering logic here; allow a hook to dim paths; ensure redaction is applied.

5.6 `pkg/diff/model.go`
- Keep orchestration: compute header/search/footer and body sizes using `lipgloss.Height` and `GetFrameSize()`.
- When search visibility toggles, recompute layout and re-apply sizes.
- Selection changes update detail content immediately.
- `NewModelWith(provider, config, options...)` merges options (already in place).

5.7 `pkg/diff/styles.go`
- Add `BadgeAdded`, `BadgeRemoved`, `BadgeUpdated`, and `FilterOn/Off` styles.
- Ensure focused vs blurred borders are consistent (as in tfplandiff).

5.8 `pkg/diff/keymap.go`
- Extract key bindings into a `keyMap` struct; mirror tfplandiff mappings; keep `help` minimal (we already have a footer line).

5.9 `pkg/diff/values.go`
- Keep `formatValue`, `censorValue`, `renderChangeLine` equivalents; generic only; no Terraform specifics.

### 6) Resizing and Layout Rules

We will continue using:
- `headerHeight := lipgloss.Height(header)`; `footerHeight := lipgloss.Height(footer)`; `searchHeight := lipgloss.Height(search.View())`, when visible
- `frameW, frameH := style.GetFrameSize()` then: `innerW := panelW - frameW`, `innerH := bodyH - frameH`
- Recompute in these events: `WindowSizeMsg`, toggling search, toggling filters (if the filters line visibility changes), and switching focus (border color only; sizes unchanged).

Pseudocode for recompute:
```
func (m *Model) recompute() {
    head := m.renderHeader()
    search := ""
    if m.showSearch { search = m.searchInput.View() }
    foot := m.renderFooter()
    m.headerHeight = lipgloss.Height(lipgloss.JoinVertical(lipgloss.Left, head, search))
    m.footerHeight = lipgloss.Height(foot)
    m.bodyHeight = max(0, m.height - m.headerHeight - m.footerHeight - 1) // reserve 1 safety line
    m.applyContentSizes()
}
```

### 7) Search Behavior Details

- Query is applied to the full set of `items` and produces `visibleItems`.
- Keep the current selection index clamped when the list shrinks.
- Optional: highlight matches in the detail view (deferred); for now, filtering is sufficient.
- Option flag `EnableSearch` can turn off both the `/` toggle and the search line.

### 8) Status Filter Line & Badges Details

- Filter line (under title) appears when `EnableStatusFilters` is true.
- Badges show counts: compute `(added, removed, updated)` per item. Place to the right of the title using `lipgloss.JoinHorizontal`.
- Toggle state should be immediately visible (e.g., ON style = bright; OFF = dimmed/strikethrough).

### 9) Keyboard and Help/Footer

- Footer text includes: `↑/↓ move  tab switch  / search  r redact  1/2/3 filter +/−/~  q quit`.
- If `EnableSearch=false`, omit `/ search` from footer.

### 10) Testing & Examples

- Add golden tests for render: build a synthetic `DataProvider` with deterministic items and assert the rendered detail contains expected badges and lines.
- Extend examples:
  - Add flags `--no-search`, `--no-filters` to show configurability
  - Provide a big directory example to validate performance and navigation
- Use `tmux` scripted tests (send keys, capture-pane) to ensure footer/help and filter toggles show correctly.

### 11) Migration Notes (tfplandiff → bobatea)

- Create a `TFProvider` adapter (as already sketched) to map Terraform resources to `DiffItem`/`Category`/`Change`.
- Keep existing tfplandiff behavior; add a flag there to run on top of bobatea diff when ready, then migrate gradually.

### 12) Risks & Mitigations

- Performance on very large diffs: start with simple list; consider virtualization later.
- Over-styling: keep one default theme; ensure good contrast in both light/dark terminals.
- API creep: keep `Config` small; expose `With*` options to avoid breaking changes.

### 13) Implementation Plan (Checklist)

1. Create new files and move code by responsibility
   - [ ] `search.go`: searchModel, filter functions
   - [ ] `filters.go`: status filter state + renderFilterLine()
   - [x] `list.go`: bubbles list setup and delegate
   - [x] `detail.go`: viewport wrapper
   - [x] `render.go`: header, badges, filter line, sections, counters
   - [x] `styles.go`: add badge/filter styles
   - [ ] `keymap.go`: key definitions
   - [ ] `values.go`: value formatting/redaction helpers
2. Wire search visibility & layout
   - [x] Toggle `/` shows/hides search; recompute sizes each time
   - [x] ESC in search hides and clears
   - [x] Filtering updates `visibleItems` and list items; selection conserved
3. Add status filter strip and toggles
   - [x] Keys `1/2/3` toggle Added/Removed/Updated
   - [x] Detail render respects toggles
   - [x] Footer/Help shows toggles
4. Add badges
   - [x] Count per item: added/removed/updated
   - [x] Render badges next to title
5. Final polish
   - [x] Check frame-size math (header+search+filters+footer)
   - [x] Verify tmux scripts show header, search, filter line, and footer
   - [ ] Add example flags to disable features

### 14) Resources

- `go-go-mento/go/cmd/tfplandiff/internal/tui/model.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/list.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/detail.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/filters.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/render.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/search.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/options.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/styles.go`
- `go-go-mento/go/cmd/tfplandiff/internal/tui/values.go`

### 15) Save Further Research

Save any new notes or experiments under `ttmp/2025-09-23/` as `0X-*.md` and keep this plan updated as features land.

---

### Changelog / Findings / Things to Remember

- Added badge and filter styles; implemented badge counts and filter line rendering. Works well with current layout.
- Footer updated to show `1/2/3` toggles; verified in tmux capture.
- Search toggle now recomputes sizes when shown/hidden; body adjusts correctly and no clipping at top/bottom.
- JSON-dir fixtures auto-generator had an initial escaping issue; fixed by writing valid JSON via `json.Marshal`.
- Next: factor list delegate and detail wrapper into their own files to simplify `model.go`, and extract keymap. Add example flags to disable features at runtime.
  - Done: list delegate and detail wrapper created; build green.
  - Pending: extract keymap; flags in examples.


