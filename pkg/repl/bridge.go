package repl

import (
	"time"

	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/rs/zerolog/log"
)

// Bridge maps Evaluator events to timeline lifecycle messages via the Shell wrapper
// to ensure viewport refresh semantics are consistent.
type Bridge struct {
	sh    *timeline.Shell
	clock func() time.Time
}

func NewBridge(sh *timeline.Shell) *Bridge {
	return &Bridge{sh: sh, clock: time.Now}
}

func (b *Bridge) NewTurnID(seq int) string {
	return time.Now().Format("20060102-150405.000000000") + ":" + itoa(seq)
}

func (b *Bridge) Emit(turnID, local, kind string, props map[string]any) timeline.EntityID {
	id := timeline.EntityID{TurnID: turnID, LocalID: local, Kind: kind}
	start := time.Now()
	b.sh.OnCreated(timeline.UIEntityCreated{
		ID:        id,
		Renderer:  timeline.RendererDescriptor{Kind: kind},
		Props:     props,
		StartedAt: b.clock(),
	})
	dur := time.Since(start)
	log.Debug().Str("op", "emit").Str("turn", turnID).Str("local", local).Str("kind", kind).Dur("dur", dur).Msg("timeline entity created and view refreshed")
	return id
}

func (b *Bridge) Patch(id timeline.EntityID, patch map[string]any) {
	start := time.Now()
	b.sh.OnUpdated(timeline.UIEntityUpdated{ID: id, Patch: patch, Version: time.Now().UnixNano(), UpdatedAt: b.clock()})
	dur := time.Since(start)
	log.Debug().Str("op", "patch").Str("kind", id.Kind).Str("turn", id.TurnID).Str("local", id.LocalID).Dur("dur", dur).Int("patch_keys", len(patch)).Msg("timeline entity patched and view refreshed")
}

func (b *Bridge) Complete(id timeline.EntityID, result map[string]any) {
	start := time.Now()
	b.sh.OnCompleted(timeline.UIEntityCompleted{ID: id, Result: result})
	dur := time.Since(start)
	log.Debug().Str("op", "complete").Str("kind", id.Kind).Str("turn", id.TurnID).Str("local", id.LocalID).Dur("dur", dur).Msg("timeline entity completed and view refreshed")
}

func itoa(i int) string {
	// small, allocation-free int to string for small numbers
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var buf [20]byte
	n := len(buf)
	for i > 0 {
		n--
		buf[n] = byte('0' + (i % 10))
		i /= 10
	}
	if neg {
		n--
		buf[n] = '-'
	}
	return string(buf[n:])
}
