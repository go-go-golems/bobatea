## Diff Component (MVP)

A minimal, reusable diff viewer for Bubble Tea apps. Focused on: two-pane list/detail, search, and redaction. Everything else is deferred to the advanced design doc.

### Background and Motivation

This component generalizes a proven pattern used in the Terraform plan diff TUI (`tfplandiff`) and makes it reusable for any structured data. The inspiration comes from a clean two-pane layout, responsive sizing, simple but effective search, and sensitive value redaction.

Reference implementation for inspiration (not a dependency):
- `go-go-mento/go/cmd/tfplandiff/internal/tui/model.go` — main Bubble Tea model, layout, focus, search integration
- `.../tui/list.go` — list panel using Bubble's list with custom item
- `.../tui/detail.go` — detail panel using viewport
- `.../tui/search.go` — substring search logic
- `.../tui/render.go` — rendering detail sections and change lines
- `.../tui/styles.go` — lipgloss styles and color system
- `.../tui/values.go` — value formatting, redaction helpers
- `.../tui/options.go` — key bindings and help

The MVP here keeps those core ideas while stripping domain coupling (Terraform) and advanced/unnecessary features for a faster first delivery.

### Minimal API

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

### Configuration

```go
 type Config struct {
     Title           string
     RedactSensitive bool
     SplitPaneRatio  float64 // default 0.35
     EnableSearch        bool
     EnableStatusFilters bool
     InitialFilter       StatusFilter
 }

 type StatusFilter struct {
     ShowAdded   bool
     ShowRemoved bool
     ShowUpdated bool
 }

 // Functional options
 func WithSearch(enabled bool) Option
 func WithStatusFilters(enabled bool, initial StatusFilter) Option
```

### Quick Start

```go
provider := NewJSONProvider(before, after)
config := diff.DefaultConfig()
config.Title = "JSON Diff"
model := diff.NewModel(provider, config)
_ = tea.NewProgram(model, tea.WithAltScreen()).Run()
```

### UX and Keys
- Up/Down: navigate list
- Tab: switch pane
- /: show search input (no leading slash in input)
- r: toggle redaction
- 1/2/3: toggle Added/Removed/Updated status filters
- q: quit

### Search
- Substring match over item name, change paths, and rendered values (lowercased).
- Search input is shown on its own header line and resizes the body content when visible.
- Filtering is applied at two levels:
  - Left list: visible items are filtered by query
  - Right detail: only matching change lines are rendered, and badges reflect filtered counts

### Rendering
- Default renderer shows:
  - Item header (name)
  - Badges with counts: `+A -R ~U` (added/removed/updated)
  - Optional filter strip: `+ ON   - ON   ~ ON` (or OFF states)
  - Categories with change lines (`- before`, `+ after`) filtered by status & search
  - Redaction applies to values when enabled

### Architecture

The component mirrors `tfplandiff` structure while remaining domain-agnostic:

- `provider.go` — minimal interfaces: `DataProvider`, `DiffItem`, `Category`, `Change`
- `model.go` — orchestrates layout, focus, search, and filters; computes sizes using `lipgloss.Height` and `GetFrameSize()`
- `list.go` — wraps a `list.Model` with an item adapter; disables built-in filter/help
- `detail.go` — viewport wrapper with `SetSize` and `SetContent`
- `renderer.go` — composes detail view: header + badges, filter line, sections; applies search/status filtering and redaction
- `styles.go` — baseline list/detail borders, plus `BadgeAdded/Removed/Updated`, `FilterOn/Off`
- `keymap.go` — key bindings similar to tfplandiff

Resizing rules:
- Header height = title + (search line if visible)
- Footer height = help text
- Body height = window height - header - footer - safety line
- Inner widths/heights = panel size - frame sizes from styles

Search behavior:
- Toggle `/` shows an empty input (no leading slash) and focuses it
- ESC hides and clears input
- As you type, list and detail outputs update live

Status filters:
- Toggle `1/2/3` to show/hide Added/Removed/Updated lines in detail
- Filter line and badges update immediately

### Reference: tfplandiff TUI code inspiration (selected snippets)

Program setup and model wiring:

