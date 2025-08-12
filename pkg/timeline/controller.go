package timeline

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
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
	// selectionVisible controls whether renderers receive selected=true in props
	selectionVisible bool
	entering         bool
}

func NewController(reg *Registry) *Controller {
	c := &Controller{store: newEntityStore(), cache: newRenderCache(), reg: reg, selected: -1}
	log.Debug().Str("component", "timeline_controller").Msg("initialized controller")
	return c
}

func (c *Controller) SetSize(w, h int) {
	c.width, c.height = w, h
	// Propagate new size to interactive models
	for _, id := range c.store.order {
		if rec, ok := c.store.get(id); ok {
			if rec.model != nil {
				rec.model.SetSize(w, h)
			}
		}
	}
}
func (c *Controller) SetTheme(theme string) {
	c.theme = theme
	// Propagate theme to interactive models
	for _, id := range c.store.order {
		if rec, ok := c.store.get(id); ok {
			if rec.model != nil {
				rec.model.OnProps(map[string]any{"theme": theme})
			}
		}
	}
}

func (c *Controller) OnCreated(e UIEntityCreated) {
    log.Debug().Str("component", "timeline_controller").Str("event", "created").Str("kind", e.ID.Kind).Str("local_id", e.ID.LocalID).Time("started_at", e.StartedAt).Int("props_len", len(e.Props)).Msg("applying created")
	rec := &entityRecord{ID: e.ID, Renderer: e.Renderer, Props: cloneMap(e.Props), StartedAt: e.StartedAt.UnixNano()}
	// Instantiate interactive model if a factory is registered
	if e.Renderer.Key != "" {
		if f, ok := c.reg.GetModelFactoryByKey(e.Renderer.Key); ok {
			rec.model = f.NewEntityModel(rec.Props)
			rec.model.SetSize(c.width, c.height)
			if c.theme != "" {
				rec.model.OnProps(map[string]any{"theme": c.theme})
			}
		}
	}
	// Fallback: try factory by kind when no key-specific factory is found
	if rec.model == nil && e.Renderer.Kind != "" {
		if f, ok := c.reg.GetModelFactoryByKind(e.Renderer.Kind); ok {
			rec.model = f.NewEntityModel(rec.Props)
			rec.model.SetSize(c.width, c.height)
			if c.theme != "" {
				rec.model.OnProps(map[string]any{"theme": c.theme})
			}
		}
	}
	c.store.add(rec)
	if c.selected < 0 {
		c.selected = 0
	}
}

func (c *Controller) OnUpdated(e UIEntityUpdated) {
	if rec, ok := c.store.get(e.ID); ok {
		log.Debug().Str("component", "timeline_controller").Str("event", "updated").Str("kind", e.ID.Kind).Str("local_id", e.ID.LocalID).Int64("version", e.Version).Int("patch_len", len(e.Patch)).Msg("applying update")
		applyPatch(rec.Props, e.Patch)
		if rec.model != nil {
			rec.model.OnProps(e.Patch)
			// Also emit a structured props-updated message for models that react in Update
			rec.model.Update(EntityPropsUpdatedMsg{ID: rec.ID, Patch: e.Patch})
		}
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
		if rec.model != nil {
			rec.model.OnCompleted(e.Result)
		}
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
		log.Debug().Str("component", "timeline_controller").Str("op", "select_next").Int("selected_index", c.selected).Int("count", len(c.store.order)).Msg("selection changed")
	}
}
func (c *Controller) SelectPrev() {
	if c.selected > 0 {
		c.selected--
		log.Debug().Str("component", "timeline_controller").Str("op", "select_prev").Int("selected_index", c.selected).Int("count", len(c.store.order)).Msg("selection changed")
	}
}

// SelectLast selects the last entity if any exist.
func (c *Controller) SelectLast() {
	if len(c.store.order) > 0 {
		c.selected = len(c.store.order) - 1
		log.Debug().Str("component", "timeline_controller").Str("op", "select_last").Int("selected", c.selected).Int("count", len(c.store.order)).Msg("selection changed")
	}
}

// SelectedIndex returns the current selected index or -1 if none.
func (c *Controller) SelectedIndex() int { return c.selected }

