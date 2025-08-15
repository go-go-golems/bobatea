package renderers

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
	"sort"
	"strings"
)

// PlainModel is a debugging model that prints all props.
type PlainModel struct {
	props    map[string]any
	width    int
	selected bool
}

func (m *PlainModel) Init() tea.Cmd { return nil }
func (m *PlainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case timeline.EntityCopyTextMsg:
		return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.View()} }
	case timeline.EntityCopyCodeMsg:
		return m, func() tea.Msg { return timeline.CopyTextRequestedMsg{Text: m.View()} }
	}
	return m, nil
}
func (m *PlainModel) View() string {
	st := chatstyle.DefaultStyles()
	sty := st.UnselectedMessage
	if m.selected {
		sty = st.SelectedMessage
	}
	keys := make([]string, 0, len(m.props))
	for k := range m.props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+toString(m.props[k]))
	}
	return sty.Width(m.width - sty.GetHorizontalPadding()).Render("[entity] " + strings.Join(parts, " "))
}
func (m *PlainModel) OnProps(patch map[string]any) {
	if m.props == nil {
		m.props = map[string]any{}
	}
	for k, v := range patch {
		m.props[k] = v
	}
	if v, ok := patch["selected"].(bool); ok {
		m.selected = v
	}
}

// Removed OnCompleted/SetSize/Focus/Blur; handled via messages

type PlainFactory struct{}

func (PlainFactory) Key() string  { return "renderer.plain.v1" }
func (PlainFactory) Kind() string { return "plain" }
func (PlainFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	m := &PlainModel{}
	m.OnProps(initialProps)
	return m
}

func toString(v any) string { return fmt.Sprintf("%v", v) }
