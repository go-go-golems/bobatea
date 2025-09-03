package adapters

import (
	"context"
	"fmt"

	replhelp "github.com/go-go-golems/bobatea/pkg/repl/help"
)

// MultiBackend composes several Backends into one logical backend.
// TopLevel pages are merged; GetBySlug queries each backend in order; Query merges results from all supporting backends.
type MultiBackend struct {
	backends []replhelp.Backend
}

func NewMultiBackend(backends ...replhelp.Backend) *MultiBackend {
	return &MultiBackend{backends: backends}
}

func (m *MultiBackend) TopLevel(ctx context.Context) (*replhelp.TopLevelPage, error) {
	merged := &replhelp.TopLevelPage{}
	for _, b := range m.backends {
		page, err := b.TopLevel(ctx)
		if err != nil {
			continue
		}
		merged.AllGeneralTopics = append(merged.AllGeneralTopics, page.AllGeneralTopics...)
		merged.AllExamples = append(merged.AllExamples, page.AllExamples...)
		merged.AllApplications = append(merged.AllApplications, page.AllApplications...)
		merged.AllTutorials = append(merged.AllTutorials, page.AllTutorials...)
	}
	return merged, nil
}

func (m *MultiBackend) GetBySlug(ctx context.Context, slug string) (*replhelp.Section, error) {
	for _, b := range m.backends {
		s, err := b.GetBySlug(ctx, slug)
		if err == nil && s != nil {
			return s, nil
		}
	}
	return nil, fmt.Errorf("section not found: %s", slug)
}

func (m *MultiBackend) Query(ctx context.Context, dsl string) ([]*replhelp.Section, bool, error) {
	var out []*replhelp.Section
	var anySupported bool
	for _, b := range m.backends {
		res, ok, err := b.Query(ctx, dsl)
		if err != nil {
			return nil, true, err
		}
		if ok {
			anySupported = true
		}
		out = append(out, res...)
	}
	return out, anySupported, nil
}