// SetSelectionVisible toggles whether renderers should highlight selection.
func (c *Controller) SetSelectionVisible(v bool) {
	c.selectionVisible = v
	log.Debug().Str("component", "timeline_controller").Str("op", "set_selection_visible").Bool("visible", v).Int("selected_index", c.selected).Msg("selection visibility updated")
}

func (c *Controller) View() string {
	log.Trace().
		Str("component", "timeline_controller").
		Str("phase", "view").
		Int("entity_count", len(c.store.order)).
		Int("selected_index", c.selected).
		Bool("selection_visible", c.selectionVisible).
		Bool("entering", c.entering).
		Msg("render start")
	var b strings.Builder
	hits := 0
	misses := 0
	for _, id := range c.store.order {
		rec, _ := c.store.get(id)
		if rec.model != nil {
			// For interactive models, update selection/focus and render
			sel := c.selectionVisible && c.selected >= 0 && keyID(id) == keyID(c.store.order[c.selected])
			rec.model.OnProps(map[string]any{"selected": sel})
			if sel {
				rec.model.Update(EntitySelectedMsg{ID: rec.ID})
			} else {
				rec.model.Update(EntityUnselectedMsg{ID: rec.ID})
			}
			if sel && c.entering {
				rec.model.Focus()
			} else {
				rec.model.Blur()
			}
			s := rec.model.View()
			b.WriteString(s)
			b.WriteByte('\n')
			continue
		}
		r := c.pickRenderer(rec)
		// Clone props and annotate selection/focus
		annotated := cloneMap(rec.Props)
		if c.selectionVisible && c.selected >= 0 {
			// Identify current entity index by comparing keys
			if keyID(id) == keyID(c.store.order[c.selected]) {
				annotated["selected"] = true
				if rec.model != nil {
					rec.model.Focus()
				} else if c.entering {
					annotated["focused"] = true
				}
			}
		}
		ck := cacheKey{RendererKey: r.Key(), EntityKey: keyID(id), Width: c.width, Theme: c.theme, PropsHash: r.RelevantPropsHash(annotated)}
		if s, _, ok := c.cache.get(ck); ok {
			b.WriteString(s)
			b.WriteByte('\n')
			hits++
			continue
		}
		s, h, _ := r.Render(annotated, c.width, c.theme)
		if h <= 0 {
			h = lipLines(s)
		}
		c.cache.set(ck, s, h)
		b.WriteString(s)
		b.WriteByte('\n')
		misses++
	}
	out := b.String()
	log.Trace().
		Str("component", "timeline_controller").
		Str("phase", "view").
		Int("cache_hits", hits).
		Int("cache_misses", misses).
		Int("output_len", len(out)).
		Msg("render done")
	return out
}

// ViewAndSelectedPosition returns the full rendered view and the offset/height of the selected entity
func (c *Controller) ViewAndSelectedPosition() (string, int, int) {
	view := c.View()
	if c.selected < 0 || c.selected >= len(c.store.order) {
		return view, 0, 0
	}
	// naive computation: split by lines and sum heights of entities until selected
	offset := 0
	for idx, id := range c.store.order {
		rec, _ := c.store.get(id)
		// If we have an interactive model, render and measure directly
		if rec.model != nil {
			sel := idx == c.selected && c.selectionVisible
			rec.model.OnProps(map[string]any{"selected": sel})
			if sel {
				rec.model.Update(EntitySelectedMsg{ID: rec.ID})
			} else {
				rec.model.Update(EntityUnselectedMsg{ID: rec.ID})
			}
			s := rec.model.View()
			h := lipLines(s)
			if idx == c.selected {
				return view, offset, h
			}
			offset += h
			continue
		}
		// Stateless renderer path with cache
		r := c.pickRenderer(rec)
		annotated := cloneMap(rec.Props)
		if idx == c.selected {
			annotated["selected"] = true
		}
		ck := cacheKey{RendererKey: r.Key(), EntityKey: keyID(id), Width: c.width, Theme: c.theme, PropsHash: r.RelevantPropsHash(annotated)}
		if s, h, ok := c.cache.get(ck); ok {
			if h <= 0 {
				h = lipLines(s)
			}
			if idx == c.selected {
				return view, offset, h
			}
			offset += h
			continue
		}
		s, h, _ := r.Render(annotated, c.width, c.theme)
		if h <= 0 {
			h = lipLines(s)
		}
		if idx == c.selected {
			return view, offset, h
		}
		offset += h
	}
	return view, 0, 0
}

