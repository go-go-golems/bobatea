package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
	"github.com/rs/zerolog/log"
	"strings"
)

// ToolWeatherModel renders a weather tool call/result as a Bubble Tea entity model
type ToolWeatherModel struct {
	location string
	units    string
	result   string
	spin     int
	width    int
	selected bool
}

func (m *ToolWeatherModel) Init() tea.Cmd { return nil }
func (m *ToolWeatherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case timeline.EntitySelectedMsg:
		m.selected = true
	case timeline.EntityUnselectedMsg:
		m.selected = false
	case timeline.EntityPropsUpdatedMsg:
		if v.Patch != nil {
			m.OnProps(v.Patch)
		}
	case timeline.EntitySetSizeMsg:
		m.width = v.Width
		return m, nil
	}
	return m, nil
}
func (m *ToolWeatherModel) View() string {
	st := chatstyle.DefaultStyles()
	sty := st.UnselectedMessage
	if m.selected {
		sty = st.SelectedMessage
	}
	frameW, _ := sty.GetFrameSize()
	inner := m.width - frameW
	if inner < 0 {
		inner = 0
	}

	units := m.units
	if units == "" {
		units = "celsius"
	}
	title := fmt.Sprintf("[Weather] %s (%s)", safeText(m.location), units)
	lines := []string{truncate(title, inner)}
	if m.result != "" {
		lines = append(lines, truncate("- Result: "+m.result, inner))
	} else {
		lines = append(lines, truncate(fmt.Sprintf("- Status: fetching %s", spinnerFrame(m.spin)), inner))
	}
	content := strings.Join(lines, "\n")
	return sty.Width(m.width - sty.GetHorizontalPadding()).Render(content)
}
func (m *ToolWeatherModel) OnProps(patch map[string]any) {
	if v, ok := patch["location"].(string); ok {
		m.location = v
	}
	if v, ok := patch["units"].(string); ok {
		m.units = v
	}
	if v, ok := patch["result"].(string); ok {
		m.result = v
	}
	if v, ok := patch["spin"].(int); ok {
		m.spin = v
	}
	if v, ok := patch["selected"].(bool); ok {
		m.selected = v
	}
}

// Removed OnCompleted/SetSize/Focus/Blur; handled via messages

type ToolWeatherFactory struct{}

func (ToolWeatherFactory) Key() string  { return "renderer.tool.get_weather.v1" }
func (ToolWeatherFactory) Kind() string { return "tool_call" }
func (ToolWeatherFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	m := &ToolWeatherModel{}
	m.OnProps(initialProps)
	return m
}

// ToolWebSearchModel renders a web_search tool call/result as a Bubble Tea model
type ToolWebSearchModel struct {
	query    string
	results  []string
	result   string
	spin     int
	width    int
	selected bool
}

func (m *ToolWebSearchModel) Init() tea.Cmd { return nil }
func (m *ToolWebSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case timeline.EntitySelectedMsg:
		m.selected = true
	case timeline.EntityUnselectedMsg:
		m.selected = false
	case timeline.EntityPropsUpdatedMsg:
		if v.Patch != nil {
			m.OnProps(v.Patch)
		}
	case timeline.EntitySetSizeMsg:
		m.width = v.Width
		return m, nil
	}
	return m, nil
}
func (m *ToolWebSearchModel) View() string {
	st := chatstyle.DefaultStyles()
	sty := st.UnselectedMessage
	if m.selected {
		sty = st.SelectedMessage
	}
	frameW, _ := sty.GetFrameSize()
	inner := m.width - frameW
	if inner < 0 {
		inner = 0
	}

	title := fmt.Sprintf("[Web Search] %s", safeText(m.query))
	lines := []string{truncate(title, inner)}
	if len(m.results) > 0 {
		lines = append(lines, truncate("- Results:", inner))
		for i, link := range m.results {
			lines = append(lines, truncate(fmt.Sprintf("  %d) %s", i+1, link), inner))
		}
	} else if m.result != "" {
		lines = append(lines, truncate("- Summary: "+m.result, inner))
	} else {
		lines = append(lines, truncate(fmt.Sprintf("- Status: querying %s", spinnerFrame(m.spin)), inner))
	}
	content := strings.Join(lines, "\n")
	return sty.Width(m.width - sty.GetHorizontalPadding()).Render(content)
}
func (m *ToolWebSearchModel) OnProps(patch map[string]any) {
	if v, ok := patch["query"].(string); ok {
		m.query = v
	}
	if v, ok := patch["results"].([]string); ok {
		m.results = v
	}
	if v, ok := patch["result"].(string); ok {
		m.result = v
	}
	if v, ok := patch["spin"].(int); ok {
		m.spin = v
	}
	if v, ok := patch["selected"].(bool); ok {
		m.selected = v
	}
}

// Removed OnCompleted/SetSize/Focus/Blur; handled via messages

type ToolWebSearchFactory struct{}

func (ToolWebSearchFactory) Key() string  { return "renderer.tool.web_search.v1" }
func (ToolWebSearchFactory) Kind() string { return "tool_call" }
func (ToolWebSearchFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	m := &ToolWebSearchModel{}
	m.OnProps(initialProps)
	return m
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func safeText(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func spinnerFrame(idx int) string {
	// Use bubbles spinner frames without a running model; just pick by index
	frames := spinner.Line.Frames
	if len(frames) == 0 {
		return "-"
	}
	if idx < 0 {
		idx = 0
	}
	return frames[idx%len(frames)]
}

// Checkbox interactive model
type CheckboxModel struct {
	label    string
	checked  bool
	width    int
	focused  bool
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
		if v.Patch != nil {
			m.OnProps(v.Patch)
		}
		log.Debug().Str("component", "checkbox_model").Msg("EntityPropsUpdatedMsg received")
		return m, nil
	}
	if !m.focused {
		return m, nil
	}
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
	if m.selected {
		sty = st.SelectedMessage
	}
	if m.focused {
		sty = st.FocusedMessage
	}
	frameW, _ := sty.GetFrameSize()
	inner := m.width - frameW
	if inner < 0 {
		inner = 0
	}
	box := "[ ]"
	if m.checked {
		box = "[x]"
	}
	title := fmt.Sprintf("[Checkbox] %s %s", box, safeText(m.label))
	content := truncate(title, inner)
	return sty.Width(m.width - sty.GetHorizontalPadding()).Render(content)
}
func (m *CheckboxModel) OnProps(patch map[string]any) {
	if v, ok := patch["label"].(string); ok {
		m.label = v
	}
	if v, ok := patch["checked"].(bool); ok {
		m.checked = v
	}
}
func (m *CheckboxModel) OnCompleted(_ map[string]any) {}
func (m *CheckboxModel) SetSize(w int, _ int)         { m.width = w }
func (m *CheckboxModel) Focus()                       { m.focused = true }
func (m *CheckboxModel) Blur()                        { m.focused = false }

type CheckboxFactory struct{}

func (CheckboxFactory) Key() string  { return "renderer.test.checkbox.v1" }
func (CheckboxFactory) Kind() string { return "" }
func (CheckboxFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	m := &CheckboxModel{}
	m.OnProps(initialProps)
	return m
}
