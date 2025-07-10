package main

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
	"github.com/go-go-golems/bobatea/pkg/repl"
	ggjengine "github.com/go-go-golems/go-go-goja/engine"
)

// JSEvaluator implements JavaScript evaluation using goja
type JSEvaluator struct {
	runtime *goja.Runtime
}

// NewJSEvaluator creates a new JavaScript evaluator
func NewJSEvaluator() (*JSEvaluator, error) {
	// Create a Goja runtime with Node-style require() and native modules enabled
	vm, _ := ggjengine.New()

	// Override console.log to write directly to stdout without timestamps
	consoleObj := vm.NewObject()
	_ = consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}
		fmt.Println(args...)
		return goja.Undefined()
	})
	_ = vm.Set("console", consoleObj)

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

func main() {
	// Create the evaluator
	evaluator, err := NewJSEvaluator()
	if err != nil {
		log.Fatal(err)
	}

	// Create configuration
	config := repl.DefaultConfig()
	config.Title = "JavaScript REPL"
	config.Placeholder = "Enter JavaScript code or /command"

	// Create the model
	model := repl.NewModel(evaluator, config)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
