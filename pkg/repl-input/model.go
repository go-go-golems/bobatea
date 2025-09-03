package replinput

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/listbox"
	"github.com/mattn/go-runewidth"
)

// SlashCommand represents a single slash command with help text.
type SlashCommand struct {
	Name string
	Help string
}

// Suggestion represents a single suggestion row.
type Suggestion struct {
	ID        string
	Primary   string // command without leading '/'
	Secondary string // help text
}

// listItem adapts Suggestion to listbox.Item for keyboard navigation.
type listItem struct{ s Suggestion }

func (li listItem) Display() string { return li.s.Primary }
func (li listItem) ID() string      { return li.s.ID }

// SelectedMsg is emitted when the user submits the input (Enter outside slash mode).
type SelectedMsg struct{ Value string }

// ErrMsg carries completion errors.
type ErrMsg error

// Completioner provides suggestions for a query string.
type Completioner func(ctx context.Context, query string) ([]Suggestion, error)

// Model is the REPL input component with slash-command autocomplete.
type Model struct {
	input          textinput.Model
	list           listbox.ListModel
	width          int
	maxSuggestions int

	// behavior
	completioner Completioner
	slashMode    bool

	// cached suggestions for custom rendering
	suggestions []Suggestion

	// styles
	sideBarStyle       lipgloss.Style
	inputBgStyle       lipgloss.Style
	underlineStyle     lipgloss.Style
	suggestionStyle    lipgloss.Style
	suggestionSelStyle lipgloss.Style

	// static commands (optional)
	staticCommands []SlashCommand

	// prompt string shown before the input
	prompt string
}

// New creates a new REPL input model.
func New(completioner Completioner, width, maxSuggestions int) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message… or /command"
	ti.Focus()
	// Avoid double prompt: we render our own prompt in View
	ti.Prompt = ""
	// Disable built-in suggestions to avoid interference with our own
	ti.ShowSuggestions = false

	li := listbox.New(maxSuggestions)
	// Match sample: no visible pointer arrow; we highlight via background
	li.SetPointer("  ")

	m := Model{
		input:          ti,
		list:           li,
		width:          width,
		maxSuggestions: maxSuggestions,
		completioner:   completioner,
		prompt:         "> ",
	}
	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	// vertical bar style; keep default colors
	m.sideBarStyle = lipgloss.NewStyle()
	// input background with subtle color difference
	bg := lipgloss.AdaptiveColor{Light: "#F7F7F7", Dark: "#101010"}
	m.inputBgStyle = lipgloss.NewStyle().Background(bg)
	// thin underline style
	m.underlineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#D0D0D0", Dark: "#2A2A2A"})
	// suggestion styles
	m.suggestionStyle = lipgloss.NewStyle()
	m.suggestionSelStyle = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#EFEFEF", Dark: "#1A1A1A"})

	// Keep textinput styles neutral (no background); we paint background externally
	m.input.PromptStyle = lipgloss.NewStyle()
	m.input.TextStyle = lipgloss.NewStyle()
	m.input.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#999999", Dark: "#666666"})
	m.input.CompletionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#777777"})
	// Align cursor styling with our background so the cell under cursor doesn't invert
	m.input.Cursor.Style = lipgloss.NewStyle().
		Background(bg).
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"})
	m.input.Cursor.TextStyle = lipgloss.NewStyle().
		Background(bg).
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"})
}

// SetCommands installs a static set of commands and uses a default completioner.
func (m *Model) SetCommands(cmds []SlashCommand) {
	m.staticCommands = cmds
	if m.completioner == nil {
		m.completioner = func(ctx context.Context, query string) ([]Suggestion, error) {
			q := strings.ToLower(strings.TrimSpace(query))
			// if empty query, show all
			var out []Suggestion
			for _, c := range m.staticCommands {
				name := strings.ToLower(c.Name)
				if q == "" || strings.Contains(name, q) {
					out = append(out, Suggestion{
						ID:        "cmd-" + c.Name,
						Primary:   c.Name,
						Secondary: c.Help,
					})
				}
			}
			return out, nil
		}
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Clamp to minimum reasonable width to avoid negative layout
		if msg.Width < 10 {
			m.width = 10
		} else {
			m.width = msg.Width
		}

		// keep textinput width in sync (inner width minus prompt)
		inner := m.getInnerWidth()
		if inner < 0 {
			inner = 0
		}
		contentW := inner - lipgloss.Width(m.prompt)
		if contentW < 0 {
			contentW = 0
		}
		m.input.Width = contentW

		// Update styles that depend on width and ensure backgrounds are consistent
		m.initStyles()

		// bubble down to list so it can compute truncation based on new width
		newList, cmd := m.list.Update(msg)
		if lm, ok := newList.(listbox.ListModel); ok {
			m.list = lm
		}

		// Return model and any command; Bubble Tea will re-render
		return m, cmd

	case listbox.SelectedMsg:
		// Accept suggestion into input and exit slash mode
		if it, ok := msg.Item.(listItem); ok {
			m.input.SetValue("/" + it.s.Primary + " ")
		}
		m.input.CursorEnd()
		m.suggestions = nil
		m.slashMode = false
		// clear list items
		return m, m.list.SetItems(nil)

	case tea.KeyMsg:
		// For navigation within suggestions, let list consume first
		var listCmd tea.Cmd
		if m.slashMode {
			newList, cmd := m.list.Update(msg)
			m.list = newList.(listbox.ListModel)
			listCmd = cmd
		}

		prev := m.input.Value()
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)

		// Determine mode and trigger completion if input changed
		val := m.input.Value()
		m.slashMode = strings.HasPrefix(val, "/")

		var cmds []tea.Cmd
		if listCmd != nil {
			cmds = append(cmds, listCmd)
		}
		if inputCmd != nil {
			cmds = append(cmds, inputCmd)
		}

		if val != prev {
			if m.slashMode {
				q := strings.TrimPrefix(val, "/")
				cmds = append(cmds, m.completionCmd(q))
			} else {
				// exit suggestions when leaving slash mode
				m.suggestions = nil
				cmds = append(cmds, m.list.SetItems(nil))
			}
		}

		// Submission outside slash mode
		if msg.Type == tea.KeyEnter && !m.slashMode {
			submitted := m.input.Value()
			// keep the input UX: clear after submit
			m.input.SetValue("")
			return m, func() tea.Msg { return SelectedMsg{Value: submitted} }
		}

		// Dismiss suggestions
		if (msg.Type == tea.KeyEsc || msg.String() == "ctrl+c") && m.slashMode {
			m.slashMode = false
			m.suggestions = nil
			cmds = append(cmds, m.list.SetItems(nil))
		}

		return m, tea.Batch(cmds...)

	case suggestionsMsg:
		m.suggestions = msg
		// adapt to list items for nav
		items := make([]listbox.Item, 0, len(msg))
		for _, s := range msg {
			items = append(items, listItem{s: s})
		}
		return m, m.list.SetItems(items)

	case ErrMsg:
		// ignore for now
		return m, nil
	}

	// default: bubble to list so it can maintain state
	newList, cmd := m.list.Update(msg)
	m.list = newList.(listbox.ListModel)
	return m, cmd
}

