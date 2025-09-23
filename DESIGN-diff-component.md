# Design Document: Reusable Diff Component for Bobatea

**Document Type:** Technical Design Document  
**Date:** 2025-09-23  
**Purpose:** Design and implementation plan for a reusable structured data diff component

## Executive Summary

This document outlines the design and implementation plan for a generic, reusable diff visualization component that will be added to the bobatea package. The component will follow bobatea's established patterns and conventions while providing powerful diff visualization capabilities for any structured data type.

## Design Principles

### Following Bobatea Conventions

1. **Package Structure**: Located in `pkg/diff/` following bobatea's component organization
2. **Interface-Driven Design**: Core abstractions through Go interfaces for extensibility
3. **Message-Based Communication**: Clean Tea message API for embedding
4. **Comprehensive Documentation**: User documentation in `docs/diff.md`, examples in `examples/diff/`
5. **Theming Support**: Built-in themes and custom styling capabilities
6. **Configuration-Driven**: Flexible configuration with sensible defaults

### Core Design Goals

1. **Generic and Reusable**: Work with any structured data diff scenario
2. **High Performance**: Handle large diffs efficiently
3. **Rich User Experience**: Comprehensive navigation, search, and filtering
4. **Embeddable**: Clean integration into larger applications
5. **Extensible**: Plugin architecture for custom renderers and filters

## Package Structure

### File Organization

```
pkg/diff/
├── doc.go                      # Package documentation
├── model.go                    # Main Bubble Tea model
├── provider.go                 # Data provider interfaces
├── renderer.go                 # Rendering system interfaces
├── messages.go                 # Tea message definitions
├── config.go                   # Configuration types and defaults
├── styles.go                   # Styling and theming
├── filters.go                  # Filtering and search system
├── export.go                   # Export functionality
├── navigation.go               # Navigation and focus management
├── diff_test.go               # Core tests
├── providers/                 # Built-in data providers
│   ├── json.go               # JSON diff provider
│   ├── yaml.go               # YAML diff provider
│   ├── terraform.go          # Terraform plan provider
│   └── generic.go            # Generic object diff provider
├── renderers/                # Built-in renderers
│   ├── json.go              # JSON-specific renderer
│   ├── yaml.go              # YAML-specific renderer
│   ├── code.go              # Code diff renderer
│   └── table.go             # Tabular data renderer
└── examples/                 # Usage examples (imports)
    ├── simple_test.go       # Basic usage examples
    ├── advanced_test.go     # Advanced configuration examples
    └── embedding_test.go    # Embedding examples
```

### External Structure

```
examples/diff/                  # Working example applications
├── basic-usage/
│   └── main.go                # Simple JSON diff example
├── terraform-plan/
│   └── main.go                # Terraform plan diff
├── database-schema/
│   └── main.go                # Database schema diff
├── custom-provider/
│   └── main.go                # Custom data provider example
├── custom-renderer/
│   └── main.go                # Custom renderer example
└── embedded-in-app/
    └── main.go                # Multi-component application
```

## Core Interfaces

### Data Provider System

```go
// Primary interface for providing diff data
type DataProvider interface {
    GetItems() []DiffItem
    GetTitle() string
    GetSummary() string
}

// Represents a single diffable item (resource, file, object)
type DiffItem interface {
    GetID() string
    GetDisplayName() string
    GetDescription() string
    GetCategories() []Category
    GetActions() []Action
    GetMetadata() map[string]interface{}
}

// Represents a category of changes within an item
type Category interface {
    GetName() string
    GetDisplayName() string
    GetChanges() []Change
    IsVisible() bool
    GetIcon() string
    GetPriority() int
}

// Represents a single change
type Change interface {
    GetPath() string
    GetStatus() ChangeStatus
    GetBeforeValue() interface{}
    GetAfterValue() interface{}
    GetMetadata() map[string]interface{}
    IsSensitive() bool
    GetDisplayType() string
}
```

### Rendering System

```go
// Primary rendering interface
type Renderer interface {
    RenderItemSummary(item DiffItem, opts RenderOptions) string
    RenderItemDetail(item DiffItem, opts RenderOptions) string
    RenderChange(change Change, opts RenderOptions) string
    RenderCategoryHeader(category Category, opts RenderOptions) string
    SupportsType(dataType string) bool
}

// Rendering configuration
type RenderOptions struct {
    RedactSensitive  bool
    ShowMetadata     bool
    SearchQuery      string
    HighlightSearch  bool
    CategoryFilters  map[string]bool
    StatusFilters    map[ChangeStatus]bool
    MaxValueLength   int
    IndentSize       int
    ShowLineNumbers  bool
    ColorScheme      ColorScheme
}

// Multi-renderer system for different data types
type RendererRegistry interface {
    Register(name string, renderer Renderer)
    GetRenderer(dataType string) Renderer
    GetDefaultRenderer() Renderer
}
```

### Filtering and Search

