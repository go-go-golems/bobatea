package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	replhelp "github.com/go-go-golems/bobatea/pkg/repl/help"
)

// MarkdownBackend provides a simple in-memory backend for plain markdown sections.
type MarkdownBackend struct {
	mu       sync.RWMutex
	sections map[string]*replhelp.Section
}

func NewMarkdownBackend() *MarkdownBackend {
	return &MarkdownBackend{sections: map[string]*replhelp.Section{}}
}

func (m *MarkdownBackend) AddSection(s *replhelp.Section) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s == nil || strings.TrimSpace(s.Slug) == "" {
		return
	}
	m.sections[s.Slug] = s
}

// AddFile loads a markdown file; if withFrontmatter is false, it treats the entire file as content.
// It derives the slug from the filename and first heading as title when available.
func (m *MarkdownBackend) AddFile(path string, withFrontmatter bool) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(b)
	slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	title := deriveTitle(content)
	if strings.TrimSpace(title) == "" {
		title = slug
	}
	s := &replhelp.Section{
		Slug:    slug,
		Title:   title,
		Short:   "",
		Content: content,
		Type:    replhelp.TypeTopic,
		Order:   100,
	}
	m.AddSection(s)
	return nil
}

func deriveTitle(md string) string {
	for _, line := range strings.Split(md, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(l, "# "))
		}
	}
	return ""
}

func (m *MarkdownBackend) TopLevel(ctx context.Context) (*replhelp.TopLevelPage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var topics []*replhelp.Section
	for _, s := range m.sections {
		ss := s
		topics = append(topics, ss)
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].Order < topics[j].Order || (topics[i].Order == topics[j].Order && topics[i].Slug < topics[j].Slug) })
	return &replhelp.TopLevelPage{AllGeneralTopics: topics}, nil
}

func (m *MarkdownBackend) GetBySlug(ctx context.Context, slug string) (*replhelp.Section, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.sections[slug]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MarkdownBackend) Query(ctx context.Context, dsl string) ([]*replhelp.Section, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	q := strings.TrimSpace(dsl)
	if q == "" {
		return nil, true, nil
	}
	var out []*replhelp.Section
	for _, s := range m.sections {
		if strings.Contains(strings.ToLower(s.Slug), strings.ToLower(q)) || strings.Contains(strings.ToLower(s.Title), strings.ToLower(q)) || strings.Contains(strings.ToLower(s.Content), strings.ToLower(q)) {
			ss := s
			out = append(out, ss)
		}
	}
	return out, true, nil
}


