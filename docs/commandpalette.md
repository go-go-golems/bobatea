# Command Palette Documentation

The bobatea command palette is a VSCode-style command launcher component for Bubble Tea applications, providing fuzzy search, customizable commands, and overlay UI functionality.

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

## Overview

The command palette component provides a modern, keyboard-driven command interface similar to VSCode's Command Palette (Ctrl+Shift+P). It features:

- **Fuzzy search** - Find commands quickly with fuzzy matching
- **Overlay UI** - Non-intrusive popup that overlays your application
- **Custom commands** - Register any commands with descriptions and actions
- **Keyboard navigation** - Full keyboard support with intuitive shortcuts
- **Customizable styling** - Configure appearance with lipgloss styles
- **Simple integration** - Easy to embed in any Bubble Tea application

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/commandpalette"
)

type model struct {
    commandPalette commandpalette.Model
    messages       []string
}

func initialModel() model {
    cp := commandpalette.New()
    
    // Register commands
    cp.RegisterCommand("help", "Show help information", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "help"}
        }
    })
    
    cp.RegisterCommand("quit", "Exit application", func() tea.Cmd {
        return tea.Quit
    })
    
    return model{
        commandPalette: cp,
        messages:       []string{"Welcome! Press Ctrl+P to open command palette."},
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.commandPalette.SetSize(msg.Width, msg.Height)
        return m, nil
        
    case commandpalette.ExecutedMsg:
        switch msg.Command {
        case "help":
            m.messages = append(m.messages, "Help: Available commands - help, quit")
        }
        return m, nil
        
    case tea.KeyMsg:
        // Command palette handles input when visible
        if m.commandPalette.IsVisible() {
            var cmd tea.Cmd
            m.commandPalette, cmd = m.commandPalette.Update(msg)
            return m, cmd
        }
        
        // Application keys
        switch msg.String() {
        case "ctrl+p":
            m.commandPalette.Show()
            return m, nil
        case "ctrl+c", "q":
            return m, tea.Quit
        }
    }
    
    return m, nil
}

func (m model) View() string {
    view := "Messages:\n"
    for _, msg := range m.messages {
        view += "• " + msg + "\n"
    }
    view += "\nPress Ctrl+P for command palette"
    
    // Overlay command palette if visible
    if m.commandPalette.IsVisible() {
        return m.commandPalette.View()
    }
    
    return view
}

