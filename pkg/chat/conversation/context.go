package conversation

import (
	"encoding/json"
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

type ManagerImpl struct {
	Messages       []*Message
	ConversationID uuid.UUID
}

var _ Manager = (*ManagerImpl)(nil)

type ManagerOption func(*ManagerImpl)

func WithMessages(messages []*Message) ManagerOption {
	return func(m *ManagerImpl) {
		m.AddMessages(messages...)
	}
}

func WithManagerConversationID(conversationID uuid.UUID) ManagerOption {
	return func(m *ManagerImpl) {
		m.ConversationID = conversationID
	}
}

func NewManager(options ...ManagerOption) *ManagerImpl {
	ret := &ManagerImpl{
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

func (ret *ManagerImpl) setMessageIds() {
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

func (c *ManagerImpl) GetConversation() Conversation {
	return c.Messages
}

func (c *ManagerImpl) AddMessages(messages ...*Message) {
	c.Messages = append(c.Messages, messages...)
	c.setMessageIds()
}

func (c *ManagerImpl) PrependMessages(messages ...*Message) {
	c.Messages = append(messages, c.Messages...)
	c.setMessageIds()
}

func (c *ManagerImpl) SaveToFile(s string) error {
	// TODO(manuel, 2023-11-14) For now only json
	msgs := c.GetConversation()
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
