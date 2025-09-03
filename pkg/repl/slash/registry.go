package slash

import "sync"

type registry struct {
	mu       sync.RWMutex
	commands map[string]*Command
}

func NewRegistry() Registry {
	return &registry{commands: map[string]*Command{}}
}

func (r *registry) Register(cmd *Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name] = cmd
	return nil
}

func (r *registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.commands, name)
}

func (r *registry) Get(name string) *Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.commands[name]
}

func (r *registry) List() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Command, 0, len(r.commands))
	for _, c := range r.commands {
		out = append(out, c)
	}
	return out
}
