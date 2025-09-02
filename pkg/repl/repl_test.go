package repl

import (
	"context"
	"testing"

	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestExampleEvaluator(t *testing.T) {
	evaluator := NewExampleEvaluator()

	// Test basic functionality
	assert.Equal(t, "Example", evaluator.GetName())
	assert.Equal(t, "example> ", evaluator.GetPrompt())
	assert.True(t, evaluator.SupportsMultiline())
	assert.Equal(t, ".txt", evaluator.GetFileExtension())

	// Test evaluation
	ctx := context.Background()

	// Test echo
	result, err := evaluator.Evaluate(ctx, "echo hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)

	// Test math
	result, err = evaluator.Evaluate(ctx, "5 + 3")
	assert.NoError(t, err)
	assert.Equal(t, "8", result)

	// Test default
	result, err = evaluator.Evaluate(ctx, "random input")
	assert.NoError(t, err)
	assert.Equal(t, "You typed: random input", result)
}

func TestHistory(t *testing.T) {
	history := NewHistory(5)

	// Test adding entries
	history.Add("input1", "output1", false)
	history.Add("input2", "output2", true)

	entries := history.GetEntries()
	assert.Len(t, entries, 2)
	assert.Equal(t, "input1", entries[0].Input)
	assert.Equal(t, "output1", entries[0].Output)
	assert.False(t, entries[0].IsErr)
	assert.Equal(t, "input2", entries[1].Input)
	assert.Equal(t, "output2", entries[1].Output)
	assert.True(t, entries[1].IsErr)

	// Test navigation
	assert.False(t, history.IsNavigating())

	up1 := history.NavigateUp()
	assert.Equal(t, "input2", up1)
	assert.True(t, history.IsNavigating())

	up2 := history.NavigateUp()
	assert.Equal(t, "input1", up2)

	down1 := history.NavigateDown()
	assert.Equal(t, "input2", down1)

	down2 := history.NavigateDown()
	assert.Equal(t, "", down2)
	assert.False(t, history.IsNavigating())

	// Test clear
	history.Clear()
	assert.Len(t, history.GetEntries(), 0)
}

func TestModel(t *testing.T) {
	evaluator := NewExampleEvaluator()
	config := DefaultConfig()
	bus, err := eventbus.NewInMemoryBus()
	assert.NoError(t, err)
	model := NewModel(evaluator, config, bus.Publisher)

	// Test initialization
	assert.Equal(t, evaluator, model.evaluator)
	assert.Equal(t, config, model.config)
	// multiline should mirror config
	assert.Equal(t, config.StartMultiline, model.multiline)
}

func TestConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "REPL", config.Title)
	assert.Equal(t, "> ", config.Prompt)
	assert.Equal(t, "Enter code or /command", config.Placeholder)
	assert.Equal(t, 80, config.Width)
	assert.False(t, config.StartMultiline)
	// Default config for external editor is disabled in timeline-based REPL
	assert.False(t, config.EnableExternalEditor)
	assert.True(t, config.EnableHistory)
	assert.Equal(t, 1000, config.MaxHistorySize)
}

func TestStyles(t *testing.T) {
	styles := DefaultStyles()

	// Just test that styles are created
	assert.NotNil(t, styles.Title)
	assert.NotNil(t, styles.Prompt)
	assert.NotNil(t, styles.Result)
	assert.NotNil(t, styles.Error)
	assert.NotNil(t, styles.Info)
	assert.NotNil(t, styles.HelpText)

	// Test themes
	assert.Contains(t, BuiltinThemes, "default")
	assert.Contains(t, BuiltinThemes, "dark")
	assert.Contains(t, BuiltinThemes, "light")
}
