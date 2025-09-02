package repl

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/timeline"
)

// replEventMsg is the envelope published by the REPL model.
type replEventMsg struct {
	TurnID string    `json:"turn_id"`
	Event  Event     `json:"event"`
	Time   time.Time `json:"time"`
}

// uiEnvelope wraps timeline messages for the ui.entities topic.
type uiEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// RegisterReplToTimelineTransformer subscribes to repl.events and publishes timeline lifecycle messages to ui.entities.
func RegisterReplToTimelineTransformer(bus *eventbus.Bus) {
	// minimal state to avoid duplicate stdout/stderr create per turn
	var mu sync.Mutex
	type turnState struct {
		stdout, stderr bool
		seq            int
	}
	byTurn := map[string]*turnState{}

	publish := func(msgType string, v any) error {
		b, _ := json.Marshal(v)
		env, _ := json.Marshal(uiEnvelope{Type: msgType, Payload: b})
		return bus.Publisher.Publish(eventbus.TopicUIEntities, message.NewMessage(watermill.NewUUID(), env))
	}

	bus.AddHandler("repl-to-ui", eventbus.TopicReplEvents, func(msg *message.Message) error {
		defer msg.Ack()
		var in replEventMsg
		if err := json.Unmarshal(msg.Payload, &in); err != nil {
			return err
		}
		turnID := in.TurnID
		mu.Lock()
		st := byTurn[turnID]
		if st == nil {
			st = &turnState{}
			byTurn[turnID] = st
		}
		mu.Unlock()

		//nolint:exhaustive
		switch in.Event.Kind {
		case EventInput:
			e := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: "input", Kind: "markdown"}, Renderer: timeline.RendererDescriptor{Kind: "markdown"}, Props: in.Event.Props, StartedAt: time.Now()}
			return publish("timeline.created", e)
		case EventStdout:
			mu.Lock()
			needCreate := !st.stdout
			if needCreate {
				st.stdout = true
			}
			mu.Unlock()
			if needCreate {
				c := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: "stdout", Kind: "text"}, Renderer: timeline.RendererDescriptor{Kind: "text"}, Props: map[string]any{"text": "", "streaming": true}, StartedAt: time.Now()}
				if err := publish("timeline.created", c); err != nil {
					return err
				}
			}
			u := timeline.UIEntityUpdated{ID: timeline.EntityID{TurnID: turnID, LocalID: "stdout", Kind: "text"}, Patch: ensureAppendPatch(in.Event.Props), Version: time.Now().UnixNano(), UpdatedAt: time.Now()}
			return publish("timeline.updated", u)
		case EventStderr:
			mu.Lock()
			needCreate := !st.stderr
			if needCreate {
				st.stderr = true
			}
			mu.Unlock()
			if needCreate {
				c := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: "stderr", Kind: "text"}, Renderer: timeline.RendererDescriptor{Kind: "text"}, Props: map[string]any{"text": "", "streaming": true, "is_error": true}, StartedAt: time.Now()}
				if err := publish("timeline.created", c); err != nil {
					return err
				}
			}
			p := ensureAppendPatch(in.Event.Props)
			p["is_error"] = true
			u := timeline.UIEntityUpdated{ID: timeline.EntityID{TurnID: turnID, LocalID: "stderr", Kind: "text"}, Patch: p, Version: time.Now().UnixNano(), UpdatedAt: time.Now()}
			return publish("timeline.updated", u)
		case EventResultMarkdown:
			mu.Lock()
			st.seq++
			local := fmt.Sprintf("result-%d", st.seq)
			mu.Unlock()
			c := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "markdown"}, Renderer: timeline.RendererDescriptor{Kind: "markdown"}, Props: in.Event.Props, StartedAt: time.Now()}
			if err := publish("timeline.created", c); err != nil {
				return err
			}
			return publish("timeline.completed", timeline.UIEntityCompleted{ID: c.ID, Result: nil})
		case EventLog:
			mu.Lock()
			st.seq++
			local := fmt.Sprintf("log-%d", st.seq)
			mu.Unlock()
			c := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "log_event"}, Renderer: timeline.RendererDescriptor{Kind: "log_event"}, Props: in.Event.Props, StartedAt: time.Now()}
			if err := publish("timeline.created", c); err != nil {
				return err
			}
			return publish("timeline.completed", timeline.UIEntityCompleted{ID: c.ID, Result: nil})
		case EventStructuredLog:
			mu.Lock()
			st.seq++
			local := fmt.Sprintf("slog-%d", st.seq)
			mu.Unlock()
			c := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "structured_log_event"}, Renderer: timeline.RendererDescriptor{Kind: "structured_log_event"}, Props: in.Event.Props, StartedAt: time.Now()}
			if err := publish("timeline.created", c); err != nil {
				return err
			}
			return publish("timeline.completed", timeline.UIEntityCompleted{ID: c.ID, Result: nil})
		case EventInspector:
			mu.Lock()
			st.seq++
			local := fmt.Sprintf("inspect-%d", st.seq)
			mu.Unlock()
			c := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "structured_data"}, Renderer: timeline.RendererDescriptor{Kind: "structured_data"}, Props: in.Event.Props, StartedAt: time.Now()}
			if err := publish("timeline.created", c); err != nil {
				return err
			}
			return publish("timeline.completed", timeline.UIEntityCompleted{ID: c.ID, Result: nil})
		default:
			mu.Lock()
			st.seq++
			local := fmt.Sprintf("event-%d", st.seq)
			mu.Unlock()
			c := timeline.UIEntityCreated{ID: timeline.EntityID{TurnID: turnID, LocalID: local, Kind: "text"}, Renderer: timeline.RendererDescriptor{Kind: "text"}, Props: in.Event.Props, StartedAt: time.Now()}
			if err := publish("timeline.created", c); err != nil {
				return err
			}
			return publish("timeline.completed", timeline.UIEntityCompleted{ID: c.ID, Result: nil})
		}
	})
}
