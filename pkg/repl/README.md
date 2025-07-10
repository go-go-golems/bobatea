# REPL Package

A powerful, generic, and embeddable REPL (Read-Eval-Print Loop) component for Bubble Tea applications with pluggable evaluators, comprehensive theming, and advanced features.

## Features

- **🔌 Pluggable Evaluators** - Generic interface supports any language or custom logic
- **🎨 Comprehensive Theming** - Multiple built-in themes and full customization
- **📚 Command History** - Persistent history with navigation and search
- **📝 Multiline Support** - Optional multiline input mode for complex expressions
- **⚡ External Editor** - Seamless integration with $EDITOR for complex input
- **🛠️ Slash Commands** - Built-in commands plus extensible custom command system
- **🎯 Embeddable Design** - Clean message-based API for easy integration
- **⌨️ Rich Keyboard Support** - Comprehensive keyboard shortcuts and navigation
- **🚀 Non-blocking Evaluation** - Async evaluation with loading states
- **🔒 Error Handling** - Robust error handling and recovery

## Quick Start

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// Simple evaluator example
type EchoEvaluator struct{}

func (e *EchoEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	return fmt.Sprintf("Echo: %s", code), nil
}

func (e *EchoEvaluator) GetPrompt() string { return "echo> " }
func (e *EchoEvaluator) GetName() string { return "Echo" }
func (e *EchoEvaluator) SupportsMultiline() bool { return false }
func (e *EchoEvaluator) GetFileExtension() string { return ".txt" }

