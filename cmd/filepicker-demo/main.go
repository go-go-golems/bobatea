package main

import (
	"flag"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/filepicker"
)

func main() {
	var (
		advanced    = flag.Bool("advanced", false, "Use the advanced file picker")
		startPath   = flag.String("path", ".", "Starting directory path")
		showPreview = flag.Bool("preview", true, "Show file preview")
		showHidden  = flag.Bool("hidden", false, "Show hidden files")
		newAPI      = flag.Bool("new-api", false, "Use the new options API")
	)
	flag.Parse()

	if *advanced {
		if *newAPI {
			runNewAPIFilePicker(*startPath, *showPreview, *showHidden)
		} else {
			runAdvancedFilePicker(*startPath, *showPreview, *showHidden)
		}
	} else {
		runBasicFilePicker()
	}
}

func runAdvancedFilePicker(startPath string, showPreview, showHidden bool) {
	// Create the advanced file picker using the new options pattern
	picker := filepicker.New(
		filepicker.WithStartPath(startPath),
		filepicker.WithShowPreview(showPreview),
		filepicker.WithShowHidden(showHidden),
	)

	// Create a program and run it
	p := tea.NewProgram(picker, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// Handle the result
	selectedFiles, hasSelection := picker.GetSelected()
	if !hasSelection {
		fmt.Println("File selection cancelled.")
		return
	}

	if len(selectedFiles) == 0 {
		fmt.Println("No files selected.")
		return
	}

	fmt.Printf("Selected files:\n")
	for _, file := range selectedFiles {
		fmt.Printf("  - %s\n", file)
	}
}

func runNewAPIFilePicker(startPath string, showPreview, showHidden bool) {
	// Create the file picker with comprehensive options to demonstrate the new API
	picker := filepicker.New(
		filepicker.WithStartPath(startPath),
		filepicker.WithShowPreview(showPreview),
		filepicker.WithShowHidden(showHidden),
		filepicker.WithShowIcons(true),
		filepicker.WithShowSizes(true),
		filepicker.WithDetailedView(true),
		filepicker.WithSortMode(filepicker.SortByName),
		filepicker.WithPreviewWidth(35),
		filepicker.WithMaxHistorySize(100),
	)

	// Create a program and run it
	p := tea.NewProgram(picker, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// Handle the result
	selectedFiles, hasSelection := picker.GetSelected()
	if !hasSelection {
		fmt.Println("File selection cancelled.")
		return
	}

	if len(selectedFiles) == 0 {
		fmt.Println("No files selected.")
		return
	}

	fmt.Printf("Selected files (using new API):\n")
	for _, file := range selectedFiles {
		fmt.Printf("  - %s\n", file)
	}
}

func runBasicFilePicker() {
	fp := filepicker.NewModel()

	// Create a simple wrapper model
	model := &wrapperModel{filepicker: fp}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type wrapperModel struct {
	filepicker filepicker.Model
	selected   bool
	cancelled  bool
}

func (m *wrapperModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m *wrapperModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case filepicker.SelectFileMsg:
		m.selected = true
		fmt.Printf("Selected file: %s\n", msg.Path)
		return m, tea.Quit
	case filepicker.CancelFilePickerMsg:
		m.cancelled = true
		fmt.Println("File selection cancelled.")
		return m, tea.Quit
	}

	var cmd tea.Cmd
	updatedModel, cmd := m.filepicker.Update(msg)
	m.filepicker = updatedModel.(filepicker.Model)
	return m, cmd
}

func (m *wrapperModel) View() string {
	if m.selected || m.cancelled {
		return ""
	}
	return m.filepicker.View()
}
