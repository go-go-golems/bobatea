package filepicker

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	ResetFileInput key.Binding
	Accept         key.Binding
	Help           key.Binding
	Exit           key.Binding

	// filepicker forward
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding

	// buttons
	NextButton  key.Binding
	LeftButton  key.Binding
	RightButton key.Binding
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
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "Exit"),
		),

		// forward to filepicker
		GoToTop:  key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "Go to top")),
		GoToLast: key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "Go to last")),
		Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("down", "Move down")),
		Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("up", "Move up")),
		PageUp:   key.NewBinding(key.WithKeys("pageup"), key.WithHelp("pageup", "Page up")),
		PageDown: key.NewBinding(key.WithKeys("pagedown"), key.WithHelp("pagedown", "Page down")),
		Back:     key.NewBinding(key.WithKeys("left"), key.WithHelp("left", "Back")),

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
		k.GoToTop,
		k.GoToLast,
		k.Back,
		k.ResetFileInput,
		k.Accept,
		k.Exit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Help,
			k.Accept,
			k.GoToTop,
			k.GoToLast,
			k.Back,
			k.ResetFileInput,
			k.Accept,
			k.Exit,
		},
	}
}

func (k *KeyMap) UpdateKeyBindings(state state, input string) {
	k.ResetFileInput.SetEnabled(state == stateList && input != "")
	k.Accept.SetEnabled(state == stateList)
	k.Help.SetEnabled(true)
	k.Exit.SetEnabled(true)

	// filepicker forward
	k.GoToTop.SetEnabled(state == stateList)
	k.GoToLast.SetEnabled(state == stateList)
	k.Down.SetEnabled(state == stateList)
	k.Up.SetEnabled(state == stateList)
	k.PageUp.SetEnabled(state == stateList)
	k.PageDown.SetEnabled(state == stateList)
	k.Back.SetEnabled(state == stateList)

	// buttons
	k.NextButton.SetEnabled(state == stateConfirm)
	k.LeftButton.SetEnabled(state == stateConfirm)
	k.RightButton.SetEnabled(state == stateConfirm)
}
