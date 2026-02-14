package repl

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

type helpBarDebounceMsg struct {
	RequestID uint64
}

type helpBarResultMsg struct {
	RequestID uint64
	Payload   HelpBarPayload
	Err       error
}

type helpDrawerDebounceMsg struct {
	RequestID uint64
}

type helpDrawerResultMsg struct {
	RequestID uint64
	Doc       HelpDrawerDocument
	Err       error
}

type completionOverlayLayout struct {
	PopupX       int
	PopupY       int
	PopupWidth   int
	VisibleRows  int
	ContentWidth int
}

type helpDrawerOverlayLayout struct {
	PanelX        int
	PanelY        int
	PanelWidth    int
	PanelHeight   int
	ContentWidth  int
	ContentHeight int
}
