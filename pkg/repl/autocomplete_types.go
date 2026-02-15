package repl

import (
	"context"

	"github.com/go-go-golems/bobatea/pkg/tui/widgets/suggest"
)

// CompletionReason describes why a completion request was triggered.
type CompletionReason = suggest.Reason

const (
	CompletionReasonDebounce CompletionReason = suggest.ReasonDebounce
	CompletionReasonShortcut CompletionReason = suggest.ReasonShortcut
	CompletionReasonManual   CompletionReason = suggest.ReasonManual
)

// CompletionRequest captures the current input context for suggestion lookup.
type CompletionRequest = suggest.Request

// CompletionResult represents what should be shown and how to apply it.
type CompletionResult = suggest.Result

// InputCompleter resolves completion suggestions for REPL input snapshots.
type InputCompleter interface {
	CompleteInput(ctx context.Context, req CompletionRequest) (CompletionResult, error)
}
