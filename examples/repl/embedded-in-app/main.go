package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/repl"
)

// Application modes
const (
	ModeMenu = "menu"
	ModeREPL = "repl"
	ModeLog  = "log"
)

// LogEntry represents a log entry
type LogEntry struct {
	Level   string
	Message string
	Time    string
}

// SimpleEvaluator for demonstration
type SimpleEvaluator struct {
	commandCount int
}

func NewSimpleEvaluator() *SimpleEvaluator {
	return &SimpleEvaluator{commandCount: 0}
}

func (e *SimpleEvaluator) Evaluate(ctx context.Context, code string) (string, error) {
	e.commandCount++
	code = strings.TrimSpace(code)

	if code == "" {
		return "Empty command", nil
	}

	// Simulate some processing
	if strings.HasPrefix(code, "process ") {
		item := strings.TrimPrefix(code, "process ")
		return fmt.Sprintf("Processing '%s'... Done! (command #%d)", item, e.commandCount), nil
	}

	if strings.HasPrefix(code, "status") {
		return fmt.Sprintf("System status: OK (executed %d commands)", e.commandCount), nil
	}

	if strings.HasPrefix(code, "error") {
		return "", fmt.Errorf("simulated error for testing")
	}

	return fmt.Sprintf("Command '%s' executed successfully (command #%d)", code, e.commandCount), nil
}

func (e *SimpleEvaluator) GetPrompt() string        { return "app> " }
func (e *SimpleEvaluator) GetName() string          { return "App Command Processor" }
func (e *SimpleEvaluator) SupportsMultiline() bool  { return false }
func (e *SimpleEvaluator) GetFileExtension() string { return ".cmd" }

// Application model
type AppModel struct {
	// Application state
	currentMode string
	width       int
	height      int

	// Components
	repl       repl.Model
	textInput  textinput.Model
	logEntries []LogEntry

	// UI state
	selectedMenuItem int
	menuItems        []string

	// Styles
	styles AppStyles
}

type AppStyles struct {
	Header    lipgloss.Style
	Menu      lipgloss.Style
	MenuItem  lipgloss.Style
	Selected  lipgloss.Style
	StatusBar lipgloss.Style
	LogEntry  lipgloss.Style
	LogLevel  lipgloss.Style
	Border    lipgloss.Style
	Title     lipgloss.Style
}

func NewAppStyles() AppStyles {
	return AppStyles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1),

		Menu: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1),

		MenuItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 2),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 2),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Background(lipgloss.Color("235")).
			Padding(0, 1),

		LogEntry: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),

		LogLevel: lipgloss.NewStyle().
			Bold(true).
			Width(7),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			Padding(0, 1),
	}
}

