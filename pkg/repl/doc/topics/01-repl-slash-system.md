---
Title: REPL Slash Command System
Slug: repl-slash-system
Short: Register, complete, and dispatch /commands in bobatea's timeline-first REPL.
Topics:
- repl
- slash-commands
- completion
- ui
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# REPL Slash Command System

## Overview

The bobatea REPL provides a built-in, backend-agnostic slash command system so applications can register `/commands`, get argument/flag parsing, and tab-completion. This removes the need to hand-parse `/...` inputs in your `Evaluator` and standardizes the UX across apps.

Key packages:

- `github.com/go-go-golems/bobatea/pkg/repl/slash` — Registry, Dispatcher, API types
- `github.com/go-go-golems/bobatea/pkg/repl` — REPL shell integration
- `github.com/go-go-golems/bobatea/pkg/autocomplete` — Suggestion type used for completion

## Features

- Register `/name` commands with summary, usage, schema
- Dispatch happens before `Evaluator.EvaluateStream`
- Tab completion for command names, args, and flags
- Configurable keybinding for focus toggle (so Tab stays free for completion)

## Quick Start

```go
import (
  "context"
  "github.com/go-go-golems/bobatea/pkg/repl"
  "github.com/go-go-golems/bobatea/pkg/repl/slash"
)

// Within your app bootstrapping
bus, model, prog, _ := repl.NewTimelineRepl(myEvaluator, repl.DefaultConfig())

// Register a simple command
model.SlashRegistry().Register(&slash.Command{
  Name:    "hello",
  Summary: "Say hello",
  Usage:   "/hello [name]",
  Run: func(ctx context.Context, in slash.Input, emit slash.Emitter) error {
    name := "world"
    if len(in.Positionals) > 0 { name = in.Positionals[0] }
    emit("repl_result_markdown", map[string]any{"markdown": "Hello, " + name + "!"})
    return nil
  },
})
```

Typing `/hello bob` will render a markdown entity: “Hello, bob!”.

## API Reference

```go
package slash

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

type Emitter func(kind string, props map[string]any)

type Handler func(ctx context.Context, in Input, emit Emitter) error

// Completion
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
```

## How Dispatch Works

The REPL model calls the slash `Dispatcher` first for `/...` inputs. If handled, the evaluator is not invoked. Handlers emit UI events (e.g., markdown) via the `Emitter` function using event kinds from `repl.EventKind` (`repl_result_markdown`, `repl_log`, etc).

## Completion

- Press Tab when input begins with `/` to resolve suggestions. The REPL inserts the best suggestion by replacing the token under the caret.
- For richer UX (popup selection), you can wire an `autocomplete.Model` and feed it with `Dispatcher.Complete` results.

Example completer: suggest enum values for a flag

```go
cmd.Complete = func(ctx context.Context, st slash.CompletionState) ([]autocomplete.Suggestion, error) {
  if st.Phase == slash.PhaseFlagValue && st.CurrentFlag == "color" {
    vals := []string{"red","green","blue"}
    var out []autocomplete.Suggestion
    for _, v := range vals {
      if st.Partial == "" || strings.HasPrefix(v, st.Partial) {
        out = append(out, autocomplete.Suggestion{Id: v, Value: v, DisplayText: v})
      }
    }
    return out, nil
  }
  return nil, nil
}
```

## Integrating Into Your App

1) Construct the REPL using `repl.NewTimelineRepl` to get the `Model`.
2) Call `model.SlashRegistry()` and register your commands.
3) Optionally set `config.FocusToggleKey = "esc"` so Tab is reserved for completion.
4) Run the bus and UI program.

```go
cfg := repl.DefaultConfig()
cfg.FocusToggleKey = "esc" // reserve Tab for completion
bus, model, prog, _ := repl.NewTimelineRepl(myEvaluator, cfg)
reg := model.SlashRegistry()
// register commands...
go bus.Run(context.Background())
prog.Run()
```

## Best Practices

- Keep commands small and orthogonal.
- Provide `Usage` and a minimal `Schema` for good help pages.
- Prefer markdown output for human-readable results.
- Use structured logs (`repl_log`) for diagnostics that should not clutter the main transcript.

## See Also

- `repl-help-system` — how to surface slash commands through the REPL help backend and show documentation pages for commands.


