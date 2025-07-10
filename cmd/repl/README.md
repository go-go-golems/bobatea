# REPL Demo

A comprehensive demonstration of the extracted REPL functionality from the bobatea package.

## Overview

This demo showcases the powerful and flexible REPL (Read-Eval-Print Loop) system that can be embedded in any Go application. It demonstrates multiple evaluators, theming capabilities, and configuration options.

## Features

- **Multiple Evaluators**: JavaScript, Math, and Example evaluators
- **Theming Support**: Built-in themes (default, dark, light) and custom themes
- **Configuration Options**: Customizable width, history, external editor support
- **Custom Commands**: Extensible slash command system
- **Multi-evaluator Switching**: Switch between different evaluators at runtime
- **Command History**: Navigate through previous commands
- **External Editor**: Open code in your preferred editor
- **Multiline Support**: Handle multiline input (where supported)

## Usage

### Main Demo Commands

```bash
# Run JavaScript REPL
go run ./cmd/repl-demo js

# Run Math calculator REPL
go run ./cmd/repl-demo math

# Run Example REPL (echo and basic operations)
go run ./cmd/repl-demo example

# Run Multi-evaluator REPL (switch between evaluators)
go run ./cmd/repl-demo multi

# List available themes
go run ./cmd/repl-demo themes
```

### Command Line Options

All REPL commands support the following flags:

- `--theme`: Theme to use (default, dark, light)
- `--width`: Terminal width (default: 80)
- `--title`: Custom title for the REPL
- `--no-history`: Disable command history
- `--no-editor`: Disable external editor support

Examples:
```bash
# JavaScript REPL with dark theme and custom title
go run ./cmd/repl-demo js --theme dark --title "My JS REPL" --width 100

# Math REPL without history
go run ./cmd/repl-demo math --no-history

# Example REPL without external editor
go run ./cmd/repl-demo example --no-editor
```

## Examples

The `examples/` directory contains standalone examples demonstrating different usage patterns:

### 1. Simple Embedding (`simple-embedding.go`)
The most basic way to embed a REPL in your application.

```bash
go run ./cmd/repl-demo/examples/simple-embedding.go
```

### 2. Custom Evaluator (`custom-evaluator.go`)
Shows how to create a custom evaluator with variable storage.

```bash
go run ./cmd/repl-demo/examples/custom-evaluator.go
```

Features:
- Variable assignment: `var=value`
- Variable lookup: `var`
- Special commands: `time`, `vars`, `clear`
- Custom slash command: `/info`

### 3. Multi-evaluator Switcher (`multi-evaluator-switcher.go`)
Demonstrates switching between different evaluators at runtime.

```bash
go run ./cmd/repl-demo/examples/multi-evaluator-switcher.go
```

Commands:
- `/switch <evaluator>`: Switch evaluator
- `/list`: List available evaluators

### 4. Custom Theme (`custom-theme.go`)
Shows how to create and use custom themes.

```bash
go run ./cmd/repl-demo/examples/custom-theme.go
```

Features:
- Custom "Cyberpunk" theme
- Custom "Minimalist" theme
- Runtime theme switching: `/theme <name>`

## Built-in Commands

All REPLs support these slash commands:

- `/help`: Show help information
- `/clear`: Clear the screen history
- `/quit` or `/exit`: Exit the REPL
- `/multiline`: Toggle multiline mode (if supported)
- `/edit`: Open external editor (if enabled)

## Keyboard Shortcuts

- **Ctrl+C**: Exit REPL
- **Ctrl+J**: Add line in multiline mode
- **Ctrl+E**: Open external editor
- **Up/Down**: Navigate command history

## JavaScript Evaluator

The JavaScript evaluator provides:
- Full ES5/ES6 support via Goja
- Console methods (`console.log`, `console.error`, `console.warn`)
- Variable persistence across evaluations
- Multiline input support
- Module system (where enabled)

Example JavaScript code:
```javascript
let x = 42;
console.log("Hello, World!");
function greet(name) { return "Hello, " + name; }
greet("REPL");
```

## Math Evaluator

The Math evaluator supports:
- Basic arithmetic operations: `+`, `-`, `*`, `/`
- Floating point numbers
- Error handling for division by zero

Example math expressions:
```
3.14 + 2.86
10 / 2
15 * 4
100 - 25
```

## Example Evaluator

The Example evaluator provides:
- Echo functionality: `echo <text>`
- Basic addition: `<number> + <number>`
- Multiline support
- Simple demonstration of evaluator interface

## Theming

### Built-in Themes

1. **Default**: Balanced colors for general use
2. **Dark**: Optimized for dark terminals
3. **Light**: Optimized for light terminals

### Custom Themes

You can create custom themes by defining a `repl.Theme` struct:

```go
customTheme := repl.Theme{
    Name: "MyTheme",
    Styles: repl.Styles{
        Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("32")),
        Prompt:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")),
        Result:   lipgloss.NewStyle().Foreground(lipgloss.Color("36")),
        Error:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")),
        Info:     lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("214")),
        HelpText: lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240")),
    },
}
```

## Architecture

The REPL system is built with a clean architecture:

1. **Evaluator Interface**: Pluggable evaluation backends
2. **Model**: Bubble Tea model handling UI state
3. **Configuration**: Flexible configuration system
4. **Theming**: Separation of logic and presentation
5. **History**: Command history management
6. **External Editor**: Integration with external editors

## Building Your Own Evaluator

To create a custom evaluator, implement the `repl.Evaluator` interface:

```go
type Evaluator interface {
    Evaluate(ctx context.Context, code string) (string, error)
    GetPrompt() string
    GetName() string
    SupportsMultiline() bool
    GetFileExtension() string
}
```

## Building and Running

Ensure you have Go 1.21+ installed, then:

```bash
# Build the demo
go build ./cmd/repl-demo

# Run directly
go run ./cmd/repl-demo <command>
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea): TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss): Styling
- [Cobra](https://github.com/spf13/cobra): CLI framework
- [Goja](https://github.com/dop251/goja): JavaScript engine
- [go-go-goja](https://github.com/go-go-golems/go-go-goja): Enhanced Goja

## License

This demo is part of the bobatea package and follows the same license.
