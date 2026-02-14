package javascript

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/go-go-golems/bobatea/pkg/repl"
	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsparse"
	"github.com/pkg/errors"
)

var (
	helpBarSymbolSignatures = map[string]string{
		"console":        "console: object (log, error, warn, info, debug, table)",
		"console.log":    "console.log(...args): void",
		"console.error":  "console.error(...args): void",
		"console.warn":   "console.warn(...args): void",
		"console.info":   "console.info(...args): void",
		"console.debug":  "console.debug(...args): void",
		"console.table":  "console.table(data, columns?): void",
		"Math":           "Math: object (numeric helpers)",
		"Math.max":       "Math.max(...values): number",
		"Math.min":       "Math.min(...values): number",
		"Math.random":    "Math.random(): number",
		"Math.floor":     "Math.floor(value): number",
		"Math.ceil":      "Math.ceil(value): number",
		"Math.round":     "Math.round(value): number",
		"JSON":           "JSON: object (parse, stringify)",
		"JSON.parse":     "JSON.parse(text): any",
		"JSON.stringify": "JSON.stringify(value): string",
		"fs":             "fs: module alias (file system APIs)",
		"fs.readFile":    "fs.readFile(path, [options], callback): void",
		"fs.writeFile":   "fs.writeFile(path, data, [options], callback): void",
		"fs.existsSync":  "fs.existsSync(path): bool",
		"fs.mkdirSync":   "fs.mkdirSync(path, [options]): string | undefined",
		"path":           "path: module alias (path utilities)",
		"path.join":      "path.join(...parts): string",
		"path.resolve":   "path.resolve(...parts): string",
		"path.dirname":   "path.dirname(path): string",
		"path.basename":  "path.basename(path): string",
		"path.extname":   "path.extname(path): string",
		"url":            "url: module alias (URL utilities)",
		"url.parse":      "url.parse(input): URLRecord",
		"url.URL":        "url.URL(input): URL",
	}
)

// Evaluator implements the REPL evaluator interface for JavaScript
type Evaluator struct {
	runtime            *goja.Runtime
	runtimeMu          sync.Mutex
	tsParser           *jsparse.TSParser
	tsMu               sync.Mutex
	runtimeDeclaredMu  sync.RWMutex
	runtimeDeclaredIDs map[string]jsparse.CompletionCandidate
	config             Config
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
		runtime:            runtime,
		runtimeDeclaredIDs: map[string]jsparse.CompletionCandidate{},
		config:             config,
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
	e.runtimeMu.Lock()
	result, err := e.runtime.RunString(code)
	if err != nil {
		e.runtimeMu.Unlock()
		return "", errors.Wrap(err, "JavaScript execution failed")
	}

	// Convert result to string
	var output string
	if result != nil && !goja.IsUndefined(result) {
		output = result.String()
	} else {
		output = "undefined"
	}
	e.runtimeMu.Unlock()
	e.observeRuntimeDeclarations(code)

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

	e.tsMu.Lock()
	root := e.tsParser.Parse([]byte(input))
	e.tsMu.Unlock()
	if root == nil {
		return repl.CompletionResult{Show: false}, nil
	}

	analysis := jsparse.Analyze("repl-input.js", input, nil)
	row, col := byteOffsetToRowCol(input, cursor)
	ctx := analysis.CompletionContextAt(root, row, col)
	if ctx.Kind == jsparse.CompletionNone {
		return repl.CompletionResult{Show: false}, nil
	}

	e.runtimeMu.Lock()
	candidates := jsparse.AugmentREPLCandidates(
		e.runtime,
		input,
		ctx,
		jsparse.ResolveCandidates(ctx, analysis.Index, root),
		e.runtimeIdentifierHints(),
	)
	e.runtimeMu.Unlock()
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

// GetHelpBar resolves contextual one-line symbol help for the JS REPL input.
func (e *Evaluator) GetHelpBar(_ context.Context, req repl.HelpBarRequest) (repl.HelpBarPayload, error) {
	if e.tsParser == nil {
		return repl.HelpBarPayload{Show: false}, nil
	}

	input := req.Input
	cursor := clampCursor(req.CursorByte, len(input))

	e.tsMu.Lock()
	root := e.tsParser.Parse([]byte(input))
	e.tsMu.Unlock()
	if root == nil {
		return repl.HelpBarPayload{Show: false}, nil
	}

	analysis := jsparse.Analyze("repl-input.js", input, nil)
	row, col := byteOffsetToRowCol(input, cursor)
	ctx := analysis.CompletionContextAt(root, row, col)

	token, _, _ := tokenAtCursor(input, cursor)
	token = strings.TrimSpace(token)
	if token == "" && ctx.Kind == jsparse.CompletionNone {
		return repl.HelpBarPayload{Show: false}, nil
	}
	if req.Reason == repl.HelpBarReasonDebounce && ctx.Kind == jsparse.CompletionIdentifier && len(strings.TrimSpace(ctx.PartialText)) < 2 {
		return repl.HelpBarPayload{Show: false}, nil
	}

	aliases := jsparse.ExtractRequireAliases(input)
	candidates := []jsparse.CompletionCandidate{}
	if ctx.Kind != jsparse.CompletionNone {
		candidates = jsparse.ResolveCandidates(ctx, analysis.Index, root)
		if ctx.Kind == jsparse.CompletionProperty {
			if moduleName, ok := aliases[ctx.BaseExpr]; ok {
				candidates = append(candidates, jsparse.FilterCandidatesByPrefix(jsparse.NodeModuleCandidates(moduleName), ctx.PartialText)...)
			}
		}
	}
	candidates = jsparse.DedupeAndSortCandidates(candidates)

	if payload, ok := e.helpBarFromContext(ctx, candidates, aliases); ok {
		return payload, nil
	}

	return e.helpBarFromTokenFallback(token), nil
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
	e.runtimeMu.Lock()
	defer e.runtimeMu.Unlock()
	return e.runtime.Set(name, value)
}

// GetVariable gets a variable from the JavaScript runtime
func (e *Evaluator) GetVariable(name string) (interface{}, error) {
	e.runtimeMu.Lock()
	defer e.runtimeMu.Unlock()
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

	e.runtimeMu.Lock()
	_, err := e.runtime.RunString(content)
	e.runtimeMu.Unlock()
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
	e.runtimeDeclaredMu.Lock()
	e.runtimeDeclaredIDs = map[string]jsparse.CompletionCandidate{}
	e.runtimeDeclaredMu.Unlock()
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

func (e *Evaluator) observeRuntimeDeclarations(code string) {
	candidates := jsparse.ExtractTopLevelBindingCandidates(code)
	if len(candidates) == 0 {
		return
	}
	e.runtimeDeclaredMu.Lock()
	defer e.runtimeDeclaredMu.Unlock()
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate.Label) == "" {
			continue
		}
		e.runtimeDeclaredIDs[candidate.Label] = candidate
	}
}

