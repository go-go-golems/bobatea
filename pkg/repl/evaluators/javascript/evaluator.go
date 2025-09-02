package javascript

import (
	"context"
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/bobatea/pkg/repl"
	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/pkg/errors"
)

// Evaluator implements the REPL evaluator interface for JavaScript
type Evaluator struct {
	runtime *goja.Runtime
	config  Config
}

// Config holds configuration for the JavaScript evaluator
type Config struct {
	EnableModules     bool
	EnableConsoleLog  bool
	EnableNodeModules bool
	CustomModules     map[string]interface{}
}

// DefaultConfig returns a default configuration for JavaScript evaluation
func DefaultConfig() Config {
	return Config{
		EnableModules:     true,
		EnableConsoleLog:  true,
		EnableNodeModules: true,
		CustomModules:     make(map[string]interface{}),
	}
}

// New creates a new JavaScript evaluator with the given configuration
func New(config Config) (*Evaluator, error) {
	var runtime *goja.Runtime

	if config.EnableModules {
		// Create runtime with module support using go-go-goja engine
		runtime, _ = ggjengine.New()
	} else {
		// Create basic runtime without modules
		runtime = goja.New()
	}

	evaluator := &Evaluator{
		runtime: runtime,
		config:  config,
	}

	// Set up console.log override if enabled
	if config.EnableConsoleLog {
		if err := evaluator.setupConsole(); err != nil {
			return nil, errors.Wrap(err, "failed to setup console")
		}
	}

	// Register custom modules if provided
	for name, module := range config.CustomModules {
		if err := evaluator.registerModule(name, module); err != nil {
			return nil, errors.Wrapf(err, "failed to register custom module %s", name)
		}
	}

	return evaluator, nil
}

// NewWithDefaults creates a new JavaScript evaluator with default configuration
func NewWithDefaults() (*Evaluator, error) {
	return New(DefaultConfig())
}

// setupConsole overrides console.log to provide clean REPL output
func (e *Evaluator) setupConsole() error {
	consoleObj := e.runtime.NewObject()

	// Override console.log to write directly without timestamps
	err := consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}
		fmt.Println(args...)
		return goja.Undefined()
	})
	if err != nil {
		return errors.Wrap(err, "failed to set console.log")
	}

	// Add other console methods
	err = consoleObj.Set("error", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}
		fmt.Printf("Error: %v\n", args...)
		return goja.Undefined()
	})
	if err != nil {
		return errors.Wrap(err, "failed to set console.error")
	}

	err = consoleObj.Set("warn", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}
		fmt.Printf("Warning: %v\n", args...)
		return goja.Undefined()
	})
	if err != nil {
		return errors.Wrap(err, "failed to set console.warn")
	}

	return e.runtime.Set("console", consoleObj)
}

// registerModule registers a custom module with the runtime
func (e *Evaluator) registerModule(name string, module interface{}) error {
	return e.runtime.Set(name, module)
}

// Evaluate executes the given JavaScript code and returns the result
func (e *Evaluator) Evaluate(ctx context.Context, code string) (string, error) {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Execute the JavaScript code
	result, err := e.runtime.RunString(code)
	if err != nil {
		return "", errors.Wrap(err, "JavaScript execution failed")
	}

	// Convert result to string
	var output string
	if result != nil && !goja.IsUndefined(result) {
		output = result.String()
	} else {
		output = "undefined"
	}

	return output, nil
}

// EvaluateStream adapts Evaluate to the streaming interface used by the timeline-based REPL.
func (e *Evaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
	out, err := e.Evaluate(ctx, code)
	if err != nil {
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": fmt.Sprintf("Error: %v", err)}})
		return nil
	}
	emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": out}})
	return nil
}

// GetPrompt returns the prompt string for JavaScript evaluation
func (e *Evaluator) GetPrompt() string {
	return "js>"
}

// GetName returns the name of this evaluator
func (e *Evaluator) GetName() string {
	return "JavaScript"
}

// SupportsMultiline returns true since JavaScript supports multiline input
func (e *Evaluator) SupportsMultiline() bool {
	return true
}

