# Bobatea Conversation UI

Bobatea Conversation UI is a Go library for rendering a conversation tree as a [Bubble Tea](https://github.com/charmbracelet/bubbletea) text UI. It provides a simple way to display conversation messages in a terminal-based interface, intended to be nested within another model.

## Message Types

The library supports handling various message types for streaming operations:

- `StreamStartMsg`: Sent when a streaming operation begins. The UI appends a new message to indicate the assistant has started processing.
- `StreamStatusMsg`: Provides status updates during streaming. Can be used to show loading indicators.
- `StreamCompletionMsg`: Sent when new data, such as a message completion, is available. The UI updates the message content.
- `StreamDoneMsg`: Signals the successful completion of streaming. The UI finalizes the message content.
- `StreamCompletionError`: Indicates an error occurred during streaming. The UI can display an error state.

These message types are used to communicate between the backend and the UI during streaming operations. The backend sends these messages to the UI through the Bubble Tea scheduler, and the UI updates the conversation tree accordingly. This allows for real-time updates and a smooth user experience as the assistant generates responses.

## Usage

There is an example on how to use the library in `cmd/conversation/main.go`.