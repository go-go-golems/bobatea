package filepicker

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDirectorySelection(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filepicker-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test WithDirectorySelection option
	model := New(
		WithStartPath(tempDir),
		WithDirectorySelection(true),
	)

	if !model.GetDirectorySelectionMode() {
		t.Error("Expected directory selection mode to be enabled")
	}

	// Test SetDirectorySelectionMode
	model.SetDirectorySelectionMode(false)
	if model.GetDirectorySelectionMode() {
		t.Error("Expected directory selection mode to be disabled after SetDirectorySelectionMode(false)")
	}

	model.SetDirectorySelectionMode(true)
	if !model.GetDirectorySelectionMode() {
		t.Error("Expected directory selection mode to be enabled after SetDirectorySelectionMode(true)")
	}

	// Test window size handling
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(windowMsg)
	model = updatedModel.(*AdvancedModel)

	// Test that model loads properly
	if model.currentPath != tempDir {
		t.Errorf("Expected currentPath to be %s, got %s", tempDir, model.currentPath)
	}

	// Check that there are files/directories loaded
	if len(model.files) == 0 {
		t.Error("Expected files to be loaded")
	}

	// Test View method renders without error
	view := model.View()
	if view == "" {
		t.Error("Expected View to return non-empty string")
	}

	// Test that title shows directory selection mode
	if !contains(view, "Directory Selection Mode") {
		t.Error("Expected view to show 'Directory Selection Mode' in title")
	}
}

func TestDirectorySelectionKeyBindings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filepicker-dir-keybind-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	model := New(
		WithStartPath(tempDir),
		WithDirectorySelection(false),
	)

	// Set window size
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(windowMsg)
	model = updatedModel.(*AdvancedModel)

	// Test toggle directory selection mode key (tab)
	tabKeyMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ = model.Update(tabKeyMsg)
	model = updatedModel.(*AdvancedModel)

	if !model.GetDirectorySelectionMode() {
		t.Error("Expected directory selection mode to be enabled after tab key")
	}

	// Test select current directory key (s)
	sKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updatedModel, cmd := model.Update(sKeyMsg)
	model = updatedModel.(*AdvancedModel)

	// Should quit and select current directory
	if cmd == nil {
		t.Error("Expected command to be returned when selecting current directory")
	}

	selected, hasSelection := model.GetSelected()
	if !hasSelection {
		t.Error("Expected selection after pressing 's' key")
	}
	if len(selected) != 1 || selected[0] != tempDir {
		t.Errorf("Expected selected path to be %s, got %v", tempDir, selected)
	}
}

func TestCompatibilityWithDirectorySelection(t *testing.T) {
	// Test that the compatibility wrapper works with directory selection
	tempDir, err := os.MkdirTemp("", "filepicker-compat-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test using the compatibility API with new options
	model := NewModelWithOptions(
		WithStartPath(tempDir),
		WithDirectorySelection(true),
	)

	if !model.GetDirectorySelectionMode() {
		t.Error("Expected directory selection mode to be enabled")
	}

	// Test that embedded methods work
	model.SetDirectorySelectionMode(false)
	if model.GetDirectorySelectionMode() {
		t.Error("Expected directory selection mode to be disabled")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
