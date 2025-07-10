# REPL Examples

This directory contains comprehensive examples demonstrating the features and capabilities of the bobatea REPL component. Each example focuses on different aspects of the REPL system, from basic usage to advanced integration patterns.

## Examples Overview

### 1. [Basic Usage](basic-usage/)
**Perfect for beginners**

A simple example showing the most basic usage of the REPL component with a minimal evaluator.

**Features demonstrated:**
- Basic evaluator implementation
- Simple configuration
- Default theming
- Built-in commands

**Use this when:**
- You're new to the REPL component
- You need a simple starting point
- You want to understand the basic architecture

### 2. [Custom Evaluator](custom-evaluator/)
**Intermediate level**

An advanced calculator demonstrating sophisticated evaluator logic with state management.

**Features demonstrated:**
- Complex evaluator with variables and functions
- State management between evaluations
- Mathematical expression parsing
- Custom command implementation
- Error handling

**Use this when:**
- You need to implement complex evaluation logic
- You want to maintain state between evaluations
- You're building a domain-specific language
- You need advanced error handling

### 3. [Embedded in App](embedded-in-app/)
**Advanced integration**

A multi-mode application showing how to embed the REPL as part of a larger application.

**Features demonstrated:**
- Multi-mode application architecture
- REPL integration with other components
- Message handling and routing
- State management across modes
- Logging and monitoring integration

**Use this when:**
- You want to embed REPL in a larger application
- You need multiple application modes
- You want to integrate with logging systems
- You're building a complex development environment

### 4. [Custom Theme](custom-theme/)
**Theming and styling**

A theme switcher demonstrating the comprehensive theming capabilities of the REPL.

**Features demonstrated:**
- Custom theme creation
- Dynamic theme switching
- Theme management systems
- Visual styling with lipgloss
- Theme-related commands

**Use this when:**
- You want to customize the appearance
- You need multiple themes
- You're building a themed application
- You want to create a unique visual experience

## Quick Start

### Running Any Example

```bash
# Navigate to the example directory
cd bobatea/examples/repl/basic-usage

# Run the example
go run main.go
```

### Prerequisites

Make sure you have Go installed and the bobatea package available:

```bash
go mod tidy
```

## Learning Path

We recommend following this learning path:

1. **Start with [Basic Usage](basic-usage/)** - Understand the fundamentals
2. **Progress to [Custom Evaluator](custom-evaluator/)** - Learn advanced evaluation
3. **Explore [Custom Theme](custom-theme/)** - Master styling and theming
4. **Study [Embedded in App](embedded-in-app/)** - Learn integration patterns

## Common Patterns

### Basic Evaluator Pattern

```go
type MyEvaluator struct{}

func (e *MyEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
    // Your evaluation logic here
    return "result", nil
}

func (e *MyEvaluator) GetPrompt() string { return "my> " }
func (e *MyEvaluator) GetName() string { return "My Evaluator" }
func (e *MyEvaluator) SupportsMultiline() bool { return false }
func (e *MyEvaluator) GetFileExtension() string { return ".my" }
```

### REPL Setup Pattern

```go
func main() {
    // Create evaluator
    evaluator := &MyEvaluator{}
    
    // Create configuration
    config := repl.DefaultConfig()
    config.Title = "My REPL"
    
    // Create model
    model := repl.NewModel(evaluator, config)
    
    // Set theme (optional)
    model.SetTheme(repl.BuiltinThemes["dark"])
    
    // Run program
    p := tea.NewProgram(model, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Custom Commands Pattern

```go
model.AddCustomCommand("mycommand", func(args []string) tea.Cmd {
    return func() tea.Msg {
        return repl.EvaluationCompleteMsg{
            Input:  "/mycommand",
            Output: "Command executed successfully",
            Error:  nil,
        }
    }
})
```

### Theme Creation Pattern

```go
customTheme := repl.Theme{
    Name: "My Theme",
    Styles: repl.Styles{
        Title: lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("15")).
            Background(lipgloss.Color("62")),
        
        Prompt: lipgloss.NewStyle().
            Foreground(lipgloss.Color("39")).
            Bold(true),
        
        Result: lipgloss.NewStyle().
            Foreground(lipgloss.Color("46")),
        
        Error: lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Bold(true),
        
        Info: lipgloss.NewStyle().
            Foreground(lipgloss.Color("226")).
            Italic(true),
        
        HelpText: lipgloss.NewStyle().
            Foreground(lipgloss.Color("243")).
            Italic(true),
    },
}

model.SetTheme(customTheme)
```

## Example Comparison

| Feature | Basic Usage | Custom Evaluator | Embedded in App | Custom Theme |
|---------|-------------|------------------|-----------------|--------------|
| **Difficulty** | Beginner | Intermediate | Advanced | Intermediate |
| **Evaluator Complexity** | Simple | Complex | Medium | Simple |
| **State Management** | None | Variables | Multi-mode | Theme state |
| **Custom Commands** | None | Few | Many | Theme commands |
| **Integration** | Standalone | Standalone | Multi-component | Standalone |
| **Theming** | Default | Dark | Dark | Multiple |
| **Lines of Code** | ~80 | ~400 | ~600 | ~500 |

## Common Use Cases

### Command Line Tools
- Use **Basic Usage** for simple command processors
- Use **Custom Evaluator** for complex domain-specific languages
- Use **Custom Theme** for branded command tools

### Development Environments
- Use **Embedded in App** for IDEs and development tools
- Use **Custom Evaluator** for language-specific REPLs
- Use **Custom Theme** for consistent styling

### Interactive Applications
- Use **Embedded in App** for multi-mode applications
- Use **Custom Theme** for user-customizable interfaces
- Use **Custom Evaluator** for application-specific commands

### Educational Tools
- Use **Basic Usage** for learning programming concepts
- Use **Custom Evaluator** for mathematical or scientific tools
- Use **Custom Theme** for engaging visual experiences

## Tips and Best Practices

### Performance
- Keep evaluators lightweight for responsive interaction
- Use context cancellation for long-running evaluations
- Consider async evaluation for complex operations

### User Experience
- Provide clear error messages
- Implement helpful command suggestions
- Use appropriate themes for your target audience
- Add comprehensive help documentation

### Code Organization
- Separate evaluator logic from UI concerns
- Use dependency injection for testability
- Implement proper error handling
- Document your evaluator interface

### Testing
- Write unit tests for your evaluators
- Test with different input scenarios
- Verify theme rendering in different terminals
- Test integration with your application

## Next Steps

After exploring these examples:

1. **Read the [Complete Documentation](../../docs/repl.md)** - Comprehensive guide
2. **Study the [Package Documentation](../../pkg/repl/README.md)** - API reference
3. **Build your own evaluator** - Create something unique
4. **Contribute back** - Share your examples with the community

## Getting Help

If you need help with any of these examples:

1. Check the individual example README files
2. Review the main [REPL documentation](../../docs/repl.md)
3. Study the [API reference](../../pkg/repl/README.md)
4. Look at the source code in [pkg/repl/](../../pkg/repl/)

Each example is designed to be self-contained and runnable, with comprehensive documentation explaining the concepts and implementation details.
