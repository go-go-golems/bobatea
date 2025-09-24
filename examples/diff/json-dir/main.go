package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
    "fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/examples/diff/internal/xdiff"
	"github.com/go-go-golems/bobatea/pkg/diff"
)

type dirProvider struct{
	title string
	items []diff.DiffItem
}

func (p *dirProvider) Title() string { return p.title }
func (p *dirProvider) Items() []diff.DiffItem { return p.items }

type item struct{ id,name string; cats []diff.Category }
func (i item) ID() string { return i.id }
func (i item) Name() string { return i.name }
func (i item) Categories() []diff.Category { return i.cats }

type cat struct{ name string; changes []diff.Change }
func (c cat) Name() string { return c.name }
func (c cat) Changes() []diff.Change { return c.changes }

type ch struct{ path string; status diff.ChangeStatus; before,after any; sensitive bool }
func (x ch) Path() string { return x.path }
func (x ch) Status() diff.ChangeStatus { return x.status }
func (x ch) Before() any { return x.before }
func (x ch) After() any { return x.after }
func (x ch) Sensitive() bool { return x.sensitive }

func buildItem(rel string, before, after map[string]any, sensitive map[string]struct{}) diff.DiffItem {
	added, removed, updated := xdiff.MapDiff(before, after)
	var changes []diff.Change
	for _, r := range removed { _, s := sensitive[r.Path]; changes = append(changes, ch{path: r.Path, status: diff.ChangeStatusRemoved, before: r.BeforeAny, sensitive: s}) }
	for _, a := range added { _, s := sensitive[a.Path]; changes = append(changes, ch{path: a.Path, status: diff.ChangeStatusAdded, after: a.AfterAny, sensitive: s}) }
	for _, u := range updated { _, s := sensitive[u.Path]; changes = append(changes, ch{path: u.Path, status: diff.ChangeStatusUpdated, before: u.BeforeAny, after: u.AfterAny, sensitive: s}) }
	sort.SliceStable(changes, func(i,j int) bool { return changes[i].Path() < changes[j].Path() })
	return item{id: rel, name: rel, cats: []diff.Category{ cat{name: "json", changes: changes} }}
}

func collect(dir string) (map[string]map[string]any, error) {
	out := map[string]map[string]any{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() { return nil }
		if !strings.HasSuffix(d.Name(), ".json") { return nil }
		flat, err := xdiff.LoadAndFlatten(path)
		if err != nil { return err }
		rel, _ := filepath.Rel(dir, path)
		out[rel] = flat
		return nil
	})
	return out, err
}

func main(){
	var beforeDir, afterDir string
	flag.StringVar(&beforeDir, "before", "./examples/diff/json-dir/before", "Before directory")
	flag.StringVar(&afterDir, "after", "./examples/diff/json-dir/after", "After directory")
	flag.Parse()

    // Auto-generate sample fixtures if missing
    ensureSampleJSONDir(beforeDir, afterDir)

	b, err := collect(beforeDir); if err != nil { log.Fatal(err) }
	a, err := collect(afterDir); if err != nil { log.Fatal(err) }

	// union of file keys
	keys := map[string]struct{}{}
	for k := range b { keys[k] = struct{}{} }
	for k := range a { keys[k] = struct{}{} }
	fileList := make([]string, 0, len(keys))
	for k := range keys { fileList = append(fileList, k) }
	sort.Strings(fileList)

	sensitive := map[string]struct{}{ "password": {}, "api_key": {}, "secrets.token": {} }
	var items []diff.DiffItem
	for _, file := range fileList {
		items = append(items, buildItem(file, b[file], a[file], sensitive))
	}

	prov := &dirProvider{ title: "JSON Directory Diff", items: items }
	cfg := diff.DefaultConfig(); cfg.Title = "JSON Directory Diff"; cfg.RedactSensitive = true

	m := diff.NewModel(prov, cfg)
	if _, err := tea.NewProgram(m, tea.WithContext(context.Background()), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}

func ensureSampleJSONDir(beforeDir, afterDir string) {
    if _, err := os.Stat(beforeDir); os.IsNotExist(err) {
        _ = os.MkdirAll(beforeDir, 0o755)
    }
    if _, err := os.Stat(afterDir); os.IsNotExist(err) {
        _ = os.MkdirAll(afterDir, 0o755)
    }
    // Only create files if directory is empty
    createIfMissing(filepath.Join(beforeDir, "app.json"), `{"name":"svc","replicas":1,"env":{"LOG_LEVEL":"info"}}\n`)
    createIfMissing(filepath.Join(afterDir, "app.json"), `{"name":"svc","replicas":2,"env":{"LOG_LEVEL":"debug"}}\n`)
    // Add a second sample file
    createIfMissing(filepath.Join(beforeDir, "db.json"), `{"host":"db","port":5432}`)
    createIfMissing(filepath.Join(afterDir, "db.json"), `{"host":"db","port":5432,"pool":{"size":10}}`)
}

func createIfMissing(path, content string) {
    if _, err := os.Stat(path); err == nil {
        return
    }
    if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
        _, _ = fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", path, err)
    }
}
