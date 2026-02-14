package repl

import "time"

type helpBarModel struct {
	provider HelpBarProvider

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	visible bool
	payload HelpBarPayload

	lastErr     error
	lastReqID   uint64
	lastReqKind HelpBarReason
}
