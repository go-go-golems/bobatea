package renderers

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	chatstyle "github.com/go-go-golems/bobatea/pkg/timeline/chatstyle"
)

// WebSearchModel renders an aggregated panel for a single web_search ItemID.
// Props/Patch keys expected:
//   - status: string (searching|in_progress|completed|failed)
//   - query: string
//   - opened_urls: []string (or patch with key "opened_urls.append": string)
//   - results: []map[string]any (or patch with key "results.append": []map[string]any)
//   - error: string
type WebSearchModel struct {
	width   int
	status  string
	query   string
	opened  []string
	results []map[string]any
	stderr  string
	style   *chatstyle.Style
}

func (m *WebSearchModel) Init() tea.Cmd { return nil }

func (m *WebSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case timeline.EntityPropsUpdatedMsg:
		if v.Patch != nil {
			m.OnProps(v.Patch)
		}
		return m, nil
	case timeline.EntitySetSizeMsg:
		m.width = v.Width
		return m, nil
	}
	return m, nil
}

func (m *WebSearchModel) View() string {
	if m.style == nil {
		m.style = chatstyle.DefaultStyles()
	}

	// Header: state + query
	state := m.status
	if state == "" {
		state = "searching"
	}
	titleCaser := cases.Title(language.English)
	left := statusIcon(state) + " " + titleCaser.String(state)
	if m.query != "" {
		left += ": " + m.query
	}

	frame := m.style.UnselectedMessage
	frameW, _ := frame.GetFrameSize()
	contentWidth := m.width - frameW - frame.GetHorizontalPadding()
	if contentWidth < 0 {
		contentWidth = 0
	}

	head := m.style.MetadataStyle.Width(contentWidth).Render(left)

	// Body: opened pages + results
	bodyLines := []string{}
	if len(m.opened) > 0 {
		bodyLines = append(bodyLines, "Opened pages:")
		for _, u := range m.opened {
			bodyLines = append(bodyLines, "  • "+u)
		}
	}
	if len(m.results) > 0 {
		if len(bodyLines) > 0 {
			bodyLines = append(bodyLines, "")
		}
		bodyLines = append(bodyLines, "Results:")
		for _, r := range m.results {
			title, _ := r["title"].(string)
			url, _ := r["url"].(string)
			snippet, _ := r["snippet"].(string)
			line := "  • " + safe(title)
			if url != "" {
				line += "  (" + url + ")"
			}
			bodyLines = append(bodyLines, line)
			if snippet != "" {
				for _, ln := range wrap(snippet, contentWidth-4) {
					bodyLines = append(bodyLines, "    "+ln)
				}
			}
		}
	}
	if m.stderr != "" {
		if len(bodyLines) > 0 {
			bodyLines = append(bodyLines, "")
		}
		bodyLines = append(bodyLines, "Error: "+m.stderr)
	}

	body := head
	if len(bodyLines) > 0 {
		body += "\n\n" + strings.Join(bodyLines, "\n")
	}

	return frame.Width(m.width - frame.GetHorizontalPadding()).Render(body)
}

func (m *WebSearchModel) OnProps(patch map[string]any) {
	if v, ok := patch["status"].(string); ok {
		m.status = v
	}
	if v, ok := patch["query"].(string); ok {
		m.query = v
	}
	if v, ok := patch["error"].(string); ok {
		m.stderr = v
	}
	if v, ok := patch["opened_urls"].([]string); ok {
		m.opened = v
	}
	// Support []any -> []string
	if v, ok := patch["opened_urls"].([]any); ok {
		m.opened = m.opened[:0]
		for _, it := range v {
			if s, ok := it.(string); ok {
				m.opened = append(m.opened, s)
			}
		}
	}
	if v, ok := patch["opened_urls.append"].(string); ok && v != "" {
		m.opened = append(m.opened, v)
	}
	if v, ok := patch["results"].([]map[string]any); ok {
		m.results = v
	}
	if v, ok := patch["results"].([]any); ok {
		m.results = m.results[:0]
		for _, it := range v {
			if m2, ok := it.(map[string]any); ok {
				m.results = append(m.results, m2)
			}
		}
	}
	if v, ok := patch["results.append"].([]map[string]any); ok {
		m.results = append(m.results, v...)
	}
	if v, ok := patch["results.append"].([]any); ok {
		for _, it := range v {
			if m2, ok := it.(map[string]any); ok {
				m.results = append(m.results, m2)
			}
		}
	}
}

type WebSearchFactory struct{}

func (WebSearchFactory) Key() string  { return "renderer.web_search_panel.v1" }
func (WebSearchFactory) Kind() string { return "web_search" }
func (WebSearchFactory) NewEntityModel(initialProps map[string]any) timeline.EntityModel {
	m := &WebSearchModel{}
	m.OnProps(initialProps)
	return m
}

// helpers
func statusIcon(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "in_progress", "searching":
		return "🔎"
	default:
		return "🔎"
	}
}

func safe(s string) string {
	if s == "" {
		return "(untitled)"
	}
	return s
}

func wrap(s string, width int) []string {
	if width <= 4 || s == "" {
		return []string{s}
	}
	words := strings.Fields(s)
	lines := []string{}
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
			continue
		}
		if len(cur)+1+len(w) <= width {
			cur += " " + w
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
