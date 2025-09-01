package renderers

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
)

// TextModel renders plain text with optional streaming and error styling.
type TextModel struct {
    width     int
    selected  bool
    isError   bool
    streaming bool
    text      strings.Builder
}

func (m *TextModel) Init() tea.Cmd { return nil }

func (m *TextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
        txt := m.text.String()
        return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: txt} }
    case timeline.EntityCopyCodeMsg:
        // Fallback to plain text copy
        txt := m.text.String()
        return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: txt} }
    }
    return m, nil
}

func (m *TextModel) onProps(patch map[string]any) {
    if v, ok := patch["selected"].(bool); ok {
        m.selected = v
    }
    if v, ok := patch["is_error"].(bool); ok {
        m.isError = v
    }
    if v, ok := patch["streaming"].(bool); ok {
        m.streaming = v
    }
    if v, ok := patch["text"].(string); ok {
        // replace content
        m.text.Reset()
        m.text.WriteString(v)
    }
    if v, ok := patch["append"].(string); ok {
        m.text.WriteString(v)
    }
}

func (m *TextModel) View() string {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if m.selected {
        sty = st.SelectedMessage
    }
    if m.isError {
        if m.selected {
            sty = st.ErrorSelected
        } else {
            sty = st.ErrorMessage
        }
    }
    content := m.text.String()
    if m.selected {
        content = "▶ " + content
    }
    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(content)
}

type TextFactory struct{}

func (TextFactory) Key() string  { return "renderer.text.v1" }
func (TextFactory) Kind() string { return "text" }
func (TextFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &TextModel{}
    m.onProps(initialProps)
    return m
}
