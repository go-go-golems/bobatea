# Custom Evaluator Example

This example demonstrates how to create a sophisticated custom evaluator for the bobatea REPL component. It implements an advanced calculator with variables, functions, and state management.

## What it does

- **Advanced Calculator** - Supports complex mathematical expressions
- **Variables** - Store and reuse values (e.g., `x = 42`)
- **Built-in Functions** - Mathematical functions like `sin`, `cos`, `sqrt`, etc.
- **Constants** - Pre-defined constants like `pi` and `e`
- **State Management** - Maintains variables and calculation history
- **Custom Commands** - Additional slash commands for calculator operations

## Running the example

```bash
go run main.go
```

## Key Features Demonstrated

- **Complex Evaluator Logic** - Shows how to implement stateful evaluation
- **Variable Management** - Demonstrates variable storage and retrieval
- **Function Processing** - Shows how to handle function calls in expressions
- **Custom Commands** - Adds REPL-specific commands
- **Error Handling** - Robust error handling for mathematical operations
- **State Persistence** - Maintains state between evaluations

## What you'll see

When you run this example, you'll get a powerful calculator REPL that supports:

### Basic Arithmetic
```
calc> 2 + 3 * 4
14

calc> (5 + 3) * 2
16

calc> 10 / 3
3.33333
```

### Variables
```
calc> x = 42
x = 42

calc> y = x * 2
y = 84

calc> x + y
126
```

### Functions
```
calc> sin(pi/2)
1

calc> sqrt(16)
4

calc> log(e)
1
```

### State Management
```
calc> vars
Variables:
  pi = 3.14159
  e = 2.71828
  x = 42
  y = 84
  ans = 126

calc> funcs
Available functions:
  sin(x)
  cos(x)
  sqrt(x)
  log(x)
  ...

calc> hist
Recent expressions:
  1: 2 + 3 * 4
  2: x = 42
  3: y = x * 2
  4: x + y
```

### Custom Commands
```
calc> /clear-vars
All variables cleared (except pi and e)

calc> /reset
Calculator state reset
```

## Code Structure

### AdvancedCalculatorEvaluator

The main evaluator struct maintains state:

```go
type AdvancedCalculatorEvaluator struct {
    variables map[string]float64
    functions map[string]func(float64) float64
    history   []string
}
```

### Key Methods

#### Variable Assignment
```go
func (e *AdvancedCalculatorEvaluator) handleAssignment(code string) (string, error) {
    parts := strings.Split(code, "=")
    varName := strings.TrimSpace(parts[0])
    expr := strings.TrimSpace(parts[1])
    
    result, err := e.evaluateExpression(expr)
    if err != nil {
        return "", err
    }
    
    e.variables[varName] = result
    return fmt.Sprintf("%s = %.6g", varName, result), nil
}
```

#### Expression Evaluation
```go
func (e *AdvancedCalculatorEvaluator) evaluateExpression(expr string) (float64, error) {
    // Replace variables with their values
    for varName, value := range e.variables {
        expr = strings.ReplaceAll(expr, varName, fmt.Sprintf("%.15g", value))
    }
    
    // Handle function calls
    for funcName, fn := range e.functions {
        // Process function calls in the expression
    }
    
    // Evaluate the remaining expression
    return e.evaluateBasicExpression(expr)
}
```

### Built-in Functions

The calculator includes mathematical functions:

```go
calc.functions["sin"] = math.Sin
calc.functions["cos"] = math.Cos
calc.functions["tan"] = math.Tan
calc.functions["sqrt"] = math.Sqrt
calc.functions["log"] = math.Log
calc.functions["abs"] = math.Abs
calc.functions["ceil"] = math.Ceil
calc.functions["floor"] = math.Floor
```

### Constants

Pre-defined mathematical constants:

```go
calc.variables["pi"] = math.Pi
calc.variables["e"] = math.E
```

## Try These Examples

Once the calculator is running, try:

### Basic Math
- `2 + 3 * 4` - Order of operations
- `(5 + 3) * 2` - Parentheses
- `10 / 3` - Division
- `2 ^ 3` - Power operator

### Variables
- `x = 10` - Variable assignment
- `y = x * 2` - Using variables
- `ans` - Previous result (automatically stored)

### Functions
- `sin(pi/2)` - Sine function
- `sqrt(16)` - Square root
- `log(e)` - Natural logarithm
- `abs(-5)` - Absolute value

### State Commands
- `vars` - Show all variables
- `funcs` - Show available functions
- `hist` - Show calculation history

### Custom Commands
- `/clear-vars` - Clear all variables
- `/reset` - Reset calculator state
- `/help` - Show built-in help

## Advanced Features

### Expression Parsing
The evaluator includes a simplified expression parser that handles:
- Basic arithmetic operations (`+`, `-`, `*`, `/`, `^`)
- Variable substitution
- Function calls
- Parentheses (basic support)

### Error Handling
Comprehensive error handling for:
- Division by zero
- Invalid expressions
- Undefined variables
- Invalid function calls

### State Management
- Variables persist between evaluations
- History tracking for debugging
- Result storage in `ans` variable

## Extending the Calculator

You can extend this calculator by:

1. **Adding more functions**:
```go
calc.functions["factorial"] = func(x float64) float64 {
    if x < 0 || x != math.Floor(x) {
        return math.NaN()
    }
    result := 1.0
    for i := 1; i <= int(x); i++ {
        result *= float64(i)
    }
    return result
}
```

2. **Adding more operators**:
```go
// Add modulo operator
if strings.Contains(expr, "%") {
    // Handle modulo operation
}
```

3. **Adding more commands**:
```go
model.AddCustomCommand("save", func(args []string) tea.Cmd {
    // Save variables to file
})
```

## Next Steps

After trying this example, check out:

- [Embedded in App Example](../embedded-in-app/) - Integration with larger applications
- [Custom Theme Example](../custom-theme/) - Styling and theming
- [Basic Usage Example](../basic-usage/) - Simpler implementation patterns
