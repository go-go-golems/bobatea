# REPL Documentation

A powerful, generic, and embeddable REPL (Read-Eval-Print Loop) component for Bubble Tea applications that supports pluggable evaluators, theming, history, and advanced features.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Installation and Quick Start](#installation-and-quick-start)
4. [Evaluator Interface](#evaluator-interface)
5. [Configuration](#configuration)
6. [Theming and Styling](#theming-and-styling)
7. [Embedding Examples](#embedding-examples)
8. [Advanced Usage](#advanced-usage)
9. [Built-in Features](#built-in-features)
10. [Message System](#message-system)
11. [Best Practices](#best-practices)
12. [API Reference](#api-reference)

## Overview

The bobatea REPL component provides a fully-featured, customizable REPL interface that can be embedded in any Bubble Tea application. It follows a pluggable architecture where evaluators can be swapped out to support different languages, tools, or custom logic.

### Key Features

- **Generic evaluator interface** - Works with any language or custom evaluator
- **Configurable behavior** - Customizable prompts, themes, and settings
- **Command history** - Navigation through previous commands with persistence
- **Multiline support** - Optional multiline input mode for complex expressions
- **External editor integration** - Open $EDITOR for complex input
- **Slash commands** - Built-in commands plus custom command support
- **Multiple themes** - Built-in themes (default, dark, light) and custom styling
- **Embeddable design** - Clean message-based API for integration
- **Keyboard shortcuts** - Comprehensive keyboard navigation
- **Real-time evaluation** - Non-blocking evaluation with loading states

## Architecture

The REPL system is built around several key components:

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

### Core Components

#### model.go
- **Main REPL model** - Manages UI state and coordination
- **Event handling** - Processes keyboard input and commands
- **Message routing** - Handles communication between components
- **State management** - Tracks evaluation state, mode, and settings

#### evaluator.go
- **Evaluator interface** - Defines the contract for language evaluators
- **Configuration types** - Config struct and factory functions
- **Result handling** - Structures for evaluation results

#### messages.go
- **Message types** - Tea messages for REPL communication
- **Event definitions** - Evaluation complete, quit, external editor messages
- **Inter-component communication** - Clean message-based API

#### history.go
- **Command history** - Manages command storage and navigation
- **Persistence** - Optional history file support
- **Navigation** - Up/down arrow key support

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
	"os"

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
	// Create evaluator and configuration
	evaluator := &EchoEvaluator{}
	config := repl.DefaultConfig()
	config.Title = "Echo REPL"
	
	// Create and run the REPL
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
	
	// Comprehensive configuration
	config := repl.Config{
		Title:                "My Language REPL",
		Prompt:               "mylang> ",
		Placeholder:          "Enter code or /help for commands",
		Width:                100,
		StartMultiline:       true,
		EnableExternalEditor: true,
		EnableHistory:        true,
		MaxHistorySize:       2000,
	}
	
	model := repl.NewModel(evaluator, config)
	
	// Apply custom theme
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
	
	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Evaluator Interface

The `Evaluator` interface is the heart of the REPL system, allowing you to plug in any evaluation logic:

```go
type Evaluator interface {
	// Evaluate executes the given code and returns the result
	Evaluate(ctx context.Context, code string) (string, error)
	
	// GetPrompt returns the prompt string for this evaluator
	GetPrompt() string
	
	// GetName returns the name of this evaluator (for display)
	GetName() string
	
	// SupportsMultiline returns true if this evaluator supports multiline input
	SupportsMultiline() bool
	
	// GetFileExtension returns the file extension for external editor
	GetFileExtension() string
}
```

### Creating Custom Evaluators

#### Simple Calculator Example

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
	code = strings.TrimSpace(code)
	
	// Handle variable assignment
	if strings.Contains(code, "=") {
		return e.handleAssignment(code)
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

#### Complex Language Evaluator

```go
type ScriptEvaluator struct {
	interpreter *MyInterpreter
	context     *EvaluationContext
}

func NewScriptEvaluator() *ScriptEvaluator {
	return &ScriptEvaluator{
		interpreter: NewMyInterpreter(),
		context:     NewEvaluationContext(),
	}
}

func (e *ScriptEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	// Parse the code
	ast, err := e.interpreter.Parse(code)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}
	
	// Check for incomplete multiline input
	if e.interpreter.IsIncomplete(ast) {
		return "", fmt.Errorf("incomplete input")
	}
	
	// Evaluate in context
	result, err := e.interpreter.Evaluate(ctx, ast, e.context)
	if err != nil {
		return "", fmt.Errorf("evaluation error: %w", err)
	}
	
	return result.String(), nil
}

func (e *ScriptEvaluator) GetPrompt() string { return "script> " }
func (e *ScriptEvaluator) GetName() string { return "ScriptLang" }
func (e *ScriptEvaluator) SupportsMultiline() bool { return true }
func (e *ScriptEvaluator) GetFileExtension() string { return ".script" }
```

#### Shell Command Evaluator

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

## Configuration

The `Config` struct provides comprehensive configuration options:

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
```

### Configuration Examples

#### Minimal Configuration

```go
config := repl.Config{
	Title: "My REPL",
}
```

#### Complete Configuration

```go
config := repl.Config{
	Title:                "Advanced REPL",
	Prompt:               ">>> ",
	Placeholder:          "Enter code, /help for commands",
	Width:                120,
	StartMultiline:       false,
	EnableExternalEditor: true,
	EnableHistory:        true,
	MaxHistorySize:       5000,
}
```

#### Default Configuration

```go
config := repl.DefaultConfig()
// Returns:
// Config{
//     Title:                "REPL",
//     Prompt:               "> ",
//     Placeholder:          "Enter code or /command",
//     Width:                80,
//     StartMultiline:       false,
//     EnableExternalEditor: true,
//     EnableHistory:        true,
//     MaxHistorySize:       1000,
// }
```

## Theming and Styling

The REPL provides a comprehensive theming system with built-in themes and custom styling support.

### Built-in Themes

```go
// Apply built-in themes
model.SetTheme(repl.BuiltinThemes["default"])
model.SetTheme(repl.BuiltinThemes["dark"])
model.SetTheme(repl.BuiltinThemes["light"])
```

### Custom Styling

#### Creating Custom Styles

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

#### Creating Custom Themes

```go
retroTheme := repl.Theme{
	Name: "Retro",
	Styles: repl.Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("53")).
			Padding(0, 1),
		
		Prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("207")).
			Bold(true),
		
		Result: lipgloss.NewStyle().
			Foreground(lipgloss.Color("119")),
		
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("201")).
			Bold(true),
		
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("221")).
			Italic(true),
		
		HelpText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true),
	},
}

model.SetTheme(retroTheme)
```

### Theme Switching

```go
// Dynamic theme switching
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "F1":
			m.repl.SetTheme(repl.BuiltinThemes["default"])
		case "F2":
			m.repl.SetTheme(repl.BuiltinThemes["dark"])
		case "F3":
			m.repl.SetTheme(repl.BuiltinThemes["light"])
		}
	}
	
	var cmd tea.Cmd
	m.repl, cmd = m.repl.Update(msg)
	return m, cmd
}
```

## Embedding Examples

The REPL is designed to be embedded in larger applications using Bubble Tea's message system.

### Basic Embedding

```go
type AppModel struct {
	repl     repl.Model
	mode     string // "repl" or "other"
	otherUI  SomeOtherComponent
}

func (m AppModel) Init() tea.Cmd {
	return m.repl.Init()
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
			if m.mode == "repl" {
				m.mode = "other"
			} else {
				m.mode = "repl"
			}
		}
	}
	
	// Route to appropriate component
	if m.mode == "repl" {
		var cmd tea.Cmd
		m.repl, cmd = m.repl.Update(msg)
		return m, cmd
	} else {
		var cmd tea.Cmd
		m.otherUI, cmd = m.otherUI.Update(msg)
		return m, cmd
	}
}

