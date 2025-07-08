package filepicker

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCompatibilityAPI(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filepicker-compat-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test NewModel() creates compatible structure
	model := NewModel()

	// Test that the compatibility fields exist and have correct types
	if model.Filepicker.DirAllowed != true {
		t.Error("Expected DirAllowed to be true by default")
	}

	if model.Filepicker.FileAllowed != true {
		t.Error("Expected FileAllowed to be true by default")
	}

	if model.Filepicker.Height != 10 {
		t.Error("Expected Height to be 10 by default")
	}

	// Test setting CurrentDirectory
	model.Filepicker.CurrentDirectory = tempDir

	// Test Update method works
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, cmd := model.Update(windowMsg)

	// Verify the model is updated correctly
	compatModel := updatedModel.(Model)
	if compatModel.Filepicker.CurrentDirectory != tempDir {
		t.Errorf("Expected CurrentDirectory to be %s, got %s",
			tempDir, compatModel.Filepicker.CurrentDirectory)
	}

	// Verify the command is returned (can be nil)
	_ = cmd // Commands can be nil, that's normal

	// Test Init method
	initCmd := model.Init()
	if initCmd == nil {
		// This is acceptable - Init might return nil
	}

	// Test View method
	view := model.View()
	if view == "" {
		t.Error("Expected View to return non-empty string")
	}
}

func TestCompatibilityMessages(t *testing.T) {
	// Test that the compatibility messages have correct structure
	selectMsg := SelectFileMsg{Path: "/test/path"}
	if selectMsg.Path != "/test/path" {
		t.Error("SelectFileMsg.Path not set correctly")
	}

	cancelMsg := CancelFilePickerMsg{}
	_ = cancelMsg // Just verify it compiles
}

func TestAdvancedModelDirectly(t *testing.T) {
	// Test that the advanced model can be used directly
	tempDir, err := os.MkdirTemp("", "filepicker-advanced-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create advanced model directly
	advModel := NewAdvancedModel(tempDir)

	if advModel.currentPath != tempDir {
		t.Errorf("Expected currentPath to be %s, got %s", tempDir, advModel.currentPath)
	}

	// Test setter methods
	advModel.SetShowPreview(false)
	if advModel.showPreview {
		t.Error("Expected showPreview to be false after SetShowPreview(false)")
	}

	advModel.SetShowHidden(true)
	if !advModel.showHidden {
		t.Error("Expected showHidden to be true after SetShowHidden(true)")
	}

	// Test GetSelected and GetError methods
	selected, hasSelection := advModel.GetSelected()
	if hasSelection {
		t.Error("Expected no selection initially")
	}
	if len(selected) != 0 {
		t.Error("Expected empty selection initially")
	}

	if err := advModel.GetError(); err != nil {
		t.Errorf("Expected no error initially, got: %v", err)
	}
}

func TestModelEmbedding(t *testing.T) {
	// Test that Model correctly embeds AdvancedModel
	model := NewModel()

	// These should work because Model embeds *AdvancedModel
	selected, hasSelection := model.GetSelected()
	if hasSelection {
		t.Error("Expected no selection initially")
	}
	if len(selected) != 0 {
		t.Error("Expected empty selection initially")
	}

	// Test that we can call advanced methods directly
	model.SetShowPreview(false)
	if model.showPreview {
		t.Error("Expected showPreview to be false")
	}
}
