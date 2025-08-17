package renderers

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
    "github.com/rs/zerolog/log"
)

// AgentModeModel renders agent mode analysis and optional switch info
type AgentModeModel struct {
    title    string
    from     string
    to       string
    analysis string
    width    int
    selected bool
    focused  bool
    style    *chatstyle.Style
}

func (m *AgentModeModel) Init() tea.Cmd { return nil }

func (m *AgentModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case timeline.EntitySelectedMsg:
        m.selected = true
    case timeline.EntityUnselectedMsg:
        m.selected = false
        m.focused = false
    case timeline.EntityPropsUpdatedMsg:
        if v.Patch != nil {
            m.OnProps(v.Patch)
        }
    case timeline.EntitySetSizeMsg:
        m.width = v.Width
        return m, nil
    case timeline.EntityFocusMsg:
        m.focused = true
    case timeline.EntityBlurMsg:
        m.focused = false
    }
    return m, nil
}

func (m *AgentModeModel) View() string {
    if m.style == nil {
        m.style = chatstyle.DefaultStyles()
    }
    // Use info-like color
    header := m.title
    if header == "" {
        header = "Agent Mode"
    }
    if m.from != "" || m.to != "" {
        header += " — " + strings.TrimSpace(m.from+" → "+m.to)
    }
    body := header
    if strings.TrimSpace(m.analysis) != "" {
        body += "\n\n" + m.analysis
    }

    sty := m.style.UnselectedMessage
    if m.selected {
        sty = m.style.SelectedMessage
    }
    if m.focused && !m.selected {
        sty = m.style.FocusedMessage
    }
    // slight accent for headers
    pink := lipgloss.Color("213")
    sty = sty.BorderForeground(pink)

    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(body)
}

func (m *AgentModeModel) OnProps(patch map[string]any) {
    if v, ok := patch["title"].(string); ok {
        m.title = v
    }
    if v, ok := patch["from"].(string); ok {
        m.from = v
    }
    if v, ok := patch["to"].(string); ok {
        m.to = v
    }
    if v, ok := patch["analysis"].(string); ok {
        m.analysis = v
    }
}

type AgentModeFactory struct{}

func (AgentModeFactory) Key() string  { return "renderer.agent_mode.v1" }
func (AgentModeFactory) Kind() string { return "agent_mode" }
func (AgentModeFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    log.Debug().Str("component", "renderer").Str("kind", "agent_mode").Interface("props", initialProps).Msg("NewEntityModel")
    m := &AgentModeModel{}
    m.OnProps(initialProps)
    return m
}


