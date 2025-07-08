package filepicker

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGlobFiltering(t *testing.T) {
	// Create a test file picker
	fp := &AdvancedModel{
		files: []File{
			{Name: "..", Path: "../", IsDir: true, ModTime: time.Now()},
			{Name: "main.go", Path: "main.go", IsDir: false, Size: 100, ModTime: time.Now()},
			{Name: "test.go", Path: "test.go", IsDir: false, Size: 200, ModTime: time.Now()},
			{Name: "readme.txt", Path: "readme.txt", IsDir: false, Size: 50, ModTime: time.Now()},
			{Name: "config.yaml", Path: "config.yaml", IsDir: false, Size: 300, ModTime: time.Now()},
			{Name: "test_file.txt", Path: "test_file.txt", IsDir: false, Size: 150, ModTime: time.Now()},
		},
	}

	tests := []struct {
		name        string
		globPattern string
		searchQuery string
		expectedLen int
		shouldMatch []string
	}{
		{
			name:        "No filter",
			globPattern: "",
			searchQuery: "",
			expectedLen: 6, // All files including ".."
			shouldMatch: []string{"..", "main.go", "test.go", "readme.txt", "config.yaml", "test_file.txt"},
		},
		{
			name:        "Go files only",
			globPattern: "*.go",
			searchQuery: "",
			expectedLen: 3, // ".." + 2 go files
			shouldMatch: []string{"..", "main.go", "test.go"},
		},
		{
			name:        "Text files only",
			globPattern: "*.txt",
			searchQuery: "",
			expectedLen: 3, // ".." + 2 txt files
			shouldMatch: []string{"..", "readme.txt", "test_file.txt"},
		},
		{
			name:        "Test prefix",
			globPattern: "test*",
			searchQuery: "",
			expectedLen: 3, // ".." + 2 test files
			shouldMatch: []string{"..", "test.go", "test_file.txt"},
		},
		{
			name:        "Combined search and glob",
			globPattern: "*.go",
			searchQuery: "main",
			expectedLen: 2, // ".." + 1 matching file
			shouldMatch: []string{"..", "main.go"},
		},
		{
			name:        "Search only",
			globPattern: "",
			searchQuery: "test",
			expectedLen: 3, // ".." + 2 test files
			shouldMatch: []string{"..", "test.go", "test_file.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp.globPattern = tt.globPattern
			fp.searchQuery = tt.searchQuery
			fp.filterFiles()

			if len(fp.filteredFiles) != tt.expectedLen {
				t.Errorf("Expected %d filtered files, got %d", tt.expectedLen, len(fp.filteredFiles))
				for i, file := range fp.filteredFiles {
					t.Logf("Filtered file %d: %s", i, file.Name)
				}
			}

			// Check that expected files are present
			foundFiles := make(map[string]bool)
			for _, file := range fp.filteredFiles {
				foundFiles[file.Name] = true
			}

			for _, expected := range tt.shouldMatch {
				if !foundFiles[expected] {
					t.Errorf("Expected file '%s' not found in filtered results", expected)
				}
			}
		})
	}
}

func TestGlobPatternMatching(t *testing.T) {
	tests := []struct {
		pattern  string
		filename string
		expected bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "test.txt", false},
		{"test_*", "test_file.txt", true},
		{"test_*", "main.go", false},
		{"*.{go,txt}", "main.go", false}, // filepath.Match doesn't support braces
		{"*config*", "my_config_file.yaml", true},
		{"README*", "README.md", true},
		{"readme*", "README.md", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.filename, func(t *testing.T) {
			matched, err := filepath.Match(tt.pattern, tt.filename)
			if err != nil {
				t.Fatalf("Error matching pattern '%s' against '%s': %v", tt.pattern, tt.filename, err)
			}
			if matched != tt.expected {
				t.Errorf("Pattern '%s' against '%s': expected %v, got %v", tt.pattern, tt.filename, tt.expected, matched)
			}
		})
	}
}

func TestWithGlobPattern(t *testing.T) {
	pattern := "*.go"
	fp := New(WithGlobPattern(pattern))

	if fp.globPattern != pattern {
		t.Errorf("Expected glob pattern '%s', got '%s'", pattern, fp.globPattern)
	}
}
