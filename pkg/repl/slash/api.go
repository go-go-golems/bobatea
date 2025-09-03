package slash

import (
    "context"

    "github.com/go-go-golems/bobatea/pkg/autocomplete"
)

type ArgType string

const (
    ArgString ArgType = "string"
    ArgNumber ArgType = "number"
    ArgBool   ArgType = "bool"
    ArgEnum   ArgType = "enum"
    ArgFile   ArgType = "file"
    ArgDir    ArgType = "dir"
)

type ArgSpec struct {
    Name        string
    Type        ArgType
    Required    bool
    Variadic    bool
    Enum        []string
    Description string
}

type FlagSpec struct {
    Name        string
    Type        ArgType
    Default     any
    Enum        []string
    Description string
    Negatable   bool
}

type Schema struct {
    Positionals []ArgSpec
    Flags       []FlagSpec
}

type Input struct {
    Raw         string
    Name        string
    Positionals []string
    Flags       map[string][]string
}

// Emitter is a generic event emitter used by slash handlers, decoupled from REPL types.
// kind should match a REPL EventKind (e.g., "repl_result_markdown").
type Emitter func(kind string, props map[string]any)

type Handler func(ctx context.Context, in Input, emit Emitter) error

type CompletionPhase int

const (
    PhaseName CompletionPhase = iota
    PhasePositional
    PhaseFlag
    PhaseFlagValue
)

type CompletionState struct {
    Raw             string
    Caret           int
    Phase           CompletionPhase
    Name            string
    PositionalIndex int
    CurrentFlag     string
    Partial         string
    Parsed          Input
    TokenStart      int
    TokenEnd        int
}

type Completer func(ctx context.Context, st CompletionState) ([]autocomplete.Suggestion, error)

type Command struct {
    Name     string
    Summary  string
    Usage    string
    Schema   Schema
    Run      Handler
    Complete Completer
}

type Registry interface {
    Register(cmd *Command) error
    Unregister(name string)
    Get(name string) *Command
    List() []*Command
}

type Dispatcher interface {
    TryHandle(ctx context.Context, input string, emit Emitter) bool
    Complete(ctx context.Context, raw string, caret int) ([]autocomplete.Suggestion, CompletionState, error)
}


