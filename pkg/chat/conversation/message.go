package conversation

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

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

type Conversation []*Message

type MessageOption func(*Message)

func WithID(id uuid.UUID) MessageOption {
	return func(m *Message) {
		m.ID = id
	}
}

func WithParentID(parentID uuid.UUID) MessageOption {
	return func(m *Message) {
		m.ParentID = parentID
	}
}

func WithConversationID(conversationID uuid.UUID) MessageOption {
	return func(m *Message) {
		m.ConversationID = conversationID
	}
}

func WithMetadata(metadata map[string]interface{}) MessageOption {
	return func(m *Message) {
		m.Metadata = metadata
	}
}

func WithTime(time time.Time) MessageOption {
	return func(m *Message) {
		m.Time = time
	}
}

func NewMessage(text string, role string, options ...MessageOption) *Message {
	m := &Message{
		Text:           text,
		Time:           time.Now(),
		Role:           role,
		ID:             uuid.Nil,
		ConversationID: uuid.Nil,
		ParentID:       uuid.Nil,
	}

	for _, option := range options {
		option(m)
	}

	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	if m.ConversationID == uuid.Nil {
		m.ConversationID = uuid.New()
	}

	return m
}

const RoleSystem = "system"
const RoleAssistant = "assistant"
const RoleUser = "user"
const RoleTool = "tool"

// GetSinglePrompt concatenates all the messages together with a prompt in front.
// It just concatenates all the messages together with a prompt in front (if there are more than one message).
func (messages Conversation) GetSinglePrompt() string {
	if len(messages) == 0 {
		return ""
	}

	if len(messages) == 1 {
		return messages[0].Text
	}

	prompt := ""
	for _, message := range messages {
		prompt += fmt.Sprintf("[%s]: %s\n", message.Role, message.Text)
	}

	return prompt
}
