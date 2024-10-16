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
