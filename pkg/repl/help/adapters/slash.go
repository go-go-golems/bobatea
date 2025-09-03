package adapters

import (
	"context"
	"fmt"
	"sort"
	"strings"

	replhelp "github.com/go-go-golems/bobatea/pkg/repl/help"
	"github.com/go-go-golems/bobatea/pkg/repl/slash"
)

// SlashBackend adapts a slash command Registry into replhelp.Backend
type SlashBackend struct {
	Reg slash.Registry
}

func NewSlashBackend(reg slash.Registry) *SlashBackend { return &SlashBackend{Reg: reg} }

func (s *SlashBackend) TopLevel(ctx context.Context) (*replhelp.TopLevelPage, error) {
	cmds := s.Reg.List()
	// stable order by name
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })

	var topics []*replhelp.Section
	for _, c := range cmds {
		topics = append(topics, commandToSection(c))
	}
	return &replhelp.TopLevelPage{AllGeneralTopics: topics}, nil
}

func (s *SlashBackend) GetBySlug(ctx context.Context, slug string) (*replhelp.Section, error) {
	name := strings.TrimPrefix(slug, "slash-")
	if name == slug {
		// also accept "/name" or plain name
		name = strings.TrimPrefix(name, "/")
	}
	if name == "" {
		return nil, fmt.Errorf("not found")
	}
	if cmd := s.Reg.Get(name); cmd != nil {
		return commandToSection(cmd), nil
	}
	return nil, fmt.Errorf("not found")
}

func (s *SlashBackend) Query(ctx context.Context, dsl string) ([]*replhelp.Section, bool, error) {
	// Provide a minimal text search over names and summaries when quoted text is used
	q := strings.TrimSpace(dsl)
	if q == "" {
		return nil, true, nil
	}
	// crude contains search, ignore DSL fields
	cmds := s.Reg.List()
	var out []*replhelp.Section
	for _, c := range cmds {
		if strings.Contains(c.Name, q) || strings.Contains(c.Summary, q) || strings.Contains(c.Usage, q) {
			out = append(out, commandToSection(c))
		}
	}
	return out, true, nil
}

func commandToSection(c *slash.Command) *replhelp.Section {
	title := "/" + c.Name
	short := c.Summary
	content := buildCommandContent(c)
	return &replhelp.Section{
		Slug:           "slash-" + c.Name,
		Title:          title,
		Short:          short,
		Content:        content,
		Type:           replhelp.TypeTopic,
		Topics:         []string{"slash", "commands"},
		Flags:          []string{},
		Commands:       []string{c.Name},
		ShowPerDefault: true,
		Order:          50,
	}
}

func buildCommandContent(c *slash.Command) string {
	var b strings.Builder
	b.WriteString("# /")
	b.WriteString(c.Name)
	if strings.TrimSpace(c.Summary) != "" {
		b.WriteString(" — ")
		b.WriteString(c.Summary)
	}
	b.WriteString("\n\n")
	if strings.TrimSpace(c.Usage) != "" {
		b.WriteString("## Usage\n\n")
		b.WriteString("``````\n")
		b.WriteString(c.Usage)
		b.WriteString("\n``````\n\n")
	}
	if len(c.Schema.Positionals) > 0 || len(c.Schema.Flags) > 0 {
		b.WriteString("## Arguments and Flags\n\n")
		if len(c.Schema.Positionals) > 0 {
			b.WriteString("### Positionals\n\n")
			for _, a := range c.Schema.Positionals {
				b.WriteString("- ")
				b.WriteString(a.Name)
				if a.Required {
					b.WriteString(" (required)")
				}
				if a.Variadic {
					b.WriteString(" (variadic)")
				}
				if strings.TrimSpace(a.Description) != "" {
					b.WriteString(": ")
					b.WriteString(a.Description)
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
		if len(c.Schema.Flags) > 0 {
			b.WriteString("### Flags\n\n")
			for _, f := range c.Schema.Flags {
				b.WriteString("- --")
				b.WriteString(f.Name)
				if strings.TrimSpace(f.Description) != "" {
					b.WriteString(": ")
					b.WriteString(f.Description)
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
