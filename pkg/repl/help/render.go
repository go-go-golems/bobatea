package replhelp

import (
	"strings"
	"sort"
)

// Renderer renders help content to markdown strings.
type Renderer interface {
	RenderTopLevel(page *TopLevelPage) string
	RenderSection(section *Section, related map[string][]*Section) string
	RenderQueryResults(results []*Section) string
}

// DefaultRenderer returns a conservative markdown renderer that fits the timeline UI.
func DefaultRenderer() Renderer { return defaultRenderer{} }

type defaultRenderer struct{}

// GroupedRenderer is a new interface for grouped multi-backend rendering.
// This is a breaking refactor: callers should implement these for best UX.
type GroupedRenderer interface {
	RenderTopLevelGrouped(groups []TopLevelGroup) string
	RenderQueryResultsGrouped(results []ScopedSection) string
}

// Ensure defaultRenderer implements GroupedRenderer with a sensible grouped layout.
func (defaultRenderer) RenderTopLevelGrouped(groups []TopLevelGroup) string {
    var b strings.Builder
    b.WriteString("# Help\n\n")
    for _, g := range groups {
        // Group header using registration title/description
        title := strings.TrimSpace(g.BackendTitle)
        if title == "" { title = g.BackendID }
        b.WriteString("## ")
        b.WriteString(title)
        b.WriteString("\n\n")
        if d := strings.TrimSpace(g.BackendDesc); d != "" {
            b.WriteString(d)
            b.WriteString("\n\n")
        }
        if g.Page == nil { continue }
        // Sub-groups: topics/examples/applications/tutorials
        if len(g.Page.AllGeneralTopics) > 0 {
            b.WriteString("### General Topics\n\n")
            for _, s := range g.Page.AllGeneralTopics { writeSectionListItem(&b, s) }
            b.WriteString("\n")
        }
        if len(g.Page.AllExamples) > 0 {
            b.WriteString("### Examples\n\n")
            for _, s := range g.Page.AllExamples { writeSectionListItem(&b, s) }
            b.WriteString("\n")
        }
        if len(g.Page.AllApplications) > 0 {
            b.WriteString("### Applications\n\n")
            for _, s := range g.Page.AllApplications { writeSectionListItem(&b, s) }
            b.WriteString("\n")
        }
        if len(g.Page.AllTutorials) > 0 {
            b.WriteString("### Tutorials\n\n")
            for _, s := range g.Page.AllTutorials { writeSectionListItem(&b, s) }
            b.WriteString("\n")
        }
    }
    return b.String()
}

func (defaultRenderer) RenderQueryResultsGrouped(results []ScopedSection) string {
    if len(results) == 0 {
        return "No results found."
    }
    var b strings.Builder
    b.WriteString("# Help Results\n\n")
    // Group by backend ID for nicer presentation
    byID := map[string][]*Section{}
    for _, ss := range results { byID[ss.BackendID] = append(byID[ss.BackendID], ss.Section) }
    // Deterministic order: sort keys
    var ids []string
    for id := range byID { ids = append(ids, id) }
    sort.Strings(ids)
    for _, id := range ids {
        b.WriteString("## ")
        b.WriteString(id)
        b.WriteString("\n\n")
        for _, s := range byID[id] {
            writeSectionListItem(&b, s)
            b.WriteString("  To view: /help ")
            b.WriteString(s.Slug)
            b.WriteString("\n")
        }
        b.WriteString("\n")
    }
    return b.String()
}

func (defaultRenderer) RenderTopLevel(page *TopLevelPage) string {
	var b strings.Builder
	b.WriteString("# Available Help\n\n")

	if len(page.AllGeneralTopics) > 0 {
		b.WriteString("## General Topics\n\n")
		for _, s := range page.AllGeneralTopics {
			writeSectionListItem(&b, s)
		}
		b.WriteString("\n")
	}
	if len(page.AllExamples) > 0 {
		b.WriteString("## Examples\n\n")
		for _, s := range page.AllExamples {
			writeSectionListItem(&b, s)
		}
		b.WriteString("\n")
	}
	if len(page.AllApplications) > 0 {
		b.WriteString("## Applications\n\n")
		for _, s := range page.AllApplications {
			writeSectionListItem(&b, s)
		}
		b.WriteString("\n")
	}
	if len(page.AllTutorials) > 0 {
		b.WriteString("## Tutorials\n\n")
		for _, s := range page.AllTutorials {
			writeSectionListItem(&b, s)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func writeSectionListItem(b *strings.Builder, s *Section) {
	b.WriteString("- **")
	b.WriteString(strings.TrimSpace(s.Slug))
	b.WriteString("**")
	if t := strings.TrimSpace(s.Title); t != "" {
		b.WriteString(" — ")
		b.WriteString(t)
	}
	if sh := strings.TrimSpace(s.Short); sh != "" {
		b.WriteString("\n  ")
		b.WriteString(sh)
	}
	b.WriteString("\n")
}

func (defaultRenderer) RenderSection(section *Section, related map[string][]*Section) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(section.Content))
	b.WriteString("\n")

	if related != nil {
		if topics := related["topics"]; len(topics) > 0 {
			b.WriteString("\n## Related Topics\n\n")
			for _, s := range topics {
				writeSectionListItem(&b, s)
			}
		}
		if examples := related["examples"]; len(examples) > 0 {
			b.WriteString("\n## Examples\n\n")
			for _, s := range examples {
				writeSectionListItem(&b, s)
			}
		}
		if apps := related["applications"]; len(apps) > 0 {
			b.WriteString("\n## Applications\n\n")
			for _, s := range apps {
				writeSectionListItem(&b, s)
			}
		}
		if tutorials := related["tutorials"]; len(tutorials) > 0 {
			b.WriteString("\n## Tutorials\n\n")
			for _, s := range tutorials {
				writeSectionListItem(&b, s)
			}
		}
	}

	return b.String()
}

func (defaultRenderer) RenderQueryResults(results []*Section) string {
	if len(results) == 0 {
		return "No results found."
	}
	var b strings.Builder
	b.WriteString("# Help Results\n\n")
	for _, s := range results {
		writeSectionListItem(&b, s)
		b.WriteString("  To view: /help ")
		b.WriteString(s.Slug)
		b.WriteString("\n")
	}
	return b.String()
}