func main() {
    p := tea.NewProgram(initialModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Chat Application Example

For a complete example, see the [demo chat application](../cmd/command-palette-demo/) which shows how to integrate the command palette into a chat interface.

## API Reference

### Types

#### Command
```go
type Command struct {
    Name        string              // Command name (used for searching)
    Description string              // Human-readable description
    Action      func() tea.Cmd      // Function to execute when selected
}
```

#### Model
```go
type Model struct {
    // Internal state - use methods to interact
}
```

#### Styles
```go
type Styles struct {
    Palette            lipgloss.Style  // Main palette container
    Header             lipgloss.Style  // "Command Palette" title
    Query              lipgloss.Style  // Search input area
    Command            lipgloss.Style  // Normal command styling
    SelectedCommand    lipgloss.Style  // Selected command styling
    CommandName        lipgloss.Style  // Command name styling
    CommandDescription lipgloss.Style  // Command description styling
    Help               lipgloss.Style  // Help text styling
}
```

### Messages

#### ExecutedMsg
```go
type ExecutedMsg struct {
    Command string      // The name of the executed command
    Data    interface{} // Optional data returned by the command
}
```
Sent when a command is executed from the palette.

### Constructor

#### New()
```go
func New() Model
```
Creates a new command palette model with default styling.

#### WithStyles()
```go
func (m Model) WithStyles(styles Styles) Model
```
Creates a copy of the model with custom styles applied.

### Methods

#### Command Registration
```go
func (m *Model) RegisterCommand(name, description string, action func() tea.Cmd)
```
Registers a new command with the palette. Commands are searchable by name.

#### Visibility Control
```go
func (m *Model) Show()         // Show the command palette
func (m *Model) Hide()         // Hide the command palette
func (m Model) IsVisible() bool // Check if palette is visible
```

#### Size Management
```go
func (m *Model) SetSize(width, height int)
```
Sets the dimensions for proper overlay positioning.

#### Bubble Tea Interface
```go
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)
func (m Model) View() string
```

### Default Styles

The package provides sensible defaults:

```go
func DefaultStyles() Styles {
    return Styles{
        Palette: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("62")).
            Background(lipgloss.Color("235")).
            Padding(1).
            Margin(2, 4),

        Header: lipgloss.NewStyle().
            Foreground(lipgloss.Color("212")).
            Bold(true).
            Margin(0, 0, 1, 0),

        Query: lipgloss.NewStyle().
            Foreground(lipgloss.Color("86")).
            Background(lipgloss.Color("240")).
            Padding(0, 1).
            Margin(0, 0, 1, 0),

        Command: lipgloss.NewStyle().
            Padding(0, 1),

        SelectedCommand: lipgloss.NewStyle().
            Background(lipgloss.Color("62")).
            Foreground(lipgloss.Color("230")).
            Padding(0, 1),

        CommandName: lipgloss.NewStyle().
            Foreground(lipgloss.Color("86")).
            Bold(true),

        CommandDescription: lipgloss.NewStyle().
            Foreground(lipgloss.Color("243")),

        Help: lipgloss.NewStyle().
            Foreground(lipgloss.Color("241")).
            Italic(true),
    }
}
```

## Keyboard Shortcuts

### Opening/Closing
| Key | Action |
|-----|--------|
| `Ctrl+P` | Open command palette |
| `Esc` | Close command palette |

### Navigation
| Key | Action |
|-----|--------|
| `↑` / `Ctrl+K` | Move selection up |
| `↓` / `Ctrl+J` | Move selection down |
| `Enter` | Execute selected command |

### Search
| Key | Action |
|-----|--------|
| `Backspace` | Delete last character from search |
| `[a-z0-9]` | Add character to search query |

### The palette automatically filters commands as you type, using fuzzy matching to find relevant commands quickly.

## Configuration

### Custom Styling

```go
package main

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/go-go-golems/bobatea/pkg/commandpalette"
)

func createStyledPalette() commandpalette.Model {
    // Create custom styles
    customStyles := commandpalette.Styles{
        Palette: lipgloss.NewStyle().
            Border(lipgloss.ThickBorder()).
            BorderForeground(lipgloss.Color("#FF6B9D")).
            Background(lipgloss.Color("#1A1B26")).
            Padding(2).
            Margin(3, 6),

        Header: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#BB9AF7")).
            Bold(true).
            Underline(true),

        Query: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#7AA2F7")).
            Background(lipgloss.Color("#24283B")).
            Padding(0, 2).
            Border(lipgloss.RoundedBorder()),

        SelectedCommand: lipgloss.NewStyle().
            Background(lipgloss.Color("#7AA2F7")).
            Foreground(lipgloss.Color("#1A1B26")).
            Bold(true).
            Padding(0, 1),

        CommandName: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#9ECE6A")).
            Bold(true),

        CommandDescription: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#565F89")),
    }

    return commandpalette.New().WithStyles(customStyles)
}
```

### Size Configuration

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Important: Update palette size when window resizes
        m.commandPalette.SetSize(msg.Width, msg.Height)
        return m, nil
    }
    // ... rest of update logic
}
```

## Integration Examples

### File Editor Integration

```go
type EditorModel struct {
    commandPalette commandpalette.Model
    editor         *editor.Model
    currentFile    string
}

func (m EditorModel) setupCommands() {
    // File operations
    m.commandPalette.RegisterCommand("file:open", "Open file", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "file:open"}
        }
    })
    
    m.commandPalette.RegisterCommand("file:save", "Save current file", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "file:save"}
        }
    })
    
    m.commandPalette.RegisterCommand("file:save-as", "Save file as...", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "file:save-as"}
        }
    })
    
    // Edit operations
    m.commandPalette.RegisterCommand("edit:find", "Find in file", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "edit:find"}
        }
    })
    
    m.commandPalette.RegisterCommand("edit:replace", "Find and replace", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "edit:replace"}
        }
    })
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case commandpalette.ExecutedMsg:
        switch msg.Command {
        case "file:open":
            return m, m.openFileDialog()
        case "file:save":
            return m, m.saveCurrentFile()
        case "edit:find":
            return m, m.showFindDialog()
        }
    }
    
    // Handle command palette updates
    if m.commandPalette.IsVisible() {
        var cmd tea.Cmd
        m.commandPalette, cmd = m.commandPalette.Update(msg)
        return m, cmd
    }
    
    // Handle editor updates
    var cmd tea.Cmd
    m.editor, cmd = m.editor.Update(msg)
    return m, cmd
}
```

### Git Client Integration

```go
type GitModel struct {
    commandPalette commandpalette.Model
    repo           *git.Repository
    status         string
}

