package repl

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the UI state for the REPL
type Model struct {
	evaluator      Evaluator
	config         Config
	styles         Styles
	history        *History
	textInput      textinput.Model
	multilineMode  bool
	multilineText  []string
	width          int
	quitting       bool
	evaluating     bool
	customCommands map[string]func([]string) tea.Cmd
}

// NewModel creates a new REPL model
func NewModel(evaluator Evaluator, config Config) Model {
	if config.Prompt == "" {
		config.Prompt = evaluator.GetPrompt()
	}

	ti := textinput.New()
	ti.Placeholder = config.Placeholder
	ti.Focus()
	ti.Width = config.Width
	ti.Prompt = config.Prompt

	history := NewHistory(config.MaxHistorySize)

	return Model{
		evaluator:      evaluator,
		config:         config,
		styles:         DefaultStyles(),
		history:        history,
		textInput:      ti,
		multilineMode:  config.StartMultiline,
		multilineText:  []string{},
		width:          config.Width,
		quitting:       false,
		evaluating:     false,
		customCommands: make(map[string]func([]string) tea.Cmd),
	}
}

// Init initializes the REPL model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles UI events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.textInput.Width = msg.Width - 10

	case EvaluationCompleteMsg:
		m.evaluating = false
		if m.config.EnableHistory {
			m.history.Add(msg.Input, msg.Output, msg.Error != nil)
		}

	case ExternalEditorCompleteMsg:
		if msg.Error != nil {
			if m.config.EnableHistory {
				m.history.Add("/edit", fmt.Sprintf("Editor error: %v", msg.Error), true)
			}
		} else {
			lines := strings.Split(strings.TrimSpace(msg.Content), "\n")
			if len(lines) > 1 {
				m.multilineMode = true
				m.multilineText = lines
				m.textInput.Reset()
			} else {
				m.multilineMode = false
				m.multilineText = []string{}
				m.textInput.SetValue(lines[0])
			}
		}

	case ClearHistoryMsg:
		if m.config.EnableHistory {
			m.history.Clear()
		}

	case QuitMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		// Handle special key combinations
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyCtrlE:
			if m.config.EnableExternalEditor {
				return m, m.handleExternalEditor()
			}

		case tea.KeyCtrlJ:
			if m.evaluator.SupportsMultiline() {
				if !m.multilineMode {
					m.multilineMode = true
					m.multilineText = []string{m.textInput.Value()}
				} else {
					m.multilineText = append(m.multilineText, m.textInput.Value())
				}
				m.textInput.Reset()
				return m, nil
			}
		}

		// Handle regular keys
		switch msg.String() {
		case "up":
			if m.config.EnableHistory {
				if entry := m.history.NavigateUp(); entry != "" {
					m.textInput.SetValue(entry)
				}
			}
			return m, nil

		case "down":
			if m.config.EnableHistory {
				entry := m.history.NavigateDown()
				m.textInput.SetValue(entry)
			}
			return m, nil

		case "enter":
			input := m.textInput.Value()

			if m.multilineMode {
				if input == "" {
					// Execute multiline code
					fullInput := strings.Join(m.multilineText, "\n")
					m.multilineMode = false
					m.multilineText = []string{}
					m.textInput.Reset()
					return m, m.processInput(fullInput)
				} else {
					// Add line to multiline input
					m.multilineText = append(m.multilineText, input)
					m.textInput.Reset()
					return m, nil
				}
			} else {
				if input == "" {
					return m, nil
				}
				m.textInput.Reset()
				if m.config.EnableHistory {
					m.history.ResetNavigation()
				}
				return m, m.processInput(input)
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	var sb strings.Builder

	// Title
	title := m.config.Title
	if title == "" {
		title = fmt.Sprintf("%s REPL", m.evaluator.GetName())
	}
	sb.WriteString(m.styles.Title.Render(fmt.Sprintf(" %s ", title)))
	sb.WriteString("\n\n")

	// History
	if m.config.EnableHistory {
		for _, entry := range m.history.GetEntries() {
			// Input
			sb.WriteString(m.styles.Prompt.Render(m.config.Prompt))
			sb.WriteString(m.wrapText(entry.Input, m.width-len(m.config.Prompt)))
			sb.WriteString("\n")

			// Output
			if entry.IsErr {
				sb.WriteString(m.wrapText(m.styles.Error.Render(entry.Output), m.width))
			} else {
				sb.WriteString(m.wrapText(m.styles.Result.Render(entry.Output), m.width))
			}
			sb.WriteString("\n\n")
		}
	}

	// Multiline input display
	if m.multilineMode {
		sb.WriteString(m.styles.Info.Render("Multiline Mode (press Enter on empty line to execute):\n"))
		for _, line := range m.multilineText {
			sb.WriteString(m.styles.Prompt.Render("... "))
			sb.WriteString(m.wrapText(line, m.width-5))
			sb.WriteString("\n")
		}
	}

	// Input field
	sb.WriteString(m.textInput.View())
	sb.WriteString("\n\n")

	// Status/Help text
	if m.evaluating {
		sb.WriteString(m.styles.Info.Render("Evaluating..."))
	} else {
		helpText := m.buildHelpText()
		sb.WriteString(m.styles.HelpText.Render(helpText))
	}
	sb.WriteString("\n")

	if m.quitting {
		sb.WriteString("\n")
		sb.WriteString(m.styles.Info.Render("Exiting..."))
		sb.WriteString("\n")
	}

	return sb.String()
}

// SetStyles updates the styles for the REPL
func (m *Model) SetStyles(styles Styles) {
	m.styles = styles
}

// SetTheme updates the theme for the REPL
func (m *Model) SetTheme(theme Theme) {
	m.styles = theme.Styles
}

// AddCustomCommand adds a custom slash command
func (m *Model) AddCustomCommand(name string, handler func([]string) tea.Cmd) {
	m.customCommands[name] = handler
}

// SetWidth sets the width of the REPL model
func (m *Model) SetWidth(width int) {
	m.width = width
	m.textInput.Width = width - 10
}

// GetHistory returns the history object
func (m *Model) GetHistory() *History {
	return m.history
}

// processInput handles user input and returns appropriate command
func (m Model) processInput(input string) tea.Cmd {
	if strings.HasPrefix(input, "/") {
		return m.handleSlashCommand(input)
	}

	// Handle evaluation
	m.evaluating = true
	return func() tea.Msg {
		ctx := context.Background()
		output, err := m.evaluator.Evaluate(ctx, input)

		var outputStr string
		if err != nil {
			outputStr = err.Error()
		} else {
			outputStr = output
		}

		return EvaluationCompleteMsg{
			Input:  input,
			Output: outputStr,
			Error:  err,
		}
	}
}

// handleSlashCommand processes slash commands
func (m Model) handleSlashCommand(input string) tea.Cmd {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	// Check for custom commands first
	if handler, exists := m.customCommands[cmd]; exists {
		return handler(args)
	}

	// Handle built-in commands
	switch cmd {
	case "help":
		return m.showHelp()

	case "clear":
		return func() tea.Msg {
			return ClearHistoryMsg{}
		}

	case "quit", "exit":
		return func() tea.Msg {
			return QuitMsg{}
		}

	case "multiline":
		if m.evaluator.SupportsMultiline() {
			m.multilineMode = !m.multilineMode
			status := "disabled"
			if m.multilineMode {
				status = "enabled"
			}
			return func() tea.Msg {
				return EvaluationCompleteMsg{
					Input:  input,
					Output: fmt.Sprintf("Multiline mode %s", status),
					Error:  nil,
				}
			}
		}
		return func() tea.Msg {
			return EvaluationCompleteMsg{
				Input:  input,
				Output: "Multiline mode not supported by this evaluator",
				Error:  fmt.Errorf("multiline not supported"),
			}
		}

	case "edit":
		if m.config.EnableExternalEditor {
			return m.handleExternalEditor()
		}
		return func() tea.Msg {
			return EvaluationCompleteMsg{
				Input:  input,
				Output: "External editor not enabled",
				Error:  fmt.Errorf("external editor disabled"),
			}
		}

	default:
		return func() tea.Msg {
			return EvaluationCompleteMsg{
				Input:  input,
				Output: fmt.Sprintf("Unknown command: %s", cmd),
				Error:  fmt.Errorf("unknown command: %s", cmd),
			}
		}
	}
}

// handleExternalEditor opens content in external editor
func (m Model) handleExternalEditor() tea.Cmd {
	var content string
	if m.multilineMode && len(m.multilineText) > 0 {
		content = strings.Join(m.multilineText, "\n")
	} else {
		content = m.textInput.Value()
	}

	return func() tea.Msg {
		editedContent, err := m.openExternalEditor(content)
		return ExternalEditorCompleteMsg{
			Content: editedContent,
			Error:   err,
		}
	}
}

// openExternalEditor opens content in external editor
func (m Model) openExternalEditor(content string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		for _, candidate := range []string{"nano", "vim", "vi"} {
			if _, err := exec.LookPath(candidate); err == nil {
				editor = candidate
				break
			}
		}
		if editor == "" {
			return "", fmt.Errorf("no suitable editor found. Set $EDITOR environment variable")
		}
	}

	// Create temporary file
	ext := m.evaluator.GetFileExtension()
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("repl-*%s", ext))
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tmpFile.Close()

	// Launch editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read edited content
	editedFile, err := os.Open(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}
	defer editedFile.Close()

	editedBytes, err := io.ReadAll(editedFile)
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	return string(editedBytes), nil
}

