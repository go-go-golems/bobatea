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

type prov struct{ title string; items []diff.DiffItem }
func (p *prov) Title() string { return p.title }
func (p *prov) Items() []diff.DiffItem { return p.items }

type it struct{ id,name string; cats []diff.Category }
func (i it) ID() string { return i.id }
func (i it) Name() string { return i.name }
func (i it) Categories() []diff.Category { return i.cats }

type ca struct{ name string; changes []diff.Change }
func (c ca) Name() string { return c.name }
func (c ca) Changes() []diff.Change { return c.changes }

type chg struct{ path string; status diff.ChangeStatus; before,after any; sensitive bool }
func (c chg) Path() string { return c.path }
func (c chg) Status() diff.ChangeStatus { return c.status }
func (c chg) Before() any { return c.before }
func (c chg) After() any { return c.after }
func (c chg) Sensitive() bool { return c.sensitive }

func buildItem(rel string, b, a map[string]any, sens map[string]struct{}) diff.DiffItem {
	ad, rm, up := xdiff.MapDiff(b, a)
	var cs []diff.Change
	for _, r := range rm { _, s := sens[r.Path]; cs = append(cs, chg{path: r.Path, status: diff.ChangeStatusRemoved, before: r.BeforeAny, sensitive: s}) }
	for _, adx := range ad { _, s := sens[adx.Path]; cs = append(cs, chg{path: adx.Path, status: diff.ChangeStatusAdded, after: adx.AfterAny, sensitive: s}) }
	for _, u := range up { _, s := sens[u.Path]; cs = append(cs, chg{path: u.Path, status: diff.ChangeStatusUpdated, before: u.BeforeAny, after: u.AfterAny, sensitive: s}) }
	sort.SliceStable(cs, func(i,j int) bool { return cs[i].Path() < cs[j].Path() })
	return it{id: rel, name: rel, cats: []diff.Category{ ca{name: "k8s", changes: cs} }}
}

func collectYAML(dir string) (map[string]map[string]any, error) {
	out := map[string]map[string]any{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() { return nil }
		name := d.Name()
		if !(strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")) { return nil }
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
	flag.StringVar(&beforeDir, "before", "./examples/diff/k8s-dir/before", "Before k8s manifests dir")
	flag.StringVar(&afterDir, "after", "./examples/diff/k8s-dir/after", "After k8s manifests dir")
	flag.Parse()

    ensureSampleK8sDir(beforeDir, afterDir)

	b, err := collectYAML(beforeDir); if err != nil { log.Fatal(err) }
	a, err := collectYAML(afterDir); if err != nil { log.Fatal(err) }

	keys := map[string]struct{}{}
	for k:= range b { keys[k] = struct{}{} }
	for k:= range a { keys[k] = struct{}{} }
	files := make([]string, 0, len(keys))
	for k := range keys { files = append(files, k) }
	sort.Strings(files)

	sens := map[string]struct{}{
		"data.password": {},
		"data.apiKey": {},
		"spec.template.spec.containers[0].env[API_KEY]": {},
	}

	var items []diff.DiffItem
	for _, f := range files {
		items = append(items, buildItem(f, b[f], a[f], sens))
	}

	p := &prov{ title: "Kubernetes Manifests Diff", items: items }
	cfg := diff.DefaultConfig(); cfg.Title = "Kubernetes Manifests Diff"; cfg.RedactSensitive = true
	m := diff.NewModel(p, cfg)
	if _, err := tea.NewProgram(m, tea.WithContext(context.Background()), tea.WithAltScreen()).Run(); err != nil { log.Fatal(err) }
}

func ensureSampleK8sDir(beforeDir, afterDir string) {
    if _, err := os.Stat(beforeDir); os.IsNotExist(err) { _ = os.MkdirAll(beforeDir, 0o755) }
    if _, err := os.Stat(afterDir); os.IsNotExist(err) { _ = os.MkdirAll(afterDir, 0o755) }
    createIfMissing(filepath.Join(beforeDir, "deployment.yaml"), "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: svc\nspec:\n  replicas: 1\n")
    createIfMissing(filepath.Join(afterDir, "deployment.yaml"), "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: svc\nspec:\n  replicas: 2\n")
}

func createIfMissing(path, content string) {
    if _, err := os.Stat(path); err == nil { return }
    if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
        _, _ = fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", path, err)
    }
}
