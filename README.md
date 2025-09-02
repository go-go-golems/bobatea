# GO GO BOBA TEA

This package contains either modified versions of common Bubble Tea bubbles or custom UI widgets with enhanced functionality and performance.

## Components

### Core UI Widgets

- **[textarea](pkg/textarea/)** - A modified version of Bubble Tea's textarea that fixes performance issues using memoization (see https://github.com/charmbracelet/bubbles/issues/301)

- **[filepicker](pkg/filepicker/)** - A powerful, feature-rich file selection component with multi-selection, file operations, search, and advanced navigation ([Documentation](docs/filepicker.md))

- **[REPL](pkg/repl/)** - A generic, embeddable REPL (Read-Eval-Print Loop) component with pluggable evaluators, theming, and advanced features ([Documentation](docs/repl.md), [Timeline REPL Tutorial](docs/timeline-repl-integration.md))

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
// Quick run: timeline-first REPL with in-memory bus
import (
    "github.com/go-go-golems/bobatea/pkg/repl"
)

type MyEvaluator struct{}

func (e *MyEvaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
    emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": "You said: " + code}})
    return nil
}
func (e *MyEvaluator) GetPrompt() string        { return "my> " }
func (e *MyEvaluator) GetName() string          { return "MyEval" }
func (e *MyEvaluator) SupportsMultiline() bool  { return false }
func (e *MyEvaluator) GetFileExtension() string { return ".txt" }

func main() {
    cfg := repl.DefaultConfig()
    cfg.Title = "My REPL"
    if err := repl.RunTimelineRepl(&MyEvaluator{}, cfg); err != nil { panic(err) }
}
```

### Integrating via Watermill in a larger app

If you need to embed the REPL model inside a broader Bubble Tea app and publish events from other subsystems:

1) Create a shared in-memory bus and register the REPL transformer and UI forwarder.
2) Construct the REPL model with `bus.Publisher`.
3) Anywhere in your app, publish semantic REPL events to `repl.events` to enrich the timeline.

```go
bus, _ := eventbus.NewInMemoryBus()
repl.RegisterReplToTimelineTransformer(bus)

eval := &MyEvaluator{}
cfg := repl.DefaultConfig()
m := repl.NewModel(eval, cfg, bus.Publisher)
p := tea.NewProgram(m)
timeline.RegisterUIForwarder(bus, p)

// From elsewhere in your app (e.g., a worker), emit an event:
payload, _ := json.Marshal(struct {
    TurnID string    `json:"turn_id"`
    Event  repl.Event `json:"event"`
    Time   time.Time `json:"time"`
}{TurnID: "ext-1", Event: repl.Event{Kind: repl.EventLog, Props: map[string]any{"level": "info", "message": "external note"}}, Time: time.Now()})
bus.Publisher.Publish(eventbus.TopicReplEvents, message.NewMessage(watermill.NewUUID(), payload))
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

