package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/filepicker"
)

func main() {
	// Create test directory structure
	testDir := filepath.Join(os.TempDir(), "filepicker-test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			log.Printf("Error removing test directory: %v", err)
		}
	}()

	// Create test files and directories
	if err := os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("test file 1"), 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file2.go"), []byte("package main\n\nfunc main() {}"), 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "README.md"), []byte("# Test Directory\n\nThis is a test."), 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(testDir, "subdir1"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(testDir, "subdir2"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(testDir, "empty-dir"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "subdir1", "nested.txt"), []byte("nested file"), 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "subdir2", "another.go"), []byte("package subdir2"), 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("🗂️  Testing Directory Selection Mode\n")
	fmt.Printf("Test directory: %s\n", testDir)
	fmt.Printf("\n📁 Directory structure created:\n")
	fmt.Printf("  file1.txt\n")
	fmt.Printf("  file2.go\n")
	fmt.Printf("  README.md\n")
	fmt.Printf("  subdir1/\n")
	fmt.Printf("    nested.txt\n")
	fmt.Printf("  subdir2/\n")
	fmt.Printf("    another.go\n")
	fmt.Printf("  empty-dir/\n")

	fmt.Printf("\n🔧 Starting in Directory Selection Mode\n")
	fmt.Printf("Instructions:\n")
	fmt.Printf("  Tab       - Toggle between file and directory selection modes\n")
	fmt.Printf("  Space     - Select/deselect items (only directories in dir mode)\n")
	fmt.Printf("  Enter     - Navigate into directories (files ignored in dir mode)\n")
	fmt.Printf("  ?         - Show help (help text changes based on mode)\n")
	fmt.Printf("  s         - Select current directory\n")
	fmt.Printf("  a         - Select all visible items\n")
	fmt.Printf("  ctrl+a    - Select all files (dirs only in dir mode)\n")
	fmt.Printf("  q         - Quit\n")
	fmt.Printf("\n🎯 Test Focus:\n")
	fmt.Printf("  1. Verify Tab toggles mode and updates title\n")
	fmt.Printf("  2. Verify Space only works on directories in dir mode\n")
	fmt.Printf("  3. Verify Enter on files does nothing in dir mode\n")
	fmt.Printf("  4. Verify help text updates based on mode\n")
	fmt.Printf("  5. Verify selection behavior differences\n")
	fmt.Printf("\nPress any key to continue...")

	// Wait for keypress
	_, _ = fmt.Scanln()

	// Create the filepicker in directory selection mode
	picker := filepicker.New(
		filepicker.WithStartPath(testDir),
		filepicker.WithDirectorySelection(true),
		filepicker.WithShowPreview(true),
		filepicker.WithDetailedView(true),
		filepicker.WithShowHidden(false),
	)

	// Run the filepicker
	p := tea.NewProgram(picker, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// Show results
	results, hasSelection := picker.GetSelected()
	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("Has selection: %v\n", hasSelection)
	fmt.Printf("Directory selection mode was: %v\n", picker.IsDirectorySelectionMode())
	if hasSelection {
		fmt.Printf("Selected items (%d):\n", len(results))
		for i, item := range results {
			fmt.Printf("  %d. %s\n", i+1, item)
		}
	} else {
		fmt.Printf("No items selected or cancelled\n")
	}
}
