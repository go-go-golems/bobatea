package renderers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
	"github.com/rs/zerolog/log"
)

// ToolCallResultModel renders the tool call result in light pink
type ToolCallResultModel struct {
	result   string
	width    int
	selected bool
	focused  bool
	style    *chatstyle.Style
}

func (m *ToolCallResultModel) Init() tea.Cmd { return nil }

func (m *ToolCallResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case timeline.EntitySelectedMsg:
		m.selected = true
	case timeline.EntityUnselectedMsg:
		m.selected = false
		m.focused = false
	case timeline.EntityPropsUpdatedMsg:
		if v.Patch != nil {
			m.OnProps(v.Patch)
		}
	case timeline.EntitySetSizeMsg:
		m.width = v.Width
		return m, nil
	case timeline.EntityFocusMsg:
		m.focused = true
	case timeline.EntityBlurMsg:
		m.focused = false
	}
	return m, nil
}

func (m *ToolCallResultModel) View() string {
	if m.style == nil {
		m.style = chatstyle.DefaultStyles()
	}
	// Light pink foreground
	pink := lipgloss.Color("212")
	base := m.style.UnselectedMessage.Foreground(pink)
	if m.selected {
		base = m.style.SelectedMessage.Foreground(pink)
	}
	if m.focused && !m.selected {
		base = m.style.FocusedMessage.Foreground(pink)
	}
	return base.Width(m.width - base.GetHorizontalPadding()).Render(m.result)
}

func (m *ToolCallResultModel) OnProps(patch map[string]any) {
	if v, ok := patch["result"].(string); ok {
		m.result = v
	}
}

type ToolCallResultFactory struct{}

func (ToolCallResultFactory) Key() string  { return "renderer.tool_call_result.v1" }
func (ToolCallResultFactory) Kind() string { return "tool_call_result" }
func (ToolCallResultFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	log.Debug().Str("component", "renderer").Str("kind", "tool_call_result").Interface("props", initialProps).Msg("NewEntityModel")
	m := &ToolCallResultModel{}
	m.OnProps(initialProps)
	return m
}
