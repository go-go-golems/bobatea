package repl

import (
	"context"

	"github.com/go-go-golems/bobatea/pkg/tui/widgets/contextpanel"
)

// HelpDrawerTrigger describes why a help drawer request was triggered.
type HelpDrawerTrigger = contextpanel.Trigger

const (
	HelpDrawerTriggerToggleOpen    HelpDrawerTrigger = contextpanel.TriggerToggleOpen
	HelpDrawerTriggerTyping        HelpDrawerTrigger = contextpanel.TriggerTyping
	HelpDrawerTriggerManualRefresh HelpDrawerTrigger = contextpanel.TriggerManualRefresh
)

// HelpDrawerRequest captures current input context for rich help drawer lookup.
type HelpDrawerRequest = contextpanel.Request

// HelpDrawerDocument describes what the REPL help drawer should render.
type HelpDrawerDocument = contextpanel.Document

// HelpDrawerProvider resolves contextual rich help for REPL input snapshots.
type HelpDrawerProvider interface {
	GetHelpDrawer(ctx context.Context, req HelpDrawerRequest) (HelpDrawerDocument, error)
}
