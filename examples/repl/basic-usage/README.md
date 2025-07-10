# Basic REPL Usage Example

This example demonstrates the most basic usage of the bobatea REPL component with a simple echo evaluator.

## What it does

- Creates a simple evaluator that echoes back user input
- Handles numbers by showing both decimal and hexadecimal representations
- Responds to greetings with a friendly message
- Uses default configuration with minimal customization

## Running the example

```bash
go run main.go
```

## Key Features Demonstrated

- **Basic Evaluator Implementation** - Shows how to implement the `Evaluator` interface
- **Simple Configuration** - Uses `DefaultConfig()` with basic customization
- **Input Processing** - Demonstrates basic string processing and formatting
- **Error Handling** - Shows simple error handling patterns

## What you'll see

When you run this example, you'll get a REPL that:

1. **Echoes input**: Type any text and it will be echoed back
2. **Handles numbers**: Type a number like `42` and see it displayed as both decimal and hex
3. **Responds to greetings**: Type "hello" and get a friendly response
4. **Supports built-in commands**: Try `/help`, `/quit`, `/clear`

## Code Structure

### EchoEvaluator

```go
type EchoEvaluator struct{}

func (e *EchoEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
    // Your evaluation logic here
    return fmt.Sprintf("You said: %s", code), nil
}

func (e *EchoEvaluator) GetPrompt() string { return "echo> " }
func (e *EchoEvaluator) GetName() string { return "Echo" }
func (e *EchoEvaluator) SupportsMultiline() bool { return false }
func (e *EchoEvaluator) GetFileExtension() string { return ".txt" }
```

### Configuration

```go
config := repl.DefaultConfig()
config.Title = "Basic Echo REPL"
config.Placeholder = "Type something to echo back..."
config.Width = 80
```

### Model Creation

```go
model := repl.NewModel(evaluator, config)
p := tea.NewProgram(model, tea.WithAltScreen())
```

## Try These Commands

Once the REPL is running, try:

- `hello` - Get a greeting
- `42` - See number formatting
- `hello world` - Basic echo
- `/help` - Show built-in help
- `/quit` - Exit the REPL
- `/clear` - Clear the history

## Next Steps

After trying this basic example, check out:

- [Custom Evaluator Example](../custom-evaluator/) - More complex evaluation logic
- [Embedded in App Example](../embedded-in-app/) - Integration with larger applications
- [Custom Theme Example](../custom-theme/) - Styling and theming