// showHelp returns help message
func (m Model) showHelp() tea.Cmd {
	helpText := `Available commands:
/help      - Show this help
/clear     - Clear the screen
/quit      - Exit the REPL`

	if m.evaluator.SupportsMultiline() {
		helpText += `
/multiline - Toggle multiline mode`
	}

	if m.config.EnableExternalEditor {
		helpText += `
/edit      - Open current content in external editor (same as Ctrl+E)`
	}

	helpText += `

Keyboard shortcuts:`

	if m.evaluator.SupportsMultiline() {
		helpText += `
Ctrl+J     - Add line in multiline mode`
	}

	if m.config.EnableExternalEditor {
		helpText += `
Ctrl+E     - Open external editor`
	}

	helpText += `
Ctrl+C     - Exit REPL`

	if m.config.EnableHistory {
		helpText += `
Up/Down    - Navigate command history`
	}

	// Add custom command help
	if len(m.customCommands) > 0 {
		helpText += `

Custom commands:`
		for name := range m.customCommands {
			helpText += fmt.Sprintf(`
/%s       - Custom command`, name)
		}
	}

	return func() tea.Msg {
		return EvaluationCompleteMsg{
			Input:  "/help",
			Output: helpText,
			Error:  nil,
		}
	}
}

// buildHelpText creates the help text for the status bar
func (m Model) buildHelpText() string {
	helpText := fmt.Sprintf("Type %s code or /help for commands", m.evaluator.GetName())

	if m.multilineMode {
		helpText = "Multiline mode: Enter empty line to execute"
		if m.config.EnableExternalEditor {
			helpText += ", Ctrl+E to edit"
		}
		if m.config.EnableHistory {
			helpText += ", ↑/↓ for history"
		}
	} else {
		if m.evaluator.SupportsMultiline() {
			helpText += " (Ctrl+J for multiline"
		}
		if m.config.EnableExternalEditor {
			if m.evaluator.SupportsMultiline() {
				helpText += ", Ctrl+E to edit"
			} else {
				helpText += " (Ctrl+E to edit"
			}
		}
		if m.config.EnableHistory {
			helpText += ", ↑/↓ for history"
		}
		if m.evaluator.SupportsMultiline() || m.config.EnableExternalEditor || m.config.EnableHistory {
			helpText += ")"
		}
	}

	return helpText
}

// wrapText wraps text to fit within the given width
func (m Model) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var sb strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if len(line) <= width {
			sb.WriteString(line)
		} else {
			// Wrap the line
			currentWidth := 0
			words := strings.Fields(line)
			for j, word := range words {
				wordLen := len(word)
				if currentWidth+wordLen > width {
					// Start a new line with proper indentation
					sb.WriteString("\n    ")
					currentWidth = 4 // Account for indentation
				} else if j > 0 {
					sb.WriteString(" ")
					currentWidth++
				}
				sb.WriteString(word)
				currentWidth += wordLen
			}
		}

		// Add newline between original lines, but not after the last one
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
