package main

import (
	"context"
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/diff"
)

type staticProvider struct{ items []diff.DiffItem }

func (p *staticProvider) Title() string          { return "JSON Diff" }
func (p *staticProvider) Items() []diff.DiffItem { return p.items }

type staticItem struct {
	id   string
	name string
	cats []diff.Category
}

func (i staticItem) ID() string                  { return i.id }
func (i staticItem) Name() string                { return i.name }
func (i staticItem) Categories() []diff.Category { return i.cats }

type staticCategory struct {
	name    string
	changes []diff.Change
}

func (c staticCategory) Name() string           { return c.name }
func (c staticCategory) Changes() []diff.Change { return c.changes }

type staticChange struct {
	path      string
	status    diff.ChangeStatus
	beforeVal any
	afterVal  any
	sensitive bool
}

func (c staticChange) Path() string              { return c.path }
func (c staticChange) Status() diff.ChangeStatus { return c.status }
func (c staticChange) Before() any               { return c.beforeVal }
func (c staticChange) After() any                { return c.afterVal }
func (c staticChange) Sensitive() bool           { return c.sensitive }

func main() {
	var (
		noSearch  = flag.Bool("no-search", false, "disable search functionality")
		noFilters = flag.Bool("no-filters", false, "disable status filters")
	)
	flag.Parse()

	items := []diff.DiffItem{
		staticItem{
			id:   "user.settings",
			name: "user.settings",
			cats: []diff.Category{
				staticCategory{
					name: "env",
					changes: []diff.Change{
						staticChange{path: "APP_ENV", status: diff.ChangeStatusUpdated, beforeVal: "staging", afterVal: "prod"},
						staticChange{path: "DEBUG", status: diff.ChangeStatusRemoved, beforeVal: true},
					},
				},
				staticCategory{
					name: "attrs",
					changes: []diff.Change{
						staticChange{path: "password", status: diff.ChangeStatusUpdated, beforeVal: "hunter2", afterVal: "correcthorsebatterystaple", sensitive: true},
						staticChange{path: "quota", status: diff.ChangeStatusAdded, afterVal: 1024},
					},
				},
			},
		},
		staticItem{
			id:   "service.config",
			name: "service.config",
			cats: []diff.Category{
				staticCategory{
					name: "env",
					changes: []diff.Change{
						staticChange{path: "LOG_LEVEL", status: diff.ChangeStatusUpdated, beforeVal: "info", afterVal: "debug"},
						staticChange{path: "API_KEY", status: diff.ChangeStatusAdded, afterVal: "abcd-efgh", sensitive: true},
					},
				},
			},
		},
	}

	provider := &staticProvider{items: items}
	config := diff.DefaultConfig()
	config.Title = "JSON Diff"

	// Apply flags
	if *noSearch {
		config.EnableSearch = false
	}
	if *noFilters {
		config.EnableStatusFilters = false
	}

	m := diff.NewModel(provider, config)
	if _, err := tea.NewProgram(m, tea.WithContext(context.Background()), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}
