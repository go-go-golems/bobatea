package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/filepicker"
)

func main() {
	// Test different glob patterns
	var pattern string
	if len(os.Args) > 1 {
		pattern = os.Args[1]
	}

	// Create file picker with optional glob pattern
	var opts []filepicker.Option
	if pattern != "" {
		opts = append(opts, filepicker.WithGlobPattern(pattern))
		fmt.Printf("Starting with glob pattern: %s\n", pattern)
	}
	
	fp := filepicker.New(opts...)

	p := tea.NewProgram(fp, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	model := finalModel.(*filepicker.AdvancedModel)
	selected, ok := model.GetSelected()
	if ok {
		fmt.Printf("Selected files:\n")
		for _, file := range selected {
			fmt.Printf("  %s\n", file)
		}
	} else {
		fmt.Println("No files selected or cancelled")
	}
}