func (m GitModel) setupGitCommands() {
    m.commandPalette.RegisterCommand("git:status", "Show git status", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "git:status"}
        }
    })
    
    m.commandPalette.RegisterCommand("git:add", "Stage all changes", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "git:add"}
        }
    })
    
    m.commandPalette.RegisterCommand("git:commit", "Commit staged changes", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "git:commit"}
        }
    })
    
    m.commandPalette.RegisterCommand("git:push", "Push to remote", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "git:push"}
        }
    })
    
    m.commandPalette.RegisterCommand("git:pull", "Pull from remote", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "git:pull"}
        }
    })
    
    m.commandPalette.RegisterCommand("git:log", "Show commit history", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "git:log"}
        }
    })
}
```

### Application Settings

```go
type AppModel struct {
    commandPalette commandpalette.Model
    settings       *Settings
    theme          string
}

func (m AppModel) setupSettingsCommands() {
    m.commandPalette.RegisterCommand("settings:theme:dark", "Switch to dark theme", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "settings:theme", Data: "dark"}
        }
    })
    
    m.commandPalette.RegisterCommand("settings:theme:light", "Switch to light theme", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "settings:theme", Data: "light"}
        }
    })
    
    m.commandPalette.RegisterCommand("settings:font:increase", "Increase font size", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "settings:font", Data: "increase"}
        }
    })
    
    m.commandPalette.RegisterCommand("settings:font:decrease", "Decrease font size", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "settings:font", Data: "decrease"}
        }
    })
    
    m.commandPalette.RegisterCommand("settings:save", "Save current settings", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "settings:save"}
        }
    })
    
    m.commandPalette.RegisterCommand("settings:reset", "Reset to defaults", func() tea.Cmd {
        return func() tea.Msg {
            return commandpalette.ExecutedMsg{Command: "settings:reset"}
        }
    })
}
```

## Advanced Features

### Dynamic Command Registration

Commands can be registered and unregistered dynamically based on application state:

```go
type DynamicModel struct {
    commandPalette commandpalette.Model
    fileOpen       bool
    hasSelection   bool
}

