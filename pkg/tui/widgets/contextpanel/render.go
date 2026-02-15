package contextpanel

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (w *Widget) RenderPanel(layout OverlayLayout, opts RenderOptions) string {
	if layout.ContentWidth <= 0 || layout.ContentHeight <= 0 {
		return ""
	}

	title := "Help Drawer"
	subtitle := "No contextual help provider content yet"
	bodyLines := []string{}

	doc := w.doc
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
	if w.err != nil {
		subtitle = "Error"
		bodyLines = append(bodyLines, w.err.Error())
	}
	if w.loading {
		if hasDoc {
			subtitle = strings.TrimSpace(subtitle + " (refreshing)")
		} else {
			subtitle = "Loading..."
		}
	}
	if w.pinned {
		subtitle = strings.TrimSpace(subtitle + " [pinned]")
	}

	toggleKey := keyOr(opts.ToggleBinding, "alt+h")
	refreshKey := keyOr(opts.RefreshBinding, "ctrl+r")
	pinKey := keyOr(opts.PinBinding, "ctrl+g")
	footer := fmt.Sprintf("%s toggle • %s refresh • %s pin", toggleKey, refreshKey, pinKey)
	if opts.FooterRenderer != nil {
		footer = opts.FooterRenderer(footer)
	}

	titleStyle := titleStyleDefault()
	if opts.TitleStyle != nil {
		titleStyle = *opts.TitleStyle
	}
	subtitleStyle := subtitleStyleDefault()
	if opts.SubtitleStyle != nil {
		subtitleStyle = *opts.SubtitleStyle
	}
	panelStyle := panelStyleDefault()
	if opts.PanelStyle != nil {
		panelStyle = *opts.PanelStyle
	}

	content := []string{
		titleStyle.Render(title),
		subtitleStyle.Render(subtitle),
	}
	if len(bodyLines) > 0 {
		content = append(content, strings.Join(bodyLines, "\n"))
	}
	content = append(content, footer)

	return panelStyle.
		Width(layout.PanelWidth).
		Height(layout.PanelHeight).
		Render(strings.Join(content, "\n\n"))
}

func panelStyleDefault() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
}

func titleStyleDefault() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33"))
}

func subtitleStyleDefault() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))
}

func keyOr(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
