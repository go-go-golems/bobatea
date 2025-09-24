package diff

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines the key bindings for the diff component
type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	PageUp        key.Binding
	PageDown      key.Binding
	Tab           key.Binding
	Search        key.Binding
	Escape        key.Binding
	ToggleRedact  key.Binding
	FilterAdded   key.Binding
	FilterRemoved key.Binding
	FilterUpdated key.Binding
	Quit          key.Binding
}

// newKeyMap returns a new keyMap with default bindings
func newKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("pgdn/f", "page down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		ToggleRedact: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "toggle redaction"),
		),
		FilterAdded: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "toggle added"),
		),
		FilterRemoved: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "toggle removed"),
		),
		FilterUpdated: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "toggle updated"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Tab, k.Search, k.ToggleRedact, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.PageUp, k.PageDown, k.Tab, k.Search},
		{k.ToggleRedact, k.FilterAdded, k.FilterRemoved, k.FilterUpdated},
		{k.Escape, k.Quit},
	}
}
