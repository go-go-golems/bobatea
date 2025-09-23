# Diff Component Documentation

A powerful, generic, and embeddable diff visualization component for Bubble Tea applications that supports pluggable data providers, custom renderers, theming, and advanced filtering.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Installation and Quick Start](#installation-and-quick-start)
4. [Data Provider Interface](#data-provider-interface)
5. [Configuration](#configuration)
6. [Theming and Styling](#theming-and-styling)
7. [Embedding Examples](#embedding-examples)
8. [Advanced Usage](#advanced-usage)
9. [Built-in Features](#built-in-features)
10. [Message System](#message-system)
11. [Best Practices](#best-practices)
12. [API Reference](#api-reference)

## Overview

The bobatea diff component provides a fully-featured, customizable diff visualization interface that can be embedded in any Bubble Tea application. It follows a pluggable architecture where data providers can be swapped out to support different types of structured data comparisons.

### Key Features

- **Generic data provider interface** - Works with any structured data type
- **Configurable behavior** - Customizable layout, filters, and navigation
- **Real-time search and filtering** - Find specific changes across large diffs
- **Multiple view modes** - List view, detail view, and split-pane layouts
- **Category-based organization** - Group related changes for better navigation
- **Custom rendering system** - Pluggable renderers for different data types
- **Multiple themes** - Built-in themes (default, dark, light) and custom styling
- **Embeddable design** - Clean message-based API for integration
- **Keyboard shortcuts** - Comprehensive keyboard navigation
- **Export capabilities** - Export diffs in various formats

## Architecture

The diff system is built around several key components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Model       │    │  DataProvider   │    │   Renderer      │
│  (UI State)     │────│   (Interface)   │    │  (Interface)    │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └─────────────────────────────────────────────────┘
                                 │
              ┌─────────────────────────────────────────┐
              │                                         │
    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
    │    Filters      │    │     Search      │    │     Styles      │
    │   (Component)   │    │   (Component)   │    │   (Theming)     │
    │                 │    │                 │    │                 │
    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Components

#### model.go
- **Main diff model** - Manages UI state and coordination
- **Layout management** - Handles list/detail split-pane layout
- **Event handling** - Processes keyboard input and navigation
- **State management** - Tracks selection, focus, and view modes

#### provider.go
- **DataProvider interface** - Defines the contract for data sources
- **DiffItem interface** - Represents individual diff items
- **Change interface** - Represents individual changes within items
- **Category system** - Groups changes by type/significance

#### renderer.go
- **Renderer interface** - Defines custom rendering contracts
- **Built-in renderers** - JSON, YAML, generic object renderers
- **Formatting system** - Value formatting and display logic

#### messages.go
- **Message types** - Tea messages for diff communication
- **Event definitions** - Selection change, filter update, export messages
- **Inter-component communication** - Clean message-based API

#### filters.go
- **Filtering system** - Status, category, and custom filters
- **Search functionality** - Multi-field search across diff content
- **Filter persistence** - Save and restore filter states

#### styles.go
- **Styling system** - Lipgloss-based styling configuration
- **Theme support** - Multiple built-in themes
- **Customization** - Custom style configuration

## Installation and Quick Start

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/diff"
)

// Simple JSON diff provider
type JSONDiffProvider struct {
	items []diff.DiffItem
}

func NewJSONDiffProvider(before, after map[string]interface{}) *JSONDiffProvider {
	items := computeJSONDiff(before, after)
	return &JSONDiffProvider{items: items}
}

func (p *JSONDiffProvider) GetItems() []diff.DiffItem {
	return p.items
}

func (p *JSONDiffProvider) GetTitle() string {
	return "JSON Diff"
}

func (p *JSONDiffProvider) GetSummary() string {
	return fmt.Sprintf("%d items changed", len(p.items))
}

func main() {
	// Create sample data
	before := map[string]interface{}{
		"name":    "old-value",
		"version": 1,
		"config":  map[string]interface{}{
			"debug": false,
		},
	}
	
	after := map[string]interface{}{
		"name":    "new-value",
		"version": 2,
		"config":  map[string]interface{}{
			"debug": true,
			"log_level": "info",
		},
	}
	
	// Create provider and configuration
	provider := NewJSONDiffProvider(before, after)
	config := diff.DefaultConfig()
	config.Title = "JSON Configuration Diff"
	
	// Create and run the diff viewer
	model := diff.NewModel(provider, config)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
```

### Advanced Configuration

```go
func main() {
	provider := &MyDataProvider{}
	
	// Comprehensive configuration
	config := diff.Config{
		Title:               "Database Schema Diff",
		Width:               120,
		Height:              40,
		ShowSummary:         true,
		EnableSearch:        true,
		EnableFilters:       true,
		EnableExport:        true,
		StartInDetailView:   false,
		DefaultCategoryFilter: []string{"tables", "indexes"},
		SplitPaneRatio:      0.4, // 40% for list, 60% for detail
	}
	
	model := diff.NewModel(provider, config)
	
	// Apply custom theme
	model.SetTheme(diff.BuiltinThemes["dark"])
	
	// Add custom renderer
	renderer := &SchemaRenderer{
		ShowTypes:     true,
		ShowConstraints: true,
		HighlightSecurity: true,
	}
	model.SetRenderer(renderer)
	
	// Add custom filters
	model.AddFilter("security", func(item diff.DiffItem) bool {
		return containsSecurityChanges(item)
	})
	
	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Data Provider Interface

The `DataProvider` interface is the heart of the diff system, allowing you to plug in any data source:

```go
// DataProvider provides diff data to the component
type DataProvider interface {
	// GetItems returns all diff items
	GetItems() []DiffItem
	
	// GetTitle returns the overall title for the diff
	GetTitle() string
	
	// GetSummary returns a summary description
	GetSummary() string
}

// DiffItem represents a single item that contains changes
type DiffItem interface {
	// GetID returns a unique identifier for this item
	GetID() string
	
	// GetDisplayName returns the name to show in the list
	GetDisplayName() string
	
	// GetDescription returns a brief description
	GetDescription() string
	
	// GetCategories returns the categories of changes in this item
	GetCategories() []Category
	
	// GetActions returns the high-level actions (create, update, delete)
	GetActions() []Action
	
	// GetMetadata returns additional metadata
	GetMetadata() map[string]interface{}
}

// Category represents a group of related changes
type Category interface {
	// GetName returns the category name (e.g., "environment", "attributes")
	GetName() string
	
	// GetDisplayName returns the human-readable name
	GetDisplayName() string
	
	// GetChanges returns all changes in this category
	GetChanges() []Change
	
	// IsVisible returns whether this category should be shown
	IsVisible() bool
	
	// GetIcon returns an optional icon for display
	GetIcon() string
}

// Change represents a single change within a category
type Change interface {
	// GetPath returns the path to the changed field
	GetPath() string
	
	// GetStatus returns the change status
	GetStatus() ChangeStatus
	
	// GetBeforeValue returns the value before the change
	GetBeforeValue() interface{}
	
	// GetAfterValue returns the value after the change
	GetAfterValue() interface{}
	
	// GetMetadata returns change-specific metadata
	GetMetadata() map[string]interface{}
	
	// IsSensitive returns true if this change contains sensitive data
	IsSensitive() bool
}

// ChangeStatus represents the type of change
type ChangeStatus string

const (
	StatusAdded   ChangeStatus = "added"
	StatusUpdated ChangeStatus = "updated"
	StatusRemoved ChangeStatus = "removed"
	StatusMoved   ChangeStatus = "moved"
)

// Action represents a high-level action on an item
type Action string

const (
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionReplace Action = "replace"
)
```

### Creating Custom Data Providers

#### Terraform Plan Provider

```go
type TerraformProvider struct {
	resources []diff.DiffItem
}

func NewTerraformProvider(planFile string) (*TerraformProvider, error) {
	plan, err := loadTerraformPlan(planFile)
	if err != nil {
		return nil, err
	}
	
	resources := make([]diff.DiffItem, 0, len(plan.ResourceChanges))
	for _, rc := range plan.ResourceChanges {
		item := &TerraformResource{
			address: rc.Address,
			actions: rc.Change.Actions,
			before:  rc.Change.Before,
			after:   rc.Change.After,
		}
		resources = append(resources, item)
	}
	
	return &TerraformProvider{resources: resources}, nil
}

func (p *TerraformProvider) GetItems() []diff.DiffItem {
	return p.resources
}

func (p *TerraformProvider) GetTitle() string {
	return "Terraform Plan"
}

func (p *TerraformProvider) GetSummary() string {
	return fmt.Sprintf("%d resource changes", len(p.resources))
}

type TerraformResource struct {
	address string
	actions []string
	before  interface{}
	after   interface{}
}

func (r *TerraformResource) GetID() string {
	return r.address
}

func (r *TerraformResource) GetDisplayName() string {
	return r.address
}

func (r *TerraformResource) GetDescription() string {
	return strings.Join(r.actions, ", ")
}

func (r *TerraformResource) GetCategories() []diff.Category {
	return []diff.Category{
		&TerraformCategory{
			name:    "configuration",
			changes: computeConfigChanges(r.before, r.after),
		},
	}
}

func (r *TerraformResource) GetActions() []diff.Action {
	actions := make([]diff.Action, len(r.actions))
	for i, action := range r.actions {
		actions[i] = diff.Action(action)
	}
	return actions
}

func (r *TerraformResource) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"resource_type": extractResourceType(r.address),
	}
}
```

#### Database Schema Provider

```go
type SchemaProvider struct {
	tables []diff.DiffItem
}

func NewSchemaProvider(beforeSchema, afterSchema *Schema) *SchemaProvider {
	tables := computeSchemaDiff(beforeSchema, afterSchema)
	return &SchemaProvider{tables: tables}
}

func (p *SchemaProvider) GetItems() []diff.DiffItem {
	return p.tables
}

func (p *SchemaProvider) GetTitle() string {
	return "Database Schema Changes"
}

func (p *SchemaProvider) GetSummary() string {
	return fmt.Sprintf("%d table changes", len(p.tables))
}

type SchemaTable struct {
	name         string
	columnChanges []diff.Change
	indexChanges  []diff.Change
	constraintChanges []diff.Change
}

func (t *SchemaTable) GetCategories() []diff.Category {
	categories := []diff.Category{}
	
	if len(t.columnChanges) > 0 {
		categories = append(categories, &SchemaCategory{
			name:    "columns",
			display: "Columns",
			changes: t.columnChanges,
			icon:    "📊",
		})
	}
	
	if len(t.indexChanges) > 0 {
		categories = append(categories, &SchemaCategory{
			name:    "indexes",
			display: "Indexes",
			changes: t.indexChanges,
			icon:    "🔍",
		})
	}
	
	if len(t.constraintChanges) > 0 {
		categories = append(categories, &SchemaCategory{
			name:    "constraints",
			display: "Constraints",
			changes: t.constraintChanges,
			icon:    "🔗",
		})
	}
	
	return categories
}
```

#### Git Diff Provider

```go
type GitProvider struct {
	files []diff.DiffItem
}

func NewGitProvider(repoPath, fromCommit, toCommit string) (*GitProvider, error) {
	gitDiff, err := getGitDiff(repoPath, fromCommit, toCommit)
	if err != nil {
		return nil, err
	}
	
	files := make([]diff.DiffItem, 0, len(gitDiff.Files))
	for _, file := range gitDiff.Files {
		item := &GitFile{
			path:   file.Path,
			status: file.Status,
			hunks:  file.Hunks,
		}
		files = append(files, item)
	}
	
	return &GitProvider{files: files}, nil
}

type GitFile struct {
	path   string
	status string
	hunks  []GitHunk
}

func (f *GitFile) GetCategories() []diff.Category {
	return []diff.Category{
		&GitCategory{
			name:    "changes",
			display: "Line Changes",
			changes: f.convertHunksToChanges(),
			icon:    "📝",
		},
	}
}
```

## Configuration

The `Config` struct provides comprehensive configuration options:

```go
type Config struct {
	Title                string   // Title displayed at the top
	Width                int      // Component width
	Height               int      // Component height
	ShowSummary          bool     // Show summary information
	EnableSearch         bool     // Enable search functionality
	EnableFilters        bool     // Enable filtering
	EnableExport         bool     // Enable export functionality
	StartInDetailView    bool     // Start with detail view focused
	DefaultCategoryFilter []string // Default categories to show
	SplitPaneRatio       float64  // Ratio for list vs detail pane
	RedactSensitive      bool     // Redact sensitive values
	MaxItemsInList       int      // Maximum items to show in list
}
```

### Configuration Examples

#### Minimal Configuration

```go
config := diff.Config{
	Title: "My Diff",
}
```

#### Complete Configuration

```go
config := diff.Config{
	Title:                "Advanced Diff Viewer",
	Width:                140,
	Height:              50,
	ShowSummary:         true,
	EnableSearch:        true,
	EnableFilters:       true,
	EnableExport:        true,
	StartInDetailView:   false,
	DefaultCategoryFilter: []string{"critical", "high"},
	SplitPaneRatio:      0.35,
	RedactSensitive:     true,
	MaxItemsInList:      1000,
}
```

#### Default Configuration

```go
config := diff.DefaultConfig()
// Returns:
// Config{
//     Title:                "Diff Viewer",
//     Width:                100,
//     Height:              30,
//     ShowSummary:          true,
//     EnableSearch:         true,
//     EnableFilters:        true,
//     EnableExport:         false,
//     StartInDetailView:    false,
//     DefaultCategoryFilter: []string{},
//     SplitPaneRatio:       0.4,
//     RedactSensitive:      false,
//     MaxItemsInList:       500,
// }
```

## Theming and Styling

The diff component provides a comprehensive theming system with built-in themes and custom styling support.

### Built-in Themes

```go
// Apply built-in themes
model.SetTheme(diff.BuiltinThemes["default"])
model.SetTheme(diff.BuiltinThemes["dark"])
model.SetTheme(diff.BuiltinThemes["light"])
model.SetTheme(diff.BuiltinThemes["monochrome"])
```

### Custom Styling

#### Creating Custom Styles

```go
customStyles := diff.Styles{
	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("24")).
		Padding(0, 1),
	
	Summary: lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true),
	
	ListBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")),
	
	DetailBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")),
	
	SelectedItem: lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33")).
		Bold(true),
	
	Added: lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")),
	
	Removed: lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")),
	
	Updated: lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")),
	
	SensitiveValue: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true),
}

model.SetStyles(customStyles)
```

#### Status Color Customization

```go
statusColors := diff.StatusColors{
	Added:   lipgloss.Color("46"),   // Bright green
	Removed: lipgloss.Color("196"),  // Bright red
	Updated: lipgloss.Color("226"),  // Bright yellow
	Moved:   lipgloss.Color("51"),   // Bright cyan
}

model.SetStatusColors(statusColors)
```

## Embedding Examples

The diff component is designed to be embedded in larger applications using Bubble Tea's message system.

### Basic Embedding

```go
type AppModel struct {
	diff     diff.Model
	mode     string // "diff" or "other"
	otherUI  SomeOtherComponent
}

func (m AppModel) Init() tea.Cmd {
	return m.diff.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case diff.SelectionChangedMsg:
		// Handle diff selection changes
		return m.handleDiffSelection(msg)
	
	case diff.FilterChangedMsg:
		// Handle filter changes
		return m.handleFilterChange(msg)
	
	case diff.ExportRequestMsg:
		// Handle export requests
		return m.handleExportRequest(msg)
	
	case tea.KeyMsg:
		if msg.String() == "tab" {
			// Switch between diff and other modes
			if m.mode == "diff" {
				m.mode = "other"
			} else {
				m.mode = "diff"
			}
		}
	}
	
	// Route to appropriate component
	if m.mode == "diff" {
		var cmd tea.Cmd
		m.diff, cmd = m.diff.Update(msg)
		return m, cmd
	} else {
		var cmd tea.Cmd
		m.otherUI, cmd = m.otherUI.Update(msg)
		return m, cmd
	}
}

func (m AppModel) View() string {
	if m.mode == "diff" {
		return m.diff.View()
	}
	return m.otherUI.View()
}
```

### Multi-Tab Application

```go
type MultiTabApp struct {
	diffViewer   diff.Model
	fileManager  FileManagerModel
	terminal     TerminalModel
	activeTab    int
	tabs         []string
}

func NewMultiTabApp() MultiTabApp {
	return MultiTabApp{
		tabs: []string{"Files", "Diff", "Terminal"},
		activeTab: 1, // Start with diff tab
	}
}

func (m MultiTabApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case diff.SelectionChangedMsg:
		// Show selected file in file manager
		return m.showFileInManager(msg.Item)
	
	case FileSelectedMsg:
		// Load file diff into diff viewer
		return m.loadFileDiff(msg.Path)
	
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+1":
			m.activeTab = 0 // Files
		case "ctrl+2":
			m.activeTab = 1 // Diff
		case "ctrl+3":
			m.activeTab = 2 // Terminal
		}
	}
	
	// Update active tab
	var cmd tea.Cmd
	switch m.activeTab {
	case 0:
		m.fileManager, cmd = m.fileManager.Update(msg)
	case 1:
		m.diffViewer, cmd = m.diffViewer.Update(msg)
	case 2:
		m.terminal, cmd = m.terminal.Update(msg)
	}
	
	return m, cmd
}

func (m MultiTabApp) View() string {
	// Render tab headers
	tabs := m.renderTabs()
	
	// Render active tab content
	var content string
	switch m.activeTab {
	case 0:
		content = m.fileManager.View()
	case 1:
		content = m.diffViewer.View()
	case 2:
		content = m.terminal.View()
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, tabs, content)
}
```

### Development Environment Integration

```go
type IDEModel struct {
	diffViewer   diff.Model
	codeEditor   EditorModel
	projectTree  TreeModel
	debugger     DebuggerModel
	layout       LayoutManager
}

func (ide IDEModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case diff.SelectionChangedMsg:
		// Open changed file in editor
		return ide.openFileInEditor(msg.Item)
	
	case CodeEditedMsg:
		// Update diff if file was changed
		return ide.refreshDiffForFile(msg.File)
	
	case diff.ExportRequestMsg:
		// Export diff to patch file
		return ide.exportToPatch(msg.Format)
	}
	
	// Update all components
	var cmds []tea.Cmd
	var cmd tea.Cmd
	
	ide.diffViewer, cmd = ide.diffViewer.Update(msg)
	cmds = append(cmds, cmd)
	
	ide.codeEditor, cmd = ide.codeEditor.Update(msg)
	cmds = append(cmds, cmd)
	
	ide.projectTree, cmd = ide.projectTree.Update(msg)
	cmds = append(cmds, cmd)
	
	return ide, tea.Batch(cmds...)
}
```

## Advanced Usage

### Custom Renderers

Create custom renderers for specialized data types:

```go
// JSON renderer with syntax highlighting
type JSONRenderer struct {
	IndentSize    int
	ShowTypes     bool
	ColorScheme   map[string]lipgloss.Color
}

func (r *JSONRenderer) RenderChange(change diff.Change, opts diff.RenderOptions) string {
	before := r.formatJSON(change.GetBeforeValue())
	after := r.formatJSON(change.GetAfterValue())
	
	if opts.RedactSensitive && change.IsSensitive() {
		before = r.redactValue(before)
		after = r.redactValue(after)
	}
	
	return r.renderSideBySide(before, after, change.GetStatus())
}

func (r *JSONRenderer) formatJSON(value interface{}) string {
	data, _ := json.MarshalIndent(value, "", strings.Repeat(" ", r.IndentSize))
	return r.applySyntaxHighlighting(string(data))
}

// Database schema renderer
type SchemaRenderer struct {
	ShowConstraints   bool
	ShowIndexes       bool
	HighlightSecurity bool
}

func (r *SchemaRenderer) RenderChange(change diff.Change, opts diff.RenderOptions) string {
	switch change.GetPath() {
	case "columns":
		return r.renderColumnChange(change, opts)
	case "indexes":
		return r.renderIndexChange(change, opts)
	case "constraints":
		return r.renderConstraintChange(change, opts)
	default:
		return r.renderGenericChange(change, opts)
	}
}

// Code diff renderer with syntax highlighting
type CodeRenderer struct {
	Language      string
	ShowLineNumbers bool
	TabSize        int
}

func (r *CodeRenderer) RenderChange(change diff.Change, opts diff.RenderOptions) string {
	beforeLines := strings.Split(change.GetBeforeValue().(string), "\n")
	afterLines := strings.Split(change.GetAfterValue().(string), "\n")
	
	return r.renderUnifiedDiff(beforeLines, afterLines, opts)
}
```

### Custom Filters

```go
// Security-focused filter
model.AddFilter("security", func(item diff.DiffItem) bool {
	for _, category := range item.GetCategories() {
		for _, change := range category.GetChanges() {
			if isSecurityRelated(change.GetPath()) {
				return true
			}
		}
	}
	return false
})

// Performance impact filter
model.AddFilter("performance", func(item diff.DiffItem) bool {
	metadata := item.GetMetadata()
	if impact, ok := metadata["performance_impact"]; ok {
		return impact.(string) != "none"
	}
	return false
})

// Size-based filter
model.AddFilter("large-changes", func(item diff.DiffItem) bool {
	changeCount := 0
	for _, category := range item.GetCategories() {
		changeCount += len(category.GetChanges())
	}
	return changeCount > 10
})
```

### Export Functionality

```go
// Export to JSON
model.AddExporter("json", func(items []diff.DiffItem, opts diff.ExportOptions) ([]byte, error) {
	export := struct {
		Timestamp string              `json:"timestamp"`
		Title     string              `json:"title"`
		Summary   string              `json:"summary"`
		Items     []diff.DiffItem     `json:"items"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Title:     opts.Title,
		Summary:   opts.Summary,
		Items:     items,
	}
	
	return json.MarshalIndent(export, "", "  ")
})

// Export to HTML report
model.AddExporter("html", func(items []diff.DiffItem, opts diff.ExportOptions) ([]byte, error) {
	tmpl := template.Must(template.New("report").Parse(htmlTemplate))
	
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct {
		Title   string
		Summary string
		Items   []diff.DiffItem
		Options diff.ExportOptions
	}{
		Title:   opts.Title,
		Summary: opts.Summary,
		Items:   items,
		Options: opts,
	})
	
	return buf.Bytes(), err
})

// Export to unified diff format
model.AddExporter("patch", func(items []diff.DiffItem, opts diff.ExportOptions) ([]byte, error) {
	var buf bytes.Buffer
	
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("--- %s\n", item.GetDisplayName()))
		buf.WriteString(fmt.Sprintf("+++ %s\n", item.GetDisplayName()))
		
		for _, category := range item.GetCategories() {
			for _, change := range category.GetChanges() {
				buf.WriteString(renderUnifiedChange(change))
			}
		}
	}
	
	return buf.Bytes(), nil
})
```

## Built-in Features

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `↑/k` | Move up in list |
| `↓/j` | Move down in list |
| `Tab` | Switch focus between panes |
| `Enter` | Select item / Toggle detail |
| `/` | Start search |
| `Esc` | Clear search / Exit modes |
| `f` | Toggle filters panel |
| `e` | Export current view |
| `r` | Toggle sensitive data redaction |
| `1/2/3` | Toggle status filters (added/updated/removed) |
| `Ctrl+C` | Quit |

### Search Functionality

```go
// Multi-field search across all diff content
// Searches in:
// - Item display names
// - Change paths
// - Before/after values
// - Metadata

// Example search queries:
// "password"           - Find all password-related changes
// "status:added"       - Find only added items
// "category:env"       - Find only environment changes
// "path:config.*"      - Find changes in config paths (regex)
// "value:prod"         - Find changes involving "prod" values
```

### Filtering System

```go
// Built-in filters
- Status filters: added, updated, removed, moved
- Category filters: env, config, security, etc.
- Sensitivity filter: show/hide sensitive changes
- Size filters: large changes, small changes

// Custom filters can be added programmatically
model.AddFilter("name", filterFunction)
```

### Navigation Features

- Two-pane layout with resizable split
- Focus management between list and detail
- Keyboard and mouse navigation
- Context-sensitive help
- Breadcrumb navigation for nested data

## Message System

The diff component communicates through a clean message-based API:

### Message Types

```go
// Selection changed
type SelectionChangedMsg struct {
	Item     diff.DiffItem
	Category diff.Category
	Change   diff.Change
}

// Filter state changed
type FilterChangedMsg struct {
	FilterName string
	Enabled    bool
	Query      string
}

// Export requested
type ExportRequestMsg struct {
	Format  string
	Items   []diff.DiffItem
	Options diff.ExportOptions
}

// Search query changed
type SearchChangedMsg struct {
	Query   string
	Results []diff.DiffItem
}

// View mode changed
type ViewModeChangedMsg struct {
	Mode string // "list", "detail", "split"
}
```

### Message Handling

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case diff.SelectionChangedMsg:
		// Handle item selection
		m.selectedItem = msg.Item
		return m.updateRelatedComponents(msg.Item)
	
	case diff.FilterChangedMsg:
		// Handle filter changes
		return m.updateFilterState(msg)
	
	case diff.ExportRequestMsg:
		// Handle export request
		return m.performExport(msg)
	
	case diff.SearchChangedMsg:
		// Handle search results
		return m.updateSearchResults(msg)
	}
	
	// Forward to diff component
	var cmd tea.Cmd
	m.diff, cmd = m.diff.Update(msg)
	return m, cmd
}
```

## Best Practices

### Performance

1. **Lazy Loading**: Load large diffs progressively
2. **Virtualization**: Use virtual scrolling for large lists
3. **Debounced Search**: Debounce search input to avoid excessive filtering
4. **Caching**: Cache rendered content for better performance

```go
// Example: Lazy loading provider
type LazyProvider struct {
	loader   func(offset, limit int) ([]diff.DiffItem, error)
	cache    map[int][]diff.DiffItem
	pageSize int
}

func (p *LazyProvider) GetItems() []diff.DiffItem {
	// Load items on demand
	return p.loadPage(0)
}
```

### Memory Management

1. **Limit Data**: Set reasonable limits on displayed data
2. **Clean Up**: Clear unused cached data
3. **Efficient Rendering**: Avoid unnecessary re-renders

### User Experience

1. **Progressive Disclosure**: Show summary first, details on demand
2. **Clear Status**: Use clear visual indicators for change types
3. **Helpful Errors**: Provide actionable error messages
4. **Responsive Design**: Adapt to different terminal sizes

### Security

1. **Data Sanitization**: Sanitize displayed data
2. **Sensitive Data**: Properly handle sensitive information
3. **Export Safety**: Validate export data before writing

## API Reference

### Types

```go
// Main diff model
type Model struct {
	// ... (see source for full definition)
}

// Configuration
type Config struct {
	Title                string
	Width                int
	Height               int
	ShowSummary          bool
	EnableSearch         bool
	EnableFilters        bool
	EnableExport         bool
	StartInDetailView    bool
	DefaultCategoryFilter []string
	SplitPaneRatio       float64
	RedactSensitive      bool
	MaxItemsInList       int
}

// Styling
type Styles struct {
	Title        lipgloss.Style
	Summary      lipgloss.Style
	ListBorder   lipgloss.Style
	DetailBorder lipgloss.Style
	SelectedItem lipgloss.Style
	Added        lipgloss.Style
	Removed      lipgloss.Style
	Updated      lipgloss.Style
	SensitiveValue lipgloss.Style
}
```

### Functions

```go
// Create new diff model
func NewModel(provider DataProvider, config Config) Model

// Create default configuration
func DefaultConfig() Config

// Create default styles
func DefaultStyles() Styles
```

### Model Methods

```go
// Bubble Tea interface
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string

// Configuration
func (m *Model) SetTheme(theme Theme)
func (m *Model) SetStyles(styles Styles)
func (m *Model) SetRenderer(renderer Renderer)
func (m *Model) SetSize(width, height int)

// Filtering and Search
func (m *Model) AddFilter(name string, filter func(diff.DiffItem) bool)
func (m *Model) RemoveFilter(name string)
func (m *Model) SetSearch(query string)
func (m *Model) ClearSearch()

// Export
func (m *Model) AddExporter(format string, exporter ExportFunc)
func (m *Model) Export(format string, options ExportOptions) ([]byte, error)

// State
func (m *Model) GetSelectedItem() diff.DiffItem
func (m *Model) GetFilteredItems() []diff.DiffItem
func (m *Model) GetViewMode() string
func (m *Model) SetViewMode(mode string)
```

### Built-in Themes

```go
var BuiltinThemes = map[string]Theme{
	"default":    Theme{...},
	"dark":       Theme{...},
	"light":      Theme{...},
	"monochrome": Theme{...},
}
```

---

This documentation provides a comprehensive guide to using the bobatea diff component. For more examples and advanced usage patterns, see the [examples directory](../examples/diff/) and the [package source code](../pkg/diff/).