func (m *DynamicModel) updateAvailableCommands() {
    // Create a new palette to clear existing commands
    m.commandPalette = commandpalette.New()
    
    // Always available commands
    m.commandPalette.RegisterCommand("help", "Show help", func() tea.Cmd {
        return func() tea.Msg { return commandpalette.ExecutedMsg{Command: "help"} }
    })
    
    // File-dependent commands
    if m.fileOpen {
        m.commandPalette.RegisterCommand("file:save", "Save file", func() tea.Cmd {
            return func() tea.Msg { return commandpalette.ExecutedMsg{Command: "file:save"} }
        })
        
        m.commandPalette.RegisterCommand("file:close", "Close file", func() tea.Cmd {
            return func() tea.Msg { return commandpalette.ExecutedMsg{Command: "file:close"} }
        })
    } else {
        m.commandPalette.RegisterCommand("file:new", "New file", func() tea.Cmd {
            return func() tea.Msg { return commandpalette.ExecutedMsg{Command: "file:new"} }
        })
        
        m.commandPalette.RegisterCommand("file:open", "Open file", func() tea.Cmd {
            return func() tea.Msg { return commandpalette.ExecutedMsg{Command: "file:open"} }
        })
    }
    
    // Selection-dependent commands
    if m.hasSelection {
        m.commandPalette.RegisterCommand("edit:copy", "Copy selection", func() tea.Cmd {
            return func() tea.Msg { return commandpalette.ExecutedMsg{Command: "edit:copy"} }
        })
        
        m.commandPalette.RegisterCommand("edit:cut", "Cut selection", func() tea.Cmd {
            return func() tea.Msg { return commandpalette.ExecutedMsg{Command: "edit:cut"} }
        })
    }
}
```

### Command Categories

Organize commands with prefixes for better discoverability:

```go
func (m *Model) setupCategorizedCommands() {
    // File operations
    m.commandPalette.RegisterCommand("file:new", "Create new file", fileNewCmd)
    m.commandPalette.RegisterCommand("file:open", "Open existing file", fileOpenCmd)
    m.commandPalette.RegisterCommand("file:save", "Save current file", fileSaveCmd)
    m.commandPalette.RegisterCommand("file:save-as", "Save file with new name", fileSaveAsCmd)
    
    // Edit operations
    m.commandPalette.RegisterCommand("edit:undo", "Undo last action", editUndoCmd)
    m.commandPalette.RegisterCommand("edit:redo", "Redo last action", editRedoCmd)
    m.commandPalette.RegisterCommand("edit:find", "Find text", editFindCmd)
    m.commandPalette.RegisterCommand("edit:replace", "Find and replace", editReplaceCmd)
    
    // View operations
    m.commandPalette.RegisterCommand("view:zoom-in", "Increase zoom level", viewZoomInCmd)
    m.commandPalette.RegisterCommand("view:zoom-out", "Decrease zoom level", viewZoomOutCmd)
    m.commandPalette.RegisterCommand("view:toggle-sidebar", "Show/hide sidebar", viewToggleSidebarCmd)
    
    // Git operations (when in a git repository)
    if m.isGitRepo {
        m.commandPalette.RegisterCommand("git:status", "Show repository status", gitStatusCmd)
        m.commandPalette.RegisterCommand("git:commit", "Commit changes", gitCommitCmd)
        m.commandPalette.RegisterCommand("git:push", "Push to remote", gitPushCmd)
    }
}
```

### Command with Parameters

Commands can accept and use data:

```go
func (m *Model) setupParameterizedCommands() {
    m.commandPalette.RegisterCommand("theme:set", "Change application theme", func() tea.Cmd {
        return func() tea.Msg {
            // This could prompt for theme selection or cycle through themes
            themes := []string{"dark", "light", "blue", "green"}
            currentIndex := m.getCurrentThemeIndex()
            nextTheme := themes[(currentIndex+1)%len(themes)]
            
            return commandpalette.ExecutedMsg{
                Command: "theme:set",
                Data:    nextTheme,
            }
        }
    })
    
    m.commandPalette.RegisterCommand("goto:line", "Go to specific line number", func() tea.Cmd {
        return func() tea.Msg {
            // This could open a prompt for line number
            return commandpalette.ExecutedMsg{
                Command: "goto:line",
                Data:    "prompt", // Signal to show line number prompt
            }
        }
    })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case commandpalette.ExecutedMsg:
        switch msg.Command {
        case "theme:set":
            if theme, ok := msg.Data.(string); ok {
                m.setTheme(theme)
            }
        case "goto:line":
            if msg.Data == "prompt" {
                return m, m.showLineNumberPrompt()
            }
        }
    }
    return m, nil
}
```

### Overlay Integration

The command palette works as an overlay that doesn't disrupt your main interface:

```go
func (m Model) View() string {
    // Render your main application view
    mainView := m.renderMainInterface()
    
    // Overlay command palette if visible
    if m.commandPalette.IsVisible() {
        return lipgloss.Place(
            m.width, m.height,
            lipgloss.Center, lipgloss.Center,
            m.commandPalette.View(),
            lipgloss.WithWhitespaceChars(" "),
            lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{
                Light: "#D9DCCF",
                Dark:  "#383838",
            }),
        )
    }
    
    return mainView
}
```

## Customization

### Theme Integration

Create themes that work with your application's color scheme:

```go
// Dark theme
func DarkTheme() commandpalette.Styles {
    return commandpalette.Styles{
        Palette: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("#6B7280")).
            Background(lipgloss.Color("#1F2937")).
            Padding(1).
            Margin(2, 4),

        Header: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#F3F4F6")).
            Bold(true),

        Query: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#E5E7EB")).
            Background(lipgloss.Color("#374151")).
            Padding(0, 1),

        SelectedCommand: lipgloss.NewStyle().
            Background(lipgloss.Color("#3B82F6")).
            Foreground(lipgloss.Color("#FFFFFF")).
            Padding(0, 1),

        CommandName: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#60A5FA")).
            Bold(true),

        CommandDescription: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#9CA3AF")),
    }
}

