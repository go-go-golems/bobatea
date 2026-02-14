package repl

import "time"

type completionModel struct {
	provider InputCompleter

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	visible     bool
	selection   int
	replaceFrom int
	replaceTo   int
	scrollTop   int
	visibleRows int

	maxVisible int
	pageSize   int
	maxWidth   int
	maxHeight  int
	minWidth   int
	margin     int
	offsetX    int
	offsetY    int
	noBorder   bool
	placement  CompletionOverlayPlacement
	horizontal CompletionOverlayHorizontalGrow

	lastResult  CompletionResult
	lastError   error
	lastReqID   uint64
	lastReqKind CompletionReason
}
