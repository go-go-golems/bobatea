package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/filepicker"
)

// App states
type appState int

const (
	stateMenu appState = iota
	stateConfig
	stateResults
)

// Demo scenarios
type scenario struct {
	title       string
	description string
	id          string
}

func (s scenario) FilterValue() string { return s.title }
func (s scenario) Title() string       { return s.title }
func (s scenario) Description() string { return s.description }

var scenarios = []list.Item{
	scenario{
		title:       "Basic File Selection",
		description: "Simple file selection with default settings",
		id:          "basic",
	},
	scenario{
		title:       "Directory Selection Mode",
		description: "Select directories instead of files",
		id:          "directory",
	},
	scenario{
		title:       "Glob Pattern Demo",
		description: "Filter files using glob patterns (*.go, test_*, etc.)",
		id:          "glob",
	},
	scenario{
		title:       "Jailed Directory Demo",
		description: "Restrict navigation to a specific directory tree",
		id:          "jail",
	},
	scenario{
		title:       "Multi-Selection Demo",
		description: "Select multiple files with preview enabled",
		id:          "multi",
	},
	scenario{
		title:       "Combined Features Demo",
		description: "All features enabled: preview, detailed view, glob, etc.",
		id:          "combined",
	},
}

// Configuration for different scenarios
type config struct {
	startPath     string
	globPattern   string
	jailDirectory string
	showPreview   bool
	showHidden    bool
	directoryMode bool
	detailedView  bool
	customOptions bool
}

// Main application model
type model struct {
	state        appState
	list         list.Model
	textInput    textinput.Model
	configuring  string // which config field we're editing
	config       config
	results      []string
	hasSelection bool
	error        error
	width        int
	height       int
}

// Key bindings
var (
	quitKeys = key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	)
	backKeys = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	)
)

func initialModel() model {
	// Create list of scenarios
	l := list.New(scenarios, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Filepicker Demo Scenarios"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 0, 1, 0)

	// Create text input for configuration
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 256
	ti.Width = 50

	// Default config
	wd, _ := os.Getwd()
	cfg := config{
		startPath:     wd,
		showPreview:   true,
		showHidden:    false,
		directoryMode: false,
		detailedView:  true,
	}

	return model{
		state:     stateMenu,
		list:      l,
		textInput: ti,
		config:    cfg,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.updateMenu(msg)
		case stateConfig:
			return m.updateConfig(msg)
		case stateResults:
			return m.updateResults(msg)
		}
	}

	switch m.state {
	case stateMenu:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case stateConfig:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	case stateResults:
		// Results state is handled above
		return m, nil
	}

	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, quitKeys):
		return m, tea.Quit
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if selected := m.list.SelectedItem(); selected != nil {
			scenario := selected.(scenario)
			m = m.setupScenario(scenario.id)
			return m, nil
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		// Custom configuration
		m.state = stateConfig
		m.configuring = "startPath"
		m.textInput.SetValue(m.config.startPath)
		m.textInput.Focus()
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, quitKeys):
		return m, tea.Quit
	case key.Matches(msg, backKeys):
		m.state = stateMenu
		m.textInput.Blur()
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		value := strings.TrimSpace(m.textInput.Value())
		switch m.configuring {
		case "startPath":
			m.config.startPath = value
			m.configuring = "globPattern"
			m.textInput.SetValue(m.config.globPattern)
		case "globPattern":
			m.config.globPattern = value
			m.configuring = "jailDirectory"
			m.textInput.SetValue(m.config.jailDirectory)
		case "jailDirectory":
			m.config.jailDirectory = value
			// Done with text inputs, run custom scenario
			m.config.customOptions = true
			m = m.setupScenario("custom")
			return m, nil
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, quitKeys):
		return m, tea.Quit
	case key.Matches(msg, backKeys), key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		m.state = stateMenu
		m.results = nil
		m.hasSelection = false
		return m, nil
	}
	return m, nil
}

