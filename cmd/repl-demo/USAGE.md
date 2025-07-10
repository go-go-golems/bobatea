# REPL Demo Usage Guide

## Quick Start

```bash
# Navigate to the demo directory
cd bobatea/cmd/repl-demo/

# Run the JavaScript REPL
go run . js

# Run the Math calculator
go run . math

# Run the Example REPL
go run . example

# Run the Multi-evaluator REPL
go run . multi

# List available themes
go run . themes
```

## Using the Makefile

```bash
# Main demo commands
make js          # JavaScript REPL
make math        # Math REPL
make example     # Example REPL
make multi       # Multi-evaluator REPL
make themes      # List themes

# Examples
make run-simple  # Simple embedding example
make run-custom  # Custom evaluator example
make run-multi   # Multi-evaluator switcher
make run-theme   # Custom theme example
make run-config  # Configuration demo

# Development
make build       # Build the demo
make test-examples # Test all examples build
make clean       # Clean build artifacts
make help        # Show all commands
```

## Command Line Options

### Global Flags
- `--theme <name>`: Set theme (default, dark, light)
- `--width <number>`: Set terminal width (default: 80)
- `--title <string>`: Set custom title
- `--no-history`: Disable command history
- `--no-editor`: Disable external editor support

### Examples
```bash
# JavaScript REPL with dark theme
go run . js --theme dark --title "My JS Console"

# Math REPL with wider display
go run . math --width 120

# Example REPL without history
go run . example --no-history
```

## REPL Commands

### Built-in Commands (all REPLs)
- `/help`: Show help information
- `/clear`: Clear screen history
- `/quit` or `/exit`: Exit the REPL
- `/multiline`: Toggle multiline mode (where supported)
- `/edit`: Open external editor (if enabled)

### Keyboard Shortcuts
- **Ctrl+C**: Exit REPL
- **Ctrl+J**: Add line in multiline mode
- **Ctrl+E**: Open external editor
- **Up/Down**: Navigate command history

### JavaScript REPL
```javascript
// Variables and functions
let x = 42;
const name = "World";
function greet(n) { return "Hello, " + n; }

// Console methods
console.log("Hello, World!");
console.error("This is an error");
console.warn("This is a warning");

// Multiline support (Ctrl+J to add lines)
function factorial(n) {
  if (n <= 1) return 1;
  return n * factorial(n - 1);
}
```

### Math REPL
```
# Basic operations
5 + 3
10.5 - 2.3
4 * 7
15 / 3

# Floating point
3.14159 + 2.71828
22 / 7
```

### Example REPL
```
# Echo functionality
echo Hello, World!

# Basic math
5 + 3
10 + 15

# Any other text
This will be echoed back
```

## Custom Evaluator Examples

### Variable Storage (custom-evaluator.go)
```
# Set variables
name=John
age=30
city=New York

# Get variables
name
age

# Special commands
time    # Current time
vars    # List all variables
clear   # Clear all variables

# Custom command
/info   # Show evaluator info
```

### Multi-evaluator Switcher
```
# Switch between evaluators
/switch js       # Switch to JavaScript
/switch math     # Switch to Math
/switch example  # Switch to Example

# List available evaluators
/list
```

### Theme Switcher (custom-theme.go)
```
# Switch themes
/theme cyberpunk    # Custom cyberpunk theme
/theme minimalist   # Custom minimalist theme
/theme dark         # Built-in dark theme
/theme light        # Built-in light theme

# List themes
/theme
```

## Integration Examples

### Simple Embedding
```go
evaluator := repl.NewExampleEvaluator()
config := repl.DefaultConfig()
model := repl.NewModel(evaluator, config)
tea.NewProgram(model, tea.WithAltScreen()).Run()
```

### Custom Evaluator
```go
type MyEvaluator struct{}

func (m *MyEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
    // Your evaluation logic here
    return "Result: " + code, nil
}

func (m *MyEvaluator) GetPrompt() string { return "my> " }
func (m *MyEvaluator) GetName() string { return "MyEvaluator" }
func (m *MyEvaluator) SupportsMultiline() bool { return false }
func (m *MyEvaluator) GetFileExtension() string { return ".txt" }
```

### Custom Theme
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
model.SetTheme(customTheme)
```

## Troubleshooting

### Build Issues
```bash
# Update dependencies
go mod tidy

# Clean and rebuild
make clean
make build
```

### External Editor Not Working
- Set the `EDITOR` environment variable: `export EDITOR=nano`
- Or use `vim`, `vi`, or your preferred editor
- Ensure the editor is installed and in PATH

### Theme Not Applying
- Check theme name spelling
- Use `/theme` command to list available themes
- Verify terminal supports colors

### Command History Not Working
- Ensure `--no-history` flag is not set
- Check if history size is set too low
- History is session-based (not persistent across restarts)

## Tips

1. **Multiline Mode**: Use Ctrl+J to add lines, then press Enter on empty line to execute
2. **External Editor**: Set `EDITOR` environment variable for your preferred editor
3. **Custom Commands**: Use `/` prefix for commands vs regular input
4. **Themes**: Try different themes to find what works best for your terminal
5. **Width**: Adjust width based on your terminal size for better formatting
