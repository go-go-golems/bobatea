package repl

import "context"

// HelpBarReason describes why a help bar request was triggered.
type HelpBarReason string

const (
	HelpBarReasonDebounce HelpBarReason = "debounce"
	HelpBarReasonShortcut HelpBarReason = "shortcut"
	HelpBarReasonManual   HelpBarReason = "manual"
)

// HelpBarRequest captures current input context for help bar lookup.
type HelpBarRequest struct {
	Input      string
	CursorByte int
	Reason     HelpBarReason
	Shortcut   string
	RequestID  uint64
}

// HelpBarPayload describes what the REPL help bar should render.
type HelpBarPayload struct {
	Show      bool
	Text      string
	Kind      string
	Severity  string
	Ephemeral bool
}

// HelpBarProvider resolves contextual help bar text for REPL input snapshots.
type HelpBarProvider interface {
	GetHelpBar(ctx context.Context, req HelpBarRequest) (HelpBarPayload, error)
}
