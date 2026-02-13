package javascript

import (
	"context"
	"sync"
	"testing"

	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		evaluator, err := NewWithDefaults()
		require.NoError(t, err)
		assert.NotNil(t, evaluator)
		assert.NotNil(t, evaluator.runtime)
		assert.True(t, evaluator.config.EnableModules)
		assert.True(t, evaluator.config.EnableConsoleLog)
		assert.True(t, evaluator.config.EnableNodeModules)
	})

	t.Run("custom configuration", func(t *testing.T) {
		config := Config{
			EnableModules:     false,
			EnableConsoleLog:  false,
			EnableNodeModules: false,
			CustomModules: map[string]interface{}{
				"test": "value",
			},
		}
		evaluator, err := New(config)
		require.NoError(t, err)
		assert.NotNil(t, evaluator)
		assert.False(t, evaluator.config.EnableModules)
		assert.False(t, evaluator.config.EnableConsoleLog)
		assert.False(t, evaluator.config.EnableNodeModules)
		assert.Equal(t, "value", evaluator.config.CustomModules["test"])
	})
}

func TestEvaluator_Evaluate(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("basic arithmetic", func(t *testing.T) {
		result, err := evaluator.Evaluate(ctx, "2 + 3")
		require.NoError(t, err)
		assert.Equal(t, "5", result)
	})

	t.Run("string operations", func(t *testing.T) {
		result, err := evaluator.Evaluate(ctx, "'Hello, ' + 'World!'")
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", result)
	})

	t.Run("variable assignment and retrieval", func(t *testing.T) {
		_, err := evaluator.Evaluate(ctx, "let x = 42")
		require.NoError(t, err)

		result, err := evaluator.Evaluate(ctx, "x")
		require.NoError(t, err)
		assert.Equal(t, "42", result)
	})

	t.Run("function definition and call", func(t *testing.T) {
		_, err := evaluator.Evaluate(ctx, "function square(n) { return n * n; }")
		require.NoError(t, err)

		result, err := evaluator.Evaluate(ctx, "square(5)")
		require.NoError(t, err)
		assert.Equal(t, "25", result)
	})

	t.Run("undefined result", func(t *testing.T) {
		result, err := evaluator.Evaluate(ctx, "undefined")
		require.NoError(t, err)
		assert.Equal(t, "undefined", result)
	})

	t.Run("syntax error", func(t *testing.T) {
		_, err := evaluator.Evaluate(ctx, "let = invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "JavaScript execution failed")
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := evaluator.Evaluate(ctx, "2 + 3")
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestEvaluator_InterfaceImplementation(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	t.Run("GetPrompt", func(t *testing.T) {
		assert.Equal(t, "js>", evaluator.GetPrompt())
	})

	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "JavaScript", evaluator.GetName())
	})

	t.Run("SupportsMultiline", func(t *testing.T) {
		assert.True(t, evaluator.SupportsMultiline())
	})

	t.Run("GetFileExtension", func(t *testing.T) {
		assert.Equal(t, ".js", evaluator.GetFileExtension())
	})
}

func TestEvaluator_SetGetVariable(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	t.Run("set and get variable", func(t *testing.T) {
		err := evaluator.SetVariable("testVar", "test value")
		require.NoError(t, err)

		value, err := evaluator.GetVariable("testVar")
		require.NoError(t, err)
		assert.Equal(t, "test value", value)
	})

	t.Run("get nonexistent variable", func(t *testing.T) {
		_, err := evaluator.GetVariable("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variable nonexistent not found")
	})
}

func TestEvaluator_LoadScript(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("load valid script", func(t *testing.T) {
		script := `
			function multiply(a, b) {
				return a * b;
			}
			let result = multiply(3, 4);
		`
		err := evaluator.LoadScript(ctx, "test.js", script)
		require.NoError(t, err)

		// Verify the script was loaded
		value, err := evaluator.GetVariable("result")
		require.NoError(t, err)
		assert.Equal(t, int64(12), value)
	})

	t.Run("load invalid script", func(t *testing.T) {
		script := "invalid javascript syntax {"
		err := evaluator.LoadScript(ctx, "invalid.js", script)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load script invalid.js")
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		script := "let x = 1;"
		err := evaluator.LoadScript(ctx, "test.js", script)
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestEvaluator_Reset(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	ctx := context.Background()

	// Set a variable
	_, err = evaluator.Evaluate(ctx, "let x = 42")
	require.NoError(t, err)

	// Verify variable exists
	result, err := evaluator.Evaluate(ctx, "x")
	require.NoError(t, err)
	assert.Equal(t, "42", result)

	// Reset the evaluator
	err = evaluator.Reset()
	require.NoError(t, err)

	// Verify variable is gone
	_, err = evaluator.Evaluate(ctx, "x")
	require.Error(t, err)
}

func TestEvaluator_IsValidCode(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	t.Run("valid code", func(t *testing.T) {
		assert.True(t, evaluator.IsValidCode("2 + 3"))
		assert.True(t, evaluator.IsValidCode("function test() { return 42; }"))
		assert.True(t, evaluator.IsValidCode("let x = 'hello'"))
	})

	t.Run("invalid code", func(t *testing.T) {
		assert.False(t, evaluator.IsValidCode("let = invalid"))
		assert.False(t, evaluator.IsValidCode("function { invalid }"))
		assert.False(t, evaluator.IsValidCode("invalid syntax {"))
	})
}

func TestEvaluator_GetAvailableModules(t *testing.T) {
	t.Run("with modules enabled", func(t *testing.T) {
		config := Config{
			EnableModules:     true,
			EnableConsoleLog:  true,
			EnableNodeModules: true,
			CustomModules: map[string]interface{}{
				"custom1": "value1",
				"custom2": "value2",
			},
		}
		evaluator, err := New(config)
		require.NoError(t, err)

		modules := evaluator.GetAvailableModules()
		assert.Contains(t, modules, "custom1")
		assert.Contains(t, modules, "custom2")
		assert.Contains(t, modules, "database")
		assert.Contains(t, modules, "http")
	})

	t.Run("with modules disabled", func(t *testing.T) {
		config := Config{
			EnableModules:     false,
			EnableConsoleLog:  true,
			EnableNodeModules: false,
			CustomModules: map[string]interface{}{
				"custom1": "value1",
			},
		}
		evaluator, err := New(config)
		require.NoError(t, err)

		modules := evaluator.GetAvailableModules()
		assert.Contains(t, modules, "custom1")
		assert.NotContains(t, modules, "database")
		assert.NotContains(t, modules, "http")
	})
}

func TestEvaluator_GetHelpText(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	helpText := evaluator.GetHelpText()
	assert.Contains(t, helpText, "JavaScript REPL")
	assert.Contains(t, helpText, "Goja")
	assert.Contains(t, helpText, "console.log")
	assert.Contains(t, helpText, "Module system")
	assert.Contains(t, helpText, "Examples:")
}

func TestEvaluator_UpdateConfig(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	newConfig := Config{
		EnableModules:     false,
		EnableConsoleLog:  false,
		EnableNodeModules: false,
		CustomModules: map[string]interface{}{
			"newModule": "newValue",
		},
	}

	err = evaluator.UpdateConfig(newConfig)
	require.NoError(t, err)

	assert.Equal(t, newConfig, evaluator.GetConfig())
}

func TestEvaluator_Console(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	ctx := context.Background()

	// Test console.log (should not error, but we can't easily capture output in tests)
	_, err = evaluator.Evaluate(ctx, "console.log('Hello, World!')")
	require.NoError(t, err)

	// Test console.error
	_, err = evaluator.Evaluate(ctx, "console.error('Error message')")
	require.NoError(t, err)

	// Test console.warn
	_, err = evaluator.Evaluate(ctx, "console.warn('Warning message')")
	require.NoError(t, err)
}

func TestEvaluator_CustomModules(t *testing.T) {
	config := Config{
		EnableModules:     true,
		EnableConsoleLog:  true,
		EnableNodeModules: true,
		CustomModules: map[string]interface{}{
			"math": map[string]interface{}{
				"pi":  3.14159,
				"add": func(a, b float64) float64 { return a + b },
			},
		},
	}

	evaluator, err := New(config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("access custom module property", func(t *testing.T) {
		result, err := evaluator.Evaluate(ctx, "math.pi")
		require.NoError(t, err)
		assert.Equal(t, "3.14159", result)
	})

	t.Run("call custom module function", func(t *testing.T) {
		result, err := evaluator.Evaluate(ctx, "math.add(10, 5)")
		require.NoError(t, err)
		assert.Equal(t, "15", result)
	})
}

func TestConfig_DefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.True(t, config.EnableModules)
	assert.True(t, config.EnableConsoleLog)
	assert.True(t, config.EnableNodeModules)
	assert.NotNil(t, config.CustomModules)
	assert.Len(t, config.CustomModules, 0)
}

func TestEvaluator_CompleteInput(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("property access completion on obj dot", func(t *testing.T) {
		result, err := evaluator.CompleteInput(ctx, repl.CompletionRequest{
			Input:      "console.lo",
			CursorByte: len("console.lo"),
			Reason:     repl.CompletionReasonShortcut,
		})
		require.NoError(t, err)
		assert.True(t, result.Show)
		assert.True(t, hasSuggestion(result, "log"))
		assert.Equal(t, len("console."), result.ReplaceFrom)
		assert.Equal(t, len("console.lo"), result.ReplaceTo)
	})

	t.Run("identifier completion for partial symbol", func(t *testing.T) {
		result, err := evaluator.CompleteInput(ctx, repl.CompletionRequest{
			Input:      "cons",
			CursorByte: len("cons"),
			Reason:     repl.CompletionReasonDebounce,
		})
		require.NoError(t, err)
		assert.True(t, result.Show)
		assert.True(t, hasSuggestion(result, "console"))
		assert.Equal(t, 0, result.ReplaceFrom)
		assert.Equal(t, len("cons"), result.ReplaceTo)
	})

	t.Run("module binding completion from require declaration", func(t *testing.T) {
		input := "const fs = require(\"fs\");\nfs.re"
		result, err := evaluator.CompleteInput(ctx, repl.CompletionRequest{
			Input:      input,
			CursorByte: len(input),
			Reason:     repl.CompletionReasonShortcut,
		})
		require.NoError(t, err)
		assert.True(t, result.Show)
		assert.True(t, hasSuggestion(result, "readFile"))
	})

	t.Run("incomplete input after dot still yields candidates", func(t *testing.T) {
		input := "console."
		result, err := evaluator.CompleteInput(ctx, repl.CompletionRequest{
			Input:      input,
			CursorByte: len(input),
			Reason:     repl.CompletionReasonShortcut,
		})
		require.NoError(t, err)
		assert.True(t, result.Show)
		assert.True(t, hasSuggestion(result, "log"))
		assert.Equal(t, len(input), result.ReplaceFrom)
		assert.Equal(t, len(input), result.ReplaceTo)
	})
}

func TestEvaluator_CompleteInput_ConcurrentRequests(t *testing.T) {
	evaluator, err := NewWithDefaults()
	require.NoError(t, err)

	reqs := []repl.CompletionRequest{
		{Input: ".co", CursorByte: len(".co"), Reason: repl.CompletionReasonDebounce},
		{Input: "console.lo", CursorByte: len("console.lo"), Reason: repl.CompletionReasonShortcut, Shortcut: "tab"},
		{Input: "const fs = require(\"fs\"); fs.re", CursorByte: len("const fs = require(\"fs\"); fs.re"), Reason: repl.CompletionReasonDebounce},
		{Input: "zzz", CursorByte: len("zzz"), Reason: repl.CompletionReasonShortcut, Shortcut: "tab"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		for _, req := range reqs {
			wg.Add(1)
			go func(request repl.CompletionRequest) {
				defer wg.Done()
				_, completeErr := evaluator.CompleteInput(context.Background(), request)
				assert.NoError(t, completeErr)
			}(req)
		}
	}
	wg.Wait()
}

func hasSuggestion(result repl.CompletionResult, label string) bool {
	for _, suggestion := range result.Suggestions {
		if suggestion.Value == label {
			return true
		}
	}
	return false
}
