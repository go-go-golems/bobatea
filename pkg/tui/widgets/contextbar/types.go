package contextbar

import "context"

// Reason describes why a context bar request was triggered.
type Reason string

const (
	ReasonDebounce Reason = "debounce"
	ReasonShortcut Reason = "shortcut"
	ReasonManual   Reason = "manual"
)

// Request captures current input context for context bar lookup.
type Request struct {
	Input      string
	CursorByte int
	Reason     Reason
	Shortcut   string
	RequestID  uint64
}

// Payload describes what the context bar should render.
type Payload struct {
	Show      bool
	Text      string
	Kind      string
	Severity  string
	Ephemeral bool
}

// Provider resolves contextual one-line help for input snapshots.
type Provider interface {
	GetContextBar(ctx context.Context, req Request) (Payload, error)
}

type DebounceMsg struct {
	RequestID uint64
}

type ResultMsg struct {
	RequestID uint64
	Payload   Payload
	Err       error
}
