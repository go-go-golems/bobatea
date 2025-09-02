package timeline

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
)

// RegisterUIForwarder adds a handler that forwards ui.entities to a tea.Program as Bubble Tea messages.
func RegisterUIForwarder(bus *eventbus.Bus, p *tea.Program) {
	bus.AddHandler("ui-forward", eventbus.TopicUIEntities, func(msg *message.Message) error {
		defer msg.Ack()
		var env struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(msg.Payload, &env); err != nil {
			return err
		}
		switch env.Type {
		case "timeline.created":
			var e UIEntityCreated
			if err := json.Unmarshal(env.Payload, &e); err == nil {
				p.Send(e)
			}
		case "timeline.updated":
			var e UIEntityUpdated
			if err := json.Unmarshal(env.Payload, &e); err == nil {
				p.Send(e)
			}
		case "timeline.completed":
			var e UIEntityCompleted
			if err := json.Unmarshal(env.Payload, &e); err == nil {
				p.Send(e)
			}
		case "timeline.deleted":
			var e UIEntityDeleted
			if err := json.Unmarshal(env.Payload, &e); err == nil {
				p.Send(e)
			}
		}
		return nil
	})
}
