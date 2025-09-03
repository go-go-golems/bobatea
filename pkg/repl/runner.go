package repl

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/timeline"
)

// NewTimelineRepl wires a timeline-first REPL with an in-memory Watermill bus,
// registers the REPL->timeline transformer, and the UI forwarder.
// It returns the bus, model and Bubble Tea program (not running yet).
func NewTimelineRepl(evaluator Evaluator, config Config) (*eventbus.Bus, *Model, *tea.Program, error) {
	bus, err := eventbus.NewInMemoryBus()
	if err != nil {
		return nil, nil, nil, err
	}
	RegisterReplToTimelineTransformer(bus)
	model := NewModel(evaluator, config, bus.Publisher)
	p := tea.NewProgram(model, tea.WithAltScreen())
	timeline.RegisterUIForwarder(bus, p)
	return bus, model, p, nil
}

// RunTimelineRepl constructs the timeline REPL and runs both the Watermill router
// and the Bubble Tea program until either exits. It returns any terminal error.
func RunTimelineRepl(evaluator Evaluator, config Config) error {
	bus, _, p, err := NewTimelineRepl(evaluator, config)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 2)
	go func() { errs <- bus.Run(ctx) }()
	go func() { _, e := p.Run(); cancel(); errs <- e }()
	if e := <-errs; e != nil {
		return e
	}
	return nil
}