```go
// Filter function type
type FilterFunc func(DiffItem) bool

// Search matcher interface
type SearchMatcher interface {
    Matches(item DiffItem, query string) bool
    SupportsAdvancedQuery() bool
    ParseQuery(query string) (SearchQuery, error)
}

// Filter manager
type FilterManager interface {
    AddFilter(name string, filter FilterFunc)
    RemoveFilter(name string)
    EnableFilter(name string, enabled bool)
    GetActiveFilters() []string
    Apply(items []DiffItem) []DiffItem
}
```

## Implementation Plan

### Phase 1: Core Foundation (Week 1)

**Objective**: Establish core interfaces and basic functionality

#### Tasks:
1. **Core Interfaces** (1-2 days)
   - Define `DataProvider`, `DiffItem`, `Category`, `Change` interfaces
   - Create basic message types in `messages.go`
   - Define configuration structure in `config.go`

2. **Basic Model** (2-3 days)
   - Implement main Bubble Tea model in `model.go`
   - Basic two-pane layout (list + detail)
   - Simple navigation and focus management
   - Basic keyboard handling

3. **Simple Provider** (1 day)
   - Generic object diff provider for testing
   - Basic diff computation logic

4. **Basic Rendering** (1-2 days)
   - Default renderer implementation
   - Simple text-based diff display
   - Basic styling setup

#### Deliverables:
- Working basic diff viewer
- Core interfaces defined and documented
- Simple example application
- Basic test suite

### Phase 2: Advanced Features (Week 2)

**Objective**: Add comprehensive features and polish

#### Tasks:
1. **Enhanced Rendering** (2 days)
   - JSON and YAML specific renderers
   - Syntax highlighting support
   - Side-by-side and unified diff modes
   - Value formatting improvements

2. **Search and Filtering** (2-3 days)
   - Implement search functionality
   - Status-based filtering (added/updated/removed)
   - Category filtering
   - Advanced search queries (field:value syntax)

3. **Theming System** (1-2 days)
   - Complete styling system
   - Built-in themes (default, dark, light, monochrome)
   - Color scheme management
   - Responsive styling

#### Deliverables:
- Rich rendering capabilities
- Comprehensive search and filtering
- Multiple built-in themes
- Enhanced examples

### Phase 3: Data Providers and Export (Week 3)

**Objective**: Create domain-specific providers and export functionality

#### Tasks:
1. **Built-in Providers** (3-4 days)
   - JSON diff provider with deep comparison
   - YAML diff provider
   - Terraform plan provider (migrated from existing code)
   - Generic object provider with reflection

2. **Export System** (2 days)
   - Export interface and registry
   - JSON export format
   - HTML report generation
   - Unified diff format export

3. **Advanced Navigation** (1-2 days)
   - Breadcrumb navigation
   - Jump to specific changes
   - Bookmarking and favorites

#### Deliverables:
- Production-ready data providers
- Export functionality
- Advanced navigation features
- Real-world examples

### Phase 4: Polish and Documentation (Week 4)

**Objective**: Complete documentation, examples, and performance optimization

#### Tasks:
1. **Performance Optimization** (2 days)
   - Virtual scrolling for large diffs
   - Lazy loading optimizations
   - Memory usage improvements
   - Benchmark suite

2. **Comprehensive Examples** (2 days)
   - Database schema diff example
   - Git diff visualization
   - Configuration file diff
   - Multi-provider application

3. **Documentation and Testing** (2-3 days)
   - Complete API documentation
   - Usage examples and tutorials
   - Comprehensive test coverage
   - Integration tests

#### Deliverables:
- Performance-optimized component
- Complete documentation suite
- Comprehensive examples
- Production-ready package

## Technical Implementation Details

### State Management

```go
type Model struct {
    // Core data
    provider     DataProvider
    config       Config
    styles       Styles
    renderer     Renderer
    
    // UI state
    items        []DiffItem
    filteredItems []DiffItem
    selectedIndex int
    focusedPane  FocusedPane
    
    // View state
    listModel    list.Model
    detailModel  viewport.Model
    searchModel  textinput.Model
    
    // Component state
    width        int
    height       int
    ready        bool
    
    // Feature state
    searchQuery  string
    filters      FilterManager
    viewMode     ViewMode
    showSummary  bool
    
    // Export state
    exporters    map[string]ExportFunc
}
```

### Layout System

```go
type LayoutManager struct {
    splitRatio   float64
    minPaneSize  int
    showSummary  bool
    compactMode  bool
}

func (lm *LayoutManager) CalculateLayout(width, height int) Layout {
    // Calculate pane sizes based on configuration
    // Handle responsive layout adjustments
    // Return layout with specific dimensions
}

type Layout struct {
    SummaryHeight int
    ListWidth     int
    ListHeight    int
    DetailWidth   int
    DetailHeight  int
    SearchHeight  int
}
```

### Message Flow

```go
// Core messages that the component emits
type SelectionChangedMsg struct {
    Item     DiffItem
    Index    int
    Category *Category
    Change   *Change
}

type FilterChangedMsg struct {
    ActiveFilters []string
    ResultCount   int
}

type SearchChangedMsg struct {
    Query       string
    ResultCount int
    Matches     []SearchMatch
}

type ExportRequestMsg struct {
    Format string
    Items  []DiffItem
}

type ViewModeChangedMsg struct {
    Mode     ViewMode
    Previous ViewMode
}
```

