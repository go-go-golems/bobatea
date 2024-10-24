package chat

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
func (ReplaceInputTextMsg) isUserAction() {}
func (AppendInputTextMsg) isUserAction()  {}
func (PrependInputTextMsg) isUserAction() {}
func (GetInputTextMsg) isUserAction()     {}
