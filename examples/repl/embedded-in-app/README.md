# Embedded REPL in Application Example

This example demonstrates how to embed the bobatea REPL component into a larger multi-mode application. It shows how to integrate the REPL with other UI components and manage state across different application modes.

## What it does

- **Multi-Mode Application** - Switch between different application modes
- **REPL Integration** - Embed the REPL as one of the application modes
- **State Management** - Maintain application state across mode switches
- **Inter-Component Communication** - Share data between REPL and other components
- **Comprehensive UI** - Full application interface with menus, logs, and status

## Running the example

```bash
go run main.go
```

## Key Features Demonstrated

- **Mode-Based Architecture** - Clean separation between application modes
- **REPL Embedding** - Proper integration of REPL component
- **Message Handling** - Processing REPL messages in the parent application
- **State Persistence** - Maintaining state when switching between modes
- **UI Coordination** - Coordinating multiple UI components
- **Logging Integration** - Capturing REPL activity in application logs

## Application Structure

### Modes

The application has three main modes:

1. **Menu Mode** - Main navigation menu
2. **REPL Mode** - Interactive command processor
3. **Log Mode** - View application logs

### Components

- **AppModel** - Main application state and coordination
- **SimpleEvaluator** - Custom evaluator for the REPL
- **Log System** - Captures and displays application activity
- **Menu System** - Navigation between modes
- **Status Bar** - Application status information

## What you'll see

When you run this example, you'll see:

### Main Menu
```
┌─────────────────────────────────────────────────────┐
│ Multi-Mode Application - MENU                       │
├─────────────────────────────────────────────────────┤
│ Main Menu                                           │
│                                                     │
│ ┌─────────────────────────────────────────────────┐ │
│ │ 1. Command REPL                                 │ │
│ │ 2. View Logs                                    │ │
│ │ 3. Settings                                     │ │
│ │ 4. Quit                                         │ │
│ └─────────────────────────────────────────────────┘ │
│                                                     │
│ Use ↑/↓ to navigate, Enter to select, q to quit    │
├─────────────────────────────────────────────────────┤
│ Mode: menu | Commands: 0 | Logs: 6 | Press ? for help │
└─────────────────────────────────────────────────────┘
```

### REPL Mode
```
┌─────────────────────────────────────────────────────┐
│ Multi-Mode Application - REPL                       │
├─────────────────────────────────────────────────────┤
│ Embedded Command Processor                          │
│                                                     │
│ app> process my-task                                │
│ Processing 'my-task'... Done! (command #1)         │
│                                                     │
│ app> status                                         │
│ System status: OK (executed 2 commands)            │
│                                                     │
│ app> _                                              │
│                                                     │
│ REPL Mode - ESC to return to menu, Ctrl+C to quit  │
├─────────────────────────────────────────────────────┤
│ Mode: repl | Commands: 2 | Logs: 8 | Press ? for help │
└─────────────────────────────────────────────────────┘
```

### Log Viewer
```
┌─────────────────────────────────────────────────────┐
│ Multi-Mode Application - LOG                        │
├─────────────────────────────────────────────────────┤
│ Application Logs                                    │
│                                                     │
│ ┌─────────────────────────────────────────────────┐ │
│ │ 10:00:00 INFO    Application started           │ │
│ │ 10:00:01 DEBUG   Loading configuration         │ │
│ │ 10:00:02 INFO    REPL component initialized    │ │
│ │ 10:00:03 WARN    Sample warning message        │ │
│ │ 10:00:04 ERROR   Sample error message          │ │
│ │ 10:00:05 INFO    Ready for user input          │ │
│ │ now      INFO    REPL: process my-task -> ...  │ │
│ │ now      INFO    REPL: status -> System stat.. │ │
│ └─────────────────────────────────────────────────┘ │
│                                                     │
│ Log Viewer - ESC to return to menu, q to quit      │
├─────────────────────────────────────────────────────┤
│ Mode: log | Commands: 2 | Logs: 8 | Press ? for help │
└─────────────────────────────────────────────────────┘
```

## Code Architecture

### Application Model

```go
type AppModel struct {
    // Application state
    currentMode  string
    width        int
    height       int
    
    // Components
    repl       repl.Model
    textInput  textinput.Model
    logEntries []LogEntry
    
    // UI state
    selectedMenuItem int
    menuItems        []string
    
    // Styles
    styles AppStyles
}
```

### Mode Management

```go
const (
    ModeMenu = "menu"
    ModeREPL = "repl"
    ModeLog  = "log"
)

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.currentMode {
        case ModeMenu:
            // Handle menu navigation
        case ModeREPL:
            // Handle REPL mode
        case ModeLog:
            // Handle log viewer
        }
    }
    
    // Update components based on current mode
    if m.currentMode == ModeREPL {
        var cmd tea.Cmd
        m.repl, cmd = m.repl.Update(msg)
        return m, cmd
    }
    
    return m, nil
}
```

