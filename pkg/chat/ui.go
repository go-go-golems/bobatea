package chat

import (
	"context"
	"github.com/go-go-golems/bobatea/pkg/chat/conversation"
)

import "github.com/charmbracelet/bubbletea"

// These messages are used by the backend to send new streaming data

type StreamMetadata struct {
	ID             conversation.NodeID `json:"id" yaml:"id"`
	ParentID       conversation.NodeID `json:"parent_id" yaml:"parent_id"`
	ConversationID conversation.NodeID `json:"conversation_id" yaml:"conversation_id"`
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
	Completion string
}

type StreamCompletionMsg struct {
	StreamMetadata
	Delta      string
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
