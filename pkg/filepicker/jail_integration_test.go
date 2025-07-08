package filepicker

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestJailIntegrationNavigationBlocking(t *testing.T) {
	// Create test directory structure
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")
	outsideDir := filepath.Join(tmpDir, "outside")
	subDir := filepath.Join(jailDir, "subdir")

	err := os.MkdirAll(jailDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(outsideDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create filepicker with jail starting in subdirectory
	fp := New(
		WithJailDirectory(jailDir),
		WithStartPath(subDir),
	)

	// Simulate backspace key press (should navigate to jail root)
	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	fp.updateNormal(backspaceMsg)

	if fp.currentPath != jailDir {
		t.Errorf("Expected to navigate to jail root %s, got %s", jailDir, fp.currentPath)
	}

	// Try backspace again (should be blocked)
	originalPath := fp.currentPath
	fp.updateNormal(backspaceMsg)

	if fp.currentPath != originalPath {
		t.Errorf("Expected navigation to be blocked at jail root, but path changed from %s to %s",
			originalPath, fp.currentPath)
	}
}

func TestJailIntegrationParentDirectoryHiding(t *testing.T) {
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")
	subDir := filepath.Join(jailDir, "subdir")

	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Test at jail root - should not show ".."
	fp := New(WithJailDirectory(jailDir))
	fp.currentPath = jailDir
	fp.loadDirectory()

	// Check that no ".." entry exists
	hasParentDir := false
	for _, file := range fp.files {
		if file.Name == ".." {
			hasParentDir = true
			break
		}
	}
	if hasParentDir {
		t.Error("Expected no '..' entry at jail root")
	}

	// Test in subdirectory - should show ".."
	fp.currentPath = subDir
	fp.loadDirectory()

	hasParentDir = false
	for _, file := range fp.files {
		if file.Name == ".." {
			hasParentDir = true
			break
		}
	}
	if !hasParentDir {
		t.Error("Expected '..' entry in subdirectory")
	}
}

func TestJailIntegrationHistoryNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")
	outsideDir := filepath.Join(tmpDir, "outside")
	subDir := filepath.Join(jailDir, "subdir")

	err := os.MkdirAll(jailDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(outsideDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	fp := New(WithJailDirectory(jailDir))

	// Navigate to subdirectory
	fp.currentPath = subDir
	fp.addToHistory(subDir)

	// Check history only contains valid paths
	for _, historyPath := range fp.history {
		if !fp.isWithinJail(historyPath) {
			t.Errorf("History contains path outside jail: %s", historyPath)
		}
	}

	// Manually try to add outside path (should be ignored)
	historyLenBefore := len(fp.history)
	fp.addToHistory(outsideDir)
	historyLenAfter := len(fp.history)

	if historyLenAfter != historyLenBefore {
		t.Error("Expected outside path to be rejected from history")
	}
}

func TestJailIntegrationEnterKeyDirectoryNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")
	subDir := filepath.Join(jailDir, "subdir")

	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	fp := New(WithJailDirectory(jailDir))
	fp.currentPath = jailDir
	fp.loadDirectory()

	// Find the subdirectory in the file list
	subDirIndex := -1
	for i, file := range fp.filteredFiles {
		if file.Name == "subdir" && file.IsDir {
			subDirIndex = i
			break
		}
	}

	if subDirIndex == -1 {
		t.Fatal("Could not find subdirectory in file list")
	}

	// Set cursor to subdirectory and simulate Enter key
	fp.cursor = subDirIndex
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	fp.updateNormal(enterMsg)

	if fp.currentPath != subDir {
		t.Errorf("Expected to navigate to subdirectory %s, got %s", subDir, fp.currentPath)
	}

	// Now try to navigate up with ".." (should work within jail)
	fp.loadDirectory()
	parentDirIndex := -1
	for i, file := range fp.filteredFiles {
		if file.Name == ".." {
			parentDirIndex = i
			break
		}
	}

	if parentDirIndex != -1 {
		fp.cursor = parentDirIndex
		fp.updateNormal(enterMsg)

		if fp.currentPath != jailDir {
			t.Errorf("Expected to navigate back to jail root %s, got %s", jailDir, fp.currentPath)
		}
	}
}

func TestJailIntegrationDirectorySelectionMode(t *testing.T) {
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")
	outsideDir := filepath.Join(tmpDir, "outside")
	subDir := filepath.Join(jailDir, "subdir")

	err := os.MkdirAll(jailDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(outsideDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	fp := New(
		WithJailDirectory(jailDir),
		WithDirectorySelection(true),
	)
	fp.currentPath = subDir
	fp.loadDirectory()

	// Try to select parent directory (should be blocked if it goes outside jail)
	// In this case, parent is jail root so it should be allowed
	if len(fp.filteredFiles) > 0 {
		// Look for ".." entry
		for i, file := range fp.filteredFiles {
			if file.Name == ".." {
				fp.cursor = i
				enterMsg := tea.KeyMsg{Type: tea.KeyEnter}

				// Should select jail directory (parent of subdirectory)
				model, cmd := fp.updateNormal(enterMsg)
				if cmd == nil {
					t.Error("Expected selection to trigger quit command")
				}

				fp = model.(*AdvancedModel)
				if len(fp.selectedFiles) == 0 || fp.selectedFiles[0] != jailDir {
					t.Errorf("Expected to select jail directory %s, got %v", jailDir, fp.selectedFiles)
				}
				break
			}
		}
	}
}

func TestJailIntegrationStatusDisplay(t *testing.T) {
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")

	err := os.MkdirAll(jailDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Test with jail
	fpWithJail := New(WithJailDirectory(jailDir))
	status := fpWithJail.buildStatusLine()

	if !containsSubstring(status, "Jailed") {
		t.Error("Expected status line to contain 'Jailed' indicator")
	}

	// Test without jail
	fpNoJail := New()
	status = fpNoJail.buildStatusLine()

	if containsSubstring(status, "Jailed") {
		t.Error("Expected status line to not contain 'Jailed' when no jail is set")
	}
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
