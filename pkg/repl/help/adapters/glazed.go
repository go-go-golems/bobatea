package adapters

import (
	"context"

	replhelp "github.com/go-go-golems/bobatea/pkg/repl/help"
	ghelp "github.com/go-go-golems/glazed/pkg/help"
)

// GlazedBackend adapts a Glazed HelpSystem to the generic REPL help Backend.
type GlazedBackend struct {
	HS *ghelp.HelpSystem
}

func (g *GlazedBackend) TopLevel(ctx context.Context) (*replhelp.TopLevelPage, error) {
	page := g.HS.GetTopLevelHelpPage()
	return &replhelp.TopLevelPage{
		AllGeneralTopics: convertSections(page.AllGeneralTopics),
		AllExamples:      convertSections(page.AllExamples),
		AllApplications:  convertSections(page.AllApplications),
		AllTutorials:     convertSections(page.AllTutorials),
	}, nil
}

func (g *GlazedBackend) GetBySlug(ctx context.Context, slug string) (*replhelp.Section, error) {
	s, err := g.HS.GetSectionWithSlug(slug)
	if err != nil || s == nil {
		return nil, err
	}
	return convertSection(s), nil
}

func (g *GlazedBackend) Query(ctx context.Context, dsl string) ([]*replhelp.Section, bool, error) {
	list, err := g.HS.QuerySections(dsl)
	if err != nil {
		return nil, true, err
	}
	return convertSections(list), true, nil
}

// Related implements the optional RelatedBackend interface for Glazed by leveraging
// its section-level helpers.
func (g *GlazedBackend) Related(ctx context.Context, s *replhelp.Section) (map[string][]*replhelp.Section, error) {
	base, err := g.HS.GetSectionWithSlug(s.Slug)
	if err != nil || base == nil {
		return map[string][]*replhelp.Section{}, nil
	}
	out := map[string][]*replhelp.Section{}
	if xs := base.DefaultGeneralTopic(); len(xs) > 0 {
		out["topics"] = convertSections(xs)
	}
	if xs := base.DefaultExamples(); len(xs) > 0 {
		out["examples"] = convertSections(xs)
	}
	if xs := base.DefaultApplications(); len(xs) > 0 {
		out["applications"] = convertSections(xs)
	}
	if xs := base.DefaultTutorials(); len(xs) > 0 {
		out["tutorials"] = convertSections(xs)
	}
	return out, nil
}

// Conversion helpers

func convertSections(in []*ghelp.Section) []*replhelp.Section {
	out := make([]*replhelp.Section, 0, len(in))
	for _, s := range in {
		out = append(out, convertSection(s))
	}
	return out
}

func convertSection(s *ghelp.Section) *replhelp.Section {
	t := sectionTypeToString(s.SectionType)
	return &replhelp.Section{
		Slug:           s.Slug,
		Title:          s.Title,
		Short:          s.Short,
		Content:        s.Content,
		Type:           t,
		Topics:         append([]string(nil), s.Topics...),
		Flags:          append([]string(nil), s.Flags...),
		Commands:       append([]string(nil), s.Commands...),
		ShowPerDefault: s.ShowPerDefault,
		Order:          s.Order,
	}
}

func sectionTypeToString(st ghelp.SectionType) string {
	switch st {
	case ghelp.SectionGeneralTopic:
		return replhelp.TypeTopic
	case ghelp.SectionExample:
		return replhelp.TypeExample
	case ghelp.SectionApplication:
		return replhelp.TypeApplication
	case ghelp.SectionTutorial:
		return replhelp.TypeTutorial
	default:
		return ""
	}
}
