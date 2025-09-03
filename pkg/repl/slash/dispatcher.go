package slash

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/bobatea/pkg/autocomplete"
)

type dispatcher struct {
	reg Registry
}

func NewDispatcher(reg Registry) Dispatcher { return &dispatcher{reg: reg} }

func (d *dispatcher) TryHandle(ctx context.Context, input string, emit Emitter) bool {
	s := strings.TrimSpace(input)
	if !strings.HasPrefix(s, "/") {
		return false
	}
	parsed, _ := parseInput(s)
	if parsed.Name == "" {
		// suggest commands
		var names []string
		for _, c := range d.reg.List() {
			names = append(names, c.Name)
		}
		sort.Strings(names)
		md := "Available commands:\n\n"
		for _, n := range names {
			md += fmt.Sprintf("- /%s\n", n)
		}
		emit("repl_result_markdown", map[string]any{"markdown": md})
		return true
	}
	cmd := d.reg.Get(parsed.Name)
	if cmd == nil {
		emit("repl_result_markdown", map[string]any{"markdown": fmt.Sprintf("Unknown command: /%s", parsed.Name)})
		return true
	}
	// TODO: schema validation minimal pass (future)
	if cmd.Run == nil {
		emit("repl_result_markdown", map[string]any{"markdown": fmt.Sprintf("No handler for /%s", parsed.Name)})
		return true
	}
	_ = cmd.Run(ctx, parsed, func(kind string, props map[string]any) { emit(kind, props) })
	return true
}

func (d *dispatcher) Complete(ctx context.Context, raw string, caret int) ([]autocomplete.Suggestion, CompletionState, error) {
	st, _ := completionState(raw, caret)
	// name completion: suggest command names filtered by current partial/name
	if st.Phase == PhaseName {
		var out []autocomplete.Suggestion
		partial := st.Partial
		if after, ok := strings.CutPrefix(partial, "/"); ok {
			partial = after
		}
		if partial == "" {
			partial = st.Name
		}
		for _, c := range d.reg.List() {
			if partial == "" || strings.HasPrefix(c.Name, partial) {
				out = append(out, autocomplete.Suggestion{Id: c.Name, Value: c.Name, DisplayText: "/" + c.Name})
			}
		}
		if len(out) > 0 {
			return out, st, nil
		}
		// fall through if none matched (e.g., full name typed) to potential command-specific completion
	}
	// delegate to command completer if available
	cmd := d.reg.Get(st.Name)
	if cmd != nil && cmd.Complete != nil {
		out, err := cmd.Complete(ctx, st)
		return out, st, err
	}
	return nil, st, nil
}
