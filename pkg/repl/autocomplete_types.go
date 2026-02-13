package repl

import (
	"context"

	"github.com/go-go-golems/bobatea/pkg/autocomplete"
)

// CompletionReason describes why a completion request was triggered.
type CompletionReason string

const (
	CompletionReasonDebounce CompletionReason = "debounce"
	CompletionReasonShortcut CompletionReason = "shortcut"
	CompletionReasonManual   CompletionReason = "manual"
)

// CompletionRequest captures the current input context for suggestion lookup.
type CompletionRequest struct {
	Input      string
	CursorByte int
	Reason     CompletionReason
	Shortcut   string
	RequestID  uint64
}

// CompletionResult represents what should be shown and how to apply it.
type CompletionResult struct {
	Suggestions []autocomplete.Suggestion
	ReplaceFrom int
	ReplaceTo   int
	Show        bool
}

// InputCompleter resolves completion suggestions for REPL input snapshots.
type InputCompleter interface {
	CompleteInput(ctx context.Context, req CompletionRequest) (CompletionResult, error)
}
