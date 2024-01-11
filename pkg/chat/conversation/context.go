package conversation

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// LoadFromFile loads messages from a json file or yaml file
func LoadFromFile(filename string) ([]*Message, error) {
	if strings.HasSuffix(filename, ".json") {
		return loadFromJSONFile(filename)
	} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		return loadFromYAMLFile(filename)
	} else {
		return nil, nil
	}
}

func loadFromYAMLFile(filename string) ([]*Message, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	var messages []*Message
	err = yaml.NewDecoder(f).Decode(&messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func loadFromJSONFile(filename string) ([]*Message, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	var messages []*Message
	err = json.NewDecoder(f).Decode(&messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

type Manager struct {
	Messages       []*Message
	ConversationID uuid.UUID
}

var _ ConversationManager = (*Manager)(nil)

type ManagerOption func(*Manager)

func WithMessages(messages []*Message) ManagerOption {
	return func(m *Manager) {
		m.AddMessages(messages...)
	}
}

func WithManagerConversationID(conversationID uuid.UUID) ManagerOption {
	return func(m *Manager) {
		m.ConversationID = conversationID
	}
}

func NewManager(options ...ManagerOption) *Manager {
	ret := &Manager{
		ConversationID: uuid.Nil,
	}
	for _, option := range options {
		option(ret)
	}

	if ret.ConversationID == uuid.Nil {
		ret.ConversationID = uuid.New()
	}

	ret.setMessageIds()

	return ret
}

func (ret *Manager) setMessageIds() {
	parentId := uuid.Nil
	for _, message := range ret.Messages {
		if message.ID == uuid.Nil {
			message.ID = uuid.New()
		}
		message.ConversationID = ret.ConversationID
		message.ParentID = parentId
		parentId = message.ID
	}
}

func (c *Manager) GetMessages() []*Message {
	return c.Messages
}

func (c *Manager) AddMessages(messages ...*Message) {
	c.Messages = append(c.Messages, messages...)
	c.setMessageIds()
}

func (c *Manager) PrependMessages(messages ...*Message) {
	c.Messages = append(messages, c.Messages...)
	c.setMessageIds()
}

// GetSinglePrompt is a helper to use the context manager with a completion api.
// It just concatenates all the messages together with a prompt in front (if there are more than one message).
func (c *Manager) GetSinglePrompt() string {
	messages := c.GetMessages()
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

func (c *Manager) SaveToFile(s string) error {
	// TODO(manuel, 2023-11-14) For now only json
	msgs := c.GetMessages()
	f, err := os.Create(s)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(msgs)
	if err != nil {
		return err
	}

	return nil
}
