package repl

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) computeHelpDrawerOverlayLayout(header, timelineView string) (helpDrawerOverlayLayout, bool) {
	if !m.helpDrawer.visible || m.width <= 0 || m.height <= 0 {
		return helpDrawerOverlayLayout{}, false
	}

	widthPercent := clampInt(m.helpDrawer.widthPercent, 20, 90)
	heightPercent := clampInt(m.helpDrawer.heightPercent, 20, 90)
	panelWidth := max(32, m.width*widthPercent/100)
	panelHeight := max(8, m.height*heightPercent/100)
	panelWidth = min(panelWidth, max(20, m.width-2))
	panelHeight = min(panelHeight, max(6, m.height-2))

	panelStyle := m.helpDrawerPanelStyle()
	frameWidth := panelStyle.GetHorizontalFrameSize()
	frameHeight := panelStyle.GetVerticalFrameSize()
	contentWidth := max(1, panelWidth-frameWidth)
	contentHeight := max(1, panelHeight-frameHeight)

	margin := max(0, m.helpDrawer.margin)
	headerHeight := lipgloss.Height(header)
	inputY := headerHeight + 1 + lipgloss.Height(timelineView)

	panelX := 0
	panelY := 0
	switch m.helpDrawer.dock {
	case HelpDrawerDockRight:
		panelX = m.width - margin - panelWidth
		panelY = headerHeight + 1 + margin
	case HelpDrawerDockLeft:
		panelX = margin
		panelY = headerHeight + 1 + margin
	case HelpDrawerDockBottom:
		panelX = (m.width - panelWidth) / 2
		panelY = m.height - margin - panelHeight
	case HelpDrawerDockAboveRepl:
		fallthrough
	default:
		panelX = (m.width - panelWidth) / 2
		panelY = inputY - margin - panelHeight
	}
	panelX = clampInt(panelX, 0, max(0, m.width-panelWidth))
	panelY = clampInt(panelY, 0, max(0, m.height-panelHeight))

	return helpDrawerOverlayLayout{
		PanelX:        panelX,
		PanelY:        panelY,
		PanelWidth:    panelWidth,
		PanelHeight:   panelHeight,
		ContentWidth:  contentWidth,
		ContentHeight: contentHeight,
	}, true
}

func (m *Model) renderHelpDrawerPanel(layout helpDrawerOverlayLayout) string {
	if layout.ContentWidth <= 0 || layout.ContentHeight <= 0 {
		return ""
	}

	title := "Help Drawer"
	subtitle := "No contextual help provider content yet"
	bodyLines := []string{}
	doc := m.helpDrawer.doc
	hasDoc := strings.TrimSpace(doc.Title) != "" ||
		strings.TrimSpace(doc.Subtitle) != "" ||
		strings.TrimSpace(doc.Markdown) != "" ||
		len(doc.Diagnostics) > 0 ||
		strings.TrimSpace(doc.VersionTag) != ""
	if hasDoc {
		if strings.TrimSpace(doc.Title) != "" {
			title = doc.Title
		}
		if strings.TrimSpace(doc.Subtitle) != "" {
			subtitle = doc.Subtitle
		}
		if !doc.Show {
			subtitle = "No contextual help for current input"
		}
		if strings.TrimSpace(doc.Markdown) != "" {
			bodyLines = append(bodyLines, strings.TrimSpace(doc.Markdown))
		}
		if len(doc.Diagnostics) > 0 {
			bodyLines = append(bodyLines, "")
			bodyLines = append(bodyLines, "Diagnostics:")
			for _, d := range doc.Diagnostics {
				d = strings.TrimSpace(d)
				if d == "" {
					continue
				}
				bodyLines = append(bodyLines, "- "+d)
			}
		}
		if strings.TrimSpace(doc.VersionTag) != "" {
			bodyLines = append(bodyLines, "")
			bodyLines = append(bodyLines, "Version: "+doc.VersionTag)
		}
	}
	if m.helpDrawer.err != nil {
		subtitle = "Error"
		bodyLines = append(bodyLines, m.helpDrawer.err.Error())
	}
	if m.helpDrawer.loading {
		if hasDoc {
			subtitle = strings.TrimSpace(subtitle + " (refreshing)")
		} else {
			subtitle = "Loading..."
		}
	}
	if m.helpDrawer.pinned {
		subtitle = strings.TrimSpace(subtitle + " [pinned]")
	}

	toggleKey := bindingPrimaryKey(m.keyMap.HelpDrawerToggle, "ctrl+h")
	refreshKey := bindingPrimaryKey(m.keyMap.HelpDrawerRefresh, "ctrl+r")
	pinKey := bindingPrimaryKey(m.keyMap.HelpDrawerPin, "ctrl+g")
	footer := fmt.Sprintf("%s toggle • %s refresh • %s pin", toggleKey, refreshKey, pinKey)
	content := []string{
		m.helpDrawerTitleStyle().Render(title),
		m.helpDrawerSubtitleStyle().Render(subtitle),
	}
	if len(bodyLines) > 0 {
		content = append(content, strings.Join(bodyLines, "\n"))
	}
	content = append(content, m.styles.HelpText.Render(footer))

	rendered := m.helpDrawerPanelStyle().
		Width(layout.PanelWidth).
		Height(layout.PanelHeight).
		Render(strings.Join(content, "\n\n"))
	return rendered
}

func (m *Model) helpDrawerPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
}

func (m *Model) helpDrawerTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33"))
}

func (m *Model) helpDrawerSubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))
}

func bindingPrimaryKey(b key.Binding, fallback string) string {
	if !b.Enabled() {
		return fallback
	}
	keyName := strings.TrimSpace(b.Help().Key)
	if keyName == "" {
		return fallback
	}
	return keyName
}
