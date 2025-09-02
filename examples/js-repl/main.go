package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/logutil"
	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/rs/zerolog"
)

// JSEvaluator implements JavaScript evaluation using goja
type JSEvaluator struct {
	runtime *goja.Runtime
}

// NewJSEvaluator creates a new JavaScript evaluator
func NewJSEvaluator() (*JSEvaluator, error) {
	// Create a Goja runtime with Node-style require() and native modules enabled
	vm, _ := ggjengine.New()

	return &JSEvaluator{runtime: vm}, nil
}

func (e *JSEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	result, err := e.runtime.RunString(code)
	if err != nil {
		return "", err
	}

	// Convert result to string
	if result != nil && !goja.IsUndefined(result) {
		return result.String(), nil
	}

	return "undefined", nil
}

func (e *JSEvaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
	// Override console functions to stream as timeline entities
	vm := e.runtime
	consoleObj := vm.NewObject()
	_ = consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 1 {
			v := call.Arguments[0].Export()
			switch v.(type) {
			case map[string]any, map[interface{}]interface{}, []any:
				// Structured event log, rendered as YAML (no markdown fences)
				emit(repl.Event{Kind: repl.EventStructuredLog, Props: map[string]any{"level": "info", "data": v}})
				return goja.Undefined()
			}
		}
		// Otherwise, plain log entry
		parts := make([]string, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			parts = append(parts, fmt.Sprint(arg.Export()))
		}
		s := strings.Join(parts, " ")
		emit(repl.Event{Kind: repl.EventLog, Props: map[string]any{"level": "info", "message": s}})
		return goja.Undefined()
	})
	_ = consoleObj.Set("error", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 1 {
			v := call.Arguments[0].Export()
			switch v.(type) {
			case map[string]any, map[interface{}]interface{}, []any:
				emit(repl.Event{Kind: repl.EventStructuredLog, Props: map[string]any{"level": "error", "data": v}})
				return goja.Undefined()
			}
		}
		parts := make([]string, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			parts = append(parts, fmt.Sprint(arg.Export()))
		}
		s := strings.Join(parts, " ")
		emit(repl.Event{Kind: repl.EventLog, Props: map[string]any{"level": "error", "message": s}})
		return goja.Undefined()
	})
	_ = consoleObj.Set("warn", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 1 {
			v := call.Arguments[0].Export()
			switch v.(type) {
			case map[string]any, map[interface{}]interface{}, []any:
				emit(repl.Event{Kind: repl.EventStructuredLog, Props: map[string]any{"level": "warn", "data": v}})
				return goja.Undefined()
			}
		}
		parts := make([]string, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			parts = append(parts, fmt.Sprint(arg.Export()))
		}
		s := strings.Join(parts, " ")
		emit(repl.Event{Kind: repl.EventLog, Props: map[string]any{"level": "warn", "message": s}})
		return goja.Undefined()
	})
	_ = vm.Set("console", consoleObj)

	// Execute code and emit result
	res, err := vm.RunString(code)
	if err != nil {
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": fmt.Sprintf("Error: %v", err)}})
		return nil
	}
	if res != nil && !goja.IsUndefined(res) {
		out := res.String()
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": out}})
	}
	return nil
}

func (e *JSEvaluator) GetPrompt() string {
	return "js> "
}

func (e *JSEvaluator) GetName() string {
	return "JavaScript"
}

func (e *JSEvaluator) SupportsMultiline() bool {
	return true
}

func (e *JSEvaluator) GetFileExtension() string {
	return ".js"
}

func parseLevel(s string) zerolog.Level {
	switch strings.ToLower(s) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error", "err":
		return zerolog.ErrorLevel
	default:
		return zerolog.ErrorLevel
	}
}

func main() {
	// CLI flags for logging
	ll := flag.String("log-level", "error", "log level: trace, debug, info, warn, error")
	lf := flag.String("log-file", "", "log file path (optional)")
	flag.Parse()

	level := parseLevel(*ll)
	if *lf != "" {
		logutil.InitTUILoggingToFile(level, *lf)
	} else {
		logutil.InitTUILoggingToDiscard(level)
	}

	// Create the evaluator
	evaluator, err := NewJSEvaluator()
	if err != nil {
		log.Fatal(err)
	}

	// Create configuration
	config := repl.DefaultConfig()
	config.Title = "JavaScript REPL"
	config.Placeholder = "Enter JavaScript code or /command"

	// Wire bus and forwarders
	bus, err := eventbus.NewInMemoryBus()
	if err != nil {
		log.Fatal(err)
	}
	repl.RegisterReplToTimelineTransformer(bus)

	model := repl.NewModel(evaluator, config, bus.Publisher)
	p := tea.NewProgram(model, tea.WithAltScreen())
	timeline.RegisterUIForwarder(bus, p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 2)
	go func() { errs <- bus.Run(ctx) }()
	go func() { _, e := p.Run(); cancel(); errs <- e }()
	if e := <-errs; e != nil {
		log.Fatal(e)
	}
}
