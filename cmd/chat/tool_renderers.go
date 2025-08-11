package main

import (
    "fmt"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    "github.com/rs/zerolog/log"
)

// ToolWeatherRenderer renders a weather tool call/result
type ToolWeatherRenderer struct{}

var _ timeline.Renderer = (*ToolWeatherRenderer)(nil)

func (r *ToolWeatherRenderer) Key() string  { return "renderer.tool.get_weather.v1" }
func (r *ToolWeatherRenderer) Kind() string { return "tool_call" }
func (r *ToolWeatherRenderer) RelevantPropsHash(props map[string]any) string { return fmt.Sprintf("%v", props) }
func (r *ToolWeatherRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    location, _ := props["location"].(string)
    units, _ := props["units"].(string)
    if units == "" { units = "celsius" }
    status := fmt.Sprintf("[weather] %s (units=%s)", location, units)
    if res, ok := props["result"].(string); ok && res != "" { status += ": " + res }
    log.Debug().Str("renderer", r.Key()).Int("len", len(status)).Msg("rendered")
    return status, 1, nil
}

// ToolWebSearchRenderer renders a web_search tool call/result
type ToolWebSearchRenderer struct{}

var _ timeline.Renderer = (*ToolWebSearchRenderer)(nil)

func (r *ToolWebSearchRenderer) Key() string  { return "renderer.tool.web_search.v1" }
func (r *ToolWebSearchRenderer) Kind() string { return "tool_call" }
func (r *ToolWebSearchRenderer) RelevantPropsHash(props map[string]any) string { return fmt.Sprintf("%v", props) }
func (r *ToolWebSearchRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    query, _ := props["query"].(string)
    status := fmt.Sprintf("[web_search] %s", query)
    if res, ok := props["result"].(string); ok && res != "" { status += ": " + res }
    log.Debug().Str("renderer", r.Key()).Int("len", len(status)).Msg("rendered")
    return status, 1, nil
}


