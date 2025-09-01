package renderers

import (
    "encoding/json"
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
)

// StructuredDataModel pretty-prints JSON or Go values as JSON.
type StructuredDataModel struct {
    width     int
    selected  bool
    raw       any
    rendered  string
}

func (m *StructuredDataModel) Init() tea.Cmd { return nil }

func (m *StructuredDataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case timeline.EntitySelectedMsg:
        m.selected = true
    case timeline.EntityUnselectedMsg:
        m.selected = false
    case timeline.EntitySetSizeMsg:
        m.width = v.Width
    case timeline.EntityPropsUpdatedMsg:
        if v.Patch != nil {
            m.onProps(v.Patch)
        }
    case timeline.EntityCopyTextMsg:
        return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.rendered} }
    case timeline.EntityCopyCodeMsg:
        return m, func() tea.Msg { return timeline.CopyCodeRequestedMsg{Code: m.rendered} }
    }
    return m, nil
}

func (m *StructuredDataModel) onProps(patch map[string]any) {
    if v, ok := patch["selected"].(bool); ok {
        m.selected = v
    }
    // Accept either pre-rendered JSON string as "json" or Go value as "data".
    if v, ok := patch["json"].(string); ok {
        m.raw = v
        m.rendered = prettyJSONFromString(v)
        return
    }
    if v, ok := patch["data"]; ok {
        m.raw = v
        m.rendered = prettyJSONFromAny(v)
        return
    }
}

func (m *StructuredDataModel) View() string {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if m.selected {
        sty = st.SelectedMessage
    }
    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(m.rendered)
}

type StructuredDataFactory struct{}

func (StructuredDataFactory) Key() string  { return "renderer.structured_data.v1" }
func (StructuredDataFactory) Kind() string { return "structured_data" }
func (StructuredDataFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &StructuredDataModel{}
    m.onProps(initialProps)
    return m
}

func prettyJSONFromString(s string) string {
    var v any
    if err := json.Unmarshal([]byte(s), &v); err != nil {
        // Not JSON; return as-is wrapped in a JSON string for consistency
        b, _ := json.Marshal(s)
        return string(b)
    }
    return prettyJSONFromAny(v)
}

func prettyJSONFromAny(v any) string {
    b, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return fmt.Sprintf("%v", v)
    }
    return strings.TrimSpace(string(b))
}