func lipLines(s string) int {
	if s == "" {
		return 0
	}
	n := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}

// HandleMsg routes a Bubble Tea message to the selected entity model when entering is true.
func (c *Controller) HandleMsg(msg tea.Msg) tea.Cmd {
	if !c.entering {
		return nil
	}
	if c.selected < 0 || c.selected >= len(c.store.order) {
		return nil
	}
	id := c.store.order[c.selected]
	rec, ok := c.store.get(id)
	if !ok || rec.model == nil {
		return nil
	}
	log.Debug().Str("component", "timeline_controller").Str("op", "handle_msg").
		Str("selected_local_id", id.LocalID).Str("msg_type", fmt.Sprintf("%T", msg)).Msg("routing msg to model")
	m2, cmd := rec.model.Update(msg)
	if mm, ok := m2.(EntityModel); ok {
		rec.model = mm
	}
	return cmd
}

// SendToSelected sends a message to the currently selected entity model regardless of entering state.
func (c *Controller) SendToSelected(msg tea.Msg) tea.Cmd {
	if c.selected < 0 || c.selected >= len(c.store.order) {
		return nil
	}
	id := c.store.order[c.selected]
	rec, ok := c.store.get(id)
	if !ok || rec.model == nil {
		return nil
	}
	m2, cmd := rec.model.Update(msg)
	if mm, ok := m2.(EntityModel); ok {
		rec.model = mm
	}
	return cmd
}

// EnterSelection toggles entering mode; when true, key events should go to selected entity
func (c *Controller) EnterSelection() {
	c.entering = true
	log.Debug().Str("component", "timeline_controller").Str("op", "enter_selection").Bool("entering", c.entering).Msg("entering selection mode")
}
func (c *Controller) ExitSelection() {
	c.entering = false
	log.Debug().Str("component", "timeline_controller").Str("op", "exit_selection").Bool("entering", c.entering).Msg("exiting selection mode")
}
func (c *Controller) IsEntering() bool { return c.entering }

// GetSelectedMeta returns ID, renderer and props of the selected entity
func (c *Controller) GetSelectedMeta() (EntityID, RendererDescriptor, map[string]any, bool) {
	if c.selected < 0 || c.selected >= len(c.store.order) {
		return EntityID{}, RendererDescriptor{}, nil, false
	}
	id := c.store.order[c.selected]
	rec, ok := c.store.get(id)
	if !ok {
		return EntityID{}, RendererDescriptor{}, nil, false
	}
	return rec.ID, rec.Renderer, cloneMap(rec.Props), true
}

// GetLastLLMByRole returns the most recent llm_text entity matching the role if present.
func (c *Controller) GetLastLLMByRole(role string) (EntityID, map[string]any, bool) {
	for i := len(c.store.order) - 1; i >= 0; i-- {
		id := c.store.order[i]
		rec, ok := c.store.get(id)
		if !ok {
			continue
		}
		if rec.Renderer.Kind == "llm_text" {
			if r, _ := rec.Props["role"].(string); r == role || role == "" {
				return rec.ID, cloneMap(rec.Props), true
			}
		}
	}
	return EntityID{}, nil, false
}

// UpdateSelected applies a patch to the selected entity props and invalidates cache
func (c *Controller) UpdateSelected(patch map[string]any) bool {
	if c.selected < 0 || c.selected >= len(c.store.order) {
		return false
	}
	id := c.store.order[c.selected]
	rec, ok := c.store.get(id)
	if !ok {
		return false
	}
	applyPatch(rec.Props, patch)
	c.cache.invalidateByID(id)
	return true
}

// Unselect clears the current selection index
func (c *Controller) Unselect() {
	if c.selected != -1 {
		log.Debug().Str("component", "timeline_controller").Str("op", "unselect").Int("previous", c.selected).Msg("clearing selection")
		c.selected = -1
	}
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