### Message Integration

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case repl.EvaluationCompleteMsg:
        // Log REPL activity
        level := "INFO"
        if msg.Error != nil {
            level = "ERROR"
        }
        
        m.logEntries = append(m.logEntries, LogEntry{
            Level:   level,
            Message: fmt.Sprintf("REPL: %s -> %s", msg.Input, msg.Output),
            Time:    "now",
        })
        
    case repl.QuitMsg:
        // Return to menu instead of quitting
        m.currentMode = ModeMenu
        return m, nil
    }
    
    return m, nil
}
```

### REPL Configuration

```go
func NewAppModel() AppModel {
    // Create the evaluator
    evaluator := NewSimpleEvaluator()
    
    // Create REPL configuration
    config := repl.Config{
        Title:                "Embedded Command Processor",
        Placeholder:          "Enter command (try 'process item', 'status', or 'error')",
        Width:                80,
        EnableHistory:        true,
        MaxHistorySize:       100,
    }
    
    // Create REPL model
    replModel := repl.NewModel(evaluator, config)
    replModel.SetTheme(repl.BuiltinThemes["dark"])
    
    // Add custom commands
    replModel.AddCustomCommand("app-info", func(args []string) tea.Cmd {
        return func() tea.Msg {
            return repl.EvaluationCompleteMsg{
                Input:  "/app-info",
                Output: "Multi-Mode Application v1.0.0\nModes: Menu, REPL, Log Viewer",
                Error:  nil,
            }
        }
    })
    
    return AppModel{
        repl: replModel,
        // ... other initialization
    }
}
```

## Key Integration Patterns

### 1. Message Handling

The application intercepts REPL messages and processes them:

```go
case repl.EvaluationCompleteMsg:
    // Log the activity
    // Update application state
    // Process the result
    
case repl.QuitMsg:
    // Handle REPL quit (return to menu instead of quitting app)
```

### 2. State Management

The application maintains state across mode switches:

```go
// REPL state persists when switching modes
// Log entries accumulate across sessions
// Configuration is maintained
```

### 3. Component Coordination

Multiple components work together:

```go
// REPL provides command processing
// Log viewer shows activity
// Menu provides navigation
// Status bar shows unified state
```

### 4. Custom Commands

Application-specific commands are added to the REPL:

```go
replModel.AddCustomCommand("app-info", func(args []string) tea.Cmd {
    // Return application-specific information
})
```

## Try These Commands

### In REPL Mode
- `process my-task` - Process an item
- `status` - Check system status
- `error` - Simulate an error
- `/app-info` - Get application information
- `/help` - Show REPL help

### Navigation
- **Menu Mode**: ↑/↓ to navigate, Enter to select
- **REPL Mode**: ESC to return to menu
- **Log Mode**: ESC to return to menu
- **Any Mode**: Ctrl+C to quit

## Advanced Features

### Logging Integration

All REPL activity is automatically logged:

```go
case repl.EvaluationCompleteMsg:
    level := "INFO"
    if msg.Error != nil {
        level = "ERROR"
    }
    
    m.logEntries = append(m.logEntries, LogEntry{
        Level:   level,
        Message: fmt.Sprintf("REPL: %s -> %s", msg.Input, msg.Output),
        Time:    "now",
    })
```

### State Persistence

The application maintains state when switching between modes:

- REPL history is preserved
- Log entries accumulate
- Application state is consistent

### UI Coordination

Components are coordinated through the main application model:

- Window size updates are propagated
- Themes are consistent
- Status is unified

## Extending the Example

You can extend this example by:

1. **Adding more modes**:
```go
const (
    ModeMenu = "menu"
    ModeREPL = "repl"
    ModeLog  = "log"
    ModeSettings = "settings"  // New mode
)
```

2. **Adding inter-mode communication**:
```go
// Pass data between modes
type ModeChangeMsg struct {
    FromMode string
    ToMode   string
    Data     interface{}
}
```

3. **Adding persistent storage**:
```go
// Save/load application state
func (m AppModel) saveState() error {
    // Serialize state to file
}
```

4. **Adding more REPL evaluators**:
```go
// Switch between different evaluators
func (m AppModel) switchEvaluator(name string) {
    // Change REPL evaluator
}
```

## Best Practices Demonstrated

1. **Clean Architecture** - Separation of concerns between modes
2. **Message-Driven Design** - Using Tea messages for coordination
3. **State Management** - Proper state handling across mode switches
4. **Error Handling** - Robust error handling and logging
5. **User Experience** - Consistent navigation and feedback

## Next Steps

After trying this example, check out:

- [Custom Theme Example](../custom-theme/) - Advanced styling and theming
- [Custom Evaluator Example](../custom-evaluator/) - Complex evaluator implementation
- [Basic Usage Example](../basic-usage/) - Simple REPL patterns

This example shows how to build complex applications with embedded REPL functionality while maintaining clean architecture and good user experience.
