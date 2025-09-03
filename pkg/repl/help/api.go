package replhelp

import (
    "context"
)

// Constants for canonical section types
const (
    TypeTopic        = "topic"
    TypeExample      = "example"
    TypeApplication  = "application"
    TypeTutorial     = "tutorial"
)

// Section is a generic help section used by the REPL help layer.
// It is intentionally independent from any particular backend implementation.
type Section struct {
    Slug           string
    Title          string
    Short          string
    Content        string
    Type           string   // topic|example|application|tutorial (extensible)
    Topics         []string
    Flags          []string
    Commands       []string
    ShowPerDefault bool
    Order          int
}

// TopLevelPage contains grouped top-level sections for discovery UIs.
type TopLevelPage struct {
    AllGeneralTopics []*Section
    AllExamples      []*Section
    AllApplications  []*Section
    AllTutorials     []*Section
}

// Backend is the minimal help backend interface required by the REPL help handler.
// - TopLevel and GetBySlug MUST be implemented.
// - Query is OPTIONAL. It should return ok=false if not supported.
type Backend interface {
    TopLevel(ctx context.Context) (*TopLevelPage, error)
    GetBySlug(ctx context.Context, slug string) (*Section, error)
    Query(ctx context.Context, dsl string) (results []*Section, ok bool, err error)
}

// RelatedBackend is an optional extension for backends that can compute related content
// for a given section. If a backend implements this interface, the handler can render
// related content when ShowRelated is enabled.
type RelatedBackend interface {
    Related(ctx context.Context, section *Section) (map[string][]*Section, error)
}





