package renderers

import (
    "encoding/json"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/glamour"
    "github.com/muesli/termenv"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
    "gopkg.in/yaml.v3"
    "github.com/rs/zerolog/log"
    "golang.org/x/term"
    "os"
)

// ToolCallModel renders a tool call request as YAML, syntax highlighted via glamour
type ToolCallModel struct {
    name      string
    inputYAML string
    width     int
    selected  bool
    focused   bool
    style     *chatstyle.Style
    renderer  *glamour.TermRenderer
}

func (m *ToolCallModel) Init() tea.Cmd { return nil }

func (m *ToolCallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *ToolCallModel) View() string {
    if m.style == nil {
        m.style = chatstyle.DefaultStyles()
    }
    sty := m.style.UnselectedMessage
    if m.selected {
        sty = m.style.SelectedMessage
    }
    if m.focused && !m.selected {
        sty = m.style.FocusedMessage
    }

    title := "→ " + strings.TrimSpace(m.name)
    body := title
    if m.inputYAML != "" {
        // Render YAML inside a fenced code block so glamour highlights it
        body += "\n\n```yaml\n" + m.inputYAML + "\n```"
    }

    // Use glamour if available
    rendered := body
    if m.renderer != nil {
        if out, err := m.renderer.Render(body + "\n"); err == nil {
            rendered = strings.TrimSpace(out)
        }
    }

    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(rendered)
}

func (m *ToolCallModel) OnProps(patch map[string]any) {
    if v, ok := patch["name"].(string); ok {
        m.name = v
    }
    if v, ok := patch["input"].(string); ok {
        // Convert JSON input to YAML for readability
        s := strings.TrimSpace(v)
        var anyv any
        if s != "" && (strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")) {
            if err := json.Unmarshal([]byte(s), &anyv); err == nil {
                if b, err := yaml.Marshal(anyv); err == nil {
                    m.inputYAML = strings.TrimSpace(string(b))
                } else {
                    m.inputYAML = s
                }
            } else {
                m.inputYAML = s
            }
        } else {
            m.inputYAML = s
        }
    }
}

type ToolCallFactory struct{ renderer *glamour.TermRenderer }

func (f *ToolCallFactory) Key() string  { return "renderer.tool_call.v1" }
func (f *ToolCallFactory) Kind() string { return "tool_call" }
func (f *ToolCallFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    log.Debug().Str("component", "renderer").Str("kind", f.Kind()).Interface("props", initialProps).Msg("NewEntityModel")
    m := &ToolCallModel{renderer: f.renderer}
    m.OnProps(initialProps)
    return m
}

// NewToolCallFactory creates a glamour-enabled factory similar to LLMText
func NewToolCallFactory() *ToolCallFactory {
    var style string
    if !term.IsTerminal(int(os.Stdout.Fd())) {
        style = "notty"
    } else if termenv.HasDarkBackground() {
        style = "dark"
    } else {
        style = "light"
    }
    r, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle(style),
        glamour.WithWordWrap(80),
    )
    if err != nil {
        log.Error().Err(err).Str("component", "renderer").Str("kind", "tool_call").Msg("Failed to create glamour renderer")
        r = nil
    }
    return &ToolCallFactory{renderer: r}
}


