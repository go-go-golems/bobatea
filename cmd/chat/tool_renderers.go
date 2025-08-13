package main

import (
    "fmt"
    "strings"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
    "github.com/rs/zerolog/log"
)

// ToolWeatherRenderer renders a weather tool call/result
type ToolWeatherRenderer struct{}

var _ timeline.Renderer = (*ToolWeatherRenderer)(nil)

func (r *ToolWeatherRenderer) Key() string  { return "renderer.tool.get_weather.v1" }
func (r *ToolWeatherRenderer) Kind() string { return "tool_call" }
func (r *ToolWeatherRenderer) RelevantPropsHash(props map[string]any) string {
    return fmt.Sprintf("%v|sel=%v", props, props["selected"])
}
func (r *ToolWeatherRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if sel, _ := props["selected"].(bool); sel { sty = st.SelectedMessage }
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
        // show spinner while fetching
        spinIdx := 0
        if v, ok := props["spin"].(int); ok { spinIdx = v }
        lines = append(lines, truncate(fmt.Sprintf("- Status: fetching %s", spinnerFrame(spinIdx)), inner))
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
func (r *ToolWebSearchRenderer) RelevantPropsHash(props map[string]any) string {
    return fmt.Sprintf("%v|sel=%v", props, props["selected"])
}
func (r *ToolWebSearchRenderer) Render(props map[string]any, width int, theme string) (string, int, error) {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if sel, _ := props["selected"].(bool); sel { sty = st.SelectedMessage }
    frameW, _ := sty.GetFrameSize()
    inner := width - frameW
    if inner < 0 { inner = 0 }

    query, _ := props["query"].(string)
    title := fmt.Sprintf("[Web Search] %s", safeText(query))
    lines := []string{truncate(title, inner)}

    if resultsAny, ok := props["results"].([]string); ok {
        if len(resultsAny) > 0 {
            lines = append(lines, truncate("- Results:", inner))
            for i, link := range resultsAny {
                lines = append(lines, truncate(fmt.Sprintf("  %d) %s", i+1, link), inner))
            }
        }
    } else if resListAny, ok := props["results"].([]any); ok {
        if len(resListAny) > 0 {
            lines = append(lines, truncate("- Results:", inner))
            for i, v := range resListAny {
                lines = append(lines, truncate(fmt.Sprintf("  %d) %v", i+1, v), inner))
            }
        }
    }

    if _, hasResults := props["results"]; !hasResults {
        if res, ok := props["result"].(string); ok && res != "" {
            lines = append(lines, truncate("- Summary: "+res, inner))
        } else {
            spinIdx := 0
            if v, ok := props["spin"].(int); ok { spinIdx = v }
            lines = append(lines, truncate(fmt.Sprintf("- Status: querying %s", spinnerFrame(spinIdx)), inner))
        }
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

func spinnerFrame(idx int) string {
    // Use bubbles spinner frames without a running model; just pick by index
    frames := spinner.Line.Frames
    if len(frames) == 0 { return "-" }
    if idx < 0 { idx = 0 }
    return frames[idx%len(frames)]
}

// Checkbox interactive model
type CheckboxModel struct {
    label   string
    checked bool
    width   int
    focused bool
    selected bool
}

func (m *CheckboxModel) Init() tea.Cmd { return nil }
func (m *CheckboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch v := msg.(type) {
    case timeline.EntitySelectedMsg:
        m.selected = true
        log.Debug().Str("component", "checkbox_model").Msg("EntitySelectedMsg received")
        return m, nil
    case timeline.EntityUnselectedMsg:
        m.selected = false
        m.focused = false
        log.Debug().Str("component", "checkbox_model").Msg("EntityUnselectedMsg received")
        return m, nil
    case timeline.EntityPropsUpdatedMsg:
        if v.Patch != nil { m.OnProps(v.Patch) }
        log.Debug().Str("component", "checkbox_model").Msg("EntityPropsUpdatedMsg received")
        return m, nil
    }
    if !m.focused { return m, nil }
    if k, ok := msg.(tea.KeyMsg); ok {
        log.Debug().Str("component", "checkbox_model").Str("key", k.String()).Msg("Update received key")
        if k.String() == " " {
            m.checked = !m.checked
            log.Debug().Str("component", "checkbox_model").Bool("checked", m.checked).Msg("Toggled")
        }
    }
    return m, nil
}
func (m *CheckboxModel) View() string {
    st := chatstyle.DefaultStyles()
    sty := st.UnselectedMessage
    if m.selected { sty = st.SelectedMessage }
    if m.focused { sty = st.FocusedMessage }
    frameW, _ := sty.GetFrameSize()
    inner := m.width - frameW
    if inner < 0 { inner = 0 }
    box := "[ ]"; if m.checked { box = "[x]" }
    title := fmt.Sprintf("[Checkbox] %s %s", box, safeText(m.label))
    content := truncate(title, inner)
    return sty.Width(m.width - sty.GetHorizontalPadding()).Render(content)
}
func (m *CheckboxModel) OnProps(patch map[string]any) {
    if v, ok := patch["label"].(string); ok { m.label = v }
    if v, ok := patch["checked"].(bool); ok { m.checked = v }
}
func (m *CheckboxModel) OnCompleted(_ map[string]any) {}
func (m *CheckboxModel) SetSize(w int, _ int) { m.width = w }
func (m *CheckboxModel) Focus()  { m.focused = true }
func (m *CheckboxModel) Blur()   { m.focused = false }

type CheckboxFactory struct{}
func (CheckboxFactory) Key() string  { return "renderer.test.checkbox.v1" }
func (CheckboxFactory) Kind() string { return "" }
func (CheckboxFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
    m := &CheckboxModel{}
    m.OnProps(initialProps)
    return m
}


