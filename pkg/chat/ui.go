package chat

import (
	"context"
	"github.com/google/uuid"
	"time"
)

import "github.com/charmbracelet/bubbletea"

type Backend interface {
	Interrupt()
	Kill()
	GetNextCompletion() tea.Cmd
	Start(ctx context.Context, msgs []*Message) error
	IsFinished() bool
}

type Message struct {
	Text string    `json:"text" yaml:"text"`
	Time time.Time `json:"time" yaml:"time"`
	Role string    `json:"role" yaml:"role"`

	ID             uuid.UUID `json:"id" yaml:"id"`
	ParentID       uuid.UUID `json:"parent_id" yaml:"parent_id"`
	ConversationID uuid.UUID `json:"conversation_id" yaml:"conversation_id"`

	// additional metadata for the message
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type ConversationManager interface {
	GetMessages() []*Message
	GetMessagesWithSystemPrompt() []*Message

	AddMessages(msgs ...*Message)
	SaveToFile(filename string) error
}

const RoleSystem = "system"
const RoleAssistant = "assistant"
const RoleUser = "user"
const RoleTool = "tool"