func (m AppModel) View() string {
	if m.mode == "repl" {
		return m.repl.View()
	}
	return m.otherUI.View()
}
```

### Multi-Pane Application

```go
type MultiPaneApp struct {
	repl      repl.Model
	editor    EditorModel
	files     FileManagerModel
	activePane int
}

func (m MultiPaneApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repl.EvaluationCompleteMsg:
		// Send REPL output to editor
		return m.sendToEditor(msg.Output)
	
	case FileSelectedMsg:
		// Load file into REPL
		return m.loadFileIntoREPL(msg.Path)
	
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+1":
			m.activePane = 0 // REPL
		case "ctrl+2":
			m.activePane = 1 // Editor
		case "ctrl+3":
			m.activePane = 2 // Files
		}
	}
	
	// Update active pane
	var cmd tea.Cmd
	switch m.activePane {
	case 0:
		m.repl, cmd = m.repl.Update(msg)
	case 1:
		m.editor, cmd = m.editor.Update(msg)
	case 2:
		m.files, cmd = m.files.Update(msg)
	}
	
	return m, cmd
}

func (m MultiPaneApp) View() string {
	replView := m.repl.View()
	editorView := m.editor.View()
	filesView := m.files.View()
	
	// Arrange panes horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		replView,
		editorView,
		filesView,
	)
}
```

### Development Environment

```go
type DevEnvironment struct {
	repl       repl.Model
	terminal   TerminalModel
	debugger   DebuggerModel
	layout     LayoutManager
}

