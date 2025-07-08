package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/bobatea/pkg/commandpalette"
)

// Styles for the chat demo
var (
	chatStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Margin(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	systemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)
)

type model struct {
	messages       []string
	input          string
	width          int
	height         int
	commandPalette commandpalette.Model
	themeIndex     int
	themes         []string
}

func initialModel() model {
	cp := commandpalette.New()

	m := model{
		messages: []string{
			"Welcome to the Chat REPL with Command Palette!",
			"Type your messages and press Enter to send.",
			"Press Ctrl+P to open the command palette.",
			"Try commands: help, clear, echo, time, date, about, theme, quit",
		},
		input:          "",
		commandPalette: cp,
		themeIndex:     0,
		themes:         []string{"default", "dark", "light", "colorful"},
	}

	// Register commands
	m.commandPalette.RegisterCommand("help", "Show help information", func() tea.Cmd {
		return func() tea.Msg {
			return commandpalette.ExecutedMsg{Command: "help"}
		}
	})

	m.commandPalette.RegisterCommand("clear", "Clear chat messages", func() tea.Cmd {
		return func() tea.Msg {
			return commandpalette.ExecutedMsg{Command: "clear"}
		}
	})

	m.commandPalette.RegisterCommand("quit", "Exit the application", func() tea.Cmd {
		return tea.Quit
	})

	m.commandPalette.RegisterCommand("echo", "Echo a test message", func() tea.Cmd {
		return func() tea.Msg {
			return commandpalette.ExecutedMsg{Command: "echo", Data: "Hello from command palette!"}
		}
	})

	m.commandPalette.RegisterCommand("time", "Show current time", func() tea.Cmd {
		return func() tea.Msg {
			return commandpalette.ExecutedMsg{Command: "time", Data: time.Now().Format("15:04:05")}
		}
	})

	m.commandPalette.RegisterCommand("date", "Show current date", func() tea.Cmd {
		return func() tea.Msg {
			return commandpalette.ExecutedMsg{Command: "date", Data: time.Now().Format("2006-01-02")}
		}
	})

	m.commandPalette.RegisterCommand("about", "Show application information", func() tea.Cmd {
		return func() tea.Msg {
			return commandpalette.ExecutedMsg{Command: "about"}
		}
	})

	m.commandPalette.RegisterCommand("theme", "Change application theme", func() tea.Cmd {
		return func() tea.Msg {
			return commandpalette.ExecutedMsg{Command: "theme"}
		}
	})

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.commandPalette.SetSize(msg.Width, msg.Height)
		return m, nil

	case commandpalette.ExecutedMsg:
		// Handle command execution
		switch msg.Command {
		case "help":
			m.messages = append(m.messages, "System: Available commands:")
			m.messages = append(m.messages, "  • help - Show this help")
			m.messages = append(m.messages, "  • clear - Clear chat messages")
			m.messages = append(m.messages, "  • echo - Echo a test message")
			m.messages = append(m.messages, "  • time - Show current time")
			m.messages = append(m.messages, "  • date - Show current date")
			m.messages = append(m.messages, "  • about - Show application info")
			m.messages = append(m.messages, "  • theme - Change theme")
			m.messages = append(m.messages, "  • quit - Exit application")
			m.messages = append(m.messages, "Press Ctrl+P to open command palette")

		case "clear":
			m.messages = []string{"System: Chat cleared"}

		case "echo":
			if data, ok := msg.Data.(string); ok {
				m.messages = append(m.messages, fmt.Sprintf("System: %s", data))
			}

		case "time":
			if data, ok := msg.Data.(string); ok {
				m.messages = append(m.messages, fmt.Sprintf("System: Current time is %s", data))
			}

		case "date":
			if data, ok := msg.Data.(string); ok {
				m.messages = append(m.messages, fmt.Sprintf("System: Current date is %s", data))
			}

		case "about":
			m.messages = append(m.messages, "System: VSCode-style Command Palette Demo")
			m.messages = append(m.messages, "Built with Charm Bracelet's Bubbletea framework")
			m.messages = append(m.messages, "Features: Fuzzy search, overlay UI, command registration")
			m.messages = append(m.messages, "Author: Reusable Bobatea Component")

		case "theme":
			m.themeIndex = (m.themeIndex + 1) % len(m.themes)
			currentTheme := m.themes[m.themeIndex]
			m.messages = append(m.messages, fmt.Sprintf("System: Theme changed to '%s'", currentTheme))
		}
		return m, nil

	case tea.KeyMsg:
		// If command palette is visible, let it handle the input first
		if m.commandPalette.IsVisible() {
			m.commandPalette, cmd = m.commandPalette.Update(msg)
			return m, cmd
		}

		// Handle main application keys
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+p":
			m.commandPalette.Show()
			return m, nil

		case "enter":
			if strings.TrimSpace(m.input) != "" {
				m.messages = append(m.messages, fmt.Sprintf("You: %s", m.input))

				// Simple command processing for direct input
				switch strings.ToLower(strings.TrimSpace(m.input)) {
				case "/help":
					m.messages = append(m.messages, "System: Type messages or use Ctrl+P for command palette")
				case "/clear":
					m.messages = []string{"System: Chat cleared via direct command"}
				case "/quit":
					return m, tea.Quit
				}

				m.input = ""
			}
			return m, nil

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil

		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	// Ensure we have valid dimensions
	if m.width <= 0 || m.height <= 0 {
		return "Loading..."
	}

	// Calculate available space
	chatHeight := m.height - 8 // Reserve space for input and borders
	if chatHeight < 1 {
		chatHeight = 1
	}

	// Render messages
	var messageLines []string
	for _, msg := range m.messages {
		if strings.HasPrefix(msg, "You: ") {
			messageLines = append(messageLines, userStyle.Render(msg))
		} else if strings.HasPrefix(msg, "System: ") {
			messageLines = append(messageLines, systemStyle.Render(msg))
		} else {
			messageLines = append(messageLines, messageStyle.Render(msg))
		}
	}

	// Keep only the last messages that fit in the chat area
	if len(messageLines) > chatHeight {
		messageLines = messageLines[len(messageLines)-chatHeight:]
	}

	chatContent := strings.Join(messageLines, "\n")

	// Ensure width is valid for styling
	chatWidth := m.width - 4
	if chatWidth < 1 {
		chatWidth = 1
	}

	chat := chatStyle.Width(chatWidth).Height(chatHeight).Render(chatContent)

	// Render input
	inputPrompt := "> " + m.input
	input := inputStyle.Width(chatWidth).Render(inputPrompt)

	// Help text
	help := helpStyle.Render("Press Ctrl+P for command palette • Ctrl+C or 'q' to quit • Try /help, /clear, /quit")

	baseView := lipgloss.JoinVertical(lipgloss.Left, chat, input, help)

	// If command palette is visible, overlay it
	if m.commandPalette.IsVisible() {
		paletteView := m.commandPalette.View()

		// Create a simple overlay that centers the palette over the base view
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			paletteView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}),
		)
	}

	return baseView
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