### Performance Considerations

1. **Virtual Scrolling**: For large item lists
2. **Lazy Rendering**: Render visible content only
3. **Memoization**: Cache expensive rendering operations
4. **Debounced Search**: Prevent excessive filtering on rapid input
5. **Efficient Diff Algorithms**: Use optimized diff computation

### Memory Management

```go
type Cache struct {
    renderedItems   map[string]string
    searchResults   map[string][]DiffItem
    maxCacheSize    int
    cacheHits       int64
    cacheMisses     int64
}

func (c *Cache) Get(key string) (string, bool) {
    // Thread-safe cache access
}

func (c *Cache) Set(key, value string) {
    // LRU eviction policy
    // Memory usage monitoring
}
```

## Integration Strategy

### Bobatea Package Integration

1. **Follow Existing Patterns**:
   - Use same directory structure as `pkg/repl/`
   - Follow same documentation format as `docs/repl.md`
   - Use same example structure as `examples/repl/`

2. **Dependency Management**:
   - Minimal external dependencies
   - Reuse bobatea's existing components where possible
   - Compatible with current Go version requirements

3. **Testing Strategy**:
   - Unit tests for all core functionality
   - Integration tests with different data providers
   - Example tests that demonstrate usage
   - Benchmark tests for performance validation

### Migration from tfplandiff

1. **Adapter Creation**:
   ```go
   // Create adapter to bridge existing tfplandiff types
   type TerraformAdapter struct {
       result *core.DiffResult
   }
   
   func (ta *TerraformAdapter) GetItems() []diff.DiffItem {
       items := make([]diff.DiffItem, len(ta.result.Resources))
       for i, res := range ta.result.Resources {
           items[i] = &TerraformDiffItem{resource: res}
       }
       return items
   }
   ```

2. **Gradual Migration**:
   - Keep existing tfplandiff functionality intact
   - Add new diff component as alternative
   - Migrate incrementally with feature flags
   - Deprecate old implementation once stable

## Risk Mitigation

### Technical Risks

1. **Performance with Large Diffs**:
   - Mitigation: Virtual scrolling, lazy loading, pagination
   - Monitoring: Performance benchmarks, memory profiling

2. **Memory Usage**:
   - Mitigation: Efficient caching, garbage collection optimization
   - Monitoring: Memory usage tests, leak detection

3. **Backward Compatibility**:
   - Mitigation: Adapter pattern, versioned APIs
   - Monitoring: Integration tests with existing code

### User Experience Risks

1. **Learning Curve**:
   - Mitigation: Comprehensive documentation, examples
   - Validation: User testing, feedback collection

2. **Feature Completeness**:
   - Mitigation: Phased implementation, early feedback
   - Validation: Feature parity checklist

## Success Metrics

### Technical Metrics

- [ ] **Performance**: Handle 1000+ diff items with <100ms render time
- [ ] **Memory**: Stable memory usage under 50MB for typical diffs
- [ ] **Coverage**: 90%+ test coverage for core functionality
- [ ] **Compatibility**: Works with Go 1.19+ and all supported platforms

### Feature Metrics

- [ ] **Completeness**: Full feature parity with existing tfplandiff TUI
- [ ] **Extensibility**: 3+ working custom provider examples
- [ ] **Usability**: 5+ comprehensive usage examples
- [ ] **Documentation**: Complete API documentation with tutorials

### Adoption Metrics

- [ ] **Integration**: Seamless integration into existing bobatea ecosystem
- [ ] **Examples**: Working examples for 5+ different data types
- [ ] **Migration**: Successful migration of tfplandiff without breaking changes
- [ ] **Community**: Ready for open-source contributions

## Future Enhancements

### Post-V1 Features

1. **Advanced Visualizations**:
   - Tree view for hierarchical data
   - Graph view for relationship changes
   - Timeline view for sequential changes

2. **Collaboration Features**:
   - Comment system for changes
   - Approval workflows
   - Change annotations

3. **Integration Features**:
   - Git integration for file diffs
   - API integration for live data
   - Plugin system for custom extensions

4. **Advanced Export**:
   - PDF report generation
   - Interactive HTML reports
   - API documentation generation

## Conclusion

This design provides a comprehensive plan for implementing a reusable diff component that follows bobatea's established patterns while delivering powerful, generic diff visualization capabilities. The phased implementation approach ensures steady progress while maintaining quality and allowing for feedback incorporation.

The component will serve as both a standalone tool and a foundation for more specialized diff visualization needs across the Go ecosystem, particularly benefiting projects that need to visualize structured data changes in terminal applications.

Key success factors:
- **Strong Abstractions**: Clean interfaces enable broad reusability
- **Performance Focus**: Efficient algorithms and lazy loading handle large datasets
- **Rich UX**: Comprehensive features provide professional-grade experience
- **Excellent Documentation**: Lower barrier to adoption and contribution
- **Proven Patterns**: Following bobatea conventions ensures consistency and quality

The estimated 4-week implementation timeline is aggressive but achievable with focused development and regular milestone validation.