func (d DevEnvironment) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repl.EvaluationCompleteMsg:
		// Handle different types of REPL output
		if strings.HasPrefix(msg.Input, "/debug") {
			return d.sendToDebugger(msg)
		}
		if strings.HasPrefix(msg.Input, "/run") {
			return d.sendToTerminal(msg)
		}
	}
	
	// Update all components
	var cmds []tea.Cmd
	var cmd tea.Cmd
	
	d.repl, cmd = d.repl.Update(msg)
	cmds = append(cmds, cmd)
	
	d.terminal, cmd = d.terminal.Update(msg)
	cmds = append(cmds, cmd)
	
	d.debugger, cmd = d.debugger.Update(msg)
	cmds = append(cmds, cmd)
	
	return d, tea.Batch(cmds...)
}
```

## Advanced Usage

### Custom Commands

Add custom slash commands to extend REPL functionality:

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

// Command with arguments
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

// Async command
model.AddCustomCommand("fetch", func(args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) < 1 {
			return repl.EvaluationCompleteMsg{
				Input:  "/fetch",
				Output: "Usage: /fetch <url>",
				Error:  fmt.Errorf("missing URL"),
			}
		}
		
		url := args[0]
		
		// Perform async HTTP request
		go func() {
			resp, err := http.Get(url)
			if err != nil {
				// Send error result
				return
			}
			defer resp.Body.Close()
			
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// Send error result
				return
			}
			
			// Send success result
			// Note: In real implementation, you'd use proper message passing
		}()
		
		return repl.EvaluationCompleteMsg{
			Input:  "/fetch " + url,
			Output: "Fetching...",
			Error:  nil,
		}
	}
})
```

### State Management

```go
type StatefulEvaluator struct {
	variables map[string]interface{}
	functions map[string]func([]interface{}) interface{}
	history   []string
}

func (e *StatefulEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	// Add to history
	e.history = append(e.history, code)
	
	// Handle variable operations
	if strings.HasPrefix(code, "let ") {
		return e.handleVariableDeclaration(code)
	}
	
	if strings.HasPrefix(code, "fn ") {
		return e.handleFunctionDeclaration(code)
	}
	
	// Evaluate expression with current state
	result, err := e.evaluateWithState(code)
	return result, err
}

func (e *StatefulEvaluator) handleVariableDeclaration(code string) (string, error) {
	// Parse: let x = 42
	parts := strings.Split(code, "=")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid variable declaration")
	}
	
	varName := strings.TrimSpace(strings.TrimPrefix(parts[0], "let"))
	value := strings.TrimSpace(parts[1])
	
	// Evaluate the value
	evalResult, err := e.evaluateExpression(value)
	if err != nil {
		return "", err
	}
	
	e.variables[varName] = evalResult
	return fmt.Sprintf("%s = %v", varName, evalResult), nil
}
```

### Error Handling

```go
type RobustEvaluator struct {
	maxExecutionTime time.Duration
	memoryLimit      int64
}

func (e *RobustEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, e.maxExecutionTime)
	defer cancel()
	
	// Run evaluation in goroutine to handle cancellation
	resultChan := make(chan EvaluationResult, 1)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- EvaluationResult{
					Output: "",
					Error:  fmt.Errorf("evaluation panicked: %v", r),
				}
			}
		}()
		
		result, err := e.performEvaluation(ctx, code)
		resultChan <- EvaluationResult{
			Output: result,
			Error:  err,
		}
	}()
	
	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result.Output, result.Error
	case <-ctx.Done():
		return "", fmt.Errorf("evaluation timed out after %v", e.maxExecutionTime)
	}
}
```

