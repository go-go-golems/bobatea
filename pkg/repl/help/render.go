package replhelp

import (
    "strings"
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
    b.WriteString("- ")
    b.WriteString(s.Slug)
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
            for _, s := range topics { writeSectionListItem(&b, s) }
        }
        if examples := related["examples"]; len(examples) > 0 {
            b.WriteString("\n## Examples\n\n")
            for _, s := range examples { writeSectionListItem(&b, s) }
        }
        if apps := related["applications"]; len(apps) > 0 {
            b.WriteString("\n## Applications\n\n")
            for _, s := range apps { writeSectionListItem(&b, s) }
        }
        if tutorials := related["tutorials"]; len(tutorials) > 0 {
            b.WriteString("\n## Tutorials\n\n")
            for _, s := range tutorials { writeSectionListItem(&b, s) }
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


