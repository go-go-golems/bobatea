package renderers

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	metadata any // prefer *events.LLMInferenceData, fallback to map[string]any
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
	role := m.role
	if role == "" {
		role = "assistant"
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

	// Append metadata line if available
	if m.metadata != nil {
		meta := formatMetadata(m.metadata)
		if meta != "" {
			metaRendered := m.style.MetadataStyle.Width(contentWidth).Render(meta)
			if body != "" {
				body += "\n\n"
			}
			body += metaRendered
		}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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
		return formatFromLLMInferenceData(t)
	case geppetto_events.LLMInferenceData:
		tt := t
		return formatFromLLMInferenceData(&tt)
	}
	// Fallback legacy maps
	mm, ok := md.(map[string]any)
	if !ok {
		return ""
	}
	// Mirror conversation metadata and extend with model and flexible usage extraction
	engine := firstString(mm, "engine")
	model := firstString(mm, "model")
	if engine == "" {
		engine = nestedString(mm, []string{"LLMMessageMetadata", "Engine"}, []string{"event_metadata", "llm", "engine"})
	}
	if model == "" {
		model = nestedString(mm, []string{"LLMMessageMetadata", "Model"}, []string{"event_metadata", "llm", "model"})
	}

	var tempStr string
	if tv, ok := firstFloat(mm, "temperature"); ok {
		tempStr = fmt.Sprintf("t: %.2f", tv)
	}
	if tempStr == "" {
		if tv, ok := nestedFloat(mm, []string{"LLMMessageMetadata", "Temperature"}, []string{"event_metadata", "llm", "temperature"}); ok {
			tempStr = fmt.Sprintf("t: %.2f", tv)
		}
	}

	inToks, outToks := extractUsageTokens(mm)
	if inToks == 0 && outToks == 0 {
		if m2, ok := mm["LLMMessageMetadata"].(map[string]any); ok {
			inToks, outToks = extractUsageTokens(m2)
		}
		if inToks == 0 && outToks == 0 {
			if ev, ok := mm["event_metadata"].(map[string]any); ok {
				inToks, outToks = extractUsageTokens(ev)
			}
		}
	}

	parts := []string{}
	if engine != "" {
		parts = append(parts, engine)
	}
	if model != "" {
		parts = append(parts, model)
	}
	if tempStr != "" {
		parts = append(parts, tempStr)
	}
	if inToks > 0 || outToks > 0 {
		parts = append(parts, fmt.Sprintf("in: %d out: %d", inToks, outToks))
	}
	return strings.Join(parts, " ")
}

func formatFromLLMInferenceData(m *geppetto_events.LLMInferenceData) string {
	if m == nil {
		return ""
	}
	parts := []string{}
	if m.Engine != "" {
		parts = append(parts, m.Engine)
	}
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
		parts = append(parts, fmt.Sprintf("%dms", *m.DurationMs))
	}
	return strings.Join(parts, " ")
}

// Helpers to robustly extract metadata fields from loose maps
func firstString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
func nestedString(m map[string]any, paths ...[]string) string {
	for _, p := range paths {
		cur := any(m)
		ok := true
		for _, k := range p {
			mm, isMap := cur.(map[string]any)
			if !isMap {
				ok = false
				break
			}
			cur, ok = mm[k]
			if !ok {
				break
			}
		}
		if ok {
			if s, ok := cur.(string); ok {
				return s
			}
		}
	}
	return ""
}
func firstFloat(m map[string]any, key string) (float64, bool) {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return t, true
		case float32:
			return float64(t), true
		case int:
			return float64(t), true
		case int64:
			return float64(t), true
		case json.Number:
			if f, err := t.Float64(); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}
func nestedFloat(m map[string]any, paths ...[]string) (float64, bool) {
	for _, p := range paths {
		cur := any(m)
		ok := true
		for _, k := range p {
			mm, isMap := cur.(map[string]any)
			if !isMap {
				ok = false
				break
			}
			cur, ok = mm[k]
			if !ok {
				break
			}
		}
		if ok {
			switch t := cur.(type) {
			case float64:
				return t, true
			case float32:
				return float64(t), true
			case int:
				return float64(t), true
			case int64:
				return float64(t), true
			case json.Number:
				if f, err := t.Float64(); err == nil {
					return f, true
				}
			}
		}
	}
	return 0, false
}
func asInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	case float32:
		return int(t), true
	case json.Number:
		if i, err := strconv.Atoi(string(t)); err == nil {
			return i, true
		}
	}
	return 0, false
}
func extractUsageTokens(m map[string]any) (int, int) {
	// Common shapes: usage{input,output} or usage{input_tokens,output_tokens}
	if u, ok := m["usage"].(map[string]any); ok {
		// flat variants
		if in, ok1 := asInt(u["input"]); ok1 {
			if out, ok2 := asInt(u["output"]); ok2 {
				return in, out
			}
		}
		if in, ok1 := asInt(u["input_tokens"]); ok1 {
			out, _ := asInt(u["output_tokens"])
			return in, out
		}
		if in, ok1 := asInt(u["InputTokens"]); ok1 {
			out, _ := asInt(u["OutputTokens"])
			return in, out
		}
	}
	// Direct keys
	if in, ok1 := asInt(m["input_tokens"]); ok1 {
		out, _ := asInt(m["output_tokens"])
		return in, out
	}
	// Nested under LLMMessageMetadata
	if md, ok := m["LLMMessageMetadata"].(map[string]any); ok {
		if u, ok := md["Usage"].(map[string]any); ok {
			if in, ok1 := asInt(u["InputTokens"]); ok1 {
				out, _ := asInt(u["OutputTokens"])
				return in, out
			}
		}
	}
	return 0, 0
}
