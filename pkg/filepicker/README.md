# File Picker Components

The `filepicker` package provides an advanced file picker implementation that completely replaces the original basic file picker while maintaining backward compatibility.

## File Picker (`Model`)

The file picker is a powerful component with both basic compatibility and advanced features.

### Features:
- **Full backward compatibility** with existing bobatea filepicker API
- **Multi-file selection** with visual indicators
- **File operations**: copy, cut, paste, delete, rename
- **Search functionality** with real-time filtering  
- **File preview** with content detection for text files
- **Multiple sort modes**: name, size, date, type
- **Navigation history** with back/forward buttons
- **Hidden file toggle**
- **Detailed/compact view modes**
- **Extensive keyboard shortcuts**
- **Single or multi-file selection modes**

### Compatibility Usage (Recommended for existing code):
```go
import "github.com/go-go-golems/bobatea/pkg/filepicker"

// Create using the existing API - this still works!
fp := filepicker.NewModel()
fp.Filepicker.DirAllowed = false
fp.Filepicker.FileAllowed = true
fp.Filepicker.CurrentDirectory = "/home/user"
fp.Filepicker.Height = 10

// Use in a bubbletea program with message handling
program := tea.NewProgram(fp)

// Handle messages the same way as before
switch msg := msg.(type) {
case filepicker.SelectFileMsg:
    selectedPath := msg.Path
case filepicker.CancelFilePickerMsg:
    // Handle cancellation
}
```

### Advanced Usage (For new code):
```go
import "github.com/go-go-golems/bobatea/pkg/filepicker"

// Create advanced file picker directly
picker := filepicker.NewAdvancedModel(".")
picker.SetShowPreview(true)
picker.SetShowHidden(false)

// Use in a bubbletea program
program := tea.NewProgram(picker)

// After the program exits, check results
selectedFiles, hasSelection := picker.GetSelected()
if hasSelection {
    // Handle selected files
}
```

### Keyboard Shortcuts:

#### Navigation:
- `↑/k` - Move up
- `↓/j` - Move down  
- `home` - Go to first item
- `end` - Go to last item
- `enter` - Select/enter directory
- `backspace` - Go up one directory
- `alt+←/h` - Go back in history
- `alt+→/l` - Go forward in history

#### Selection:
- `space` - Toggle selection
- `a` - Select all
- `A` - Deselect all
- `ctrl+a` - Select all files (not directories)

#### File Operations:
- `c` - Copy selected files
- `x` - Cut selected files
- `v` - Paste files
- `d` - Delete selected files
- `r` - Rename current file
- `n` - Create new file
- `m` - Create new directory

#### View Options:
- `tab` - Toggle file preview
- `f2` - Toggle hidden files
- `f3` - Toggle detailed view
- `f4` - Cycle sort mode
- `/` - Search files
- `f5` - Refresh directory

#### System:
- `?` - Toggle help
- `q/ctrl+c` - Quit
- `esc` - Cancel/clear search

### Configuration Methods:

- `SetShowPreview(bool)` - Enable/disable file preview
- `SetShowHidden(bool)` - Show/hide hidden files
- `GetSelected() ([]string, bool)` - Get selected files after picker exits
- `GetError() error` - Get any error that occurred

### Example:

See `cmd/filepicker-demo/main.go` for a complete example of how to use the file picker.

```bash
# Run the basic compatibility demo
go run cmd/filepicker-demo/main.go

# Run the advanced demo
go run cmd/filepicker-demo/main.go -advanced -path /some/directory
```

## Migration Guide

**No migration needed!** Existing code using the old bobatea filepicker API will continue to work without any changes. The new implementation provides the same messages (`SelectFileMsg`, `CancelFilePickerMsg`) and the same `Model` structure.

### What's Different:
- Much more powerful UI with multi-selection, search, and file operations
- Better performance and more responsive interface
- Additional features available through the advanced API
- Same backward-compatible API as before

## Integration with Larger Applications

The file picker can be easily embedded in larger applications by handling its messages:

```go
// Example of embedding in a larger model
type AppModel struct {
    filePicker filepicker.Model
    showPicker bool
    selectedFiles []string
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case filepicker.SelectFileMsg:
        m.selectedFiles = []string{msg.Path}
        m.showPicker = false
        return m, nil
    case filepicker.CancelFilePickerMsg:
        m.showPicker = false
        return m, nil
    }
    
    if m.showPicker {
        // Delegate to file picker
        updatedPicker, cmd := m.filePicker.Update(msg)
        m.filePicker = updatedPicker.(filepicker.Model)
        return m, cmd
    }
    // Handle other app logic
    return m, nil
}
```

## Advanced Features

When using `NewAdvancedModel()` directly, you get access to additional features:

- **Multi-selection**: Use `space` to select multiple files
- **File operations**: Copy (`c`), cut (`x`), paste (`v`), delete (`d`)
- **Search**: Press `/` to search files
- **Navigation history**: Use `alt+←` and `alt+→` to go back/forward
- **Sort modes**: Press `F4` to cycle through sort modes
- **File preview**: Press `tab` to toggle preview panel
- **Hidden files**: Press `F2` to toggle hidden file visibility

## Troubleshooting

### Common Issues:

1. **Compilation errors after upgrade**: Make sure to use type assertions when calling `Update()`:
   ```go
   // Old: m.filepicker, cmd = m.filepicker.Update(msg)
   // New:
   updatedModel, cmd := m.filepicker.Update(msg)
   m.filepicker = updatedModel.(filepicker.Model)
   ```

2. **Missing file preview**: File preview only works for text files under 10KB. Binary files show file type information instead.

3. **Performance with large directories**: The picker handles large directories well, but initial loading might take a moment. Use search (`/`) to quickly find files.
