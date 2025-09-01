package repl

import "context"

// EventKind enumerates structured output kinds for the timeline.
type EventKind string

const (
    EventInput          EventKind = "repl_input"          // recommended to be emitted by REPL shell
    EventResultMarkdown EventKind = "repl_result_markdown" // props: markdown|string in "markdown" or "text"
    EventStdout         EventKind = "repl_stdout"          // props: append|string or text
    EventStderr         EventKind = "repl_stderr"          // props: append|string or text, is_error=true
    EventToolCalls      EventKind = "repl_tool_calls"
    EventProgress       EventKind = "repl_progress"
    EventPerf           EventKind = "repl_perf"
    EventTable          EventKind = "repl_table"
    EventDiff           EventKind = "repl_diff"
    EventShellCmd       EventKind = "repl_shell_cmd"
    EventInspector      EventKind = "repl_inspector"
)

// Event carries a semantic payload for the UI.
type Event struct {
    Kind  EventKind
    Props map[string]any
}

// Evaluator executes code and streams events.
// Errors should be represented as events too (e.g., stderr or result with error annotation),
// with the returned error used for terminal failures.
type Evaluator interface {
    EvaluateStream(ctx context.Context, code string, emit func(Event)) error
    GetPrompt() string
    GetName() string
    SupportsMultiline() bool
    GetFileExtension() string
}

