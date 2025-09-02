package renderers

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    "github.com/rs/zerolog/log"
    "gopkg.in/yaml.v3"
)

// StructuredLogEventModel renders a styled log header with a YAML body (always visible).
type StructuredLogEventModel struct {
    level    string
    message  string
    yamlStr  string
    width    int
    selected bool
}

func (m *StructuredLogEventModel) Init() tea.Cmd { return nil }

func (m *StructuredLogEventModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case timeline.EntitySelectedMsg:
        m.selected = true
    case timeline.EntityUnselectedMsg:
        m.selected = false
    case timeline.EntitySetSizeMsg:
        m.width = v.Width
    case timeline.EntityPropsUpdatedMsg:
        if v.Patch != nil { m.onProps(v.Patch) }
    }
    return m, nil
}

func (m *StructuredLogEventModel) onProps(patch map[string]any) {
    if v, ok := patch["level"].(string); ok { m.level = v }
    if v, ok := patch["message"].(string); ok { m.message = v }
    // Prefer explicit yaml string, else marshal data/metadata/fields
    if v, ok := patch["yaml"].(string); ok {
        m.yamlStr = strings.TrimSpace(v)
        return
    }
    combined := map[string]any{}
    if v, ok := patch["data"]; ok && v != nil { combined["data"] = v }
    if v, ok := patch["metadata"]; ok && v != nil { combined["meta"] = v }
    if v, ok := patch["fields"]; ok && v != nil { combined["fields"] = v }
    if len(combined) > 0 {
        if b, err := yaml.Marshal(combined); err == nil {
            m.yamlStr = strings.TrimSpace(string(b))
        } else {
            log.Debug().Err(err).Str("component", "renderer").Str("kind", "structured_log_event").Msg("failed to marshal yaml")
            m.yamlStr = ""
        }
    }
}

func (m *StructuredLogEventModel) View() string {
    base := lipgloss.NewStyle().Padding(0, 1)
    level := strings.ToUpper(strings.TrimSpace(m.level))
    msg := strings.TrimSpace(m.message)

    var lvlColor lipgloss.Color
    switch strings.ToLower(level) {
    case "error", "err": lvlColor = lipgloss.Color("196")
    case "warn", "warning": lvlColor = lipgloss.Color("214")
    case "debug": lvlColor = lipgloss.Color("243")
    case "info": lvlColor = lipgloss.Color("39")
    default: lvlColor = lipgloss.Color("245")
    }

    lvl := lipgloss.NewStyle().Foreground(lvlColor).Bold(true).Render("[" + level + "]")
    msgColor := lipgloss.Color("245"); if m.selected { msgColor = lipgloss.Color("252") }
    msgStyled := lipgloss.NewStyle().Foreground(msgColor).Render(msg)

    header := strings.TrimSpace(strings.TrimSpace(lvl + " " + msgStyled))
    body := header
    if strings.TrimSpace(m.yamlStr) != "" {
        // Ensure exactly one newline between header and YAML, no leading newline
        body += "\n" + m.yamlStr
    }
    return base.Width(m.width - base.GetHorizontalPadding()).Render(body)
}

type StructuredLogEventFactory struct{}

func (StructuredLogEventFactory) Key() string  { return "renderer.structured_log_event.v1" }
func (StructuredLogEventFactory) Kind() string { return "structured_log_event" }
func (StructuredLogEventFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &StructuredLogEventModel{}
    m.onProps(initialProps)
    return m
}

