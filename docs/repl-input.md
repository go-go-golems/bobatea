## REPL Input Component with Slash-Command Autocomplete

This document describes the architecture and usage of a Bubble Tea component that provides a REPL-style text input with support for slash-commands (e.g., `/help`) and inline autocomplete. It integrates with the lightweight `listbox` component and the `autocomplete` package for async suggestion fetching.

### Goals
- Provide a single reusable component to capture free-form input and slash commands.
- Display a bordered REPL input prompt with a `>` indicator.
- When the user types `/`, present a suggestions list below the input showing commands and help texts.
- Support keyboard navigation: Up/Down to move, Tab/Enter to accept a suggestion, Esc to cancel suggestions.
- Allow async fetching of suggestions with basic debouncing.

### Key Types
- `type SlashCommand struct { Name, Help string }` — basic metadata for commands.
- `type Suggestion struct { ID, Primary, Secondary string }` — list entries combining command name and help text.
- `type Completioner func(ctx context.Context, query string) ([]Suggestion, error)` — pluggable suggestion source.
- `type Model struct` — Bubble Tea model that owns a `textinput.Model` and a `listbox.ListModel`.

### High-Level Architecture
1. Input handling uses `bubbles/textinput` for editing state and cursor movement.
2. Suggestions panel uses `pkg/listbox` for lightweight navigation and selection.
3. When the current input starts with `/`, the component enters "slash mode":
   - Query text is the content after the leading `/`.
   - A debounced background command calls the provided `Completioner` to fetch suggestions.
   - Returned suggestions are adapted to `listbox.Item` and rendered as:
     - Left column: `/name` in monospace-like plain text
     - Right column: help string
4. On `enter`/`tab` while a suggestion is visible, the component replaces the input with the chosen command and a trailing space (e.g., `/help `) and exits slash mode.
5. On `esc` or when input no longer starts with `/`, suggestions are hidden.

### Rendering
The component renders a simple boxed layout:

```
┃                                                                              ┃
┃ >                                                                            ┃
┃                                                                              ┃
```

When slash suggestions are available, they are shown between the top and the prompt rows:

```
┃   /new       new session                                                     ┃
┃   /help      show help                                                       ┃
┃   /share     share session                                                   ┃
┃                                                                              ┃
┃ > /                                                                          ┃
┃                                                                              ┃
```

The component handles window size messages to maintain internal width for truncation.

### Messages
- `SelectedMsg` — emitted when the user pressed Enter on the input line (not in slash mode). Contains the final string to submit.
- `CancelMsg` — emitted when the user presses Esc while suggestions are open to indicate cancellation.
- `ErrMsg` — optional errors from the completioner; currently ignored by default by the model.

### Public API
- `New(completioner Completioner, width int, maxSuggestions int) Model`
- `Model.Init() tea.Cmd`
- `Model.Update(msg tea.Msg) (tea.Model, tea.Cmd)`
- `Model.View() string`
- `Model.SetCommands(commands []SlashCommand)` — optional helper to install a default static completioner.

### Completion Strategies
You can:
- Provide a custom `Completioner` that queries dynamic sources.
- Use the built-in static completioner by calling `SetCommands`, which performs case-insensitive prefix/substring matching and prepares ID and texts.

The component uses a small debounce (about ~120ms) and a short timeout (~400ms) similar to `pkg/autocomplete` to avoid flooding completion providers.

### Usage
Create a Bubble Tea program and embed the `repl-input` model into your application model:

```go
type appModel struct {
    repl replinput.Model
}

func newApp() appModel {
    m := appModel{
        repl: replinput.New(nil /* use SetCommands later */, 60, 8),
    }
    m.repl.SetCommands([]replinput.SlashCommand{
        {Name: "new", Help: "new session"},
        {Name: "help", Help: "show help"},
        {Name: "share", Help: "share session"},
    })
    return m
}

func (m appModel) Init() tea.Cmd { return m.repl.Init() }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case replinput.SelectedMsg:
        // handle submitted input: msg.Value
        return m, nil
    }
    var cmd tea.Cmd
    rm, cmd := m.repl.Update(msg)
    m.repl = rm.(replinput.Model)
    return m, cmd
}

func (m appModel) View() string { return m.repl.View() }
```

### Notes
- The UI keeps styling minimal and dependency-free (no Glamour/lipgloss required). You can wrap the `View` output with your own framing/styling if desired.
- The list items are truncated to available width, adjusting with window resizes.


