package repl

import (
	"github.com/go-go-golems/bobatea/pkg/tui/widgets/contextbar"
	"github.com/go-go-golems/bobatea/pkg/tui/widgets/contextpanel"
)

func ensureAppendPatch(props map[string]any) map[string]any {
	if props == nil {
		return map[string]any{}
	}
	if _, ok := props["append"]; ok {
		return props
	}
	if s, ok := props["text"].(string); ok {
		return map[string]any{"append": s}
	}
	return props
}

// internal helpers
type timelineRefreshMsg struct{}

type completionDebounceMsg struct {
	RequestID uint64
}

type completionResultMsg struct {
	RequestID uint64
	Result    CompletionResult
	Err       error
}

type helpBarDebounceMsg = contextbar.DebounceMsg
type helpBarResultMsg = contextbar.ResultMsg

type helpDrawerDebounceMsg = contextpanel.DebounceMsg
type helpDrawerResultMsg = contextpanel.ResultMsg

type completionOverlayLayout struct {
	PopupX       int
	PopupY       int
	PopupWidth   int
	VisibleRows  int
	ContentWidth int
}

type helpDrawerOverlayLayout = contextpanel.OverlayLayout
