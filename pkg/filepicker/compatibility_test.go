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
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

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
	_ = initCmd // Commands can be nil, that's normal

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
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

	// Create advanced model directly using deprecated function
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

func TestNewOptionsPattern(t *testing.T) {
	// Test that the new options pattern works
	tempDir, err := os.MkdirTemp("", "filepicker-options-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

	// Test with no options (should use defaults)
	model1 := New()
	if !model1.showPreview {
		t.Error("Expected showPreview to be true by default")
	}
	if model1.showHidden {
		t.Error("Expected showHidden to be false by default")
	}

	// Test with options
	model2 := New(
		WithStartPath(tempDir),
		WithShowPreview(false),
		WithShowHidden(true),
		WithShowIcons(false),
		WithShowSizes(false),
		WithDetailedView(false),
		WithSortMode(SortBySize),
		WithPreviewWidth(30),
		WithMaxHistorySize(25),
	)

	if model2.currentPath != tempDir {
		t.Errorf("Expected currentPath to be %s, got %s", tempDir, model2.currentPath)
	}
	if model2.showPreview {
		t.Error("Expected showPreview to be false")
	}
	if !model2.showHidden {
		t.Error("Expected showHidden to be true")
	}
	if model2.showIcons {
		t.Error("Expected showIcons to be false")
	}
	if model2.showSizes {
		t.Error("Expected showSizes to be false")
	}
	if model2.detailedView {
		t.Error("Expected detailedView to be false")
	}
	if model2.sortMode != SortBySize {
		t.Error("Expected sortMode to be SortBySize")
	}
	if model2.previewWidth != 30 {
		t.Errorf("Expected previewWidth to be 30, got %d", model2.previewWidth)
	}
	if model2.maxHistorySize != 25 {
		t.Errorf("Expected maxHistorySize to be 25, got %d", model2.maxHistorySize)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that the old NewAdvancedModel still works
	tempDir, err := os.MkdirTemp("", "filepicker-compat-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

	model1 := NewAdvancedModel(tempDir)
	model2 := New(WithStartPath(tempDir))

	// Both should have the same configuration
	if model1.currentPath != model2.currentPath {
		t.Error("Expected NewAdvancedModel and New to have same currentPath")
	}
	if model1.showPreview != model2.showPreview {
		t.Error("Expected NewAdvancedModel and New to have same showPreview")
	}
	if model1.showHidden != model2.showHidden {
		t.Error("Expected NewAdvancedModel and New to have same showHidden")
	}
}

func TestNewModelWithOptions(t *testing.T) {
	// Test that the new compatibility function works
	tempDir, err := os.MkdirTemp("", "filepicker-compat-options-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}()

	// Test with options
	model := NewModelWithOptions(
		WithStartPath(tempDir),
		WithShowPreview(false),
		WithShowHidden(true),
	)

	// Should have compatibility wrapper
	if model.Filepicker.CurrentDirectory != tempDir {
		t.Errorf("Expected Filepicker.CurrentDirectory to be %s, got %s", tempDir, model.Filepicker.CurrentDirectory)
	}
	if model.Filepicker.DirAllowed != true {
		t.Error("Expected DirAllowed to be true")
	}
	if model.Filepicker.FileAllowed != true {
		t.Error("Expected FileAllowed to be true")
	}

	// Should have options applied
	if model.showPreview {
		t.Error("Expected showPreview to be false")
	}
	if !model.showHidden {
		t.Error("Expected showHidden to be true")
	}
}
