package timeline

import (
    "crypto/sha1"
    "encoding/hex"
    "fmt"
    "strings"
    "github.com/rs/zerolog/log"
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

// Deprecated stateless renderers have been removed in favor of Bubble Tea models.

func hashMap(m map[string]any) string {
    // simple non-stable across types: for demo only
    h := sha1.Sum([]byte(fmt.Sprintf("%v", m)))
    return hex.EncodeToString(h[:])
}

// Tool-specific renderers for the demo were moved into the chat cmd; keep core package lean.


