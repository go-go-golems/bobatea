package timeline

import (
    "sync"
    "github.com/rs/zerolog/log"
    tea "github.com/charmbracelet/bubbletea"
)

// EntityModel represents an interactive per-entity UI model
// It renders itself and can handle messages routed by the controller.
// For simplicity, View takes selection/focus flags.
type EntityModel interface{ tea.Model }

// EntityModelFactory constructs an EntityModel for a given renderer Key/Kind
type EntityModelFactory interface {
    Key() string
    Kind() string
    NewEntityModel(initialProps map[string]any) EntityModel
}

type Registry struct {
    mu     sync.RWMutex
    modelByKey  map[string]EntityModelFactory
    modelByKind map[string]EntityModelFactory
}

func NewRegistry() *Registry {
    log.Debug().Str("component", "timeline_registry").Msg("initialized registry")
    return &Registry{modelByKey: map[string]EntityModelFactory{}, modelByKind: map[string]EntityModelFactory{}}
}

func (r *Registry) RegisterModelFactory(factory EntityModelFactory) {
    log.Debug().Str("component", "timeline_registry").Str("op", "register_model_factory").Str("key", factory.Key()).Str("kind", factory.Kind()).Msg("registering")
    r.mu.Lock()
    defer r.mu.Unlock()
    if k := factory.Key(); k != "" {
        r.modelByKey[k] = factory
    }
    if k := factory.Kind(); k != "" {
        r.modelByKind[k] = factory
    }
    log.Debug().Str("component", "timeline_registry").Str("op", "register_model_factory").Str("key", factory.Key()).Str("kind", factory.Kind()).Msg("registered")
}

func (r *Registry) GetModelFactoryByKey(key string) (EntityModelFactory, bool) {
    log.Debug().Str("component", "timeline_registry").Str("when", "get_model_factory_by_key").Str("key", key).Msg("Getting model factory by key")
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.modelByKey[key]
    log.Debug().Str("component", "timeline_registry").Str("when", "get_model_factory_by_key").Str("key", key).Bool("ok", ok).Msg("Got model factory by key")
    return v, ok
}

func (r *Registry) GetModelFactoryByKind(kind string) (EntityModelFactory, bool) {
    log.Debug().Str("component", "timeline_registry").Str("when", "get_model_factory_by_kind").Str("kind", kind).Msg("Getting model factory by kind")
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.modelByKind[kind]
    log.Debug().Str("component", "timeline_registry").Str("when", "get_model_factory_by_kind").Str("kind", kind).Bool("ok", ok).Msg("Got model factory by kind")
    return v, ok
}


