package filepicker

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	ResetFileInput key.Binding
	Accept         key.Binding
	Help           key.Binding
	Exit           key.Binding

	CreateFile key.Binding

	// filepicker forward
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Forward  key.Binding

	// buttons
	NextButton  key.Binding
	LeftButton  key.Binding
	RightButton key.Binding
}

func (k *KeyMap) ForEach(f func(b *key.Binding)) {
	v := reflect.ValueOf(k).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Struct &&
			field.Type().Name() == "Binding" &&
			field.Type().PkgPath() == "github.com/charmbracelet/bubbles/key" {
			if addr, ok := field.Addr().Interface().(*key.Binding); ok {
				f(addr)
			}
			return
		}
		if field.Kind() == reflect.Ptr &&
			field.Type().Elem().Name() == "Binding" &&
			field.Type().Elem().PkgPath() == "github.com/charmbracelet/bubbles/key" {
			if addr, ok := field.Interface().(*key.Binding); ok {
				f(addr)
			}
			return
		}
	}
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

func (k KeyMap) DisableAll() {
	k.ForEach(func(b *key.Binding) {
		b.SetEnabled(false)
	})
}

func (k *KeyMap) UpdateKeyBindings(state state, input string) {
	k.DisableAll()

	switch state {
	case stateBrowse:
		for _, k := range []key.Binding{
			k.Accept, k.Back, k.Help, k.Exit, k.CreateFile,
			k.GoToTop, k.GoToLast, k.Down, k.Up, k.PageUp, k.PageDown, k.Back, k.Forward,
		} {
			k.SetEnabled(true)
		}

	case stateNewFile:
		for _, k := range []key.Binding{
			k.Help, k.Exit,
		} {
			k.SetEnabled(true)
		}
		if input != "" {
			k.Accept.SetEnabled(true)
			k.ResetFileInput.SetEnabled(true)
		}

	case stateConfirmNew:
		for _, k := range []key.Binding{
			k.Accept, k.NextButton, k.LeftButton, k.RightButton, k.Help, k.Exit,
		} {
			k.SetEnabled(true)
		}
	}
}
