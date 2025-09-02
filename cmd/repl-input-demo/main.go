package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/repl-input"
)

type model struct {
    repl replinput.Model
    last string
    width int
    height int
}

func newModel() model {
    m := model{
        repl: replinput.New(nil, 80, 8),
    }
    m.repl.SetCommands([]replinput.SlashCommand{
        {Name: "new", Help: "new session"},
        {Name: "help", Help: "show help"},
        {Name: "share", Help: "share session"},
        {Name: "models", Help: "list models"},
        {Name: "agents", Help: "list agents"},
        {Name: "editor", Help: "open editor"},
        {Name: "redo", Help: "redo message"},
    })
    return m
}

func (m model) Init() tea.Cmd { return m.repl.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Let the repl handle the resize
        var cmd tea.Cmd
        rm, cmd := m.repl.Update(msg)
        m.repl = rm.(replinput.Model)
        return m, cmd
    case tea.KeyMsg:
        if msg.Type == tea.KeyCtrlC {
            return m, tea.Quit
        }
    case replinput.SelectedMsg:
        m.last = msg.Value
        return m, nil
    }
    var cmd tea.Cmd
    rm, cmd := m.repl.Update(msg)
    m.repl = rm.(replinput.Model)
    return m, cmd
}

func (m model) View() string {
    header := "REPL Input Demo"
    if m.width > 0 {
        header += fmt.Sprintf(" (Terminal: %dx%d)", m.width, m.height)
    }
    
    if m.last != "" {
        return fmt.Sprintf("Submitted: %s\n\n%s", m.last, m.repl.View())
    }
    return header + "\n\n" + m.repl.View()
}

func main() {
    p := tea.NewProgram(newModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}