func (e *Evaluator) runtimeIdentifierHints() []jsparse.CompletionCandidate {
	e.runtimeDeclaredMu.RLock()
	defer e.runtimeDeclaredMu.RUnlock()
	if len(e.runtimeDeclaredIDs) == 0 {
		return nil
	}
	out := make([]jsparse.CompletionCandidate, 0, len(e.runtimeDeclaredIDs))
	for _, candidate := range e.runtimeDeclaredIDs {
		out = append(out, candidate)
	}
	return out
}

func (e *Evaluator) helpBarFromContext(
	ctx jsparse.CompletionContext,
	candidates []jsparse.CompletionCandidate,
	aliases map[string]string,
) (repl.HelpBarPayload, bool) {
	switch ctx.Kind {
	case jsparse.CompletionProperty:
		base := strings.TrimSpace(ctx.BaseExpr)
		if base == "" {
			return repl.HelpBarPayload{}, false
		}

		// Base object summary when no property token is typed yet.
		if strings.TrimSpace(ctx.PartialText) == "" {
			if txt, ok := e.helpBarSignatureFor(base, "", aliases); ok {
				return makeHelpBarPayload(txt, "signature"), true
			}
		}

		exact := jsparse.FindExactCandidate(candidates, ctx.PartialText)
		if exact != nil {
			if txt, ok := e.helpBarSignatureFor(base, exact.Label, aliases); ok {
				return makeHelpBarPayload(txt, "signature"), true
			}
			return makeHelpBarPayload(fmt.Sprintf("%s.%s - %s", base, exact.Label, normalizeCandidateDetail(exact.Detail)), "info"), true
		}
		if len(candidates) > 0 {
			c := candidates[0]
			if txt, ok := e.helpBarSignatureFor(base, c.Label, aliases); ok {
				return makeHelpBarPayload(txt, "signature"), true
			}
			return makeHelpBarPayload(fmt.Sprintf("%s.%s - %s", base, c.Label, normalizeCandidateDetail(c.Detail)), "info"), true
		}
	case jsparse.CompletionIdentifier:
		exact := jsparse.FindExactCandidate(candidates, ctx.PartialText)
		if exact != nil {
			if txt, ok := e.helpBarSignatureFor(exact.Label, "", nil); ok {
				return makeHelpBarPayload(txt, "signature"), true
			}
			if txt, ok := e.runtimeHelpForIdentifier(exact.Label); ok {
				return makeHelpBarPayload(txt, "runtime"), true
			}
			return makeHelpBarPayload(fmt.Sprintf("%s - %s", exact.Label, normalizeCandidateDetail(exact.Detail)), "info"), true
		}
		if len(candidates) > 0 {
			c := candidates[0]
			if txt, ok := e.helpBarSignatureFor(c.Label, "", nil); ok {
				return makeHelpBarPayload(txt, "signature"), true
			}
			if txt, ok := e.runtimeHelpForIdentifier(c.Label); ok {
				return makeHelpBarPayload(txt, "runtime"), true
			}
			return makeHelpBarPayload(fmt.Sprintf("%s - %s", c.Label, normalizeCandidateDetail(c.Detail)), "info"), true
		}
	case jsparse.CompletionNone, jsparse.CompletionArgument:
		return repl.HelpBarPayload{}, false
	}

	return repl.HelpBarPayload{}, false
}

