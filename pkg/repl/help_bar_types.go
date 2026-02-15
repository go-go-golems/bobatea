package repl

import (
	"context"

	"github.com/go-go-golems/bobatea/pkg/tui/widgets/contextbar"
)

// HelpBarReason describes why a help bar request was triggered.
type HelpBarReason = contextbar.Reason

const (
	HelpBarReasonDebounce HelpBarReason = contextbar.ReasonDebounce
	HelpBarReasonShortcut HelpBarReason = contextbar.ReasonShortcut
	HelpBarReasonManual   HelpBarReason = contextbar.ReasonManual
)

// HelpBarRequest captures current input context for help bar lookup.
type HelpBarRequest = contextbar.Request

// HelpBarPayload describes what the REPL help bar should render.
type HelpBarPayload = contextbar.Payload

// HelpBarProvider resolves contextual help bar text for REPL input snapshots.
type HelpBarProvider interface {
	GetHelpBar(ctx context.Context, req HelpBarRequest) (HelpBarPayload, error)
}
