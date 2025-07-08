package filepicker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJailDirectory(t *testing.T) {
	// Create a temporary directory structure for testing
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

	// Test WithJailDirectory option
	fp := New(WithJailDirectory(jailDir))
	
	if fp.jailDirectory != jailDir {
		t.Errorf("Expected jail directory %s, got %s", jailDir, fp.jailDirectory)
	}
}

func TestIsWithinJail(t *testing.T) {
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

	tests := []struct {
		path     string
		expected bool
		name     string
	}{
		{jailDir, true, "jail directory itself"},
		{subDir, true, "subdirectory of jail"},
		{outsideDir, false, "outside jail directory"},
		{tmpDir, false, "parent of jail directory"},
		{filepath.Join(subDir, "nested"), true, "nested subdirectory"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := fp.isWithinJail(test.path)
			if result != test.expected {
				t.Errorf("isWithinJail(%s) = %v, expected %v", test.path, result, test.expected)
			}
		})
	}
}

func TestIsAtJailRoot(t *testing.T) {
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")
	subDir := filepath.Join(jailDir, "subdir")
	
	err := os.MkdirAll(jailDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	fp := New(WithJailDirectory(jailDir))

	// Test when at jail root
	fp.currentPath = jailDir
	if !fp.isAtJailRoot() {
		t.Error("Expected to be at jail root")
	}

	// Test when in subdirectory
	fp.currentPath = subDir
	if fp.isAtJailRoot() {
		t.Error("Expected not to be at jail root when in subdirectory")
	}

	// Test with no jail
	fpNoJail := New()
	if fpNoJail.isAtJailRoot() {
		t.Error("Expected not to be at jail root when no jail is set")
	}
}

func TestValidateNavigationPath(t *testing.T) {
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

	tests := []struct {
		path     string
		expected bool
		name     string
	}{
		{jailDir, true, "jail directory"},
		{subDir, true, "subdirectory of jail"},
		{outsideDir, false, "outside jail"},
		{tmpDir, false, "parent of jail"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := fp.validateNavigationPath(test.path)
			if result != test.expected {
				t.Errorf("validateNavigationPath(%s) = %v, expected %v", test.path, result, test.expected)
			}
		})
	}
}

func TestJailWithStartPathOutsideJail(t *testing.T) {
	tmpDir := t.TempDir()
	jailDir := filepath.Join(tmpDir, "jail")
	outsideDir := filepath.Join(tmpDir, "outside")
	
	err := os.MkdirAll(jailDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(outsideDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create file picker with start path outside jail
	fp := New(
		WithStartPath(outsideDir),
		WithJailDirectory(jailDir),
	)

	// Should have moved to jail directory
	if fp.currentPath != jailDir {
		t.Errorf("Expected current path to be moved to jail %s, got %s", jailDir, fp.currentPath)
	}
}

func TestNoJailRestriction(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create file picker without jail
	fp := New(WithStartPath(tmpDir))

	// All paths should be valid when no jail is set
	if !fp.isWithinJail(tmpDir) {
		t.Error("Expected path to be valid when no jail is set")
	}
	if !fp.isWithinJail(subDir) {
		t.Error("Expected subdir to be valid when no jail is set")
	}
	if !fp.validateNavigationPath(tmpDir) {
		t.Error("Expected navigation to be valid when no jail is set")
	}
	if fp.isAtJailRoot() {
		t.Error("Expected not to be at jail root when no jail is set")
	}
}
