package renderers

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
	geppetto_events "github.com/go-go-golems/geppetto/pkg/events"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

// LLMTextModel is an interactive model for rendering LLM text messages.
type LLMTextModel struct {
	role     string
	text     string
	width    int
	selected bool
	focused  bool
	renderer *glamour.TermRenderer
	style    *chatstyle.Style
	metadata any // prefer *events.LLMInferenceData
	streaming bool
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
	case timeline.EntitySetSizeMsg:
		m.width = v.Width
		return m, nil
	case timeline.EntityFocusMsg:
		m.focused = true
		return m, nil
	case timeline.EntityBlurMsg:
		m.focused = false
		return m, nil
	case timeline.EntityCopyTextMsg:
		return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.text} }
	case timeline.EntityCopyCodeMsg:
		code := extractFirstCodeBlock(m.text)
		if code != "" {
			return m, func() tea.Msg { return timeline.CopyCodeRequestedMsg{Code: code} }
		}
		// Fallback to copying text when no code block present
		return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.text} }
	}
	return m, nil
}

func (m *LLMTextModel) View() string {
	if m.style == nil {
		m.style = chatstyle.DefaultStyles()
	}
	if m.role == "" {
		m.role = "assistant"
	}

	// Choose base style (selected/focused/error)
	sty := m.style.UnselectedMessage
	if m.selected {
		sty = m.style.SelectedMessage
	}
	if m.focused && !m.selected {
		sty = m.style.FocusedMessage
	}
	if looksLikeError(m.text) {
		if m.selected {
			sty = m.style.ErrorSelected
		} else {
			sty = m.style.ErrorMessage
		}
	}

	// Content width accounting for border and padding
	frameW, _ := sty.GetFrameSize()
	contentWidth := m.width - frameW - sty.GetHorizontalPadding()
	if contentWidth < 0 {
		contentWidth = 0
	}

	// Render markdown with glamour
	var body string
	if m.renderer != nil {
		if out, err := m.renderer.Render(m.text + "\n"); err == nil {
			body = strings.TrimSpace(out)
		}
	}
	if body == "" {
		body = m.text
	}

	// Build status/metadata line with left spinner and right metadata
	var statusLine string
	{
		left := ""
		if m.streaming {
			// Animate simple dot spinner: Generating., Generating.., Generating...
			phase := int((time.Now().UnixNano() / int64(time.Millisecond*300)) % 3)
			dots := strings.Repeat(".", phase+1)
			left = "Generating" + dots
		}
		rightRaw := formatMetadata(m.metadata)
		right := m.style.MetadataStyle.Render(rightRaw)

		if left != "" || rightRaw != "" {
			l := lipgloss.Width(left)
			r := lipgloss.Width(right)
			pad := contentWidth - l - r
			if pad < 1 {
				pad = 1
			}
			statusLine = left + strings.Repeat(" ", pad) + right
			statusLine = m.style.MetadataStyle.Width(contentWidth).Render(statusLine)
		}
	}

	// Combine body and status line
	if statusLine != "" {
		if body != "" {
			body += "\n\n"
		}
		body += statusLine
	}

	// Box the content
	boxed := sty.Width(m.width - sty.GetHorizontalPadding()).Render(body)
	return boxed
}

func (m *LLMTextModel) OnProps(patch map[string]any) {
	if v, ok := patch["role"].(string); ok {
		m.role = v
	}
	if v, ok := patch["text"].(string); ok {
		m.text = v
	}
	if v, ok := patch["selected"].(bool); ok {
		m.selected = v
	}
	if v, ok := patch["metadata"]; ok {
		m.metadata = v
	}
	if v, ok := patch["streaming"].(bool); ok {
		m.streaming = v
	}
}

// Removed OnCompleted/SetSize/Focus/Blur; handled via messages

// LLMTextFactory registers the model for llm_text renderer.
type LLMTextFactory struct {
	renderer *glamour.TermRenderer
}

func (f *LLMTextFactory) Key() string  { return "renderer.llm_text.simple.v1" }
func (f *LLMTextFactory) Kind() string { return "llm_text" }
func (f *LLMTextFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	m := &LLMTextModel{
		renderer: f.renderer,
	}
	m.OnProps(initialProps)
	return m
}

// NewLLMTextFactory creates a new LLMTextFactory with a shared glamour renderer
func NewLLMTextFactory() *LLMTextFactory {
	// Determine glamour style once at startup
	var determinedStyle string
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		determinedStyle = "notty"
	} else if termenv.HasDarkBackground() {
		determinedStyle = "dark"
	} else {
		determinedStyle = "light"
	}

	// Create renderer with a reasonable default width (will wrap anyway)
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(determinedStyle),
		glamour.WithWordWrap(80), // Default width, content will be wrapped by the style anyway
	)
	if err != nil {
		log.Error().Err(err).Str("component", "timeline_registry").Str("when", "factory_creation").Str("key", "renderer.llm_text.simple.v1").Str("kind", "llm_text").Msg("Failed to create glamour renderer")
		r = nil
	}

	return &LLMTextFactory{
		renderer: r,
	}
}

// max helper removed (not used in this file)

var codeBlockRe = regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\n(.*?)\n```")

func extractFirstCodeBlock(s string) string {
	m := codeBlockRe.FindStringSubmatch(s)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

func looksLikeError(s string) bool {
	t := strings.TrimSpace(strings.ToLower(s))
	return strings.HasPrefix(t, "**error**") || strings.HasPrefix(t, "error:")
}

func formatMetadata(md any) string {
	if md == nil {
		return ""
	}
	// Preferred: typed events.LLMInferenceData
	switch t := md.(type) {
	case *geppetto_events.LLMInferenceData:
		out := formatFromLLMInferenceData(t)
		return out
	case geppetto_events.LLMInferenceData:
		tt := t
		out := formatFromLLMInferenceData(&tt)
		return out
	}
	// Unknown type: ignore
	return ""
}

func formatFromLLMInferenceData(m *geppetto_events.LLMInferenceData) string {
	if m == nil {
		return ""
	}
	parts := []string{}
	// Engine omitted (duplicates model)
	if m.Model != "" {
		parts = append(parts, m.Model)
	}
	if m.Temperature != nil {
		parts = append(parts, fmt.Sprintf("t: %.2f", *m.Temperature))
	}
	if m.TopP != nil {
		parts = append(parts, fmt.Sprintf("top_p: %.2f", *m.TopP))
	}
	if m.MaxTokens != nil {
		parts = append(parts, fmt.Sprintf("max: %d", *m.MaxTokens))
	}
	if m.StopReason != nil && *m.StopReason != "" {
		parts = append(parts, "stop:"+*m.StopReason)
	}
	if m.Usage != nil && (m.Usage.InputTokens > 0 || m.Usage.OutputTokens > 0) {
		parts = append(parts, fmt.Sprintf("in: %d out: %d", m.Usage.InputTokens, m.Usage.OutputTokens))
	}
	if m.DurationMs != nil && *m.DurationMs > 0 {
		// tokens per second based on output tokens only
		if m.Usage != nil && m.Usage.OutputTokens > 0 {
			sec := float64(*m.DurationMs) / 1000.0
			if sec > 0 {
				tps := float64(m.Usage.OutputTokens) / sec
				parts = append(parts, fmt.Sprintf("tps: %.2f", tps))
			}
		}
		parts = append(parts, fmt.Sprintf("%dms", *m.DurationMs))
	}
	return strings.Join(parts, " ")
}