// GetFileExtension returns the file extension for external editor
func (e *Evaluator) GetFileExtension() string {
	return ".js"
}

// GetRuntime returns the underlying Goja runtime (for advanced usage)
func (e *Evaluator) GetRuntime() *goja.Runtime {
	return e.runtime
}

// SetVariable sets a variable in the JavaScript runtime
func (e *Evaluator) SetVariable(name string, value interface{}) error {
	return e.runtime.Set(name, value)
}

// GetVariable gets a variable from the JavaScript runtime
func (e *Evaluator) GetVariable(name string) (interface{}, error) {
	val := e.runtime.Get(name)
	if val == nil {
		return nil, fmt.Errorf("variable %s not found", name)
	}
	return val.Export(), nil
}

// LoadScript loads and executes a JavaScript file
func (e *Evaluator) LoadScript(ctx context.Context, filename string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := e.runtime.RunString(content)
	if err != nil {
		return errors.Wrapf(err, "failed to load script %s", filename)
	}

	return nil
}

// Reset resets the JavaScript runtime to a clean state
func (e *Evaluator) Reset() error {
	// Create a new runtime with the same configuration
	newEvaluator, err := New(e.config)
	if err != nil {
		return errors.Wrap(err, "failed to reset JavaScript evaluator")
	}

	e.runtime = newEvaluator.runtime
	return nil
}

// GetConfig returns the current configuration
func (e *Evaluator) GetConfig() Config {
	return e.config
}

// UpdateConfig updates the evaluator configuration
func (e *Evaluator) UpdateConfig(config Config) error {
	e.config = config

	// Re-setup console if needed
	if config.EnableConsoleLog {
		if err := e.setupConsole(); err != nil {
			return errors.Wrap(err, "failed to re-setup console")
		}
	}

	// Re-register custom modules
	for name, module := range config.CustomModules {
		if err := e.registerModule(name, module); err != nil {
			return errors.Wrapf(err, "failed to re-register custom module %s", name)
		}
	}

	return nil
}

// IsValidCode checks if the given code is syntactically valid JavaScript
func (e *Evaluator) IsValidCode(code string) bool {
	// Try to run the code in a temporary runtime to check syntax
	tempRuntime := goja.New()
	_, err := tempRuntime.RunString(code)
	return err == nil
}

// GetAvailableModules returns a list of available modules
func (e *Evaluator) GetAvailableModules() []string {
	modules := make([]string, 0)

	// Add custom modules
	for name := range e.config.CustomModules {
		modules = append(modules, name)
	}

	// Add standard modules if enabled
	if e.config.EnableModules {
		// These are typical modules available through go-go-goja
		standardModules := []string{
			"database",
			"http",
			"fs",
			"path",
			"url",
		}
		modules = append(modules, standardModules...)
	}

	return modules
}

// GetHelpText returns help text for JavaScript evaluation
func (e *Evaluator) GetHelpText() string {
	var help strings.Builder

	help.WriteString("JavaScript REPL - Powered by Goja\n\n")
	help.WriteString("Available features:\n")
	help.WriteString("- Full ES5/ES6 JavaScript support\n")
	help.WriteString("- Multiline input support\n")
	help.WriteString("- Variable persistence across evaluations\n")

	if e.config.EnableConsoleLog {
		help.WriteString("- Console logging (console.log, console.error, console.warn)\n")
	}

	if e.config.EnableModules {
		help.WriteString("- Module system support (require())\n")
		modules := e.GetAvailableModules()
		if len(modules) > 0 {
			help.WriteString("- Available modules: ")
			help.WriteString(strings.Join(modules, ", "))
			help.WriteString("\n")
		}
	}

	help.WriteString("\nExamples:\n")
	help.WriteString("  let x = 42;\n")
	help.WriteString("  console.log('Hello, World!');\n")
	help.WriteString("  function greet(name) { return 'Hello, ' + name; }\n")

	if e.config.EnableModules {
		help.WriteString("  const db = require('database');\n")
	}

	return help.String()
}

// Compile the interface implementation
var _ repl.Evaluator = (*Evaluator)(nil)
