package renderers

import (
    "os"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/glamour"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
    "github.com/muesli/termenv"
    "golang.org/x/term"
)

// MarkdownModel renders markdown with glamour.
type MarkdownModel struct {
    width     int
    selected  bool
    streaming bool
    md        string
    renderer  *glamour.TermRenderer
}

func (m *MarkdownModel) Init() tea.Cmd { return nil }

func (m *MarkdownModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
        // Copy raw markdown text
        return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.md} }
    case timeline.EntityCopyCodeMsg:
        blocks := extractAllCodeBlocks(m.md)
        if len(blocks) > 0 {
            joined := strings.Join(blocks, "\n\n")
            return m, func() tea.Msg { return timeline.CopyCodeRequestedMsg{Code: joined} }
        }
        return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.md} }
    }
    return m, nil
}

func (m *MarkdownModel) onProps(patch map[string]any) {
    if v, ok := patch["selected"].(bool); ok {
        m.selected = v
    }
    if v, ok := patch["streaming"].(bool); ok {
        m.streaming = v
    }
    // accept either "markdown" or "text"
    if v, ok := patch["markdown"].(string); ok {
        m.md = v
    } else if v, ok := patch["text"].(string); ok {
        m.md = v
    }
}

func (m *MarkdownModel) View() string {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if m.selected {
        sty = st.SelectedMessage
    }

    if m.renderer == nil {
        m.renderer = newGlamourRenderer()
    }
    var body string
    if m.renderer != nil {
        if out, err := m.renderer.Render(strings.TrimSpace(m.md) + "\n"); err == nil {
            body = strings.TrimSpace(out)
        }
    }
    if body == "" {
        body = m.md
    }

    // Box it
    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(body)
}

type MarkdownFactory struct{}

func (MarkdownFactory) Key() string  { return "renderer.markdown.v1" }
func (MarkdownFactory) Kind() string { return "markdown" }
func (MarkdownFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &MarkdownModel{}
    m.onProps(initialProps)
    return m
}

func newGlamourRenderer() *glamour.TermRenderer {
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
        return nil
    }
    return r
}