func main() {
	evaluator := &EchoEvaluator{}
	config := repl.DefaultConfig()
	config.Title = "Echo REPL"
	
	model := repl.NewModel(evaluator, config)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
```

### Advanced Configuration

```go
func main() {
	evaluator := &MyEvaluator{}
	
	config := repl.Config{
		Title:                "Advanced REPL",
		Prompt:               ">>> ",
		Placeholder:          "Enter code or /help for commands",
		Width:                120,
		StartMultiline:       true,
		EnableExternalEditor: true,
		EnableHistory:        true,
		MaxHistorySize:       2000,
	}
	
	model := repl.NewModel(evaluator, config)
	model.SetTheme(repl.BuiltinThemes["dark"])
	
	// Add custom commands
	model.AddCustomCommand("version", func(args []string) tea.Cmd {
		return func() tea.Msg {
			return repl.EvaluationCompleteMsg{
				Input:  "/version",
				Output: "MyLanguage v1.0.0",
				Error:  nil,
			}
		}
	})
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating Custom Evaluators

### Simple Calculator

```go
type CalculatorEvaluator struct {
	variables map[string]float64
}

func NewCalculatorEvaluator() *CalculatorEvaluator {
	return &CalculatorEvaluator{
		variables: make(map[string]float64),
	}
}

func (e *CalculatorEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	// Handle variable assignment: x = 42
	if strings.Contains(code, "=") {
		parts := strings.Split(code, "=")
		if len(parts) == 2 {
			varName := strings.TrimSpace(parts[0])
			valueStr := strings.TrimSpace(parts[1])
			
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return "", fmt.Errorf("invalid number: %s", valueStr)
			}
			
			e.variables[varName] = value
			return fmt.Sprintf("%s = %.2f", varName, value), nil
		}
	}
	
	// Handle expressions
	result, err := e.evaluateExpression(code)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%.2f", result), nil
}

func (e *CalculatorEvaluator) GetPrompt() string { return "calc> " }
func (e *CalculatorEvaluator) GetName() string { return "Calculator" }
func (e *CalculatorEvaluator) SupportsMultiline() bool { return false }
func (e *CalculatorEvaluator) GetFileExtension() string { return ".calc" }
```

### Shell Command Evaluator

```go
type ShellEvaluator struct {
	workingDir string
	env        []string
}

func NewShellEvaluator() *ShellEvaluator {
	wd, _ := os.Getwd()
	return &ShellEvaluator{
		workingDir: wd,
		env:        os.Environ(),
	}
}

func (e *ShellEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", code)
	cmd.Dir = e.workingDir
	cmd.Env = e.env
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	
	return string(output), nil
}

func (e *ShellEvaluator) GetPrompt() string { return "shell> " }
func (e *ShellEvaluator) GetName() string { return "Shell" }
func (e *ShellEvaluator) SupportsMultiline() bool { return true }
func (e *ShellEvaluator) GetFileExtension() string { return ".sh" }
```

## Configuration Options

```go
type Config struct {
	Title                string  // REPL title displayed at the top
	Prompt               string  // Custom prompt (overrides evaluator prompt)
	Placeholder          string  // Input placeholder text
	Width                int     // REPL width in characters
	StartMultiline       bool    // Start in multiline mode
	EnableExternalEditor bool    // Enable Ctrl+E external editor
	EnableHistory        bool    // Enable command history
	MaxHistorySize       int     // Maximum history entries
}

// Get default configuration
config := repl.DefaultConfig()
```

## Theming and Styling

### Built-in Themes

```go
// Apply built-in themes
model.SetTheme(repl.BuiltinThemes["default"])
model.SetTheme(repl.BuiltinThemes["dark"])
model.SetTheme(repl.BuiltinThemes["light"])
```

### Custom Styling

```go
customStyles := repl.Styles{
	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("32")).
		Background(lipgloss.Color("240")).
		Padding(0, 1),
	
	Prompt: lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true),
	
	Result: lipgloss.NewStyle().
		Foreground(lipgloss.Color("36")),
	
	Error: lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true),
	
	Info: lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Italic(true),
	
	HelpText: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true),
}

model.SetStyles(customStyles)
```

## Custom Commands

Add powerful custom slash commands:

```go
// Simple command
model.AddCustomCommand("time", func(args []string) tea.Cmd {
	return func() tea.Msg {
		return repl.EvaluationCompleteMsg{
			Input:  "/time",
			Output: time.Now().Format("15:04:05"),
			Error:  nil,
		}
	}
})

// Command with arguments and validation
model.AddCustomCommand("calc", func(args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) < 1 {
			return repl.EvaluationCompleteMsg{
				Input:  "/calc",
				Output: "Usage: /calc <expression>",
				Error:  fmt.Errorf("missing expression"),
			}
		}
		
		expr := strings.Join(args, " ")
		result, err := evaluateExpression(expr)
		
		return repl.EvaluationCompleteMsg{
			Input:  "/calc " + expr,
			Output: result,
			Error:  err,
		}
	}
})
```

## Embedding in Applications

### Basic Integration

```go
type AppModel struct {
	repl     repl.Model
	mode     string // "repl" or "other"
	otherUI  SomeOtherComponent
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repl.EvaluationCompleteMsg:
		// Handle REPL evaluation results
		return m.handleREPLResult(msg)
	
	case repl.QuitMsg:
		// Handle REPL quit
		return m.handleREPLQuit()
	
	case tea.KeyMsg:
		if msg.String() == "tab" {
			// Switch between REPL and other modes
			m.mode = toggleMode(m.mode)
		}
	}
	
	// Route to appropriate component
	if m.mode == "repl" {
		var cmd tea.Cmd
		m.repl, cmd = m.repl.Update(msg)
		return m, cmd
	}
	
	var cmd tea.Cmd
	m.otherUI, cmd = m.otherUI.Update(msg)
	return m, cmd
}
```

### Multi-Pane Development Environment

```go
type DevEnvironment struct {
	repl       repl.Model
	editor     EditorModel
	terminal   TerminalModel
	debugger   DebuggerModel
	activePane int
}

func (d DevEnvironment) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repl.EvaluationCompleteMsg:
		// Route REPL output to appropriate pane
		if strings.HasPrefix(msg.Input, "/debug") {
			return d.sendToDebugger(msg)
		}
		if strings.HasPrefix(msg.Input, "/run") {
			return d.sendToTerminal(msg)
		}
		if strings.HasPrefix(msg.Input, "/edit") {
			return d.sendToEditor(msg)
		}
	}
	
	// Handle pane switching
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+1":
			d.activePane = 0 // REPL
		case "ctrl+2":
			d.activePane = 1 // Editor
		case "ctrl+3":
			d.activePane = 2 // Terminal
		case "ctrl+4":
			d.activePane = 3 // Debugger
		}
	}
	
	// Update active pane
	var cmd tea.Cmd
	switch d.activePane {
	case 0:
		d.repl, cmd = d.repl.Update(msg)
	case 1:
		d.editor, cmd = d.editor.Update(msg)
	case 2:
		d.terminal, cmd = d.terminal.Update(msg)
	case 3:
		d.debugger, cmd = d.debugger.Update(msg)
	}
	
	return d, cmd
}
```

## Message System

The REPL communicates through a clean message-based API:

```go
// Evaluation completed
type EvaluationCompleteMsg struct {
	Input  string
	Output string
	Error  error
}

// REPL should quit
type QuitMsg struct{}

// Clear history
type ClearHistoryMsg struct{}

// External editor complete
type ExternalEditorCompleteMsg struct {
	Content string
	Error   error
}
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Exit REPL |
| `Ctrl+J` | Add line in multiline mode |
| `Ctrl+E` | Open external editor |
| `Up/Down` | Navigate command history |
| `Enter` | Execute code or add line |
| `Tab` | Toggle between modes (if embedded) |

## Built-in Slash Commands

| Command | Description |
|---------|-------------|
| `/help` | Show help message |
| `/clear` | Clear history |
| `/quit` | Exit REPL |
| `/multiline` | Toggle multiline mode |
| `/edit` | Open external editor |

## Architecture

The REPL follows a modular, extensible architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Model       │    │   Evaluator     │    │    History      │
│  (UI State)     │────│  (Interface)    │    │ (Command Log)   │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └─────────────────────────────────────────────────┘
                                 │
                    ┌─────────────────┐
                    │     Styles      │
                    │   (Theming)     │
                    │                 │
                    └─────────────────┘
```

### File Structure

- **`model.go`** - Main REPL model and core logic
- **`evaluator.go`** - Interface for pluggable evaluators and configuration
- **`messages.go`** - Message types for communication
- **`history.go`** - Command history management
- **`styles.go`** - Styling and theme system
- **`example_evaluator.go`** - Example evaluator implementation

## API Reference

### Core Types

```go
// Main REPL model
type Model struct { /* ... */ }

// Evaluator interface
type Evaluator interface {
	Evaluate(ctx context.Context, code string) (string, error)
	GetPrompt() string
	GetName() string
	SupportsMultiline() bool
	GetFileExtension() string
}

// Configuration
type Config struct {
	Title                string
	Prompt               string
	Placeholder          string
	Width                int
	StartMultiline       bool
	EnableExternalEditor bool
	EnableHistory        bool
	MaxHistorySize       int
}
```

### Key Functions

```go
// Create new REPL model
func NewModel(evaluator Evaluator, config Config) Model

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

// Configuration and theming
func (m *Model) SetTheme(theme Theme)
func (m *Model) SetStyles(styles Styles)
func (m *Model) SetWidth(width int)

// Custom commands
func (m *Model) AddCustomCommand(name string, handler func([]string) tea.Cmd)

// History management
func (m *Model) GetHistory() *History
func (m *Model) ClearHistory()

// State management
func (m *Model) IsMultilineMode() bool
func (m *Model) SetMultilineMode(multiline bool)
func (m *Model) IsEvaluating() bool
```

## Examples

See the [examples directory](../../examples/repl/) for complete working examples including:

- **Basic usage** - Simple REPL with custom evaluator
- **Advanced features** - Multi-pane development environment
- **Custom themes** - Theme customization and switching
- **Embedded integration** - REPL as part of larger applications

## Documentation

- **[Complete Documentation](../../docs/repl.md)** - Comprehensive guide with examples
- **[API Reference](../../docs/repl.md#api-reference)** - Detailed API documentation
- **[Examples](../../examples/repl/)** - Working example applications

This makes the REPL easy to understand, extend, and embed in other applications while providing powerful features out of the box.
