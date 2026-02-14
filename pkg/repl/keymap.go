package repl

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines REPL key bindings with mode tags for help rendering.
type KeyMap struct {
	Quit       key.Binding `keymap-mode:"*"`
	ToggleHelp key.Binding `keymap-mode:"*"`

	ToggleFocus key.Binding `keymap-mode:"*"`

	Submit      key.Binding `keymap-mode:"input"`
	HistoryPrev key.Binding `keymap-mode:"input"`
	HistoryNext key.Binding `keymap-mode:"input"`

	CompletionTrigger   key.Binding `keymap-mode:"input"`
	CompletionAccept    key.Binding `keymap-mode:"input"`
	CompletionPrev      key.Binding `keymap-mode:"input"`
	CompletionNext      key.Binding `keymap-mode:"input"`
	CompletionPageUp    key.Binding `keymap-mode:"input"`
	CompletionPageDown  key.Binding `keymap-mode:"input"`
	CompletionCancel    key.Binding `keymap-mode:"input"`
	HelpDrawerToggle    key.Binding `keymap-mode:"input"`
	HelpDrawerClose     key.Binding `keymap-mode:"input"`
	HelpDrawerRefresh   key.Binding `keymap-mode:"input"`
	HelpDrawerPin       key.Binding `keymap-mode:"input"`
	CommandPaletteOpen  key.Binding `keymap-mode:"input"`
	CommandPaletteClose key.Binding `keymap-mode:"input"`

	TimelinePrev      key.Binding `keymap-mode:"timeline"`
	TimelineNext      key.Binding `keymap-mode:"timeline"`
	TimelineEnterExit key.Binding `keymap-mode:"timeline"`
	CopyCode          key.Binding `keymap-mode:"timeline"`
	CopyText          key.Binding `keymap-mode:"timeline"`
}

// NewKeyMap returns REPL key bindings derived from config.
func NewKeyMap(
	autocompleteCfg AutocompleteConfig,
	helpDrawerCfg HelpDrawerConfig,
	commandPaletteCfg CommandPaletteConfig,
	focusToggleKey string,
) KeyMap {
	km := KeyMap{
		Quit: binding([]string{"ctrl+c", "alt+q"}, "quit"),
		ToggleHelp: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "help"),
		),
		ToggleFocus: binding([]string{focusToggleKey}, "toggle focus"),

		Submit:      binding([]string{"enter"}, "submit"),
		HistoryPrev: binding([]string{"up"}, "history prev"),
		HistoryNext: binding([]string{"down"}, "history next"),

		CompletionTrigger:   binding(autocompleteCfg.TriggerKeys, "trigger completion"),
		CompletionAccept:    binding(autocompleteCfg.AcceptKeys, "accept completion"),
		CompletionPrev:      binding([]string{"up", "ctrl+p"}, "completion prev"),
		CompletionNext:      binding([]string{"down", "ctrl+n"}, "completion next"),
		CompletionPageUp:    binding([]string{"pgup", "ctrl+b"}, "completion page up"),
		CompletionPageDown:  binding([]string{"pgdown", "ctrl+f"}, "completion page down"),
		CompletionCancel:    binding([]string{"esc"}, "close completion"),
		HelpDrawerToggle:    binding(helpDrawerCfg.ToggleKeys, "toggle drawer"),
		HelpDrawerClose:     binding(helpDrawerCfg.CloseKeys, "close drawer"),
		HelpDrawerRefresh:   binding(helpDrawerCfg.RefreshShortcuts, "refresh drawer"),
		HelpDrawerPin:       binding(helpDrawerCfg.PinShortcuts, "pin drawer"),
		CommandPaletteOpen:  binding(commandPaletteCfg.OpenKeys, "open palette"),
		CommandPaletteClose: binding(commandPaletteCfg.CloseKeys, "close palette"),

		TimelinePrev:      binding([]string{"up"}, "select prev"),
		TimelineNext:      binding([]string{"down"}, "select next"),
		TimelineEnterExit: binding([]string{"enter"}, "enter/exit item"),
		CopyCode:          binding([]string{"c"}, "copy code"),
		CopyText:          binding([]string{"y"}, "copy text"),
	}

	if len(autocompleteCfg.TriggerKeys) == 0 {
		km.CompletionTrigger.SetEnabled(false)
	}
	if len(autocompleteCfg.AcceptKeys) == 0 {
		km.CompletionAccept.SetEnabled(false)
	}
	if len(helpDrawerCfg.ToggleKeys) == 0 {
		km.HelpDrawerToggle.SetEnabled(false)
	}
	if len(helpDrawerCfg.CloseKeys) == 0 {
		km.HelpDrawerClose.SetEnabled(false)
	}
	if len(helpDrawerCfg.RefreshShortcuts) == 0 {
		km.HelpDrawerRefresh.SetEnabled(false)
	}
	if len(helpDrawerCfg.PinShortcuts) == 0 {
		km.HelpDrawerPin.SetEnabled(false)
	}
	if !commandPaletteCfg.Enabled || len(commandPaletteCfg.OpenKeys) == 0 {
		km.CommandPaletteOpen.SetEnabled(false)
	}
	if !commandPaletteCfg.Enabled || len(commandPaletteCfg.CloseKeys) == 0 {
		km.CommandPaletteClose.SetEnabled(false)
	}

	return km
}

func binding(keys []string, desc string) key.Binding {
	cleaned := make([]string, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		cleaned = append(cleaned, k)
	}
	if len(cleaned) == 0 {
		b := key.NewBinding()
		b.SetEnabled(false)
		return b
	}
	return key.NewBinding(
		key.WithKeys(cleaned...),
		key.WithHelp(cleaned[0], desc),
	)
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ToggleHelp,
		k.ToggleFocus,
		k.CompletionTrigger,
		k.CompletionAccept,
		k.CompletionCancel,
		k.HelpDrawerToggle,
		k.HelpDrawerPin,
		k.HelpDrawerRefresh,
		k.CommandPaletteOpen,
		k.CompletionPageUp,
		k.CompletionPageDown,
		k.Submit,
		k.HistoryPrev,
		k.HistoryNext,
		k.TimelinePrev,
		k.TimelineNext,
		k.CopyCode,
		k.CopyText,
		k.Quit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ToggleHelp, k.Quit},
		{k.ToggleFocus, k.Submit},
		{k.HistoryPrev, k.HistoryNext},
		{k.CompletionTrigger, k.CompletionAccept, k.CompletionCancel},
		{k.HelpDrawerToggle, k.HelpDrawerClose, k.HelpDrawerRefresh, k.HelpDrawerPin},
		{k.CommandPaletteOpen, k.CommandPaletteClose},
		{k.CompletionPrev, k.CompletionNext, k.CompletionPageUp, k.CompletionPageDown},
		{k.TimelinePrev, k.TimelineNext, k.TimelineEnterExit},
		{k.CopyCode, k.CopyText},
	}
}
