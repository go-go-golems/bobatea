package main

import (
    "fmt"
    "strings"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
    "github.com/rs/zerolog/log"
)

// ToolWeatherRenderer renders a weather tool call/result
type ToolWeatherRenderer struct{}

var _ timeline.Renderer = (*ToolWeatherRenderer)(nil)

func (r *ToolWeatherRenderer) Key() string  { return "renderer.tool.get_weather.v1" }
func (r *ToolWeatherRenderer) Kind() string { return "tool_call" }
func (r *ToolWeatherRenderer) RelevantPropsHash(props map[string]any) string { return fmt.Sprintf("%v", props) }
func (r *ToolWeatherRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    frameW, _ := sty.GetFrameSize()
    inner := width - frameW
    if inner < 0 { inner = 0 }

    location, _ := props["location"].(string)
    units, _ := props["units"].(string)
    if units == "" { units = "celsius" }
    title := fmt.Sprintf("[Weather] %s (%s)", safeText(location), units)
    lines := []string{truncate(title, inner)}

    if res, ok := props["result"].(string); ok && res != "" {
        lines = append(lines, truncate("- Result: "+res, inner))
    } else {
        lines = append(lines, truncate("- Status: in progress…", inner))
    }

    content := strings.Join(lines, "\n")
    panel := sty.Width(width - sty.GetHorizontalPadding()).Render(content)
    h := 1 + strings.Count(panel, "\n")
    log.Debug().Str("renderer", r.Key()).Int("len", len(panel)).Msg("rendered")
    return panel, h, nil
}

// ToolWebSearchRenderer renders a web_search tool call/result
type ToolWebSearchRenderer struct{}

var _ timeline.Renderer = (*ToolWebSearchRenderer)(nil)

func (r *ToolWebSearchRenderer) Key() string  { return "renderer.tool.web_search.v1" }
func (r *ToolWebSearchRenderer) Kind() string { return "tool_call" }
func (r *ToolWebSearchRenderer) RelevantPropsHash(props map[string]any) string { return fmt.Sprintf("%v", props) }
func (r *ToolWebSearchRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    frameW, _ := sty.GetFrameSize()
    inner := width - frameW
    if inner < 0 { inner = 0 }

    query, _ := props["query"].(string)
    title := fmt.Sprintf("[Web Search] %s", safeText(query))
    lines := []string{truncate(title, inner)}

    if res, ok := props["result"].(string); ok && res != "" {
        // For demo, show summarized result line
        lines = append(lines, truncate("- Summary: "+res, inner))
    } else {
        lines = append(lines, truncate("- Status: querying…", inner))
    }

    content := strings.Join(lines, "\n")
    panel := sty.Width(width - sty.GetHorizontalPadding()).Render(content)
    h := 1 + strings.Count(panel, "\n")
    log.Debug().Str("renderer", r.Key()).Int("len", len(panel)).Msg("rendered")
    return panel, h, nil
}

func truncate(s string, n int) string {
    if n <= 0 { return "" }
    if len(s) <= n { return s }
    if n <= 1 { return s[:n] }
    return s[:n-1] + "…"
}

func safeText(s string) string {
    if s == "" { return "N/A" }
    return s
}