```38:76:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/model.go
func Run(ctx context.Context, result *core.DiffResult, opts Options) error {
	resources := make([]core.ResourceDiff, 0, len(result.Resources))
	for _, res := range result.Resources {
		filtered, ok := core.FilterResource(res, opts.ComponentFilters, opts.PathFilters, opts.ShowArgs)
		if !ok {
			continue
		}
		resources = append(resources, filtered)
	}

	if len(resources) == 0 {
		fmt.Println("No resource changes matched the current filters.")
		return nil
	}

	m := model{
		resources:        resources,
		list:             newResourceListModel(resources),
		detail:           newDetailModel(),
		help:             help.New(),
		keys:             newKeyMap(),
		focus:            focusList,
		redacted:         opts.RedactValues,
		detailFilters:    newDetailFilters(),
		visibleResources: resources,
		showDetails:      false,
	}
	m.detailFilters.ShowOther = opts.ShowArgs

	input := textinput.New()
	input.Placeholder = "Search resources"
	input.Prompt = ""
	input.CharLimit = 0
	m.searchInput = input
	m.applySearchFilter("")

	p := tea.NewProgram(&m, tea.WithContext(ctx), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

List model setup with Bubble list:

```35:66:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/list.go
func newResourceListModel(resources []core.ResourceDiff) resourceListModel {
	items := make([]list.Item, len(resources))
	for idx, res := range resources {
		items[idx] = resourceItem{diff: res}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)

	styles := list.NewDefaultItemStyles()
	styles.NormalTitle = styles.NormalTitle.Foreground(colorPrimary).Bold(true)
	styles.NormalDesc = styles.NormalDesc.Foreground(colorMuted)
	styles.DimmedTitle = styles.DimmedTitle.Foreground(colorMuted)
	styles.DimmedDesc = styles.DimmedDesc.Foreground(colorMuted)
	styles.SelectedTitle = styles.SelectedTitle.
		BorderForeground(colorAccent).
		Foreground(colorAccent).
		Bold(true)
	styles.SelectedDesc = styles.SelectedDesc.Foreground(colorAccent)
	delegate.Styles = styles

	l := list.New(items, delegate, 0, 0)
	l.Title = "Resources"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.DisableQuitKeybindings()

	return resourceListModel{list: l}
}
```

Detail panel via viewport:

```12:31:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/detail.go
func newDetailModel() detailModel {
	vp := viewport.New(0, 0)
	return detailModel{viewport: vp}
}

func (m *detailModel) SetSize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	m.viewport.Width = width
	m.viewport.Height = height
}

func (m *detailModel) SetContent(content string) {
	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}
```

Search matching strategy (lowercased substring across fields):

```40:76:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/search.go
func resourceMatchesQuery(res core.ResourceDiff, lowerQuery string) bool {
	if lowerQuery == "" {
		return true
	}
	if strings.Contains(strings.ToLower(res.Address), lowerQuery) {
		return true
	}
	for _, action := range res.Actions {
		if strings.Contains(strings.ToLower(action), lowerQuery) {
			return true
		}
	}
	for _, diff := range res.EnvDiffs {
		if strings.Contains(strings.ToLower(diff.Key), lowerQuery) ||
			strings.Contains(strings.ToLower(diff.ComponentPath), lowerQuery) ||
			strings.Contains(strings.ToLower(diff.Status), lowerQuery) {
			return true
		}
		before := strings.ToLower(fmt.Sprint(diff.BeforeValue))
		after := strings.ToLower(fmt.Sprint(diff.AfterValue))
		if strings.Contains(before, lowerQuery) || strings.Contains(after, lowerQuery) {
			return true
		}
	}
	for _, diff := range res.OtherDiffs {
		if strings.Contains(strings.ToLower(diff.Path), lowerQuery) {
			return true
		}
		before := strings.ToLower(fmt.Sprint(diff.Before))
		after := strings.ToLower(fmt.Sprint(diff.After))
		if strings.Contains(before, lowerQuery) || strings.Contains(after, lowerQuery) {
			return true
		}
	}
	return false
}
```

Detail rendering composition:

```58:75:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/render.go
func renderResourceDetail(res core.ResourceDiff, redacted bool, filters detailFilters, searchQuery string, showMeta bool) string {
	if res.Address == "" {
		return detailEmptyStyle.Render("No resource selected.")
	}

	headerParts := []string{detailTitleStyle.Render(res.Address)}
	if badges := renderActionBadges(res.Actions); badges != "" {
		headerParts = append(headerParts, badges)
	}
	header := lipgloss.JoinHorizontal(lipgloss.Left, headerParts...)
	lowerQuery := strings.ToLower(strings.TrimSpace(searchQuery))

	filtersLine := renderDetailFilters(filters, showMeta)
	envSection := renderEnvSection(res.EnvDiffs, redacted, filters, lowerQuery, showMeta)
	otherSection := renderOtherSection(res.OtherDiffs, filters, lowerQuery)

	return lipgloss.JoinVertical(lipgloss.Left, header, filtersLine, envSection, otherSection)
}
```

Styles baseline for list/detail borders:

```22:31:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/styles.go
var (
	focusedBorderColor = lipgloss.AdaptiveColor{Light: "#3B82F6", Dark: "#60A5FA"}
	blurredBorderColor = lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#1F2937"}

	baseListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blurredBorderColor).
			Padding(0, 1).
			MarginRight(1)

	baseDetailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blurredBorderColor).
			Padding(0, 1)
