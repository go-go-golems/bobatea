package chat

import (
	"context"
	"github.com/charmbracelet/bubbletea"
)

// Backend abstracts initiating and stopping the backend process that is responsible
// for producing chat completions and other UI updates.
//
// Communication between the backend and the chat UI is now timeline-centric:
// implementions should inject timeline lifecycle events directly into the Bubble Tea
// program using tea.Program.Send:
//   - timeline.UIEntityCreated to create entities (e.g., assistant llm_text, tool_call)
//   - timeline.UIEntityUpdated to stream text or update properties
//   - timeline.UIEntityCompleted to finish an entity
//   - timeline.UIEntityDeleted when removing an entity
// The chat model consumes these events and renders the timeline accordingly.
//
// Typical flow:
//  1. The UI invokes Start(ctx, prompt).
//  2. Backend sends UIEntityCreated for an assistant message, then multiple UIEntityUpdated
//     with an increasing Version to stream text, followed by UIEntityCompleted.
//  3. When the backend finishes, it sends BackendFinishedMsg so the UI can unblur input.
type Backend interface {
	// Start begins the backend process with the provided context and prompt string.
	// Implementations should stream results back to the program via tea messages.
	Start(ctx context.Context, prompt string) (tea.Cmd, error)

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