func (m model) setupScenario(scenarioID string) model {
	var options []filepicker.Option

	switch scenarioID {
	case "basic":
		options = []filepicker.Option{
			filepicker.WithStartPath(m.config.startPath),
		}

	case "directory":
		options = []filepicker.Option{
			filepicker.WithStartPath(m.config.startPath),
			filepicker.WithDirectorySelection(true),
			filepicker.WithShowPreview(true),
		}

	case "glob":
		options = []filepicker.Option{
			filepicker.WithStartPath(m.config.startPath),
			filepicker.WithGlobPattern("*.go"), // Default to Go files
			filepicker.WithShowPreview(true),
		}

	case "jail":
		// Create a safe jail directory in temp
		jailDir := filepath.Join(os.TempDir(), "filepicker-demo-jail")
		if err := os.MkdirAll(filepath.Join(jailDir, "subdir1"), 0755); err != nil {
			log.Printf("Error creating jail subdirectory: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(jailDir, "subdir2"), 0755); err != nil {
			log.Printf("Error creating jail subdirectory: %v", err)
		}

		// Create some demo files
		demoFile1 := filepath.Join(jailDir, "demo.txt")
		demoFile2 := filepath.Join(jailDir, "subdir1", "readme.md")
		if err := os.WriteFile(demoFile1, []byte("This is a demo file in the jail directory."), 0644); err != nil {
			log.Printf("Error creating demo file: %v", err)
		}
		if err := os.WriteFile(demoFile2, []byte("# Demo\n\nThis is a subdirectory in the jail."), 0644); err != nil {
			log.Printf("Error creating demo file: %v", err)
		}

		options = []filepicker.Option{
			filepicker.WithJailDirectory(jailDir),
			filepicker.WithStartPath(jailDir),
			filepicker.WithShowPreview(true),
		}

	case "multi":
		options = []filepicker.Option{
			filepicker.WithStartPath(m.config.startPath),
			filepicker.WithShowPreview(true),
			filepicker.WithDetailedView(true),
		}

	case "combined":
		options = []filepicker.Option{
			filepicker.WithStartPath(m.config.startPath),
			filepicker.WithShowPreview(true),
			filepicker.WithShowHidden(false),
			filepicker.WithShowIcons(true),
			filepicker.WithShowSizes(true),
			filepicker.WithDetailedView(true),
			filepicker.WithSortMode(filepicker.SortByName),
			filepicker.WithPreviewWidth(35),
		}

	case "custom":
		options = []filepicker.Option{
			filepicker.WithStartPath(m.config.startPath),
			filepicker.WithShowPreview(m.config.showPreview),
			filepicker.WithShowHidden(m.config.showHidden),
			filepicker.WithDirectorySelection(m.config.directoryMode),
			filepicker.WithDetailedView(m.config.detailedView),
		}
		if m.config.globPattern != "" {
			options = append(options, filepicker.WithGlobPattern(m.config.globPattern))
		}
		if m.config.jailDirectory != "" {
			options = append(options, filepicker.WithJailDirectory(m.config.jailDirectory))
		}
	}

	// Run the filepicker in its own program and capture results
	return m.runFilePicker(options)
}

func (m model) runFilePicker(options []filepicker.Option) model {
	// Create the filepicker
	picker := filepicker.New(options...)

	// Run it in its own program
	p := tea.NewProgram(picker, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		m.error = err
		m.state = stateResults
		return m
	}

	// Get results
	results, hasSelection := picker.GetSelected()
	m.results = results
	m.hasSelection = hasSelection
	m.state = stateResults
	return m
}

func (m model) View() string {
	switch m.state {
	case stateMenu:
		return m.menuView()
	case stateConfig:
		return m.configView()
	case stateResults:
		return m.resultsView()
	}
	return ""
}

func (m model) menuView() string {
	style := lipgloss.NewStyle().Margin(1, 2)

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Render("🗂️  Bobatea Filepicker Interactive Demo")

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render("Choose a scenario to explore the filepicker's capabilities:\n\n" +
			"• Use ↑/↓ to navigate, Enter to select\n" +
			"• Press 'c' for custom configuration\n" +
			"• Press 'q' to quit")

	return style.Render(
		header + "\n\n" +
			help + "\n\n" +
			m.list.View())
}

func (m model) configView() string {
	style := lipgloss.NewStyle().Margin(1, 2)

	var prompt string
	switch m.configuring {
	case "startPath":
		prompt = "Enter starting directory path:"
	case "globPattern":
		prompt = "Enter glob pattern (e.g., *.go, test_*, or leave empty):"
	case "jailDirectory":
		prompt = "Enter jail directory (or leave empty for no restriction):"
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Render("⚙️  Custom Configuration")

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render("Press Enter to continue, Esc to go back")

	return style.Render(
		header + "\n\n" +
			prompt + "\n\n" +
			m.textInput.View() + "\n\n" +
			help)
}

func (m model) resultsView() string {
	style := lipgloss.NewStyle().Margin(1, 2)

	var content string
	if !m.hasSelection {
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Render("❌ Selection cancelled or no files selected")
	} else if len(m.results) == 0 {
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Render("ℹ️  No files selected")
	} else {
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")).
			Bold(true).
			Render(fmt.Sprintf("✅ Selected %d item(s):", len(m.results)))

		var items []string
		for i, file := range m.results {
			if i >= 10 { // Limit display to first 10 items
				items = append(items, fmt.Sprintf("   ... and %d more", len(m.results)-10))
				break
			}
			items = append(items, fmt.Sprintf("   • %s", file))
		}

		content = header + "\n\n" + strings.Join(items, "\n")
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render("Press Enter or Esc to return to menu, q to quit")

	return style.Render(content + "\n\n" + help)
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
