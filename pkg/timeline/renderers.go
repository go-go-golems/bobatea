package timeline

import (
    "crypto/sha1"
    "encoding/hex"
    "fmt"
    "strings"
    "github.com/rs/zerolog/log"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
)

// plainRenderer is a fallback renderer that prints props as JSON-like text.
type plainRenderer struct{}

func (p *plainRenderer) Key() string  { return "renderer.plain.v1" }
func (p *plainRenderer) Kind() string { return "plain" }
func (p *plainRenderer) RelevantPropsHash(props map[string]any) string { return hashMap(props) }
func (p *plainRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    parts := []string{"[entity]"}
    for k, v := range props {
        parts = append(parts, fmt.Sprintf("%s=%v", k, v))
    }
    s := strings.Join(parts, " ")
    log.Debug().Str("component", "timeline_renderer").Str("renderer", p.Key()).Int("width", width).Int("len", len(s)).Msg("rendered")
    return s, 1, nil
}

// LLM text renderer (simple, no markdown to keep demo self-contained)
type LLMTextRenderer struct{}

func (r *LLMTextRenderer) Key() string  { return "renderer.llm_text.simple.v1" }
func (r *LLMTextRenderer) Kind() string { return "llm_text" }
func (r *LLMTextRenderer) RelevantPropsHash(props map[string]any) string { return hashMap(map[string]any{"text": props["text"]}) }
func (r *LLMTextRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    role, _ := props["role"].(string)
    if role == "" { role = "assistant" }
    text, _ := props["text"].(string)
    st := chatstyle.DefaultStyles()
    // Use Selected=false, Focused=false for demo; can be extended by props later
    box := chatstyle.RenderBox(st, role, text, width, false, false)
    lines := 1
    if strings.Count(box, "\n") > 0 { lines = strings.Count(box, "\n") + 1 }
    log.Debug().Str("component", "timeline_renderer").Str("renderer", r.Key()).Int("width", width).Int("len", len(box)).Msg("rendered")
    return box, lines, nil
}

// Tool calls panel renderer
type ToolCallsPanelRenderer struct{}

func (r *ToolCallsPanelRenderer) Key() string  { return "renderer.tools.panel.v1" }
func (r *ToolCallsPanelRenderer) Kind() string { return "tool_calls_panel" }
func (r *ToolCallsPanelRenderer) RelevantPropsHash(props map[string]any) string { return hashMap(props) }
func (r *ToolCallsPanelRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    s := "[tools]"
    if calls, ok := props["calls"].([]any); ok {
        s += fmt.Sprintf(" %d call(s)", len(calls))
    }
    log.Debug().Str("component", "timeline_renderer").Str("renderer", r.Key()).Int("width", width).Int("len", len(s)).Msg("rendered")
    return s, 1, nil
}

func hashMap(m map[string]any) string {
    // simple non-stable across types: for demo only
    h := sha1.Sum([]byte(fmt.Sprintf("%v", m)))
    return hex.EncodeToString(h[:])
}

// Tool-specific renderers for the demo were moved into the chat cmd; keep core package lean.


