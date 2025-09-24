package diff

import (
    "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
    Up, Down key.Binding
    PageUp, PageDown key.Binding
    FocusNext key.Binding
    ToggleRedact key.Binding
}

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
        PageUp: key.NewBinding(
            key.WithKeys("pgup", "b"),
            key.WithHelp("pgup/b", "page up"),
        ),
        PageDown: key.NewBinding(
            key.WithKeys("pgdown", "f"),
            key.WithHelp("pgdn/f", "page down"),
        ),
        FocusNext: key.NewBinding(
            key.WithKeys("tab"),
            key.WithHelp("tab", "switch panel"),
        ),
        ToggleRedact: key.NewBinding(
            key.WithKeys("r"),
            key.WithHelp("r", "toggle redaction"),
        ),
    }
}


