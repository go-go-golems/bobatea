package chat

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	SelectPrevMessage key.Binding `keymap-mode:"moving-around"`
	SelectNextMessage key.Binding `keymap-mode:"moving-around"`
	UnfocusMessage    key.Binding `keymap-mode:"user-input"`
	FocusMessage      key.Binding `keymap-mode:"moving-around"`

	SubmitMessage key.Binding `keymap-mode:"user-input"`
	ScrollUp      key.Binding
	ScrollDown    key.Binding

	CancelCompletion key.Binding `keymap-mode:"stream-completion"`
	DismissError     key.Binding `keymap-mode:"error"`

	LoadFromFile key.Binding

	Regenerate         key.Binding `keymap-mode:"user-input"`
	RegenerateFromHere key.Binding `keymap-mode:"moving-around"`
	EditMessage        key.Binding `keymap-mode:"moving-around"`

	PreviousConversationThread key.Binding `keymap-mode:"moving-around"`
	NextConversationThread     key.Binding `keymap-mode:"moving-around"`

	SaveToFile             key.Binding `keymap-mode:"*"`
	SaveSourceBlocksToFile key.Binding `keymap-mode:"*"`

	CopyToClipboard             key.Binding `keymap-mode:"moving-around"`
	CopyLastResponseToClipboard key.Binding `keymap-mode:"user-input"`
	CopySourceBlocksToClipboard key.Binding `keymap-mode:"moving-around"`

	Help key.Binding `keymap-mode:"*"`
	Quit key.Binding `keymap-mode:"*"`
}

var DefaultKeyMap = KeyMap{
	SelectPrevMessage: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up")),
	SelectNextMessage: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),

	UnfocusMessage: key.NewBinding(
		key.WithKeys("esc", "ctrl+g"),
		key.WithHelp("esc", "unfocus message"),
	),
	FocusMessage: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "focus message"),
	),

	SubmitMessage: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "submit message"),
	),
	CancelCompletion: key.NewBinding(
		key.WithKeys("esc", "ctrl+g"),
		key.WithHelp("esc", "cancel completion"),
	),

	DismissError: key.NewBinding(
		key.WithKeys("esc", "ctrl+g"),
		key.WithHelp("esc", "dismiss error"),
	),

	SaveToFile: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save to file"),
	),
	SaveSourceBlocksToFile: key.NewBinding(
		key.WithKeys("alt+s"),
		key.WithHelp("alt+s", "save source"),
	),

	CopyToClipboard: key.NewBinding(
		key.WithKeys("alt+c"),
		key.WithHelp("alt+c", "copy selected"),
	),
	CopyLastResponseToClipboard: key.NewBinding(
		key.WithKeys("alt+l"),
		key.WithHelp("alt+l", "copy response"),
	),
	CopySourceBlocksToClipboard: key.NewBinding(
		key.WithKeys("alt+d"),
		key.WithHelp("alt+d", "copy selected source"),
	),

	ScrollUp: key.NewBinding(
		key.WithKeys("shift+pgup"),
		key.WithHelp("shift+pgup", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("shift+pgdown"),
		key.WithHelp("shift+pgdown", "scroll down"),
	),

	Quit: key.NewBinding(
		key.WithKeys("alt+q"),
		key.WithHelp("alt+q", "quit"),
	),

	Help: key.NewBinding(
		key.WithKeys("ctrl-?"),
		key.WithHelp("ctrl-?", "help")),

	PreviousConversationThread: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("left", "previous conversation thread"),
	),

	NextConversationThread: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("right", "next conversation thread"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help,
		k.SubmitMessage,
		k.CopyLastResponseToClipboard,
		k.CopyToClipboard,
		k.CopySourceBlocksToClipboard,
		k.FocusMessage,
		k.DismissError,
		k.CancelCompletion,
		k.SelectPrevMessage,
		k.SelectNextMessage,
		k.SaveToFile,
		k.Help,
		k.Quit,
	}
}
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.SelectPrevMessage, k.SelectNextMessage},
		{k.UnfocusMessage, k.FocusMessage},
		{k.CopyLastResponseToClipboard, k.CopyToClipboard},
		{k.CopySourceBlocksToClipboard},
	}
}
