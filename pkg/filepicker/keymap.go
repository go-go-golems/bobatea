package filepicker

import (
	"github.com/charmbracelet/bubbles/key"
	mode_keymap "github.com/go-go-golems/bobatea/pkg/mode-keymap"
)

type KeyMap struct {
	ResetFileInput key.Binding `keymap-mode:"new-file"`
	Accept         key.Binding `keymap-mode:"*"`
	Help           key.Binding `keymap-mode:"*"`
	Exit           key.Binding `keymap-mode:"*"`

	CreateFile key.Binding `keymap-mode:"browse"`

	// filepicker forward
	GoToTop  key.Binding `keymap-mode:"browse"`
	GoToLast key.Binding `keymap-mode:"browse"`
	Down     key.Binding `keymap-mode:"browse"`
	Up       key.Binding `keymap-mode:"browse"`
	PageUp   key.Binding `keymap-mode:"browse"`
	PageDown key.Binding `keymap-mode:"browse"`
	Back     key.Binding `keymap-mode:"browse"`
	Forward  key.Binding `keymap-mode:"browse"`

	// buttons
	NextButton  key.Binding `keymap-mode:"confirm-new,confirm-overwrite"`
	LeftButton  key.Binding `keymap-mode:"confirm-new,confirm-overwrite"`
	RightButton key.Binding `keymap-mode:"confirm-new,confirm-overwrite"`
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		ResetFileInput: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Reset file input"),
		),
		Accept: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Accept"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "Help"),
		),
		Exit: key.NewBinding(
			key.WithKeys("ctrl+g", "esc"),
			key.WithHelp("esc/ctrl+g", "Exit"),
		),

		CreateFile: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "Create file")),

		// forward to filepicker
		GoToTop:  key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "Go to top")),
		GoToLast: key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "Go to last")),
		Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("down", "Move down")),
		Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("up", "Move up")),
		PageUp:   key.NewBinding(key.WithKeys("pageup"), key.WithHelp("pageup", "Page up")),
		PageDown: key.NewBinding(key.WithKeys("pagedown"), key.WithHelp("pagedown", "Page down")),
		Back:     key.NewBinding(key.WithKeys("left"), key.WithHelp("left", "Back")),
		Forward:  key.NewBinding(key.WithKeys("right"), key.WithHelp("right", "Enter")),

		// buttons
		NextButton:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Next button")),
		LeftButton:  key.NewBinding(key.WithKeys("left"), key.WithHelp("left", "Left button")),
		RightButton: key.NewBinding(key.WithKeys("right"), key.WithHelp("right", "Right button")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Accept,
		k.CreateFile,
		k.Back,
		k.Forward,
		k.ResetFileInput,
		k.Exit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Help,
			k.Accept,
			k.ResetFileInput,
		},
		{
			k.Exit,
			k.CreateFile,
		}, {
			k.GoToTop,
			k.GoToLast,
		}, {
			k.NextButton,
			k.LeftButton,
			k.RightButton,
		},
	}
}

func (k *KeyMap) UpdateKeyBindings(state state, input string) {
	mode_keymap.EnableMode(k, string(state))

	if state == stateNewFile {
		inputActive := input != ""
		k.Accept.SetEnabled(inputActive)
		k.ResetFileInput.SetEnabled(inputActive)
	}
}
