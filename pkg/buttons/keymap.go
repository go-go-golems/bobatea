package buttons

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	LeftButton  key.Binding
	RightButton key.Binding
	Accept      key.Binding
	Exit        key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		LeftButton:  key.NewBinding(key.WithKeys("left"), key.WithHelp("left", "Left button")),
		RightButton: key.NewBinding(key.WithKeys("tab", "right"), key.WithHelp("right", "Right button")),
		Accept:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Accept")),
		Exit:        key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("esc", "Exit")),
	}
}
