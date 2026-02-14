package repl

import "github.com/go-go-golems/bobatea/pkg/tui/inputhistory"

// HistoryEntry aliases the shared input history entry type.
type HistoryEntry = inputhistory.HistoryEntry

// History aliases the shared input history state manager.
type History = inputhistory.History

// NewHistory keeps backward-compatible constructor usage inside repl.
func NewHistory(maxSize int) *History {
	return inputhistory.NewHistory(maxSize)
}
