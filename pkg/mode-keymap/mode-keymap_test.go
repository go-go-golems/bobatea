package mode_keymap

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/stretchr/testify/assert"
)

// Test with single key binding and check for '*' mode
func TestForEachKeyBinding_SingleKeyBindingWithAsteriskMode(t *testing.T) {
	keymap := struct {
		SomeAction key.Binding
	}{}

	called := false
	expectedModes := NewModes([]string{"*"})

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		called = true
		assert.Equal(t, expectedModes, modes, "Modes should match the '*' mode")
	})

	assert.True(t, called, "Function should be called with the '*' mode")
}

// Test with multiple key bindings and check for correct modes
func TestForEachKeyBinding_MultipleKeyBindingsWithCorrectModes(t *testing.T) {
	keymap := struct {
		ActionOne   key.Binding `keymap-mode:"mode1"`
		ActionTwo   key.Binding `keymap-mode:"mode2"`
		ActionThree key.Binding `keymap-mode:"mode3"`
	}{}

	calls := 0
	expectedModes := []Modes{
		NewModes([]string{"mode1"}),
		NewModes([]string{"mode2"}),
		NewModes([]string{"mode3"}),
	}

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		assert.Equal(t, expectedModes[calls], modes, "Modes should match the provided modes")
		calls++
	})

	assert.Equal(t, 3, calls, "Function should be called three times with the correct modes")
}

// Test with nested structs and check for correct modes
func TestForEachKeyBinding_NestedStructsWithCorrectModes(t *testing.T) {
	type NestedStruct struct {
		NestedAction key.Binding `keymap-mode:"nested"`
	}

	keymap := struct {
		TopAction key.Binding `keymap-mode:"top"`
		Nested    NestedStruct
	}{}

	calls := 0
	expectedModes := []Modes{
		NewModes([]string{"top"}),
		NewModes([]string{"nested"}),
	}

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		assert.Equal(t, expectedModes[calls], modes, "Modes should match the provided modes")
		calls++
	})

	assert.Equal(t, 2, calls, "Function should be called twice with the correct modes")
}

// Test with pointer to struct
func TestForEachKeyBinding_PointerToStruct(t *testing.T) {
	type KeymapWithPointer struct {
		ActionOne *key.Binding `keymap-mode:"mode1"`
	}

	binding := &key.Binding{}
	keymap := &KeymapWithPointer{
		ActionOne: binding,
	}

	called := false
	expectedModes := NewModes([]string{"mode1"})

	ForEachKeyBinding(keymap, func(b *key.Binding, modes Modes) {
		called = true
		assert.Equal(t, binding, b, "Binding should match the expected pointer")
		assert.Equal(t, expectedModes, modes, "Modes should match the 'mode1'")
	})

	assert.True(t, called, "Function should be called with the pointer to struct")
}

// Test with non-key.Binding fields
func TestForEachKeyBinding_NonKeyBindingFields(t *testing.T) {
	keymap := struct {
		SomeAction key.Binding
		NotBinding string `keymap-mode:"mode1"`
	}{}

	called := false
	expectedModes := NewModes([]string{"*"})

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		called = true
		assert.Equal(t, expectedModes, modes, "Modes should match the '*' mode")
	})

	assert.True(t, called, "Function should be called with the key.Binding field only")
}

// Test with '*' mode tag
func TestForEachKeyBinding_AsteriskModeTag(t *testing.T) {
	keymap := struct {
		SomeAction key.Binding `keymap-mode:"*"`
	}{}

	called := false
	expectedModes := NewModes([]string{"*"})

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		called = true
		assert.Equal(t, expectedModes, modes, "Modes should match the '*' mode")
	})

	assert.True(t, called, "Function should be called with the '*' mode tag")
}

