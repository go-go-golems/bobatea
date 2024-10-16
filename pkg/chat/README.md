# Bobatea Chat UI

Bobatea Chat UI is a Go library for building interactive terminal-based chat interfaces that communicate with language model (LLM) stream completion endpoints.

## Features

- Renders a conversation tree as a terminal-based UI using [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Supports streaming message updates from the backend, with real-time display of assistant responses
- Handles user input and sends messages to the backend for processing
- Allows navigation through the conversation history
- Provides customizable key bindings and help dialog
- Abstracts the backend interaction, allowing developers to implement their own LLM communication layer

## Installation

To use the Bobatea Chat UI library in your Go project, run:

```bash
go get github.com/go-go-golems/bobatea/pkg/chat
```

## Usage

### Keybinding

## Supported Keymap Functionality

- `Help` (default: `ctrl-?`): Displays the help dialog with available key bindings and their descriptions.
- `Quit` (default: `alt+q`): Quits the chat UI application.

- `SubmitMessage` (default: `tab`): Submits the current user input message to the backend for processing when in "
- `CancelCompletion` (default: `esc` or `ctrl+g`): Cancels the current streaming completion operation when in "
  stream-completion" mode.
  user-input" mode.

- `ScrollUp` (default: `shift+pgup`): Scrolls the conversation viewport up.
- `ScrollDown` (default: `shift+pgdown`): Scrolls the conversation viewport down.
- `DismissError` (default: `esc` or `ctrl+g`): Dismisses the displayed error message when in "error" mode.

- `SelectPrevMessage` (default: `shift+up`): Moves the selection to the previous message in the conversation history
  when in "moving-around" mode.
- `SelectNextMessage` (default: `shift+down`): Moves the selection to the next message in the conversation history when
  in "moving-around" mode.
- `UnfocusMessage` (default: `esc` or `ctrl+g`): Unfocuses the current message input and enters "moving-around" mode
  when in "user-input" mode.
- `FocusMessage` (default: `enter`): Focuses the selected message for editing or interaction when in "moving-around"
  mode.

- `LoadFromFile`: Loads a conversation from a file (not bound by default).
- `SaveToFile` (default: `ctrl+s`): Saves the current conversation to a file.
- `SaveSourceBlocksToFile` (default: `alt+s`): Saves the source code blocks from the conversation to a file.

- `CopyToClipboard` (default: `alt+c`): Copies the selected message to the clipboard when in "moving-around" mode.
- `CopyLastResponseToClipboard` (default: `alt+l`): Copies the last assistant response to the clipboard when in "
  user-input" mode.
- `CopyLastSourceBlocksToClipboard` (default: `alt+k`): Copies the source code blocks from the last assistant response
  to the clipboard when in "user-input" mode.
- `CopySourceBlocksToClipboard` (default: `alt+d`): Copies the source code blocks from the selected message to the
  clipboard when in "moving-around" mode.

Not yet implemented:

- `Regenerate`: Regenerates the current user input message (not bound by default, only available in "user-input" mode).
- `RegenerateFromHere`: Regenerates the conversation from the selected message onwards (not bound by default, only
  available in "moving-around" mode).
- `EditMessage`: Allows editing the selected message (not bound by default, only available in "moving-around" mode).
- `PreviousConversationThread` (default: `left`): Navigates to the previous conversation thread when in "moving-around"
  mode.
- `NextConversationThread` (default: `right`): Navigates to the next conversation thread when in "moving-around" mode.

## UI Modes

The Bobatea Chat UI operates in three main modes:

1. **User Input Mode**: This is the default mode when the user is actively entering a new message. The user can type
   their message and submit it to the backend for processing using the `SubmitMessage` key binding.

2. **Stream Completion Mode**: When the backend is processing a user input message and generating a response, the UI
   enters the stream completion mode. In this mode, the UI displays a loading indicator and updates the conversation
   display in real-time as it receives `StreamCompletionMsg` messages from the backend. The user can cancel the
   streaming operation using the `CancelCompletion` key binding.

3. **Move Around Mode**: This mode allows the user to navigate through the conversation history and interact with
   previous messages. The user can enter this mode by using the `UnfocusMessage` key binding while in user input mode.
   In move around mode, the user can select previous messages using the `SelectPrevMessage` and `SelectNextMessage` key
   bindings, and copy the content of the selected message using the `CopyToClipboard` or `CopySourceBlocksToClipboard`
   key bindings. The user can return to user input mode by using the `FocusMessage` key binding.

These modes provide a smooth and intuitive user experience for interacting with the chat UI, allowing users to easily
enter new messages, view generated responses, and navigate the conversation history.

## Message Types

The Bobatea Chat UI library defines several message types for communication between the backend and the UI during streaming:

- `StreamStartMsg`: Sent when a streaming operation begins, indicating that the backend has started processing a user
  input message.
- `StreamStatusMsg`: Provides status updates during streaming, allowing the UI to display loading indicators or progress
  information.
- `StreamCompletionMsg`: Sent when new data, such as a message completion or partial response, is available. The UI
  updates the conversation display with the received content.
- `StreamDoneMsg`: Signals the successful completion of the streaming operation, indicating that the backend has
  finished generating a response.
- `StreamCompletionError`: Indicates that an error occurred during the streaming process, allowing the UI to display an
  error message or take appropriate action.

In addition to these streaming-related messages, the library also defines a `BackendFinishedMsg`, which is sent by the
backend to indicate that it has completed its processing and will not send any further messages.

### Implementing the Backend

To use the Bobatea Chat UI, you need to implement the `Backend` interface, which defines methods for starting and
stopping the backend process that communicates with the LLM API. The backend is responsible for sending `Stream*Msg`
messages to the UI to update the conversation state during streaming.

When the backend finishes, it should send a `BackendFinishedMsg` to the UI to indicate that it has completed processing.

### Creating the Chat UI

To create a new chat UI, use the `InitialModel` function, passing a `conversation.Manager` and your backend implementation:

```go
manager := conversation.NewManager()
backend := &MyBackend{}

model := chat.InitialModel(manager, backend)
```

To run the chat UI, create a new Bubble Tea program with your model and start it:

```go
p := tea.NewProgram(model)
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}
```

## Customization

The chat UI appearance can be customized by modifying the `Style` struct in the `conversation` package. You can create a
new style and pass it to the `InitialModel` function using the `WithStyle` option.

Key bindings can be customized by creating a new `KeyMap` struct and updating the bindings as needed. Pass the custom
key map to `InitialModel` using the `WithKeyMap` option.

## Example

For a complete example of how to use the Bobatea Chat UI library, see the `cmd/chat/main.go` file in the repository.