// View renders the framed REPL input with optional suggestions.
func (m Model) View() string {
	innerW := m.getInnerWidth()
	var b strings.Builder

	// Render suggestions (if any), each inside side bars, with two columns
	for i, s := range m.suggestions {
		left := "/" + s.Primary
		right := s.Secondary

		// compute dynamic left column width based on suggestions, but clamp
		leftW := m.suggestionLeftWidth()
		leftTrunc := runewidth.Truncate(left, leftW, "…")

		gap := "  "
		remaining := innerW - lipgloss.Width(leftTrunc) - lipgloss.Width(gap)
		if remaining < 0 {
			remaining = 0
		}
		rightTrunc := runewidth.Truncate(right, remaining, "…")

		line := fmt.Sprintf("%-*s%s%s", leftW, leftTrunc, gap, rightTrunc)
		if i == m.list.Cursor() {
			line = m.suggestionSelStyle.Width(innerW).Render(line)
		} else {
			line = m.suggestionStyle.Width(innerW).Render(line)
		}
		b.WriteString(m.frameLine(line))
		b.WriteString("\n")
	}

	// Input row with background
	prompt := m.prompt
	contentW := innerW - lipgloss.Width(prompt)
	if contentW < 0 {
		contentW = 0
	}
	// Render the textinput as-is; its styles are neutral and cursor matches bg
	inputStr := m.input.View()
	inputStr = runewidth.Truncate(inputStr, contentW, "")

	inputLine := prompt + inputStr + strings.Repeat(" ", contentW-lipgloss.Width(inputStr))
	inputLine = m.inputBgStyle.Width(innerW).Render(inputLine)
	b.WriteString(m.frameLine(inputLine))
	b.WriteString("\n")

	// Thin underline (spans inner width)
	underline := strings.Repeat("─", innerW)
	// Ensure underline occupies the full inner width without wrapping anomalies
	underline = m.underlineStyle.Width(innerW).Render(underline)
	b.WriteString(m.frameLine(underline))

	return b.String()
}

func (m Model) getInnerWidth() int {
	w := m.width
	if w <= 0 {
		w = 80
	}
	// two side bars take 2 glyphs
	return maxInt(0, w-2)
}

func (m Model) frameLine(inner string) string {
	// pad/truncate inner to exact inner width
	innerW := m.getInnerWidth()
	innerTrunc := runewidth.Truncate(inner, innerW, "")
	pad := innerW - lipgloss.Width(innerTrunc)
	if pad < 0 {
		pad = 0
	}
	return m.sideBarStyle.Render("┃") + innerTrunc + strings.Repeat(" ", pad) + m.sideBarStyle.Render("┃")
}

// suggestionsMsg carries new suggestions from completioner
type suggestionsMsg []Suggestion

// completionCmd dispatches an async fetch with small debounce and timeout
func (m Model) completionCmd(query string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(120 * time.Millisecond)
		if m.completioner == nil {
			return suggestionsMsg(nil)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
		defer cancel()
		s, err := m.completioner(ctx, query)
		if err != nil {
			return ErrMsg(err)
		}
		// clamp to maxSuggestions
		if m.maxSuggestions > 0 && len(s) > m.maxSuggestions {
			s = s[:m.maxSuggestions]
		}
		return suggestionsMsg(s)
	}
}

func (m Model) suggestionLeftWidth() int {
	// compute max width of left column among suggestions, limited to innerW/2 and min 10
	maxW := 10
	for _, s := range m.suggestions {
		w := runewidth.StringWidth("/" + s.Primary)
		if w > maxW {
			maxW = w
		}
	}
	inner := m.getInnerWidth()
	if maxW > inner/2 {
		maxW = inner / 2
	}
	if maxW < 10 {
		maxW = 10
	}
	return maxW
}

// Helpers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
