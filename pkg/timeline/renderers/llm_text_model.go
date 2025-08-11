package renderers

import (
    tea "github.com/charmbracelet/bubbletea"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
    "github.com/go-go-golems/bobatea/pkg/timeline"
)

// LLMTextModel is an interactive model for rendering LLM text messages.
type LLMTextModel struct {
    role     string
    text     string
    width    int
    selected bool
    focused  bool
}

func (m *LLMTextModel) Init() tea.Cmd { return nil }

func (m *LLMTextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case timeline.EntitySelectedMsg:
        m.selected = true
        return m, nil
    case timeline.EntityUnselectedMsg:
        m.selected = false
        m.focused = false
        return m, nil
    case timeline.EntityPropsUpdatedMsg:
        if v.Patch != nil {
            m.OnProps(v.Patch)
        }
        return m, nil
    }
    return m, nil
}

func (m *LLMTextModel) View() string {
    st := chatstyle.DefaultStyles()
    role := m.role
    if role == "" { role = "assistant" }
    return chatstyle.RenderBox(st, role, m.text, m.width, m.selected, m.focused)
}

func (m *LLMTextModel) OnProps(patch map[string]any) {
    if v, ok := patch["role"].(string); ok { m.role = v }
    if v, ok := patch["text"].(string); ok { m.text = v }
    if v, ok := patch["selected"].(bool); ok { m.selected = v }
}

func (m *LLMTextModel) OnCompleted(result map[string]any) {
    if v, ok := result["text"].(string); ok { m.text = v }
}

func (m *LLMTextModel) SetSize(w, _ int) { m.width = w }
func (m *LLMTextModel) Focus()            { m.focused = true }
func (m *LLMTextModel) Blur()             { m.focused = false }

// LLMTextFactory registers the model for llm_text renderer.
type LLMTextFactory struct{}

func (LLMTextFactory) Key() string  { return "renderer.llm_text.simple.v1" }
func (LLMTextFactory) Kind() string { return "llm_text" }
func (LLMTextFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &LLMTextModel{}
    m.OnProps(initialProps)
    return m
}