```

Value rendering and redaction pattern:

```11:29:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/values.go
func renderChangeLine(prefix string, value any, redacted bool, style lipgloss.Style) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok && str == "" {
		label := "EMPTY"
		if strings.HasPrefix(prefix, "-") {
			label = "REMOVED"
		} else if strings.HasPrefix(prefix, "+") {
			label = "ADDED"
		}
		return style.Render(fmt.Sprintf("%s %s", prefix, label))
	}
	display := formatValue(value)
	if redacted {
		display = redactedValueStyle.Render(censorValue(value))
	}
	return style.Render(fmt.Sprintf("%s %s", prefix, display))
}
```

Keymap inspiration:

```38:64:/home/manuel/workspaces/2025-09-15/onboarding-developers/go-go-mento/go/cmd/tfplandiff/internal/tui/options.go
func newKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("pgdn/f", "page down"),
		),
		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		ToggleRedact: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "toggle redaction"),
		),
	}
}
```

### Scope and Non-Goals (v1)
- No exporter/HTML/patch output
- No plugin/registry systems
- No multiple built-in providers
- No advanced renderers (code, schema, syntax highlighting)
- No complex theming; one default theme

### Roadmap / Advanced Features
See `ttmp/2025-09-23/01-advanced-diff-design-features-for-later.md` for deferred capabilities, including exporters, plugin registries, advanced renderers, and performance work.

### Project Layout and Starting Points

Package scaffold:

```
pkg/diff/
  doc.go           # package docs
  model.go         # Bubble Tea model (list/detail/search/redaction)
  provider.go      # minimal interfaces
  renderer.go      # default renderer + RenderOptions
  config.go        # Config + defaults
  styles.go        # minimal styles/theme
  list.go          # list wrapper and item adapter
  detail.go        # viewport wrapper
  keymap.go        # key bindings
examples/diff/basic-usage/main.go
```

### Adapting tfplandiff to DataProvider (adapter sketch)

```go
type TFProvider struct{ resources []core.ResourceDiff }

func (p *TFProvider) Title() string { return "Terraform Plan" }
func (p *TFProvider) Items() []diff.DiffItem {
    out := make([]diff.DiffItem, 0, len(p.resources))
    for _, r := range p.resources {
        out = append(out, tfItem{r})
    }
    return out
}

type tfItem struct{ r core.ResourceDiff }
func (i tfItem) ID() string   { return i.r.Address }
func (i tfItem) Name() string { return i.r.Address }
func (i tfItem) Categories() []diff.Category {
    return []diff.Category{
        tfCategory{"env", i.r.EnvDiffs},
        tfOtherCategory{"attr", i.r.OtherDiffs},
    }
}
```

Start by wiring the Bubble Tea program (see model snippet above), then implement minimal rendering, search, and redaction.