// Test with specific mode tags
func TestForEachKeyBinding_SpecificModeTags(t *testing.T) {
	keymap := struct {
		ActionOne key.Binding `keymap-mode:"mode1"`
		ActionTwo key.Binding `keymap-mode:"mode2"`
	}{}

	calls := 0
	expectedModes := []Modes{
		NewModes([]string{"mode1"}),
		NewModes([]string{"mode2"}),
	}

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		assert.Equal(t, expectedModes[calls], modes, "Modes should match the provided modes")
		calls++
	})

	assert.Equal(t, 2, calls, "Function should be called twice with the specific mode tags")
}

// Test with nil pointer
func TestForEachKeyBinding_NilPointer(t *testing.T) {
	var keymap *struct {
		ActionOne *key.Binding `keymap-mode:"mode1"`
	}

	called := false

	ForEachKeyBinding(keymap, func(b *key.Binding, modes Modes) {
		called = true
	})

	assert.False(t, called, "Function should not be called with a nil pointer")
}

// Test with embedded structs
func TestForEachKeyBinding_EmbeddedStructs(t *testing.T) {
	type EmbeddedStruct struct {
		EmbeddedAction key.Binding `keymap-mode:"embedded"`
	}

	keymap := struct {
		EmbeddedStruct
		TopAction key.Binding `keymap-mode:"top"`
	}{}

	calls := 0
	expectedModes := []Modes{
		NewModes([]string{"embedded"}),
		NewModes([]string{"top"}),
	}

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		assert.Equal(t, expectedModes[calls], modes, "Modes should match the provided modes")
		calls++
	})

	assert.Equal(t, 2, calls, "Function should be called twice with the correct modes")
}

// Test with embedded pointers to structs
func TestForEachKeyBinding_EmbeddedPointersToStructs(t *testing.T) {
	type EmbeddedStruct struct {
		EmbeddedAction key.Binding `keymap-mode:"embedded"`
	}

	keymap := struct {
		*EmbeddedStruct
		TopAction key.Binding `keymap-mode:"top"`
	}{
		EmbeddedStruct: &EmbeddedStruct{},
	}

	calls := 0
	expectedModes := []Modes{
		NewModes([]string{"embedded"}),
		NewModes([]string{"top"}),
	}

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		assert.Equal(t, expectedModes[calls], modes, "Modes should match the provided modes")
		calls++
	})

	assert.Equal(t, 2, calls, "Function should be called twice with the correct modes")
}

// Test with mode tag inheritance
func TestForEachKeyBinding_ModeTagInheritance(t *testing.T) {
	type NestedStruct struct {
		NestedAction key.Binding // Inherits the "*" mode tag
	}

	keymap := struct {
		TopAction key.Binding `keymap-mode:"top"`
		Nested    NestedStruct
	}{}

	calls := 0
	expectedModes := []Modes{
		NewModes([]string{"top"}),
		NewModes([]string{"*"}), // Inherited mode tag
	}

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		assert.Equal(t, expectedModes[calls], modes, "Modes should match the provided or inherited modes")
		calls++
	})

	assert.Equal(t, 2, calls, "Function should be called twice with the correct modes, including inherited modes")
}

// Test with mode tag inheritance override
func TestForEachKeyBinding_ModeTagInheritanceOverride(t *testing.T) {
	type NestedStruct struct {
		NestedAction key.Binding `keymap-mode:"override"`
		FlopAction   key.Binding // Inherits the "nested" mode tag
	}

	keymap := struct {
		TopAction key.Binding  `keymap-mode:"top"`
		Nested    NestedStruct `keymap-mode:"nested"`
	}{}

	calls := 0
	expectedModes := []Modes{
		NewModes([]string{"top"}),
		NewModes([]string{"override"}),
		NewModes([]string{"nested"}),
	}

	ForEachKeyBinding(&keymap, func(b *key.Binding, modes Modes) {
		assert.Equal(t, expectedModes[calls], modes, "Modes should match the provided or inherited modes")
		calls++
	})

	assert.Equal(t, 3, calls, "Function should be called twice with the correct modes, including inherited modes")
}
