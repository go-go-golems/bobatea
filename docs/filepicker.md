# Filepicker Documentation

The bobatea filepicker is a powerful, feature-rich file selection component for Bubble Tea applications, providing both compatibility with the original bubbles filepicker and an advanced mode with extended functionality.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [API Reference](#api-reference)
4. [Keyboard Shortcuts](#keyboard-shortcuts)
5. [Configuration](#configuration)
6. [Integration Examples](#integration-examples)
7. [Advanced Features](#advanced-features)
8. [Customization](#customization)
9. [Troubleshooting](#troubleshooting)
10. [Migration Guide](#migration-guide)

## Overview

The filepicker component provides two modes of operation:

### Compatibility Mode
A drop-in replacement for the original bubbles filepicker that maintains API compatibility while providing enhanced performance and features under the hood.

### Advanced Mode
A full-featured file manager with:
- **Multi-selection** - Select multiple files and directories
- **Directory selection** - Dedicated mode for choosing directories
- **File operations** - Copy, cut, paste, delete, rename, create
- **Search functionality** - Real-time file filtering
- **Glob pattern filtering** - Filter files using glob patterns (*.go, test_*, etc.)
- **Jail directory** - Restrict navigation to a specific directory tree
- **Preview panel** - View file contents and metadata
- **Multiple view modes** - Normal, detailed, and hidden file visibility
- **Sorting options** - Sort by name, size, date, or type
- **Navigation history** - Browser-like back/forward navigation
- **Keyboard shortcuts** - Comprehensive keyboard navigation

## Quick Start

### Compatibility Mode

For drop-in replacement of the original bubbles filepicker:

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/bobatea/pkg/filepicker"
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    filepicker filepicker.Model
}

func (m model) Init() tea.Cmd {
    return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case filepicker.SelectFileMsg:
        fmt.Printf("Selected file: %s\n", msg.Path)
        return m, tea.Quit
    case filepicker.CancelFilePickerMsg:
        fmt.Println("File picker cancelled")
        return m, tea.Quit
    }
    
    var cmd tea.Cmd
    m.filepicker, cmd = m.filepicker.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.filepicker.View()
}

func main() {
    m := model{
        filepicker: filepicker.NewModel(),
    }
    
    if _, err := tea.NewProgram(m).Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Advanced Mode (Unified Constructor)

For full-featured file management with the unified constructor:

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/bobatea/pkg/filepicker"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Create with comprehensive options
    picker := filepicker.New(
        filepicker.WithStartPath("."),
        filepicker.WithShowPreview(true),
        filepicker.WithShowHidden(false),
        filepicker.WithDirectorySelection(false), // for file selection
        filepicker.WithGlobPattern("*.go"),       // filter Go files
        filepicker.WithJailDirectory("/home/user/project"), // restrict to project dir
    )
    
    p := tea.NewProgram(picker, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Handle results
    selectedFiles, hasSelection := picker.GetSelected()
    if !hasSelection {
        fmt.Println("Selection cancelled")
        return
    }
    
    for _, file := range selectedFiles {
        fmt.Printf("Selected: %s\n", file)
    }
}
```

### Directory Selection Mode

For selecting directories instead of files:

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/bobatea/pkg/filepicker"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Create directory picker
    picker := filepicker.New(
        filepicker.WithStartPath("/home/user"),
        filepicker.WithDirectorySelection(true), // enable directory mode
        filepicker.WithShowPreview(true),
    )
    
    p := tea.NewProgram(picker, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    selectedDirs, hasSelection := picker.GetSelected()
    if hasSelection {
        fmt.Printf("Selected directories: %v\n", selectedDirs)
    }
}
```

## API Reference

### Types

#### File
```go
type File struct {
    Name     string        // File or directory name
    Path     string        // Full path to the file
    IsDir    bool          // True if this is a directory
    Size     int64         // File size in bytes
    ModTime  time.Time     // Last modification time
    Mode     os.FileMode   // File permissions
    Selected bool          // Whether file is selected
    Hidden   bool          // Whether file is hidden
}
```

#### ViewState
```go
type ViewState int

const (
    ViewStateNormal ViewState = iota
    ViewStateConfirmDelete
    ViewStateRename
    ViewStateCreateFile
    ViewStateCreateDir
    ViewStateSearch
)
```

#### SortMode
```go
type SortMode int

const (
    SortByName SortMode = iota
    SortBySize
    SortByDate
    SortByType
)
```

#### Operation
```go
type Operation int

const (
    OpNone Operation = iota
    OpCopy
    OpCut
)
```

### Messages

#### SelectFileMsg
```go
type SelectFileMsg struct {
    Path string
}
```
Sent when a file is selected in compatibility mode.

#### CancelFilePickerMsg
```go
type CancelFilePickerMsg struct{}
```
Sent when the file picker is cancelled.

### Compatibility Model

#### NewModel()
```go
func NewModel() Model
```
Creates a new compatibility mode file picker starting in the current working directory.

#### Methods
```go
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

### Advanced Model

#### NewAdvancedModel()
```go
func NewAdvancedModel(startPath string) *AdvancedModel
```
Creates a new advanced file picker starting at the specified path.

#### Methods

##### Core Methods
```go
func (fp *AdvancedModel) Init() tea.Cmd
func (fp *AdvancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (fp *AdvancedModel) View() string
```

##### State Query Methods
```go
func (fp *AdvancedModel) SelectedFiles() []string
func (fp *AdvancedModel) Cancelled() bool
func (fp *AdvancedModel) CurrentPath() string
func (fp *AdvancedModel) CurrentFile() *File
```

##### Configuration Methods
```go
func (fp *AdvancedModel) SetSize(width, height int)
func (fp *AdvancedModel) SetShowPreview(show bool)
func (fp *AdvancedModel) SetShowHidden(show bool)
func (fp *AdvancedModel) SetDetailedView(detailed bool)
func (fp *AdvancedModel) SetSortMode(mode SortMode)
```

##### Navigation Methods
```go
func (fp *AdvancedModel) NavigateTo(path string)
func (fp *AdvancedModel) GoBack()
func (fp *AdvancedModel) GoForward()
func (fp *AdvancedModel) CanGoBack() bool
func (fp *AdvancedModel) CanGoForward() bool
```

## Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Home` | Go to first item |
| `End` | Go to last item |
| `Enter` | Select file or enter directory |
| `Backspace` | Go up one directory |
| `Alt+←` / `h` | Navigate back in history |
| `Alt+→` / `l` | Navigate forward in history |

### Selection
| Key | Action |
|-----|--------|
| `Space` | Toggle selection of current item |
| `a` | Select all items |
| `A` | Deselect all items |
| `Ctrl+a` | Select all files (not directories) |

### File Operations
| Key | Action |
|-----|--------|
| `c` | Copy selected items |
| `x` | Cut selected items |
| `v` | Paste copied/cut items |
| `d` | Delete selected items |
| `r` | Rename current item |
| `n` | Create new file |
| `m` | Create new directory |

### View Controls
| Key | Action |
|-----|--------|
| `Tab` | Toggle preview panel / directory selection mode |
| `F2` | Toggle hidden files |
| `F3` | Toggle detailed view |
| `F4` | Cycle sort mode |
| `F5` | Refresh directory |
| `/` | Search files |
| `g` | Enter glob pattern filter |
| `G` | Clear glob filter |

### System
| Key | Action |
|-----|--------|
| `?` | Toggle help |
| `Esc` | Cancel operation or quit |
| `q` / `Ctrl+c` | Quit |

## Configuration

### Basic Configuration (Unified Constructor)

```go
// Create with comprehensive options
picker := filepicker.New(
    filepicker.WithStartPath("/home/user/documents"),
    filepicker.WithShowPreview(true),
    filepicker.WithShowHidden(false),
    filepicker.WithDirectorySelection(false),
    filepicker.WithDetailedView(true),
    filepicker.WithSortMode(filepicker.SortByName),
)
```

### Advanced Configuration Options

```go
// Complete configuration example
picker := filepicker.New(
    filepicker.WithStartPath("."),
    filepicker.WithShowPreview(true),
    filepicker.WithShowHidden(false),
    filepicker.WithShowIcons(true),
    filepicker.WithShowSizes(true),
    filepicker.WithDetailedView(true),
    filepicker.WithSortMode(filepicker.SortByDate),
    filepicker.WithPreviewWidth(40),
    filepicker.WithMaxHistorySize(100),
    filepicker.WithDirectorySelection(false),
    filepicker.WithGlobPattern("*.go"),
    filepicker.WithJailDirectory("/home/user/safe-area"),
)
```

### Available Configuration Options

| Option | Description |
|--------|-------------|
| `WithStartPath(string)` | Set starting directory |
| `WithShowPreview(bool)` | Enable/disable file preview panel |
| `WithShowHidden(bool)` | Show/hide hidden files |
| `WithShowIcons(bool)` | Show file type icons |
| `WithShowSizes(bool)` | Show file sizes |
| `WithDetailedView(bool)` | Enable detailed view mode |
| `WithSortMode(SortMode)` | Set initial sort mode |
| `WithPreviewWidth(int)` | Set preview panel width |
| `WithMaxHistorySize(int)` | Set navigation history size |
| `WithDirectorySelection(bool)` | Enable directory selection mode |
| `WithGlobPattern(string)` | Set initial glob filter pattern |
| `WithJailDirectory(string)` | Restrict navigation to directory tree |

## Integration Examples

### File Editor Integration

```go
type EditorModel struct {
    filepicker *filepicker.AdvancedModel
    editor     *editor.Model
    mode       string // "picker" or "editor"
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.mode {
    case "picker":
        var cmd tea.Cmd
        m.filepicker, cmd = m.filepicker.Update(msg)
        
        if files := m.filepicker.SelectedFiles(); len(files) > 0 {
            // Switch to editor mode
            m.mode = "editor"
            return m, m.editor.LoadFile(files[0])
        }
        
        return m, cmd
        
    case "editor":
        return m.editor.Update(msg)
    }
    
    return m, nil
}
```

### Multi-File Processor

```go
type ProcessorModel struct {
    filepicker *filepicker.AdvancedModel
    processor  *processor.Model
    files      []string
}

func (m ProcessorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            if files := m.filepicker.SelectedFiles(); len(files) > 0 {
                m.files = files
                return m, m.processor.ProcessFiles(files)
            }
        }
    }
    
    var cmd tea.Cmd
    m.filepicker, cmd = m.filepicker.Update(msg)
    return m, cmd
}
```

### Configuration File Manager

```go
type ConfigManager struct {
    filepicker *filepicker.AdvancedModel
    configDir  string
}

func NewConfigManager() *ConfigManager {
    configDir := filepath.Join(os.Getenv("HOME"), ".config", "myapp")
    
    return &ConfigManager{
        filepicker: filepicker.NewAdvancedModel(configDir),
        configDir:  configDir,
    }
}

func (cm *ConfigManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "ctrl+n" {
            // Create new config file
            return cm, cm.createNewConfig()
        }
    }
    
    var cmd tea.Cmd
    cm.filepicker, cmd = cm.filepicker.Update(msg)
    return cm, cmd
}
```

## Advanced Features

### Directory Selection Mode

The filepicker includes a dedicated directory selection mode for choosing directories instead of files:

```go
// Enable directory selection mode
picker := filepicker.New(
    filepicker.WithDirectorySelection(true),
    filepicker.WithStartPath("/home/user"),
)

// Or toggle at runtime
picker.SetDirectorySelectionMode(true)

// Check current mode
isDirectoryMode := picker.GetDirectorySelectionMode()
```

**Features in Directory Selection Mode:**
- **Enter** selects the current directory (instead of navigating into it)
- **Space** toggles selection on directories (not files)
- **Tab** toggles between file and directory selection modes
- Visual indicators show directories are selectable
- Status shows "Directory Selection Mode"

### Glob Pattern Filtering

Filter files using glob patterns for precise file matching:

```go
// Create with glob pattern
picker := filepicker.New(
    filepicker.WithGlobPattern("*.go"),     // Show only Go files
    filepicker.WithStartPath("."),
)

// Common glob patterns
picker := filepicker.New(
    filepicker.WithGlobPattern("test_*"),   // Files starting with "test_"
)

picker := filepicker.New(
    filepicker.WithGlobPattern("*.{js,ts}"), // JavaScript and TypeScript files
)
```

**Interactive Glob Filtering:**
- Press **g** to enter glob pattern mode
- Type your pattern (e.g., `*.md`, `test_*`, `*.{go,mod}`)
- Press **Enter** to apply the filter
- Press **G** to clear the glob filter

### Jail Directory (Security Restriction)

Restrict navigation to a specific directory tree for security:

```go
// Create with jail directory
picker := filepicker.New(
    filepicker.WithJailDirectory("/home/user/safe-area"),
    filepicker.WithStartPath("/home/user/safe-area/documents"),
)
```

**Jail Features:**
- **Navigation Restriction**: Users cannot navigate above the jail directory
- **Visual Indicators**: 
  - Status shows "Jailed" when restriction is active
  - Path display shows relative path from jail root (`[jail]/subdir` instead of absolute path)
  - ".." entry is hidden when at jail root
- **Complete Coverage**: All navigation methods respect the jail boundary:
  - Backspace key navigation
  - ".." directory entry  
  - History navigation (back/forward)
  - Direct path setting
- **Security**: Handles symlinks and prevents escape attempts
- **Graceful Handling**: If current directory is outside jail, automatically navigates to jail

#### Jail API Usage

```go
// Basic setup with jail
fp := filepicker.New(
    filepicker.WithJailDirectory("/home/user/safe-area"),
    filepicker.WithStartPath("/home/user/safe-area/documents"),
)

// Using the compatibility Model
fp := filepicker.NewModelWithOptions(
    filepicker.WithJailDirectory("/path/to/jail"),
    filepicker.WithShowPreview(true),
)
```

#### Security Considerations

- **Path Validation**: All paths are resolved to absolute form and validated
- **Symlink Handling**: Symlinks cannot be used to escape the jail
- **History Restriction**: Navigation history only contains paths within jail
- **Thread Safety**: Thread-safe for concurrent access (no shared mutable state)
- **Zero Performance Impact**: No performance cost when jail is not used

#### Example Scenarios

**Restrict to User Documents:**
```go
fp := filepicker.New(
    filepicker.WithJailDirectory("/home/user/Documents"),
)
```

**Temporary Directory Restriction:**
```go
tmpDir := "/tmp/app-sandbox"
os.MkdirAll(tmpDir, 0755)
fp := filepicker.New(
    filepicker.WithJailDirectory(tmpDir),
)
```

**Application Data Directory:**
```go
dataDir := filepath.Join(os.Getenv("HOME"), ".myapp", "data")
fp := filepicker.New(
    filepicker.WithJailDirectory(dataDir),
    filepicker.WithStartPath(filepath.Join(dataDir, "projects")),
)
```

### Multi-Selection

The filepicker supports selecting multiple files and directories:

```go
// Get all selected files after picker exits
selectedFiles, hasSelection := picker.GetSelected()

// In compatibility mode, use SelectedFiles()
selectedFiles := fp.SelectedFiles()
```

### File Operations

Built-in file operations with confirmation:

```go
// Copy files (Ctrl+C)
fp.CopyFiles([]string{"/path/to/file.txt"})

// Cut files (Ctrl+X)
fp.CutFiles([]string{"/path/to/file.txt"})

// Paste files (Ctrl+V)
fp.PasteFiles()

// Delete files (Del) - shows confirmation dialog
fp.DeleteFiles([]string{"/path/to/file.txt"})

// Rename file (F2)
fp.RenameFile("/path/to/old.txt", "new.txt")

// Create new file (Ctrl+N)
fp.CreateFile("newfile.txt")

// Create new directory (Ctrl+Shift+N)
fp.CreateDirectory("newdir")
```

### Search and Filtering

Real-time search and filtering capabilities:

```go
// Search files by name
fp.SetSearchQuery("*.go")

// Custom file filter
fp.SetFileFilter(func(file filepicker.File) bool {
    // Only show Go files
    return filepath.Ext(file.Name) == ".go"
})

// Filter by file type
fp.SetFileTypeFilter([]string{".txt", ".md", ".go"})

// Filter by size
fp.SetSizeFilter(1024, 1024*1024) // Between 1KB and 1MB
```

### Preview Panel

Rich file preview with content and metadata:

```go
// Toggle preview panel
fp.SetShowPreview(true)

// Configure preview width
fp.SetPreviewWidth(40)

// Custom preview handler
fp.SetPreviewHandler(func(file filepicker.File) string {
    if filepath.Ext(file.Name) == ".json" {
        return formatJSON(file.Path)
    }
    return filepicker.DefaultPreview(file)
})
```

### Navigation History

Browser-like navigation with history:

```go
// Navigate back
fp.GoBack()

// Navigate forward
fp.GoForward()

// Check navigation state
canGoBack := fp.CanGoBack()
canGoForward := fp.CanGoForward()

// Get history
history := fp.GetHistory()

// Clear history
fp.ClearHistory()
```

## Customization

### Styling

Customize the appearance using lipgloss styles:

```go
// Custom styles
fp.SetStyle(filepicker.StyleConfig{
    Border:        lipgloss.RoundedBorder(),
    BorderColor:   lipgloss.Color("62"),
    Selected:      lipgloss.NewStyle().Background(lipgloss.Color("62")),
    Directory:     lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
    File:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
    Hidden:        lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
    Preview:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
    PreviewTitle:  lipgloss.NewStyle().Bold(true),
})
```

### Key Bindings

Customize keyboard shortcuts:

```go
// Custom key bindings
fp.SetKeyMap(filepicker.KeyMap{
    Up:     key.NewBinding(key.WithKeys("up", "k")),
    Down:   key.NewBinding(key.WithKeys("down", "j")),
    Select: key.NewBinding(key.WithKeys("enter", "space")),
    Back:   key.NewBinding(key.WithKeys("h", "backspace")),
    // ... more bindings
})
```

### File Icons

Add custom file type icons:

```go
// Enable icons
fp.SetShowIcons(true)

// Custom icon provider
fp.SetIconProvider(func(file filepicker.File) string {
    if file.IsDir {
        return "📁"
    }
    
    ext := filepath.Ext(file.Name)
    switch ext {
    case ".go":
        return "🐹"
    case ".js":
        return "📜"
    case ".json":
        return "📋"
    default:
        return "📄"
    }
})
```

## Troubleshooting

### Common Issues

#### File Permission Errors
```go
// Handle permission errors
fp.SetErrorHandler(func(err error) {
    if os.IsPermission(err) {
        // Handle permission error
        log.Printf("Permission denied: %v", err)
    }
})
```

#### Large Directory Performance
```go
// Enable lazy loading for large directories
fp.SetLazyLoading(true)

// Set maximum files to display
fp.SetMaxFiles(1000)

// Enable pagination
fp.SetPagination(true, 100) // 100 files per page
```

#### Memory Usage
```go
// Limit preview content size
fp.SetPreviewMaxSize(1024 * 1024) // 1MB

// Disable preview for large files
fp.SetPreviewFilter(func(file filepicker.File) bool {
    return file.Size < 1024*1024 // Only preview files < 1MB
})
```

### Debug Mode

Enable debug mode for troubleshooting:

```go
fp.SetDebugMode(true)
```

This will show additional information about:
- File loading performance
- Memory usage
- Key binding conflicts
- Error details

## Migration Guide

### From bubbles/filepicker

The bobatea filepicker is designed as a drop-in replacement:

#### Before (bubbles/filepicker)
```go
import "github.com/charmbracelet/bubbles/filepicker"

type model struct {
    filepicker filepicker.Model
}

func initialModel() model {
    fp := filepicker.New()
    fp.AllowedTypes = []string{".txt", ".md"}
    fp.CurrentDirectory = "/home/user"
    
    return model{
        filepicker: fp,
    }
}
```

#### After (bobatea/filepicker)
```go
import "github.com/go-go-golems/bobatea/pkg/filepicker"

type model struct {
    filepicker filepicker.Model
}

func initialModel() model {
    fp := filepicker.NewModel()
    
    // Advanced configuration still available
    fp.SetFileTypeFilter([]string{".txt", ".md"})
    fp.NavigateTo("/home/user")
    
    return model{
        filepicker: fp,
    }
}
```

### API Changes

#### Message Types
- `filepicker.FileSelectedMsg` → `filepicker.SelectFileMsg`
- `filepicker.FileCancelledMsg` → `filepicker.CancelFilePickerMsg`

#### Configuration
- `AllowedTypes` → `SetFileTypeFilter()`
- `CurrentDirectory` → `NavigateTo()`
- `ShowHidden` → `SetShowHidden()`

#### Methods
- `SelectedFile()` → `SelectedFiles()[0]`
- `DidSelectFile()` → `len(SelectedFiles()) > 0`
- `DidSelectDisabledFile()` → Removed (not needed)

### Upgrading to Advanced Mode

To use the full advanced features:

```go
// Replace Model with AdvancedModel
fp := filepicker.NewAdvancedModel("/starting/path")

// Enable advanced features
fp.SetShowPreview(true)
fp.SetDetailedView(true)
fp.SetSortMode(filepicker.SortByDate)

// Handle multi-selection
if selectedFiles := fp.SelectedFiles(); len(selectedFiles) > 0 {
    // Process multiple files
    for _, file := range selectedFiles {
        processFile(file)
    }
}
```

### Performance Considerations

The new implementation provides better performance for large directories:

- **Lazy loading**: Files are loaded on-demand
- **Efficient filtering**: Search and filtering are optimized
- **Memory management**: Preview content is cached and limited
- **Async operations**: File operations don't block the UI

For best performance with large directories:
```go
fp.SetLazyLoading(true)
fp.SetMaxFiles(500)
fp.SetPreviewMaxSize(64 * 1024) // 64KB
```

---

This documentation covers the comprehensive features of the bobatea filepicker. For additional examples and advanced usage patterns, see the [examples directory](../examples/) and the [test files](../pkg/filepicker/filepicker_test.go).
