package renderers

import (
	"os"
	"regexp"
	"strings"

	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

// MarkdownModel renders markdown with glamour.
type MarkdownModel struct {
    width     int
    selected  bool
    streaming bool
    md        string
    renderer  *glamour.TermRenderer
    // cache rendered output by width and content
    cachedRendered string
    cachedWidth    int
    cachedMD       string
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
        log.Trace().Interface("msg", msg).Msg("updating markdown model")
        if v.Patch != nil {
            m.onProps(v.Patch)
        }
    case timeline.EntityCopyTextMsg:
        // Copy raw markdown text
        log.Trace().Interface("msg", msg).Msg("copying markdown text")
        return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.md} }
    case timeline.EntityCopyCodeMsg:
        log.Trace().Interface("msg", msg).Msg("copying code")
        blocks := mdExtractAllCodeBlocks(m.md)
        if len(blocks) > 0 {
            joined := strings.Join(blocks, "\n\n")
            return m, func() tea.Msg { return timeline.CopyCodeRequestedMsg{Code: joined} }
        }
        return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.md} }
    }
    log.Trace().Interface("msg", msg).Msg("updating markdown model default case")
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
    log.Trace().Str("component", "markdown_model").Str("phase", "view").Int("md_len", len(m.md)).Msg("calling model.View")
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if m.selected {
        sty = st.SelectedMessage
    }

    // approximate content width
    frameW, _ := sty.GetFrameSize()
    contentWidth := m.width - frameW - sty.GetHorizontalPadding()
    if contentWidth < 0 { contentWidth = 0 }

    // Render markdown with cache
    var body string
    if m.cachedRendered != "" && m.cachedWidth == contentWidth && m.cachedMD == m.md {
        body = m.cachedRendered
    } else {
        if m.renderer != nil {
            start := time.Now()
            log.Trace().Str("component", "markdown_model").Str("phase", "view").Int("md_len", len(m.md)).Msg("calling glamour renderer")
            if out, err := m.renderer.Render(strings.TrimSpace(m.md) + "\n"); err == nil {
                body = strings.TrimSpace(out)
            }
            dur := time.Since(start)
            log.Trace().Str("component", "markdown_model").Int("content_width", contentWidth).Int("md_len", len(m.md)).Dur("render_dur", dur).Msg("glamour render")
        }
        if body == "" { body = m.md }
        m.cachedRendered = body
        m.cachedWidth = contentWidth
        m.cachedMD = m.md
    }

    // Add selection indicator
    if m.selected {
        body = "▶ " + body
    }
    // Box it
    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(body)
}

type MarkdownFactory struct{ renderer *glamour.TermRenderer }

func (MarkdownFactory) Key() string  { return "renderer.markdown.v1" }
func (MarkdownFactory) Kind() string { return "markdown" }
func (f MarkdownFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &MarkdownModel{renderer: f.renderer}
    m.onProps(initialProps)
    return m
}

// NewMarkdownFactory constructs a factory with a shared glamour renderer.
func NewMarkdownFactory() *MarkdownFactory {
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
        log.Error().Err(err).Str("component", "markdown_factory").Msg("failed to create glamour renderer")
        r = nil
    }
    return &MarkdownFactory{renderer: r}
}

var mdCodeBlockRe = regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\n(.*?)\n```")

func mdExtractAllCodeBlocks(s string) []string {
    matches := mdCodeBlockRe.FindAllStringSubmatch(s, -1)
    var blocks []string
    for _, m := range matches {
        if len(m) >= 2 && m[1] != "" {
            blocks = append(blocks, m[1])
        }
    }
    return blocks
}
