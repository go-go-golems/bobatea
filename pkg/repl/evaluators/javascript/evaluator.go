package javascript

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/go-go-golems/bobatea/pkg/repl"
	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsparse"
	"github.com/pkg/errors"
)

var (
	requireAliasPattern = regexp.MustCompile(`(?m)\b(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*require\(\s*['"]([^'"]+)['"]\s*\)`)

	nodeModuleCandidates = map[string][]jsparse.CompletionCandidate{
		"fs": {
			{Label: "readFile", Kind: jsparse.CandidateMethod, Detail: "fs method"},
			{Label: "writeFile", Kind: jsparse.CandidateMethod, Detail: "fs method"},
			{Label: "existsSync", Kind: jsparse.CandidateMethod, Detail: "fs method"},
			{Label: "mkdirSync", Kind: jsparse.CandidateMethod, Detail: "fs method"},
		},
		"path": {
			{Label: "join", Kind: jsparse.CandidateMethod, Detail: "path method"},
			{Label: "resolve", Kind: jsparse.CandidateMethod, Detail: "path method"},
			{Label: "dirname", Kind: jsparse.CandidateMethod, Detail: "path method"},
			{Label: "basename", Kind: jsparse.CandidateMethod, Detail: "path method"},
			{Label: "extname", Kind: jsparse.CandidateMethod, Detail: "path method"},
		},
		"url": {
			{Label: "URL", Kind: jsparse.CandidateFunction, Detail: "constructor"},
			{Label: "URLSearchParams", Kind: jsparse.CandidateFunction, Detail: "constructor"},
			{Label: "parse", Kind: jsparse.CandidateMethod, Detail: "url method"},
		},
	}
)

// Evaluator implements the REPL evaluator interface for JavaScript
type Evaluator struct {
	runtime  *goja.Runtime
	tsParser *jsparse.TSParser
	config   Config
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
	if parser, parserErr := jsparse.NewTSParser(); parserErr == nil {
		evaluator.tsParser = parser
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

// CompleteInput resolves JavaScript completions using jsparse CST + resolver primitives.
func (e *Evaluator) CompleteInput(_ context.Context, req repl.CompletionRequest) (repl.CompletionResult, error) {
	if e.tsParser == nil {
		return repl.CompletionResult{Show: false}, nil
	}

	input := req.Input
	cursor := clampCursor(req.CursorByte, len(input))

	root := e.tsParser.Parse([]byte(input))
	if root == nil {
		return repl.CompletionResult{Show: false}, nil
	}

	analysis := jsparse.Analyze("repl-input.js", input, nil)
	row, col := byteOffsetToRowCol(input, cursor)
	ctx := analysis.CompletionContextAt(root, row, col)
	if ctx.Kind == jsparse.CompletionNone {
		return repl.CompletionResult{Show: false}, nil
	}

	candidates := jsparse.ResolveCandidates(ctx, analysis.Index, root)
	if ctx.Kind == jsparse.CompletionProperty {
		aliases := extractRequireAliases(input)
		if moduleName, ok := aliases[ctx.BaseExpr]; ok {
			if moduleCandidates, ok := nodeModuleCandidates[moduleName]; ok {
				candidates = append(candidates, moduleCandidates...)
			}
		}
	}
	if len(candidates) == 0 {
		return repl.CompletionResult{
			Show:        false,
			ReplaceFrom: cursor,
			ReplaceTo:   cursor,
		}, nil
	}

	replaceFrom := clampCursor(cursor-len(ctx.PartialText), len(input))
	replaceTo := cursor
	if replaceFrom > replaceTo {
		replaceFrom = replaceTo
	}

	suggestions := make([]autocomplete.Suggestion, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate.Label == "" {
			continue
		}
		if _, ok := seen[candidate.Label]; ok {
			continue
		}
		seen[candidate.Label] = struct{}{}

		display := candidate.Label
		icon := candidate.Kind.Icon()
		if icon != "" {
			display = icon + " " + display
		}
		if candidate.Detail != "" {
			display += " - " + candidate.Detail
		}

		suggestions = append(suggestions, autocomplete.Suggestion{
			Id:          candidate.Label,
			Value:       candidate.Label,
			DisplayText: display,
		})
	}

	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestions[i].Value < suggestions[j].Value
	})

	show := len(suggestions) > 0
	if req.Reason == repl.CompletionReasonDebounce {
		if ctx.Kind == jsparse.CompletionIdentifier && len(ctx.PartialText) == 0 {
			show = false
		}
	}

	return repl.CompletionResult{
		Show:        show,
		Suggestions: suggestions,
		ReplaceFrom: replaceFrom,
		ReplaceTo:   replaceTo,
	}, nil
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
	e.tsParser = newEvaluator.tsParser
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

func clampCursor(cursor, upperBound int) int {
	if cursor < 0 {
		return 0
	}
	if cursor > upperBound {
		return upperBound
	}
	return cursor
}

func byteOffsetToRowCol(input string, cursor int) (int, int) {
	cursor = clampCursor(cursor, len(input))
	row, col := 0, 0
	for i := 0; i < cursor; i++ {
		if input[i] == '\n' {
			row++
			col = 0
			continue
		}
		col++
	}
	return row, col
}

func extractRequireAliases(input string) map[string]string {
	aliases := make(map[string]string)
	matches := requireAliasPattern.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		alias := match[1]
		moduleName := match[2]
		aliases[alias] = moduleName
	}
	return aliases
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
