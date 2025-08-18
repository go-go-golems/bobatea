package timeline

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
)

type entityRecord struct {
	ID        EntityID
	Renderer  RendererDescriptor
	Props     map[string]any
	StartedAt int64
	UpdatedAt int64
	Version   int64
	Completed bool
	model     EntityModel
}

type entityStore struct {
	order []EntityID
	byID  map[string]*entityRecord // key is JSON of EntityID for stability
}

func newEntityStore() *entityStore {
	log.Debug().Str("component", "timeline_store").Msg("initialized store")
	return &entityStore{byID: map[string]*entityRecord{}}
}

func keyID(id EntityID) string {
	b, _ := json.Marshal(id)
	return string(b)
}

func (s *entityStore) get(id EntityID) (*entityRecord, bool) {
	rec, ok := s.byID[keyID(id)]
	return rec, ok
}

func (s *entityStore) add(rec *entityRecord) {
	k := keyID(rec.ID)
	if _, exists := s.byID[k]; exists {
		log.Debug().Str("component", "timeline_store").Str("op", "add").Str("key", k).Msg("already exists")
		return
	}
	s.byID[k] = rec
	s.order = append(s.order, rec.ID)
	log.Debug().Str("component", "timeline_store").Str("op", "add").Str("key", k).Int("count", len(s.order)).Msg("added")
}

func (s *entityStore) remove(id EntityID) {
	k := keyID(id)
	if _, ok := s.byID[k]; !ok {
		return
	}
	delete(s.byID, k)
	// remove from order (small N typical)
	out := s.order[:0]
	for _, e := range s.order {
		if keyID(e) != k {
			out = append(out, e)
		}
	}
	s.order = out
	log.Debug().Str("component", "timeline_store").Str("op", "remove").Str("key", k).Int("count", len(s.order)).Msg("removed")
}
