## Advanced Diff Design Features (Deferred)

This document captures advanced features intentionally deferred from the MVP. Implement after the minimal diff viewer is stable and profiled.

### Ecosystem and Extensibility
- Renderer registry and plugin architecture
- Exporters (JSON, HTML, patch/unified diff) with registry
- Multiple built-in providers (YAML, Terraform, Git, DB schema)
- Adapter helpers for common domains (Terraform, Git)

### Rendering Enhancements
- Code renderer with syntax highlighting and unified/side-by-side views
- JSON/YAML pretty printers with color schemes
- Table/column-level renderers for schema diffs
- Rich status/iconography, per-category headers and counts

### UX and Navigation
- Category toggles and status filters panel
- Breadcrumbs for nested paths
- Jump-to next/previous match, bookmarking
- Resizable split panes and layout presets
- Context-sensitive help overlay

### Search and Filtering
- Fielded queries (status:added, category:env, path:regex)
- Debounced search and large-dataset optimizations
- Saved filter sets and presets

### Performance and Scale
- Virtualized lists for >10k items
- Lazy loading providers (paged data)
- Render memoization and value diff caching
- Benchmarks and profiling harness

### Theming and Styling
- Theme registry (default, dark, light, monochrome)
- Status color customization API
- Per-renderer style hooks

### API Surface (Proposed Later)
- RendererRegistry
- Export API with `AddExporter`
- Extended messages: FilterChanged, ViewModeChanged, ExportRequest
- Extended Config options (EnableExport, EnableFilters, DefaultCategoryFilter, MaxItemsInList)

### Documentation and Examples
- Examples: terraform-plan, db-schema, git-diff, custom-renderer, exporters
- “Best practices” guide with performance, UX, security sections

### Migration Path
- tfplandiff adapter package and migration guide
- Feature flags to shift from MVP to extended features progressively

### Risks and Mitigations
- Complexity creep: gate behind roadmap and flags
- Perf regressions: require benchmarks before enabling virtualized paths
- API churn: version extended APIs, keep MVP interfaces stable
