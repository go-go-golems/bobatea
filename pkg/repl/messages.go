package repl

// EvaluationCompleteMsg is sent when evaluation is complete
type EvaluationCompleteMsg struct {
    Input  string
    Output string
    Error  error
}

// HistoryNavigationMsg is sent when history navigation occurs
type HistoryNavigationMsg struct {
	Direction string // "up" or "down"
	Entry     string
}

// MultilineModeToggleMsg is sent when multiline mode is toggled
type MultilineModeToggleMsg struct {
	Enabled bool
}

// ClearHistoryMsg is sent when history should be cleared
type ClearHistoryMsg struct{}

// ExternalEditorMsg is sent when external editor should be opened
type ExternalEditorMsg struct {
	Content string
}

// ExternalEditorCompleteMsg is sent when external editor is complete
type ExternalEditorCompleteMsg struct {
	Content string
	Error   error
}

// QuitMsg is sent when the REPL should quit
type QuitMsg struct{}

// SlashCommandMsg is sent when a slash command is executed
type SlashCommandMsg struct {
    Command string
    Args    []string
}

// New messages for streaming evaluator bridge
// Note: these are also declared in model.go for internal usage but kept here for completeness.
type _ReplEventMsg struct { // unused placeholder to avoid breaking external imports
    _ string
}
