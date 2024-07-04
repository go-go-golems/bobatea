package conversation

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type ContentType string

const (
	ContentTypeChatMessage ContentType = "chat-message"
	// TODO(manuel, 2024-06-04) This needs to also handle tool call and tool response blocks (tool use block in claude API)
	// See also the comment to refactor this in openai/helpers.go, where tool use information is actually stored in the metadata of the message
	ContentTypeToolUse    ContentType = "tool-use"
	ContentTypeToolResult ContentType = "tool-result"
)

// MessageContent is an interface for different types of node content.
type MessageContent interface {
	ContentType() ContentType
	String() string
	View() string
}

type Role string

const (
	RoleSystem    Role = "system"
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
	RoleTool      Role = "tool"
)

type ChatMessageContent struct {
	Role Role   `json:"role"`
	Text string `json:"text"`
}

func (c *ChatMessageContent) ContentType() ContentType {
	return ContentTypeChatMessage
}

func (c *ChatMessageContent) String() string {
	return c.Text
}

func (c *ChatMessageContent) View() string {
	// If we are markdown, add a newline so that it becomes valid markdown to parse.
	if strings.HasPrefix(c.Text, "```") {
		c.Text = "\n" + c.Text
	}
	return fmt.Sprintf("[%s]: %s", c.Role, strings.TrimRight(c.Text, "\n"))
}

var _ MessageContent = (*ChatMessageContent)(nil)

type ToolUseContent struct {
	ToolID string          `json:"toolID"`
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	// used by openai currently (only function)
	Type string `json:"type"`
}

func (t *ToolUseContent) ContentType() ContentType {
	return ContentTypeToolUse
}

func (t *ToolUseContent) String() string {
	return fmt.Sprintf("ToolUseContent{ToolID: %s, Name: %s, Input: %s}", t.ToolID, t.Name, t.Input)
}

func (t *ToolUseContent) View() string {
	return fmt.Sprintf("ToolUseContent{ToolID: %s, Name: %s, Input: %s}", t.ToolID, t.Name, t.Input)
}

var _ MessageContent = (*ToolUseContent)(nil)

type ToolResultContent struct {
	ToolID string          `json:"toolID"`
	Result json.RawMessage `json:"result"`
}

func (t *ToolResultContent) ContentType() ContentType {
	return ContentTypeToolResult
}

func (t *ToolResultContent) String() string {
	return fmt.Sprintf("ToolResultContent{ToolID: %s, Result: %s}", t.ToolID, t.Result)
}

func (t *ToolResultContent) View() string {
	return fmt.Sprintf("ToolResultContent{ToolID: %s, Result: %s}", t.ToolID, t.Result)
}

var _ MessageContent = (*ToolResultContent)(nil)

// Message represents a single message node in the conversation tree.
type Message struct {
	ParentID   NodeID    `json:"parentID"`
	ID         NodeID    `json:"id"`
	Time       time.Time `json:"time"`
	LastUpdate time.Time `json:"lastUpdate"`

	Content  MessageContent         `json:"content"`
	Metadata map[string]interface{} `json:"metadata"` // Flexible metadata field

	// TODO(manuel, 2024-04-07) Add Parent and Sibling lists
	// omit in json
	Children []*Message `json:"-"`
}

type MessageOption func(*Message)

func WithMetadata(metadata map[string]interface{}) MessageOption {
	return func(message *Message) {
		message.Metadata = metadata
	}
}

func WithTime(time time.Time) MessageOption {
	return func(message *Message) {
		message.Time = time
	}
}

func WithParentID(parentID NodeID) MessageOption {
	return func(message *Message) {
		message.ParentID = parentID
	}
}

func WithID(id NodeID) MessageOption {
	return func(message *Message) {
		message.ID = id
	}
}

func NewMessage(content MessageContent, options ...MessageOption) *Message {
	ret := &Message{
		Content:    content,
		ID:         NodeID(uuid.New()),
		Time:       time.Now(),
		LastUpdate: time.Now(),
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}

func NewChatMessage(role Role, text string, options ...MessageOption) *Message {
	return NewMessage(&ChatMessageContent{
		Role: role,
		Text: text,
	}, options...)
}

func (mn *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		ContentType ContentType `json:"contentType"`
		*Alias
	}{
		ContentType: mn.Content.ContentType(),
		Alias:       (*Alias)(mn),
	})
}

type Conversation []*Message

// GetSinglePrompt concatenates all the messages together with a prompt in front.
// It just concatenates all the messages together with a prompt in front (if there are more than one message).
func (messages Conversation) GetSinglePrompt() string {
	if len(messages) == 0 {
		return ""
	}

	if len(messages) == 1 && messages[0].Content.ContentType() == ContentTypeChatMessage {
		return messages[0].Content.(*ChatMessageContent).Text
	}

	prompt := ""
	for _, message := range messages {
		if message.Content.ContentType() == ContentTypeChatMessage {
			message := message.Content.(*ChatMessageContent)
			prompt += fmt.Sprintf("[%s]: %s\n", message.Role, message.Text)
		}
	}

	return prompt
}
