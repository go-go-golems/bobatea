package chat

// These are the bubbletea messages for controlling the bubbletea-based bobatea chat widget.
//
// These UserActionMsg types represent different actions that can be triggered within the chat interface.
// While typically invoked through keyboard shortcuts in the TUI, these messages can also be sent
// programmatically, for example via HTTP requests to enable remote control of the chat application.
//
// Common actions include:
// - Text input manipulation (append, prepend, replace)
// - Navigation (focus/unfocus, select next/previous)
// - Clipboard operations (copy responses, source blocks)
// - Application control (quit, toggle help, cancel completion)

// New message types for user actions
type UserActionMsg interface {
	isUserAction()
}

type ToggleHelpMsg struct{}
type UnfocusMessageMsg struct{}
type QuitMsg struct{}
type FocusMessageMsg struct{}
type SelectNextMessageMsg struct{}
type SelectPrevMessageMsg struct{}
type SubmitMessageMsg struct{}
type StartBackendMsg struct{}
type CopyToClipboardMsg struct{}
type CopyLastResponseToClipboardMsg struct{}
type CopyLastSourceBlocksToClipboardMsg struct{}
type CopySourceBlocksToClipboardMsg struct{}
type SaveToFileMsg struct{}
type CancelCompletionMsg struct{}
type DismissErrorMsg struct{}
type InputTextMsg struct {
	Text string
}

// Add these new message types
type ReplaceInputTextMsg struct {
	Text string
}

type AppendInputTextMsg struct {
	Text string
}

type PrependInputTextMsg struct {
	Text string
}

type GetInputTextMsg struct{}

// Demo tool triggers
type TriggerWeatherToolMsg struct{}
type TriggerWebSearchToolMsg struct{}

func (ToggleHelpMsg) isUserAction()                      {}
func (UnfocusMessageMsg) isUserAction()                  {}
func (QuitMsg) isUserAction()                            {}
func (FocusMessageMsg) isUserAction()                    {}
func (SelectNextMessageMsg) isUserAction()               {}
func (SelectPrevMessageMsg) isUserAction()               {}
func (SubmitMessageMsg) isUserAction()                   {}
func (CopyToClipboardMsg) isUserAction()                 {}
func (CopyLastResponseToClipboardMsg) isUserAction()     {}
func (CopyLastSourceBlocksToClipboardMsg) isUserAction() {}
func (CopySourceBlocksToClipboardMsg) isUserAction()     {}
func (SaveToFileMsg) isUserAction()                      {}
func (CancelCompletionMsg) isUserAction()                {}
func (DismissErrorMsg) isUserAction()                    {}
func (InputTextMsg) isUserAction()                       {}

// Add isUserAction() methods for the new types
func (ReplaceInputTextMsg) isUserAction()     {}
func (AppendInputTextMsg) isUserAction()      {}
func (PrependInputTextMsg) isUserAction()     {}
func (GetInputTextMsg) isUserAction()         {}
func (TriggerWeatherToolMsg) isUserAction()   {}
func (TriggerWebSearchToolMsg) isUserAction() {}
