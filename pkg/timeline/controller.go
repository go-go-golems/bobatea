package timeline

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"strings"
)

type Controller struct {
	store    *entityStore
	cache    *renderCache
	reg      *Registry
	width    int
	height   int
	theme    string
	selected int
}

func NewController(reg *Registry) *Controller {
	c := &Controller{store: newEntityStore(), cache: newRenderCache(), reg: reg, selected: -1}
	log.Debug().Str("component", "timeline_controller").Msg("initialized controller")
	return c
}

func (c *Controller) SetSize(w, h int)      { c.width, c.height = w, h }
func (c *Controller) SetTheme(theme string) { c.theme = theme }

func (c *Controller) OnCreated(e UIEntityCreated) {
	log.Debug().Str("component", "timeline_controller").Str("event", "created").Str("kind", e.ID.Kind).Str("local_id", e.ID.LocalID).Time("started_at", e.StartedAt).Int("props_len", len(e.Props)).Msg("applying created")
	rec := &entityRecord{ID: e.ID, Renderer: e.Renderer, Props: cloneMap(e.Props), StartedAt: e.StartedAt.UnixNano()}
	c.store.add(rec)
	if c.selected < 0 {
		c.selected = 0
	}
}

func (c *Controller) OnUpdated(e UIEntityUpdated) {
	if rec, ok := c.store.get(e.ID); ok {
		log.Debug().Str("component", "timeline_controller").Str("event", "updated").Str("kind", e.ID.Kind).Str("local_id", e.ID.LocalID).Int64("version", e.Version).Int("patch_len", len(e.Patch)).Msg("applying update")
		applyPatch(rec.Props, e.Patch)
		rec.Version = max64(rec.Version, e.Version)
		rec.UpdatedAt = e.UpdatedAt.UnixNano()
		c.cache.invalidateByID(e.ID)
	}
}

func (c *Controller) OnCompleted(e UIEntityCompleted) {
	if rec, ok := c.store.get(e.ID); ok {
		log.Debug().Str("component", "timeline_controller").Str("event", "completed").Str("kind", e.ID.Kind).Str("local_id", e.ID.LocalID).Int("result_len", len(e.Result)).Msg("applying complete")
		if len(e.Result) > 0 {
			applyPatch(rec.Props, e.Result)
		}
		rec.Completed = true
		c.cache.invalidateByID(e.ID)
	}
}

func (c *Controller) OnDeleted(e UIEntityDeleted) {
	log.Debug().Str("component", "timeline_controller").Str("event", "deleted").Str("kind", e.ID.Kind).Str("local_id", e.ID.LocalID).Msg("applying delete")
	c.store.remove(e.ID)
	c.cache.invalidateByID(e.ID)
	if c.selected >= len(c.store.order) {
		c.selected = len(c.store.order) - 1
	}
}

func (c *Controller) SelectNext() {
	if c.selected+1 < len(c.store.order) {
		c.selected++
	}
}
func (c *Controller) SelectPrev() {
	if c.selected > 0 {
		c.selected--
	}
}

func (c *Controller) View() string {
	log.Debug().Str("component", "timeline_controller").Str("phase", "view").Int("entity_count", len(c.store.order)).Msg("render start")
	var b strings.Builder
	hits := 0
	misses := 0
	for _, id := range c.store.order {
		rec, _ := c.store.get(id)
		r := c.pickRenderer(rec)
		ck := cacheKey{RendererKey: r.Key(), EntityKey: keyID(id), Width: c.width, Theme: c.theme, PropsHash: r.RelevantPropsHash(rec.Props)}
		if s, _, ok := c.cache.get(ck); ok {
			b.WriteString(s)
			b.WriteByte('\n')
			hits++
			continue
		}
		s, h, _ := r.Render(rec.Props, c.width, c.theme)
		_ = h
		c.cache.set(ck, s, h)
		b.WriteString(s)
		b.WriteByte('\n')
		misses++
	}
	out := b.String()
	log.Debug().Str("component", "timeline_controller").Str("phase", "view").Int("cache_hits", hits).Int("cache_misses", misses).Int("output_len", len(out)).Msg("render done")
	return out
}

func (c *Controller) pickRenderer(rec *entityRecord) Renderer {
	if r, ok := c.reg.GetByKey(rec.Renderer.Key); ok {
		return r
	}
	if r, ok := c.reg.GetByKind(rec.Renderer.Kind); ok {
		return r
	}
	return &plainRenderer{}
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	b, _ := json.Marshal(m)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}

func applyPatch(dst map[string]any, patch map[string]any) {
	if patch == nil {
		return
	}
	for k, v := range patch {
		dst[k] = v
	}
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
