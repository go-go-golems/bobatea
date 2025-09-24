package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/examples/diff/internal/xdiff"
	"github.com/go-go-golems/bobatea/pkg/diff"
)

type fileProvider struct{
	title string
	items []diff.DiffItem
}

func (p *fileProvider) Title() string { return p.title }
func (p *fileProvider) Items() []diff.DiffItem { return p.items }

type simpleItem struct{
	id string
	name string
	cats []diff.Category
}
func (i simpleItem) ID() string { return i.id }
func (i simpleItem) Name() string { return i.name }
func (i simpleItem) Categories() []diff.Category { return i.cats }

type simpleCat struct{
	name string
	changes []diff.Change
}
func (c simpleCat) Name() string { return c.name }
func (c simpleCat) Changes() []diff.Change { return c.changes }

type simpleChange struct{
	path string
	status diff.ChangeStatus
	before any
	after any
	sensitive bool
}
func (c simpleChange) Path() string { return c.path }
func (c simpleChange) Status() diff.ChangeStatus { return c.status }
func (c simpleChange) Before() any { return c.before }
func (c simpleChange) After() any { return c.after }
func (c simpleChange) Sensitive() bool { return c.sensitive }

func buildItem(name string, before, after map[string]any, sensitivePaths map[string]struct{}) diff.DiffItem {
	added, removed, updated := xdiff.MapDiff(before, after)
	var changes []diff.Change
	for _, ch := range removed {
		_, sens := sensitivePaths[ch.Path]
		changes = append(changes, simpleChange{path: ch.Path, status: diff.ChangeStatusRemoved, before: ch.BeforeAny, sensitive: sens})
	}
	for _, ch := range added {
		_, sens := sensitivePaths[ch.Path]
		changes = append(changes, simpleChange{path: ch.Path, status: diff.ChangeStatusAdded, after: ch.AfterAny, sensitive: sens})
	}
	for _, ch := range updated {
		_, sens := sensitivePaths[ch.Path]
		changes = append(changes, simpleChange{path: ch.Path, status: diff.ChangeStatusUpdated, before: ch.BeforeAny, after: ch.AfterAny, sensitive: sens})
	}
	// stable order by path
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].Path() < changes[j].Path() })
	return simpleItem{id: name, name: name, cats: []diff.Category{ simpleCat{name: "json", changes: changes} }}
}

func main(){
	var beforePath string
	var afterPath string
	flag.StringVar(&beforePath, "before", "before.json", "Path to JSON before file")
	flag.StringVar(&afterPath, "after", "after.json", "Path to JSON after file")
	flag.Parse()

	before, err := xdiff.LoadAndFlatten(beforePath)
	if err != nil { log.Fatal(err) }
	after, err := xdiff.LoadAndFlatten(afterPath)
	if err != nil { log.Fatal(err) }

	sensitive := map[string]struct{}{ "password": {}, "api_key": {}, "secrets.token": {} }
	item := buildItem("json.diff", before, after, sensitive)

	prov := &fileProvider{ title: "JSON File Diff", items: []diff.DiffItem{ item } }
	cfg := diff.DefaultConfig()
	cfg.Title = "JSON File Diff"
	cfg.RedactSensitive = true

	m := diff.NewModel(prov, cfg)
	if _, err := tea.NewProgram(m, tea.WithContext(context.Background()), tea.WithAltScreen()).Run(); err != nil {
		_, _ = os.Stderr.WriteString(err.Error()+"\n")
		os.Exit(1)
	}
}
