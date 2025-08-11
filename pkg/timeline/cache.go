package timeline

import (
    "encoding/json"
    "github.com/rs/zerolog/log"
)

type cacheKey struct {
    RendererKey string
    EntityKey   string
    Width       int
    Theme       string
    PropsHash   string
}

type renderCache struct {
    m map[cacheKey]string
    h map[cacheKey]int
}

func newRenderCache() *renderCache { return &renderCache{m: map[cacheKey]string{}, h: map[cacheKey]int{}} }

func (c *renderCache) get(k cacheKey) (string, int, bool) {
    s, ok := c.m[k]
    if ok {
        log.Debug().Str("component", "timeline_cache").Str("op", "get").Str("renderer", k.RendererKey).Str("entity", k.EntityKey).Msg("hit")
    }
    if !ok { return "", 0, false }
    return s, c.h[k], true
}

func (c *renderCache) set(k cacheKey, v string, h int) {
    c.m[k] = v
    c.h[k] = h
    log.Debug().Str("component", "timeline_cache").Str("op", "set").Str("renderer", k.RendererKey).Str("entity", k.EntityKey).Int("len", len(v)).Msg("stored")
}

func (c *renderCache) invalidateByID(id EntityID) {
    idk, _ := json.Marshal(id)
    for k := range c.m {
        if k.EntityKey == string(idk) {
            delete(c.m, k)
            delete(c.h, k)
            log.Debug().Str("component", "timeline_cache").Str("op", "invalidate").Str("entity", string(idk)).Msg("invalidated")
        }
    }
}


