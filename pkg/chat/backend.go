package chat

import (
	"context"
	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/conversation"
)

// Backend abstracts initiating and stopping the backend process that is responsible
// for streaming and processing chat messages.
//
// Communication between the backend and the chat UI is facilitated through a series
// of Stream*Msg messages that the backend sends to the UI.
//
// The typical flow of communication is as follows:
//  1. The UI invokes the Start method, providing a context and the current conversation
//     messages, to begin the backend streaming process.
//  2. As the backend processes the stream, it sends Stream*Msg messages to the UI:
//     - StreamStartMsg: Indicates the streaming process has started.
//     - StreamStatusMsg: Provides updates on the status of the streaming process.
//     - StreamCompletionMsg: Contains new data from the backend, such as a new message.
//     - StreamDoneMsg: Signals the successful completion of the streaming process.
//     - StreamErrorMsg: Communicates errors that occurred during streaming.
//  3. The UI's Update method receives these messages and updates the chat model and view.
//  4. The Backend interface provides Interrupt and Kill methods to allow the UI to request
//     the backend to gracefully stop or forcefully terminate the streaming process.
//  5. Upon completion of its tasks, the backend sends a BackendFinishedMsg to indicate
//     that it has finished processing and will not send any further messages.
//
// The backend is expected to maintain the context of the conversation, ensuring that
// new messages sent as completion events are correctly associated with the last message
// in the conversation to maintain the chat's continuity.
type Backend interface {
	// Start begins the backend process with the provided context and conversation messages.
	Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error)

	// Interrupt signals the backend process to gracefully stop its current operation.
	Interrupt()

	// Kill forces the backend process to terminate immediately. This is used in
	// situations where an immediate halt of the backend process is required, such as
	// when the application is closing or an unrecoverable error has occurred.
	Kill()

	// IsFinished checks if the backend process has completed its tasks. It returns
	// true if the backend has finished processing and no further Stream*Msg messages
	// will be sent to the UI.
	IsFinished() bool
}

// BackendFinishedMsg is a message sent when the backend process has finished its operation.
type BackendFinishedMsg struct{}
