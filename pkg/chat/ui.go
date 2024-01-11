package chat

import (
	"context"
	"github.com/go-go-golems/bobatea/pkg/chat/conversation"
	"github.com/google/uuid"
)

import "github.com/charmbracelet/bubbletea"

// These messages are used by the backend to send new streaming data

type StreamMetadata struct {
	ID             uuid.UUID `json:"id" yaml:"id"`
	ParentID       uuid.UUID `json:"parent_id" yaml:"parent_id"`
	ConversationID uuid.UUID `json:"conversation_id" yaml:"conversation_id"`
}

type StreamStartMsg struct {
	StreamMetadata
}

type StreamStatusMsg struct {
	StreamMetadata
	Text string
}

type StreamDoneMsg struct {
	StreamMetadata
}

type StreamCompletionMsg struct {
	StreamMetadata
	Completion string
}

// StreamCompletionError does not imply that the stream finished
type StreamCompletionError struct {
	StreamMetadata
	Err error
}

type Backend interface {
	Start(ctx context.Context, msgs []*conversation.Message) error
	Interrupt()
	Kill()

	GetNextCompletion() tea.Cmd
	IsFinished() bool
}
