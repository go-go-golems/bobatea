package inputhistory

import "slices"

// HistoryEntry represents a single entry in input history.
type HistoryEntry struct {
	Input  string
	Output string
	IsErr  bool
}

// History manages command history navigation and bounded retention.
type History struct {
	entries      []HistoryEntry
	inputHistory []string
	currentIndex int
	maxSize      int
}

// NewHistory creates a new history manager.
func NewHistory(maxSize int) *History {
	return &History{
		entries:      make([]HistoryEntry, 0),
		inputHistory: make([]string, 0),
		currentIndex: -1,
		maxSize:      maxSize,
	}
}

// Add adds a new entry to the history.
func (h *History) Add(input, output string, isErr bool) {
	entry := HistoryEntry{
		Input:  input,
		Output: output,
		IsErr:  isErr,
	}

	h.entries = append(h.entries, entry)
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[1:]
	}

	if input != "" && (len(h.inputHistory) == 0 || h.inputHistory[len(h.inputHistory)-1] != input) {
		h.inputHistory = append(h.inputHistory, input)
		if len(h.inputHistory) > h.maxSize {
			h.inputHistory = h.inputHistory[1:]
		}
	}

	h.currentIndex = -1
}

// NavigateUp moves up in history (to older entries).
func (h *History) NavigateUp() string {
	if len(h.inputHistory) == 0 {
		return ""
	}

	if h.currentIndex == -1 {
		h.currentIndex = len(h.inputHistory) - 1
	} else if h.currentIndex > 0 {
		h.currentIndex--
	}

	return h.inputHistory[h.currentIndex]
}

// NavigateDown moves down in history (to newer entries).
func (h *History) NavigateDown() string {
	if h.currentIndex == -1 {
		return ""
	}

	if h.currentIndex < len(h.inputHistory)-1 {
		h.currentIndex++
		return h.inputHistory[h.currentIndex]
	}

	h.currentIndex = -1
	return ""
}

// Clear clears all history.
func (h *History) Clear() {
	h.entries = make([]HistoryEntry, 0)
	h.inputHistory = make([]string, 0)
	h.currentIndex = -1
}

// GetEntries returns all history entries.
func (h *History) GetEntries() []HistoryEntry {
	return slices.Clone(h.entries)
}

// GetAll returns all input history entries.
func (h *History) GetAll() []string {
	return slices.Clone(h.inputHistory)
}

// IsNavigating returns true if currently navigating history.
func (h *History) IsNavigating() bool {
	return h.currentIndex != -1
}

// ResetNavigation resets the navigation state.
func (h *History) ResetNavigation() {
	h.currentIndex = -1
}
