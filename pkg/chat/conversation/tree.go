package conversation

import (
	"encoding/json"
	fmt "fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"
)

type ContentType string

const (
	ContentTypeChatMessage ContentType = "chat-message"
)

// NodeContent is an interface for different types of node content.
type NodeContent interface {
	ContentType() ContentType
	String() string
	View() string
}

type NodeID uuid.UUID

func (id NodeID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(id))
}

func (id *NodeID) UnmarshalJSON(data []byte) error {
	var uuid uuid.UUID
	if err := json.Unmarshal(data, &uuid); err != nil {
		return err
	}
	*id = NodeID(uuid)
	return nil
}

func NewNodeID() NodeID {
	return NodeID(uuid.New())
}

// Message represents a single message in the conversation.
type Message struct {
	ParentID   NodeID    `json:"parentID"`
	ID         NodeID    `json:"id"`
	Time       time.Time `json:"time"`
	LastUpdate time.Time `json:"lastUpdate"`

	Content  NodeContent            `json:"content"`
	Metadata map[string]interface{} `json:"metadata"` // Flexible metadata field

	Children []*Message `json:"children,omitempty"`
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

func NewMessage(content NodeContent, options ...MessageOption) *Message {
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

type Conversation []*Message

type Role string

const (
	RoleSystem    Role = "system"
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
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
	return fmt.Sprintf("[%s]: %s", c.Role, strings.TrimRight(c.Text, "\n"))
}

var _ NodeContent = (*ChatMessageContent)(nil)

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

// Intermediate representation for unmarshaling.
type messageNodeAlias struct {
	ID          NodeID                 `json:"id"`
	ParentID    NodeID                 `json:"parentID"`
	Content     json.RawMessage        `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	ContentType ContentType            `json:"contentType"`
}

// UnmarshalJSON custom unmarshaler for Message.
func (mn *Message) UnmarshalJSON(data []byte) error {
	var mna messageNodeAlias
	if err := json.Unmarshal(data, &mna); err != nil {
		return err
	}

	// Determine the type of content based on ContentType.
	switch mna.ContentType {
	case ContentTypeChatMessage:
		var content *ChatMessageContent
		if err := json.Unmarshal(mna.Content, &content); err != nil {
			return err
		}
		mn.Content = content
	default:
		return errors.New("unknown content type")
	}

	mn.ID = mna.ID
	mn.ParentID = mna.ParentID
	mn.Metadata = mna.Metadata
	return nil
}

// ConversationTree holds the entire conversation.
type ConversationTree struct {
	Nodes  map[NodeID]*Message
	RootID NodeID
	LastID NodeID
}

// NewConversationTree creates a new conversation tree.
func NewConversationTree() *ConversationTree {
	return &ConversationTree{
		Nodes: make(map[NodeID]*Message),
	}
}

var NullNode NodeID = NodeID(uuid.Nil)

// InsertMessages adds a new message to the conversation tree.
func (ct *ConversationTree) InsertMessages(msgs ...*Message) {
	for _, msg := range msgs {
		ct.Nodes[msg.ID] = msg
		if ct.RootID == NullNode {
			ct.RootID = msg.ID
		}
		ct.LastID = msg.ID

		if parent, exists := ct.Nodes[msg.ParentID]; exists {
			parent.Children = append(parent.Children, msg)
		}
	}
}

func (ct *ConversationTree) AttachThread(parentID NodeID, thread Conversation) {
	for _, msg := range thread {
		msg.ParentID = parentID
		ct.Nodes[msg.ID] = msg
		if ct.RootID == NullNode {
			ct.RootID = msg.ID
		}
		ct.LastID = msg.ID

		if parent, exists := ct.Nodes[msg.ParentID]; exists {
			parent.Children = append(parent.Children, msg)
		}
		parentID = msg.ID
	}
}

func (ct *ConversationTree) AppendMessages(thread Conversation) {
	ct.AttachThread(ct.LastID, thread)
}

func (ct *ConversationTree) PrependThread(thread Conversation) {
	prevRootID := ct.RootID
	newRootID := NullNode
	for _, msg := range thread {
		ct.Nodes[msg.ID] = msg
		ct.RootID = msg.ID
		newRootID = msg.ID
		// not setting LastID on purpose

		if parent, exists := ct.Nodes[msg.ParentID]; exists {
			parent.Children = append(parent.Children, msg)
		}
	}

	if prevRootID != NullNode {
		if prevRoot, exists := ct.Nodes[prevRootID]; exists {
			prevRoot.ParentID = newRootID
		}
	}
}

// FindSiblings returns the IDs of all sibling messages.
func (ct *ConversationTree) FindSiblings(id NodeID) []NodeID {
	node, exists := ct.Nodes[id]
	if !exists {
		return nil
	}

	parent, exists := ct.Nodes[node.ParentID]
	if !exists {
		return nil
	}

	var siblings []NodeID
	for _, sibling := range parent.Children {
		if sibling.ID != id {
			siblings = append(siblings, sibling.ID)
		}
	}

	return siblings
}

// FindChildren returns the IDs of all child messages.
func (ct *ConversationTree) FindChildren(id NodeID) []NodeID {
	node, exists := ct.Nodes[id]
	if !exists {
		return nil
	}

	var children []NodeID
	for _, child := range node.Children {
		children = append(children, child.ID)
	}

	return children
}

// GetConversationThread returns the linear conversation thread starting from a given child ID.
func (ct *ConversationTree) GetConversationThread(id NodeID) Conversation {
	var thread Conversation
	for uuid.UUID(id) != uuid.Nil {
		node, exists := ct.Nodes[id]
		if !exists {
			break
		}
		thread = append([]*Message{node}, thread...)
		id = node.ParentID
	}
	return thread
}

// GetLeftMostThread returns the thread starting from id by always chosing the first sibling
// in the tree.
func (ct *ConversationTree) GetLeftMostThread(id NodeID) Conversation {
	var thread Conversation
	for id != NullNode {
		node, exists := ct.Nodes[id]
		if !exists {
			break
		}
		thread = append(thread, node)
		if len(node.Children) > 0 {
			id = node.Children[0].ID
		} else {
			id = NullNode
		}
	}
	return thread
}

// SaveToFile saves the conversation tree to a JSON file.
func (ct *ConversationTree) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(ct, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads the conversation tree from a JSON file.
func (ct *ConversationTree) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, ct)
}

func (ct *ConversationTree) GetMessage(id NodeID) (*Message, bool) {
	ret, exists := ct.Nodes[id]

	return ret, exists
}
