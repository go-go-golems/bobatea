---
Title: REPL Help System
Slug: repl-help-system
Short: Surface documentation and searchable pages in the REPL using pluggable backends.
Topics:
- repl
- help
- documentation
- dsl
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# REPL Help System

## Overview

The REPL help system turns your application’s docs into a navigable set of pages that can be shown directly in the timeline UI. It is backend-agnostic: implement a simple `Backend` to feed sections; or use the provided adapters for Glazed help and slash commands.

Key packages:

- `github.com/go-go-golems/bobatea/pkg/repl/help` — generic help API
- `github.com/go-go-golems/bobatea/pkg/repl/help/adapters` — adapters for Glazed and slash

## Concepts

- `Section` — a single documentation item (title, slug, content, metadata)
- `TopLevelPage` — collections grouped by type (topics, examples, applications, tutorials)
- `Backend` — provides `TopLevel`, `GetBySlug`, and optional `Query`

## Backend API

```go
package replhelp

type Section struct {
  Slug, Title, Short, Content string
  Type string // topic|example|application|tutorial
  Topics, Flags, Commands []string
  ShowPerDefault bool
  Order int
}

type TopLevelPage struct {
  AllGeneralTopics []*Section
  AllExamples      []*Section
  AllApplications  []*Section
  AllTutorials     []*Section
}

type Backend interface {
  TopLevel(ctx context.Context) (*TopLevelPage, error)
  GetBySlug(ctx context.Context, slug string) (*Section, error)
  Query(ctx context.Context, dsl string) (results []*Section, ok bool, err error)
}
```

## Rendering and Handler

Applications typically call the `HandleHelpCommand` function to parse a `/help` input line, query a backend, and render markdown:

```go
md := replhelp.HandleHelpCommand(ctx, replhelp.Config{
  Backend:     backend,                // any Backend
  ShowRelated: true,                   // show related sections when available
  Renderer:    replhelp.DefaultRenderer(),
}, "/help --all")
```

The returned markdown can be emitted into the REPL timeline as `repl_result_markdown`.

## Ready-to-use Adapters

### Glazed Help Adapter

If your project already uses Glazed help (frontmatter Markdown + store), use the adapter:

```go
import hadapt "github.com/go-go-golems/bobatea/pkg/repl/help/adapters"

backend := &hadapt.GlazedBackend{HS: helpSystem}
```

This supports top-level pages, slugs, and the Glazed query DSL via `QuerySections`.

### Slash Command Adapter

You can expose registered slash commands as help sections:

```go
sb := hadapt.NewSlashBackend(model.SlashRegistry())
// top-level lists all commands as topics; each command is under slug "slash-<name>"
```

### Composing Backends

Use `MultiBackend` to merge multiple sources (e.g., Glazed docs + slash commands):

```go
backend := hadapt.NewMultiBackend(
  &hadapt.GlazedBackend{HS: helpSystem},
  hadapt.NewSlashBackend(model.SlashRegistry()),
)
```

## Wiring in a Slash Command

Register a `/help` slash command that calls the handler with your backend:

```go
reg.Register(&slash.Command{
  Name:    "help",
  Summary: "Show help",
  Usage:   "/help [slug] [--all] [--query=DSL]",
  Run: func(ctx context.Context, in slash.Input, emit slash.Emitter) error {
    s := "/help"
    if len(in.Positionals) > 0 { s += " " + in.Positionals[0] }
    if _, ok := in.Flags["all"]; ok { s += " --all" }
    if qv, ok := in.Flags["query"]; ok && len(qv) > 0 { s += " --query=\"" + qv[0] + "\"" }
    md := replhelp.HandleHelpCommand(ctx, replhelp.Config{
      Backend:     backend,
      ShowRelated: true,
      Renderer:    replhelp.DefaultRenderer(),
    }, s)
    emit("repl_result_markdown", map[string]any{"markdown": md})
    return nil
  },
})
```

## Query DSL

If your backend supports queries, `--query` hands the raw DSL to `Backend.Query`. With Glazed, this includes boolean ops, field filters, grouping, and quoted text search.

Examples:

```
/help --query "type:example AND topic:database"
/help --query "command:json OR command:yaml"
/help --query "\"SQLite integration\""
```

## Writing Docs

For Glazed-backed docs, add Markdown files with YAML frontmatter (Title, Slug, SectionType, IsTopLevel, ShowPerDefault, etc.). See the style guide: `how-to-write-good-documentation-pages`.

## Best Practices

- Use meaningful slugs and short summaries.
- Keep examples and tutorials concise, link out for details.
- Use `Order` to control sort order.
- In composite backends, assign distinct slug prefixes (e.g., `slash-`) to avoid collisions.

## See Also

- `repl-slash-system` — register and complete commands that you can document via this help system


