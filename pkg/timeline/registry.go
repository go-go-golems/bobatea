package timeline

import "sync"
import "github.com/rs/zerolog/log"

type Renderer interface {
    Key() string
    Kind() string
    Render(props map[string]any, width int, theme string) (string, int, error)
    RelevantPropsHash(props map[string]any) string
}

type Registry struct {
    byKey  map[string]Renderer
    byKind map[string]Renderer
    mu     sync.RWMutex
}

func NewRegistry() *Registry {
    log.Debug().Str("component", "timeline_registry").Msg("initialized registry")
    return &Registry{byKey: map[string]Renderer{}, byKind: map[string]Renderer{}}
}

func (r *Registry) Register(renderer Renderer) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if k := renderer.Key(); k != "" {
        r.byKey[k] = renderer
    }
    if k := renderer.Kind(); k != "" {
        r.byKind[k] = renderer
    }
    log.Debug().Str("component", "timeline_registry").Str("op", "register").Str("key", renderer.Key()).Str("kind", renderer.Kind()).Msg("registered")
}

func (r *Registry) GetByKey(key string) (Renderer, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.byKey[key]
    return v, ok
}

func (r *Registry) GetByKind(kind string) (Renderer, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.byKind[kind]
    return v, ok
}


