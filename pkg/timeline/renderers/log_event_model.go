package renderers

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// LogEventModel renders a compact, borderless gray log entry with YAML-formatted metadata/fields
type LogEventModel struct {
	level    string
	message  string
	yamlStr  string
	width    int
	selected bool
	showMeta bool
}

func (m *LogEventModel) Init() tea.Cmd { return nil }

func (m *LogEventModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case timeline.EntitySelectedMsg:
		m.selected = true
	case timeline.EntityUnselectedMsg:
		m.selected = false
		// Hide metadata when unselected to keep timeline compact
		m.showMeta = false
	case timeline.EntityPropsUpdatedMsg:
		if v.Patch != nil {
			m.OnProps(v.Patch)
		}
	case timeline.EntitySetSizeMsg:
		m.width = v.Width
		return m, nil
	case tea.KeyMsg:
		// Toggle metadata visibility with TAB when in selected/interactive mode
		if m.selected && v.String() == "tab" {
			m.showMeta = !m.showMeta
			log.Debug().
				Str("component", "renderer").
				Str("kind", "log_event").
				Bool("show_meta", m.showMeta).
				Msg("toggle metadata visibility")
			return m, nil
		}
	}
	return m, nil
}

func (m *LogEventModel) View() string {
	// Borderless, unobtrusive gray style
	base := lipgloss.NewStyle().Padding(0, 1)
	// Darker gray foreground (adaptive-ish by using numbers)
	gray := lipgloss.Color("245")
	if m.selected {
		gray = lipgloss.Color("252")
	}
	base = base.Foreground(gray)

	header := strings.TrimSpace(fmt.Sprintf("[%s] %s", strings.ToUpper(strings.TrimSpace(m.level)), strings.TrimSpace(m.message)))
	body := header
	if m.showMeta && strings.TrimSpace(m.yamlStr) != "" {
		body += "\n\n" + m.yamlStr
	}
	return base.Width(m.width - base.GetHorizontalPadding()).Render(body)
}

func (m *LogEventModel) OnProps(patch map[string]any) {
	if v, ok := patch["level"].(string); ok {
		m.level = v
	}
	if v, ok := patch["message"].(string); ok {
		m.message = v
	}
	// Compose YAML from provided metadata/fields
	combined := map[string]any{}
	if v, ok := patch["metadata"]; ok && v != nil {
		combined["meta"] = v
	}
	if v, ok := patch["fields"]; ok && v != nil {
		combined["fields"] = v
	}
	if len(combined) > 0 {
		if b, err := yaml.Marshal(combined); err == nil {
			m.yamlStr = strings.TrimSpace(string(b))
		} else {
			log.Debug().Err(err).Str("component", "renderer").Str("kind", "log_event").Msg("failed to marshal yaml")
			m.yamlStr = ""
		}
	}
}

type LogEventFactory struct{}

func (LogEventFactory) Key() string  { return "renderer.log_event.v1" }
func (LogEventFactory) Kind() string { return "log_event" }
func (LogEventFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	log.Debug().Str("component", "renderer").Str("kind", "log_event").Interface("props", initialProps).Msg("NewEntityModel")
	m := &LogEventModel{}
	m.OnProps(initialProps)
	return m
}
