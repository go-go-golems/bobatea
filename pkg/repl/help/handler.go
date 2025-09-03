package replhelp

import (
    "context"
    "fmt"
    "strings"
)

// Config configures the help handler.
type Config struct {
    Backend         Backend
    ShowRelated     bool
    Renderer        Renderer
    ContextProvider func() (string, []string)
}

// HandleHelpCommand parses the input (expected to start with /help), queries the backend,
// and renders markdown output according to the renderer.
func HandleHelpCommand(ctx context.Context, cfg Config, input string) string {
    if cfg.Renderer == nil {
        cfg.Renderer = DefaultRenderer()
    }

    args, err := ParseHelpInput(input)
    if err != nil {
        return fmt.Sprintf("Error: %v", err)
    }

    // Default: show top-level if no specific request
    if args.ShowAll || (strings.TrimSpace(args.Slug) == "" && strings.TrimSpace(args.Query) == "" &&
        len(args.Types) == 0 && len(args.Topics) == 0 && len(args.Flags) == 0 && len(args.Commands) == 0 && strings.TrimSpace(args.Search) == "") {
        page, err := cfg.Backend.TopLevel(ctx)
        if err != nil {
            return fmt.Sprintf("Error: %v", err)
        }
        return cfg.Renderer.RenderTopLevel(page)
    }

    // Slug lookup has priority if provided explicitly
    if strings.TrimSpace(args.Slug) != "" {
        sec, err := cfg.Backend.GetBySlug(ctx, args.Slug)
        if err != nil || sec == nil {
            // Suggest using search
            return fmt.Sprintf("Help topic '%s' not found. Try /help --search \"%s\" or /help --query \"%s\".", args.Slug, args.Slug, args.Slug)
        }

        var related map[string][]*Section
        if cfg.ShowRelated {
            if rb, ok := cfg.Backend.(RelatedBackend); ok {
                if m, rerr := rb.Related(ctx, sec); rerr == nil {
                    related = m
                }
            }
        }
        return cfg.Renderer.RenderSection(sec, related)
    }

    // Build DSL from convenience flags / search
    if dsl, ok := BuildDSL(args); ok {
        results, supported, err := cfg.Backend.Query(ctx, dsl)
        if err != nil {
            return fmt.Sprintf("Invalid query: %v", err)
        }
        if !supported {
            return "This help backend does not support queries. Try a slug: /help <slug> or /help --all"
        }
        return cfg.Renderer.RenderQueryResults(results)
    }

    // Fallback: if nothing matched above, return top-level
    page, err := cfg.Backend.TopLevel(ctx)
    if err != nil {
        return fmt.Sprintf("Error: %v", err)
    }
    return cfg.Renderer.RenderTopLevel(page)
}





