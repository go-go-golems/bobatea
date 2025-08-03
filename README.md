# GO GO BOBA TEA

This package contains either modified versions of common Bubble Tea bubbles or custom UI widgets with enhanced functionality and performance.

## Components

### Core UI Widgets

- **[textarea](pkg/textarea/)** - A modified version of Bubble Tea's textarea that fixes performance issues using memoization (see https://github.com/charmbracelet/bubbles/issues/301)

- **[filepicker](pkg/filepicker/)** - A powerful, feature-rich file selection component with multi-selection, file operations, search, and advanced navigation ([Documentation](docs/filepicker.md))

- **[REPL](pkg/repl/)** - A generic, embeddable REPL (Read-Eval-Print Loop) component with pluggable evaluators, theming, and advanced features ([Documentation](docs/repl.md))

### Specialized Components

- **[listbox](pkg/listbox/)** - Enhanced listbox with advanced selection and filtering capabilities

- **[buttons](pkg/buttons/)** - Button components with various styles and states

- **[chat](pkg/chat/)** - Chat interface components for conversational applications

- **[overlay](pkg/overlay/)** - Modal and overlay components for layered UI

- **[autocomplete](pkg/autocomplete/)** - Autocomplete input with customizable suggestions

- **[mode-keymap](pkg/mode-keymap/)** - Mode-based keyboard mapping system

- **[sparkline](pkg/sparkline/)** - Terminal data visualization component for displaying trends in compact charts ([Documentation](docs/sparkline.md))

## Quick Start

```go
import (
    "github.com/go-go-golems/bobatea/pkg/repl"
    "github.com/go-go-golems/bobatea/pkg/filepicker"
    "github.com/go-go-golems/bobatea/pkg/textarea"
    "github.com/go-go-golems/bobatea/pkg/sparkline"
)

// Use components in your Bubble Tea application
func main() {
    // Create a REPL with custom evaluator
    evaluator := &MyEvaluator{}
    config := repl.DefaultConfig()
    replModel := repl.NewModel(evaluator, config)
    
    // Create a filepicker
    picker := filepicker.New(
        filepicker.WithStartPath("."),
        filepicker.WithShowPreview(true),
    )
    
    // Create a sparkline for data visualization
    sparklineConfig := sparkline.Config{
        Width:  40,
        Height: 6,
        Style:  sparkline.StyleBars,
        Title:  "CPU Usage",
    }
    chart := sparkline.New(sparklineConfig)
    
    // Run your application
    p := tea.NewProgram(replModel)
    p.Run()
}
```

## Features

- **🔧 Performance Optimized** - Components are optimized for large datasets and complex UIs
- **🎨 Themeable** - Comprehensive theming system with built-in and custom themes
- **📦 Composable** - Components work well together and can be easily combined
- **🔌 Extensible** - Plugin architecture for custom functionality
- **📚 Well Documented** - Comprehensive documentation and examples
- **🧪 Tested** - Thoroughly tested components with examples

## Documentation

- **[REPL Documentation](docs/repl.md)** - Complete guide for the REPL component
- **[Filepicker Documentation](docs/filepicker.md)** - Comprehensive filepicker guide
- **[Sparkline Documentation](docs/sparkline.md)** - Terminal data visualization guide
- **[Examples](examples/)** - Working example applications

## Contributing

Contributions are welcome! Please see the individual component documentation for specific contribution guidelines.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