## Built-in Features

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Exit REPL |
| `Ctrl+J` | Add line in multiline mode |
| `Ctrl+E` | Open external editor |
| `Up/Down` | Navigate command history |
| `Enter` | Execute code or add line |
| `Tab` | Toggle between modes (if embedded) |

### Slash Commands

| Command | Description |
|---------|-------------|
| `/help` | Show help message |
| `/clear` | Clear history |
| `/quit` | Exit REPL |
| `/multiline` | Toggle multiline mode |
| `/edit` | Open external editor |

### External Editor Integration

```go
// Enable external editor
config := repl.DefaultConfig()
config.EnableExternalEditor = true

// Customize editor
os.Setenv("EDITOR", "vim")
// or
os.Setenv("EDITOR", "code --wait")
```

### History Management

```go
// Enable history
config := repl.DefaultConfig()
config.EnableHistory = true
config.MaxHistorySize = 2000

// Access history programmatically
history := model.GetHistory()
commands := history.GetAll()
```

## Message System

The REPL communicates through a clean message-based API:

### Message Types

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

### Message Handling

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repl.EvaluationCompleteMsg:
		// Handle evaluation results
		if msg.Error != nil {
			// Handle error
			return m.handleEvaluationError(msg.Error)
		}
		
		// Process successful output
		return m.processREPLOutput(msg.Output)
	
	case repl.QuitMsg:
		// Handle REPL quit
		return m.handleREPLQuit()
	
	case repl.ClearHistoryMsg:
		// Handle history clear
		return m.handleHistoryClear()
	
	case repl.ExternalEditorCompleteMsg:
		// Handle external editor result
		if msg.Error != nil {
			return m.handleEditorError(msg.Error)
		}
		
		// Process editor content
		return m.processEditorContent(msg.Content)
	}
	
	// Forward to REPL
	var cmd tea.Cmd
	m.repl, cmd = m.repl.Update(msg)
	return m, cmd
}
```

## Best Practices

### Performance

1. **Async Evaluation**: For long-running evaluations, use goroutines
2. **Memory Management**: Limit history size and clean up resources
3. **Context Cancellation**: Support context cancellation in evaluators
4. **Error Handling**: Implement robust error handling and recovery

### Security

1. **Input Validation**: Validate and sanitize input in evaluators
2. **Resource Limits**: Implement timeouts and memory limits
3. **Sandboxing**: Consider sandboxing for untrusted code evaluation
4. **Permission Checks**: Verify permissions before file operations

### User Experience

1. **Helpful Error Messages**: Provide clear, actionable error messages
2. **Command Completion**: Consider implementing command completion
3. **Syntax Highlighting**: Add syntax highlighting for complex languages
4. **Progress Indicators**: Show progress for long-running operations

### Code Organization

1. **Separate Concerns**: Keep evaluator logic separate from UI
2. **Testability**: Make evaluators easily testable
3. **Documentation**: Document custom commands and evaluator behavior
4. **Error Recovery**: Implement graceful error recovery

## API Reference

### Types

```go
// Main REPL model
type Model struct {
	// ... (see source for full definition)
}

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

// Styling
type Styles struct {
	Title    lipgloss.Style
	Prompt   lipgloss.Style
	Result   lipgloss.Style
	Error    lipgloss.Style
	Info     lipgloss.Style
	HelpText lipgloss.Style
}
```

### Functions

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

// Configuration
func (m *Model) SetTheme(theme Theme)
func (m *Model) SetStyles(styles Styles)
func (m *Model) SetWidth(width int)

// Commands
func (m *Model) AddCustomCommand(name string, handler func([]string) tea.Cmd)

// History
func (m *Model) GetHistory() *History
func (m *Model) ClearHistory()

// State
func (m *Model) IsMultilineMode() bool
func (m *Model) SetMultilineMode(multiline bool)
func (m *Model) IsEvaluating() bool
```

### Built-in Themes

```go
var BuiltinThemes = map[string]Theme{
	"default": Theme{...},
	"dark":    Theme{...},
	"light":   Theme{...},
}
```

---

This documentation provides a comprehensive guide to using the bobatea REPL component. For more examples and advanced usage patterns, see the [examples directory](../examples/repl/) and the [package documentation](../pkg/repl/README.md).
