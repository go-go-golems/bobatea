# JavaScript Evaluator Usage Summary

## Overview

The JavaScript evaluator demonstrates how to implement the generic REPL evaluator interface in Bobatea. It extracts JavaScript-specific logic from the Jesus REPL and provides a pluggable evaluator that can be used with the generic REPL system.

## Evaluator Interface

The evaluator implements the `repl.Evaluator` interface:

```go
type Evaluator interface {
    // Core evaluation method
    Evaluate(ctx context.Context, code string) (string, error)
    
    // Interface properties
    GetPrompt() string
    GetName() string
    SupportsMultiline() bool
    GetFileExtension() string
}
```

## Implementation Details

### Core Methods

1. **Evaluate**: Executes JavaScript code using the Goja engine
2. **GetPrompt**: Returns `"js>"` for JavaScript evaluation
3. **GetName**: Returns `"JavaScript"` for display purposes
4. **SupportsMultiline**: Returns `true` (JavaScript supports multiline code)
5. **GetFileExtension**: Returns `".js"` for external editor support

### Key Features

- **Goja Engine**: Uses the Goja JavaScript engine for execution
- **Module Support**: Integrates with go-go-goja for Node.js-style modules
- **Console Override**: Provides clean console.log output without timestamps
- **Variable Persistence**: Variables persist across evaluations
- **Error Handling**: Comprehensive error wrapping and reporting

## Usage Examples

### Basic Usage

```go
// Create evaluator
evaluator, err := javascript.NewWithDefaults()
if err != nil {
    log.Fatal(err)
}

// Evaluate code
result, err := evaluator.Evaluate(context.Background(), "2 + 3")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result) // Output: 5
```

### Custom Configuration

```go
config := javascript.Config{
    EnableModules:     true,
    EnableConsoleLog:  true,
    EnableNodeModules: true,
    CustomModules: map[string]interface{}{
        "math": map[string]interface{}{
            "pi": 3.14159,
            "add": func(a, b float64) float64 { return a + b },
        },
    },
}

evaluator, err := javascript.New(config)
```

### Integration with Generic REPL

The evaluator is designed to work with the generic REPL system:

```go
// Create evaluator
jsEvaluator, err := javascript.NewWithDefaults()
if err != nil {
    panic(err)
}

// Use with generic REPL (when implemented)
config := repl.DefaultConfig()
config.Title = "JavaScript REPL"

// TODO: Use with REPL model
// replModel := repl.NewModel(jsEvaluator, config)
```

## Configuration Options

The `javascript.Config` struct provides flexible configuration:

```go
type Config struct {
    EnableModules     bool                    // Enable go-go-goja modules
    EnableConsoleLog  bool                    // Enable console.log override
    EnableNodeModules bool                    // Enable Node.js compatibility
    CustomModules     map[string]interface{}  // Custom module definitions
}
```

## Available Modules

When modules are enabled, the following are typically available:

- **database**: Database connectivity
- **http**: HTTP client functionality
- **fs**: File system operations
- **path**: Path manipulation
- **url**: URL parsing
- **Custom modules**: User-defined modules

## Advanced Features

### Variable Management

```go
// Set variables from Go
evaluator.SetVariable("myVar", "Hello from Go")

// Get variables
value, err := evaluator.GetVariable("myVar")
```

### Script Loading

```go
script := `
    function greet(name) {
        return 'Hello, ' + name + '!';
    }
`
err := evaluator.LoadScript(ctx, "greet.js", script)
```

### Runtime Reset

```go
// Reset to clean state
err := evaluator.Reset()
```

### Code Validation

```go
// Check syntax validity
isValid := evaluator.IsValidCode("function test() { return 42; }")
```

## Creating Custom Evaluators

The JavaScript evaluator serves as a template for creating other evaluators:

1. **Implement the Interface**: Implement all methods of `repl.Evaluator`
2. **Handle Context**: Respect context cancellation in `Evaluate`
3. **Error Handling**: Wrap errors with appropriate context
4. **Configuration**: Provide flexible configuration options
5. **Testing**: Include comprehensive unit tests

### Example Structure

```go
package mylang

import (
    "context"
    "github.com/go-go-golems/bobatea/pkg/repl"
)

type Evaluator struct {
    // Your evaluator fields
}

func (e *Evaluator) Evaluate(ctx context.Context, code string) (string, error) {
    // Your evaluation logic
}

func (e *Evaluator) GetPrompt() string {
    return "mylang>"
}

func (e *Evaluator) GetName() string {
    return "MyLanguage"
}

func (e *Evaluator) SupportsMultiline() bool {
    return true
}

func (e *Evaluator) GetFileExtension() string {
    return ".mylang"
}

// Compile-time interface check
var _ repl.Evaluator = (*Evaluator)(nil)
```

## Testing

The evaluator includes comprehensive tests:

- Unit tests for all interface methods
- Integration tests for JavaScript evaluation
- Configuration tests
- Error handling tests
- Example tests for documentation

Run tests with:

```bash
go test ./pkg/repl/evaluators/javascript/...
```

## Dependencies

- **goja**: JavaScript engine
- **go-go-goja**: Module system and Node.js compatibility
- **pkg/errors**: Error handling utilities
- **testify**: Testing framework

## Benefits of This Design

1. **Separation of Concerns**: Language-specific logic is isolated
2. **Reusability**: Can be used with any REPL UI implementation
3. **Testability**: Easy to test language features independently
4. **Extensibility**: Easy to add new language evaluators
5. **Configuration**: Flexible configuration without changing the interface

This design allows users to choose which evaluators to include in their applications, reducing dependencies and binary size when not needed.
