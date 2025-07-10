# JavaScript Evaluator for Bobatea REPL

The JavaScript evaluator provides a powerful JavaScript runtime for the Bobatea REPL system, powered by the [Goja](https://github.com/dop251/goja) JavaScript engine.

## Features

- **Full JavaScript Support**: ES5/ES6 JavaScript execution with Goja
- **Module System**: Support for Node.js-style `require()` modules via [go-go-goja](https://github.com/go-go-golems/go-go-goja)
- **Console API**: Built-in console.log, console.error, console.warn functions
- **Variable Persistence**: Variables and functions persist across evaluations
- **Multiline Support**: Full support for multiline JavaScript code
- **Custom Modules**: Easy integration of custom Go modules
- **Error Handling**: Comprehensive error reporting with context
- **External Editor**: Support for editing code in external editors (.js files)

## Installation

```bash
go get github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
)

func main() {
    // Create evaluator with default configuration
    evaluator, err := javascript.NewWithDefaults()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Evaluate JavaScript code
    result, err := evaluator.Evaluate(ctx, "2 + 3")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %s\n", result) // Output: Result: 5
}
```

## Configuration

The JavaScript evaluator supports various configuration options:

```go
config := javascript.Config{
    EnableModules:     true,  // Enable Node.js-style modules
    EnableConsoleLog:  true,  // Enable console.log/error/warn
    EnableNodeModules: true,  // Enable Node.js compatibility modules
    CustomModules: map[string]interface{}{
        "math": map[string]interface{}{
            "pi":  3.14159,
            "add": func(a, b float64) float64 { return a + b },
        },
    },
}

evaluator, err := javascript.New(config)
```

## Available Modules

When modules are enabled, the following modules are typically available:

- **database**: Database connectivity and queries
- **http**: HTTP client functionality
- **fs**: File system operations
- **path**: Path manipulation utilities
- **url**: URL parsing and manipulation

Plus any custom modules you define in the configuration.

## REPL Interface Implementation

The JavaScript evaluator implements the `repl.Evaluator` interface:

```go
type Evaluator interface {
    Evaluate(ctx context.Context, code string) (string, error)
    GetPrompt() string
    GetName() string
    SupportsMultiline() bool
    GetFileExtension() string
}
```

### Interface Methods

- **Evaluate**: Execute JavaScript code and return the result
- **GetPrompt**: Returns `"js> "` as the prompt
- **GetName**: Returns `"JavaScript"` as the evaluator name
- **SupportsMultiline**: Returns `true` (JavaScript supports multiline code)
- **GetFileExtension**: Returns `".js"` for external editor support

## Advanced Usage

### Variable Management

```go
// Set a variable from Go
err := evaluator.SetVariable("myVar", "Hello from Go!")

// Get a variable value
value, err := evaluator.GetVariable("myVar")
```

### Script Loading

```go
script := `
    function factorial(n) {
        if (n <= 1) return 1;
        return n * factorial(n - 1);
    }
    
    let result = factorial(5);
`

err := evaluator.LoadScript(ctx, "factorial.js", script)
```

### Runtime Reset

```go
// Reset the JavaScript runtime to clean state
err := evaluator.Reset()
```

### Code Validation

```go
// Check if code is syntactically valid
isValid := evaluator.IsValidCode("function test() { return 42; }")
```

### Help and Documentation

```go
// Get help text for the evaluator
helpText := evaluator.GetHelpText()

// Get available modules
modules := evaluator.GetAvailableModules()
```

## Integration with Bobatea REPL

To use the JavaScript evaluator with the Bobatea REPL system:

```go
package main

import (
    "github.com/go-go-golems/bobatea/pkg/repl"
    "github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
)

func main() {
    // Create JavaScript evaluator
    jsEvaluator, err := javascript.NewWithDefaults()
    if err != nil {
        panic(err)
    }
    
    // Create REPL with JavaScript evaluator
    config := repl.DefaultConfig()
    config.Title = "JavaScript REPL"
    
    // TODO: Use with REPL model when implemented
    // replModel := repl.NewModel(jsEvaluator, config)
}
```

## Custom Module Example

```go
// Define a custom module
mathModule := map[string]interface{}{
    "pi": 3.14159,
    "add": func(a, b float64) float64 { return a + b },
    "multiply": func(a, b float64) float64 { return a * b },
}

config := javascript.Config{
    EnableModules:     true,
    EnableConsoleLog:  true,
    EnableNodeModules: true,
    CustomModules: map[string]interface{}{
        "math": mathModule,
    },
}

evaluator, err := javascript.New(config)
if err != nil {
    log.Fatal(err)
}

// Use the custom module
result, err := evaluator.Evaluate(ctx, "math.add(math.pi, 1)")
```

## Error Handling

The evaluator provides comprehensive error handling:

```go
result, err := evaluator.Evaluate(ctx, "invalid javascript syntax")
if err != nil {
    // Error will be wrapped with context
    fmt.Printf("Error: %v\n", err)
    // Output: Error: JavaScript execution failed: SyntaxError: Unexpected token...
}
```

## Testing

Run the test suite:

```bash
go test ./pkg/repl/evaluators/javascript/...
```

## Examples

See the `example_test.go` file for comprehensive examples of usage patterns.

## Dependencies

- [Goja](https://github.com/dop251/goja): JavaScript engine
- [go-go-goja](https://github.com/go-go-golems/go-go-goja): Module system and Node.js compatibility
- [pkg/errors](https://github.com/pkg/errors): Error handling

## Contributing

When contributing to the JavaScript evaluator:

1. Follow the existing code style and patterns
2. Add comprehensive tests for new features
3. Update documentation as needed
4. Ensure all tests pass with `go test`

## License

This package is part of the Bobatea project and follows the same license terms.
