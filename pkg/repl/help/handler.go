package replhelp

import (
	"context"
	"fmt"
	"strings"
)

// Config configures the help handler.
type Config struct {
	Registrations   []BackendRegistration
	ShowRelated     bool
	Renderer        Renderer
	ContextProvider func() (string, []string)
}

// HandleHelpCommand parses the input (expected to start with /help), queries backends,
// and renders markdown output according to the renderer.
func HandleHelpCommand(ctx context.Context, cfg Config, input string) string {
	if cfg.Renderer == nil {
		cfg.Renderer = DefaultRenderer()
	}

	args, err := ParseHelpInput(input)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	renderTop := func() string {
		var groups []TopLevelGroup
		for _, r := range cfg.Registrations {
			page, err := r.Backend.TopLevel(ctx)
			if err != nil {
				continue
			}
			groups = append(groups, TopLevelGroup{BackendID: r.ID, BackendTitle: r.Title, BackendDesc: r.Description, Page: page})
		}
		if gr, ok := cfg.Renderer.(GroupedRenderer); ok {
			return gr.RenderTopLevelGrouped(groups)
		}
		flat := &TopLevelPage{}
		for _, g := range groups {
			if g.Page == nil {
				continue
			}
			flat.AllGeneralTopics = append(flat.AllGeneralTopics, g.Page.AllGeneralTopics...)
			flat.AllExamples = append(flat.AllExamples, g.Page.AllExamples...)
			flat.AllApplications = append(flat.AllApplications, g.Page.AllApplications...)
			flat.AllTutorials = append(flat.AllTutorials, g.Page.AllTutorials...)
		}
		return cfg.Renderer.RenderTopLevel(flat)
	}

	// Default: show top-level if no specific request
	if args.ShowAll || (strings.TrimSpace(args.Slug) == "" && strings.TrimSpace(args.Query) == "" &&
		len(args.Types) == 0 && len(args.Topics) == 0 && len(args.Flags) == 0 && len(args.Commands) == 0 && strings.TrimSpace(args.Search) == "") {
		return renderTop()
	}

	// Slug lookup with id:slug, prefix, then first match
	if slug := strings.TrimSpace(args.Slug); slug != "" {
		// id:slug routing
		if i := strings.IndexByte(slug, ':'); i > 0 {
			id := slug[:i]
			inner := slug[i+1:]
			for _, r := range cfg.Registrations {
				if r.ID == id {
					sec, err := r.Backend.GetBySlug(ctx, inner)
					if err != nil || sec == nil {
						return fmt.Sprintf("Help topic '%s' not found in backend '%s'.", inner, id)
					}
					var related map[string][]*Section
					if cfg.ShowRelated {
						if rb, ok := r.Backend.(RelatedBackend); ok {
							if m, e := rb.Related(ctx, sec); e == nil {
								related = m
							}
						}
					}
					return cfg.Renderer.RenderSection(sec, related)
				}
			}
			return fmt.Sprintf("Unknown backend '%s' in slug.", id)
		}
		// slug prefix routing
		for _, r := range cfg.Registrations {
			if strings.TrimSpace(r.SlugPrefix) != "" && strings.HasPrefix(slug, r.SlugPrefix) {
				if sec, err := r.Backend.GetBySlug(ctx, slug); err == nil && sec != nil {
					var related map[string][]*Section
					if cfg.ShowRelated {
						if rb, ok := r.Backend.(RelatedBackend); ok {
							if m, e := rb.Related(ctx, sec); e == nil {
								related = m
							}
						}
					}
					return cfg.Renderer.RenderSection(sec, related)
				}
			}
		}
		// first backend that matches
		for _, r := range cfg.Registrations {
			if sec, err := r.Backend.GetBySlug(ctx, slug); err == nil && sec != nil {
				var related map[string][]*Section
				if cfg.ShowRelated {
					if rb, ok := r.Backend.(RelatedBackend); ok {
						if m, e := rb.Related(ctx, sec); e == nil {
							related = m
						}
					}
				}
				return cfg.Renderer.RenderSection(sec, related)
			}
		}
		return fmt.Sprintf("Help topic '%s' not found. Try /help --search \"%s\" or /help --query \"%s\".", slug, slug, slug)
	}

	// Build DSL and fan-out query
	if dsl, ok := BuildDSL(args); ok {
		var scoped []ScopedSection
		for _, r := range cfg.Registrations {
			res, ok, err := r.Backend.Query(ctx, dsl)
			if err != nil {
				return fmt.Sprintf("Invalid query: %v", err)
			}
			if !ok {
				continue
			}
			for _, s := range res {
				scoped = append(scoped, ScopedSection{BackendID: r.ID, Section: s})
			}
		}
		if gr, ok := cfg.Renderer.(GroupedRenderer); ok {
			return gr.RenderQueryResultsGrouped(scoped)
		}
		flat := make([]*Section, 0, len(scoped))
		for _, ss := range scoped {
			flat = append(flat, ss.Section)
		}
		return cfg.Renderer.RenderQueryResults(flat)
	}

	// Fallback
	return renderTop()
}