func NewAppModel() AppModel {
	// Create the evaluator
	evaluator := NewSimpleEvaluator()

	// Create REPL configuration
	config := repl.Config{
		Title:                "Embedded Command Processor",
		Placeholder:          "Enter command (try 'process item', 'status', or 'error')",
		Width:                80,
		StartMultiline:       false,
		EnableExternalEditor: true,
		EnableHistory:        true,
		MaxHistorySize:       100,
	}

	// Create REPL model
	replModel := repl.NewModel(evaluator, config)
	replModel.SetTheme(repl.BuiltinThemes["dark"])

	// Add custom commands
	replModel.AddCustomCommand("app-info", func(args []string) tea.Cmd {
		return func() tea.Msg {
			return repl.EvaluationCompleteMsg{
				Input:  "/app-info",
				Output: "Multi-Mode Application v1.0.0\nModes: Menu, REPL, Log Viewer",
				Error:  nil,
			}
		}
	})

	// Create text input for menu mode
	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.Focus()

	return AppModel{
		currentMode:      ModeMenu,
		repl:             replModel,
		textInput:        ti,
		selectedMenuItem: 0,
		menuItems:        []string{"Command REPL", "View Logs", "Settings", "Quit"},
		logEntries: []LogEntry{
			{Level: "INFO", Message: "Application started", Time: "10:00:00"},
			{Level: "DEBUG", Message: "Loading configuration", Time: "10:00:01"},
			{Level: "INFO", Message: "REPL component initialized", Time: "10:00:02"},
			{Level: "WARN", Message: "Sample warning message", Time: "10:00:03"},
			{Level: "ERROR", Message: "Sample error message", Time: "10:00:04"},
			{Level: "INFO", Message: "Ready for user input", Time: "10:00:05"},
		},
		styles: NewAppStyles(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.repl.Init(),
		textinput.Blink,
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.repl.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch m.currentMode {
		case ModeMenu:
			switch msg.String() {
			case "up", "k":
				if m.selectedMenuItem > 0 {
					m.selectedMenuItem--
				}
			case "down", "j":
				if m.selectedMenuItem < len(m.menuItems)-1 {
					m.selectedMenuItem++
				}
			case "enter":
				return m.handleMenuSelection()
			case "q", "ctrl+c":
				return m, tea.Quit
			}

		case ModeREPL:
			switch msg.String() {
			case "esc":
				m.currentMode = ModeMenu
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}

		case ModeLog:
			switch msg.String() {
			case "esc":
				m.currentMode = ModeMenu
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

	case repl.EvaluationCompleteMsg:
		// Log REPL activity
		level := "INFO"
		if msg.Error != nil {
			level = "ERROR"
		}

		m.logEntries = append(m.logEntries, LogEntry{
			Level:   level,
			Message: fmt.Sprintf("REPL: %s -> %s", msg.Input, msg.Output),
			Time:    "now",
		})

		// Keep only last 50 log entries
		if len(m.logEntries) > 50 {
			m.logEntries = m.logEntries[len(m.logEntries)-50:]
		}

	case repl.QuitMsg:
		m.currentMode = ModeMenu
		return m, nil
	}

	// Update components based on current mode
	if m.currentMode == ModeREPL {
		var cmd tea.Cmd
		updatedModel, cmd := m.repl.Update(msg)
		if replModel, ok := updatedModel.(repl.Model); ok {
			m.repl = replModel
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.selectedMenuItem {
	case 0: // Command REPL
		m.currentMode = ModeREPL
		return m, nil
	case 1: // View Logs
		m.currentMode = ModeLog
		return m, nil
	case 2: // Settings
		// Add settings logic here
		return m, nil
	case 3: // Quit
		return m, tea.Quit
	}
	return m, nil
}

func (m AppModel) View() string {
	var content string

	switch m.currentMode {
	case ModeMenu:
		content = m.viewMenu()
	case ModeREPL:
		content = m.viewREPL()
	case ModeLog:
		content = m.viewLog()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewHeader(),
		content,
		m.viewStatusBar(),
	)
}

func (m AppModel) viewHeader() string {
	title := fmt.Sprintf("Multi-Mode Application - %s", strings.ToUpper(m.currentMode))
	return m.styles.Header.Render(title)
}

func (m AppModel) viewMenu() string {
	var items []string

	for i, item := range m.menuItems {
		style := m.styles.MenuItem
		if i == m.selectedMenuItem {
			style = m.styles.Selected
		}
		items = append(items, style.Render(fmt.Sprintf("%d. %s", i+1, item)))
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	help := m.styles.LogEntry.Render("Use ↑/↓ to navigate, Enter to select, q to quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.Title.Render("Main Menu"),
		"",
		m.styles.Menu.Render(menu),
		"",
		help,
	)
}

func (m AppModel) viewREPL() string {
	replView := m.repl.View()

	help := m.styles.LogEntry.Render("REPL Mode - ESC to return to menu, Ctrl+C to quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		replView,
		"",
		help,
	)
}

func (m AppModel) viewLog() string {
	var logLines []string

	// Show recent log entries
	start := len(m.logEntries) - 20
	if start < 0 {
		start = 0
	}

	for i := start; i < len(m.logEntries); i++ {
		entry := m.logEntries[i]

		levelStyle := m.styles.LogLevel
		switch entry.Level {
		case "ERROR":
			levelStyle = levelStyle.Foreground(lipgloss.Color("196"))
		case "WARN":
			levelStyle = levelStyle.Foreground(lipgloss.Color("214"))
		case "INFO":
			levelStyle = levelStyle.Foreground(lipgloss.Color("39"))
		case "DEBUG":
			levelStyle = levelStyle.Foreground(lipgloss.Color("243"))
		}

		logLine := fmt.Sprintf("%s %s %s",
			entry.Time,
			levelStyle.Render(entry.Level),
			entry.Message,
		)

		logLines = append(logLines, m.styles.LogEntry.Render(logLine))
	}

	logs := lipgloss.JoinVertical(lipgloss.Left, logLines...)

	help := m.styles.LogEntry.Render("Log Viewer - ESC to return to menu, q to quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.Title.Render("Application Logs"),
		"",
		m.styles.Border.Render(logs),
		"",
		help,
	)
}

func (m AppModel) viewStatusBar() string {
	status := fmt.Sprintf("Mode: %s | Commands: %d | Logs: %d | Press ? for help",
		m.currentMode,
		len(m.repl.GetHistory().GetAll()),
		len(m.logEntries),
	)

	return m.styles.StatusBar.Width(m.width).Render(status)
}

func main() {
	model := NewAppModel()

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