// Light theme
func LightTheme() commandpalette.Styles {
    return commandpalette.Styles{
        Palette: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("#D1D5DB")).
            Background(lipgloss.Color("#FFFFFF")).
            Padding(1).
            Margin(2, 4),

        Header: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#111827")).
            Bold(true),

        Query: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#374151")).
            Background(lipgloss.Color("#F9FAFB")).
            Border(lipgloss.NormalBorder()).
            BorderForeground(lipgloss.Color("#D1D5DB")).
            Padding(0, 1),

        SelectedCommand: lipgloss.NewStyle().
            Background(lipgloss.Color("#3B82F6")).
            Foreground(lipgloss.Color("#FFFFFF")).
            Padding(0, 1),

        CommandName: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#1D4ED8")).
            Bold(true),

        CommandDescription: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#6B7280")),
    }
}
```

### Custom Key Bindings

While the command palette has fixed key bindings for consistency, you can customize the trigger key in your application:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Custom trigger keys
        switch msg.String() {
        case "ctrl+p":        // VSCode style
            m.commandPalette.Show()
            return m, nil
        case "ctrl+shift+p":  // Alternative VSCode style
            m.commandPalette.Show()
            return m, nil
        case "f1":            // Windows style
            m.commandPalette.Show()
            return m, nil
        case "cmd+shift+p":   // macOS style
            m.commandPalette.Show()
            return m, nil
        }
    }
    return m, nil
}
```

### Command Icons

Add visual indicators to commands:

```go
func (m *Model) setupCommandsWithIcons() {
    m.commandPalette.RegisterCommand("📄 file:new", "Create new file", fileNewCmd)
    m.commandPalette.RegisterCommand("📂 file:open", "Open existing file", fileOpenCmd)
    m.commandPalette.RegisterCommand("💾 file:save", "Save current file", fileSaveCmd)
    m.commandPalette.RegisterCommand("🔍 edit:find", "Find text", editFindCmd)
    m.commandPalette.RegisterCommand("🔄 edit:replace", "Find and replace", editReplaceCmd)
    m.commandPalette.RegisterCommand("⚙️ settings:open", "Open settings", settingsOpenCmd)
    m.commandPalette.RegisterCommand("🎨 theme:toggle", "Toggle theme", themeToggleCmd)
    m.commandPalette.RegisterCommand("ℹ️ help:about", "About this application", helpAboutCmd)
}
```

## Troubleshooting

### Common Issues

#### Command Palette Not Showing
```go
// Ensure you're calling Show() when the trigger key is pressed
case "ctrl+p":
    m.commandPalette.Show()  // Make sure this is called
    return m, nil
```

#### Commands Not Executing
```go
// Make sure you're handling the ExecutedMsg
case commandpalette.ExecutedMsg:
    // Handle the executed command
    switch msg.Command {
    case "your-command":
        // Your command logic here
    }
    return m, nil
```

#### Overlay Not Displaying Correctly
```go
// Ensure you're setting the size properly
case tea.WindowSizeMsg:
    m.commandPalette.SetSize(msg.Width, msg.Height)
    return m, nil

// And rendering the overlay correctly
if m.commandPalette.IsVisible() {
    return lipgloss.Place(
        m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        m.commandPalette.View(),
    )
}
```

#### Key Conflicts
```go
// Handle command palette input first
if m.commandPalette.IsVisible() {
    var cmd tea.Cmd
    m.commandPalette, cmd = m.commandPalette.Update(msg)
    return m, cmd  // Return early to prevent conflicts
}

// Then handle your application keys
switch msg.String() {
case "ctrl+p":
    m.commandPalette.Show()
    return m, nil
}
```

### Performance Considerations

For applications with many commands:

```go
// Register commands lazily or conditionally
func (m *Model) registerContextualCommands() {
    // Only register file commands when a file is open
    if m.fileOpen {
        m.commandPalette.RegisterCommand("file:save", "Save file", fileSaveCmd)
    }
    
    // Only register git commands in git repositories
    if m.isGitRepo {
        m.commandPalette.RegisterCommand("git:commit", "Commit changes", gitCommitCmd)
    }
}
```

### Debug Mode

For debugging command palette issues:

```go
// Check if palette is receiving updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.commandPalette.IsVisible() {
        fmt.Printf("Palette handling: %+v\n", msg) // Debug output
        var cmd tea.Cmd
        m.commandPalette, cmd = m.commandPalette.Update(msg)
        return m, cmd
    }
    return m, nil
}
```

---

The command palette is designed to be a drop-in component that enhances any Bubble Tea application with powerful command functionality. For more examples, see the [demo application](../cmd/command-palette-demo/) and explore the [test files](../pkg/commandpalette/) for additional usage patterns.
