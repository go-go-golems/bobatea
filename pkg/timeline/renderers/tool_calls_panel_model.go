package renderers

import (
    tea "github.com/charmbracelet/bubbletea"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    "fmt"
    "strings"
)

// ToolCallsPanelModel shows a compact panel summarizing tool calls.
type ToolCallsPanelModel struct {
    calls    []any
    summary  string
    width    int
    selected bool
}

func (m *ToolCallsPanelModel) Init() tea.Cmd { return nil }
func (m *ToolCallsPanelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case timeline.EntitySelectedMsg:
        m.selected = true
    case timeline.EntityUnselectedMsg:
        m.selected = false
    case timeline.EntityPropsUpdatedMsg:
        if v.Patch != nil { m.OnProps(v.Patch) }
    }
    return m, nil
}
func (m *ToolCallsPanelModel) View() string {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if m.selected { sty = st.SelectedMessage }
    header := "[tools]"
    if len(m.calls) > 0 { header += fmt.Sprintf(" %d call(s)", len(m.calls)) }
    lines := []string{header}
    if m.summary != "" { lines = append(lines, "- "+m.summary) }
    content := strings.Join(lines, "\n")
    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(content)
}
func (m *ToolCallsPanelModel) OnProps(patch map[string]any) {
    if v, ok := patch["calls"].([]any); ok { m.calls = v }
    if v, ok := patch["summary"].(string); ok { m.summary = v }
    if v, ok := patch["selected"].(bool); ok { m.selected = v }
}
func (m *ToolCallsPanelModel) OnCompleted(_ map[string]any) {}
func (m *ToolCallsPanelModel) SetSize(w, _ int) { m.width = w }
func (m *ToolCallsPanelModel) Focus() {}
func (m *ToolCallsPanelModel) Blur()  {}

type ToolCallsPanelFactory struct{}
func (ToolCallsPanelFactory) Key() string  { return "renderer.tools.panel.v1" }
func (ToolCallsPanelFactory) Kind() string { return "tool_calls_panel" }
func (ToolCallsPanelFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &ToolCallsPanelModel{}
    m.OnProps(initialProps)
    return m
}


