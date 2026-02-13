package repl

import "context"

// HelpDrawerTrigger describes why a help drawer request was triggered.
type HelpDrawerTrigger string

const (
	HelpDrawerTriggerToggleOpen    HelpDrawerTrigger = "toggle-open"
	HelpDrawerTriggerTyping        HelpDrawerTrigger = "typing"
	HelpDrawerTriggerManualRefresh HelpDrawerTrigger = "manual-refresh"
)

// HelpDrawerRequest captures current input context for rich help drawer lookup.
type HelpDrawerRequest struct {
	Input      string
	CursorByte int
	RequestID  uint64
	Trigger    HelpDrawerTrigger
}

// HelpDrawerDocument describes what the REPL help drawer should render.
type HelpDrawerDocument struct {
	Show        bool
	Title       string
	Subtitle    string
	Markdown    string
	Diagnostics []string
	VersionTag  string
}

// HelpDrawerProvider resolves contextual rich help for REPL input snapshots.
type HelpDrawerProvider interface {
	GetHelpDrawer(ctx context.Context, req HelpDrawerRequest) (HelpDrawerDocument, error)
}
