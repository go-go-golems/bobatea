# Directory Jail Feature

The bobatea filepicker now supports directory restriction (jail) functionality to prevent users from navigating outside a specified directory boundary. This is important for security and UX in applications that need to restrict file access.

## Features

- **Navigation Restriction**: Users cannot navigate above the jail directory
- **Visual Indicators**: 
  - Status shows "Jailed" when jail is active
  - Path display shows relative path from jail root (`[jail]/subdir` instead of absolute path)
  - ".." entry is hidden when at jail root
- **Complete Coverage**: All navigation methods respect the jail boundary:
  - Backspace key navigation
  - ".." directory entry  
  - History navigation (back/forward)
  - Direct path setting
- **Security**: Handles symlinks and prevents escape attempts

## Usage

### Basic Setup

```go
import "github.com/go-go-golems/bobatea/pkg/filepicker"

// Create filepicker with jail directory
fp := filepicker.New(
    filepicker.WithJailDirectory("/home/user/safe-area"),
    filepicker.WithStartPath("/home/user/safe-area/documents"),
)
```

### With Compatibility API

```go
// Using the compatibility Model
fp := filepicker.NewModelWithOptions(
    filepicker.WithJailDirectory("/path/to/jail"),
    filepicker.WithShowPreview(true),
)
```

## API Reference

### WithJailDirectory Option

```go
func WithJailDirectory(path string) Option
```

Sets a directory restriction boundary. Navigation will be limited to this directory and its subdirectories.

**Parameters:**
- `path`: Absolute or relative path to the jail directory. Will be converted to absolute path internally.

**Behavior:**
- If the starting path is outside the jail, the filepicker will automatically navigate to the jail directory
- The jail directory must exist and be accessible
- Empty string disables jail restriction

### Internal Methods

The following methods are used internally but are available for advanced use cases:

```go
// Check if a path is within the jail boundary
func (fp *AdvancedModel) isWithinJail(path string) bool

// Check if currently at the jail root directory  
func (fp *AdvancedModel) isAtJailRoot() bool

// Validate if navigation to a path is allowed
func (fp *AdvancedModel) validateNavigationPath(path string) bool
```

## Example

Here's a complete example demonstrating the jail functionality:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/filepicker"
)

func main() {
    jailDir := "/home/user/documents"
    
    // Ensure jail directory exists
    if _, err := os.Stat(jailDir); os.IsNotExist(err) {
        log.Fatalf("Jail directory does not exist: %s", jailDir)
    }
    
    // Create filepicker with jail
    fp := filepicker.New(
        filepicker.WithJailDirectory(jailDir),
        filepicker.WithShowPreview(true),
        filepicker.WithDetailedView(true),
    )
    
    // Use in your bubbletea program...
}
```

## UI Indicators

When jail is active, users will see:

1. **Status Line**: Shows "Jailed" in the options section
2. **Path Display**: Shows `[jail]/relative/path` instead of absolute paths
3. **Navigation**: ".." entry disappears when at jail root
4. **History**: Back/forward navigation stops at jail boundary

## Security Considerations

- **Path Validation**: All paths are resolved to absolute form and validated
- **Symlink Handling**: Symlinks cannot be used to escape the jail
- **History Restriction**: Navigation history only contains paths within jail
- **Graceful Handling**: If current directory is outside jail, automatically navigates to jail

## Testing

The jail functionality includes comprehensive tests:

```bash
# Run jail-specific tests
go test ./pkg/filepicker -v -run TestJail
go test ./pkg/filepicker -v -run TestIsWithinJail
go test ./pkg/filepicker -v -run TestValidateNavigationPath

# Run all tests
go test ./pkg/filepicker -v
```

## Example Scenarios

### Restrict to User Documents
```go
fp := filepicker.New(
    filepicker.WithJailDirectory("/home/user/Documents"),
)
```

### Temporary Directory Restriction
```go
tmpDir := "/tmp/app-sandbox"
os.MkdirAll(tmpDir, 0755)
fp := filepicker.New(
    filepicker.WithJailDirectory(tmpDir),
)
```

### Application Data Directory
```go
dataDir := filepath.Join(os.Getenv("HOME"), ".myapp", "data")
fp := filepicker.New(
    filepicker.WithJailDirectory(dataDir),
    filepicker.WithStartPath(filepath.Join(dataDir, "projects")),
)
```

## Implementation Notes

- Uses `filepath.Clean()` and `filepath.Abs()` for robust path handling
- Handles Windows and Unix path separators correctly
- Validates jail directory exists before applying restriction
- Thread-safe for concurrent access (no shared mutable state)
- Zero performance impact when jail is not used