func (e *Evaluator) helpBarFromTokenFallback(token string) repl.HelpBarPayload {
	if token == "" {
		return repl.HelpBarPayload{Show: false}
	}
	// Keep fallback token-only: no arbitrary expression evaluation.
	token = strings.Trim(token, ".")
	if token == "" {
		return repl.HelpBarPayload{Show: false}
	}

	if txt, ok := helpBarSymbolSignatures[token]; ok {
		return makeHelpBarPayload(txt, "signature")
	}
	if txt, ok := e.runtimeHelpForIdentifier(token); ok {
		return makeHelpBarPayload(txt, "runtime")
	}
	return repl.HelpBarPayload{Show: false}
}

func (e *Evaluator) helpBarSignatureFor(base, property string, aliases map[string]string) (string, bool) {
	candidates := make([]string, 0, 4)

	if property == "" {
		if aliases != nil {
			if moduleName, ok := aliases[base]; ok {
				candidates = append(candidates, moduleName)
			}
		}
		candidates = append(candidates, base)
	} else {
		if aliases != nil {
			if moduleName, ok := aliases[base]; ok {
				candidates = append(candidates, moduleName+"."+property)
			}
		}
		candidates = append(candidates, base+"."+property)
	}

	for _, key := range candidates {
		if txt, ok := helpBarSymbolSignatures[key]; ok {
			return txt, true
		}
	}

	return "", false
}

func (e *Evaluator) runtimeHelpForIdentifier(name string) (string, bool) {
	if name == "" || strings.Contains(name, ".") {
		return "", false
	}

	e.runtimeMu.Lock()
	defer e.runtimeMu.Unlock()

	v := e.runtime.Get(name)
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return "", false
	}

	if _, ok := goja.AssertFunction(v); ok {
		obj := v.ToObject(e.runtime)
		displayName := name
		if n := obj.Get("name"); n != nil && !goja.IsUndefined(n) {
			if s := strings.TrimSpace(n.String()); s != "" {
				displayName = s
			}
		}
		if l := obj.Get("length"); l != nil && !goja.IsUndefined(l) {
			switch vv := l.Export().(type) {
			case int64:
				return fmt.Sprintf("%s(...): function (arity %d)", displayName, vv), true
			case int32:
				return fmt.Sprintf("%s(...): function (arity %d)", displayName, vv), true
			case int:
				return fmt.Sprintf("%s(...): function (arity %d)", displayName, vv), true
			case float64:
				return fmt.Sprintf("%s(...): function (arity %d)", displayName, int64(vv)), true
			}
		}
		return fmt.Sprintf("%s(...): function", displayName), true
	}

	obj := v.ToObject(e.runtime)
	className := strings.ToLower(strings.TrimSpace(obj.ClassName()))
	if className == "" {
		className = "value"
	}
	return fmt.Sprintf("%s: %s", name, className), true
}

func tokenAtCursor(input string, cursor int) (string, int, int) {
	cursor = clampCursor(cursor, len(input))
	start := cursor
	for start > 0 && isTokenByte(input[start-1]) {
		start--
	}
	end := cursor
	for end < len(input) && isTokenByte(input[end]) {
		end++
	}
	return input[start:end], start, end
}

func isTokenByte(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_' ||
		b == '$' ||
		b == '.'
}

func normalizeCandidateDetail(detail string) string {
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return "symbol"
	}
	return detail
}

func makeHelpBarPayload(text, kind string) repl.HelpBarPayload {
	return repl.HelpBarPayload{
		Show:     true,
		Text:     text,
		Kind:     kind,
		Severity: "info",
	}
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
