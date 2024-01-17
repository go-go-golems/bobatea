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
	Tree           *ConversationTree
	ConversationID uuid.UUID
}

var _ Manager = (*ManagerImpl)(nil)

type ManagerOption func(*ManagerImpl)

func WithMessages(messages ...*Message) ManagerOption {
	return func(m *ManagerImpl) {
		m.AppendMessages(messages...)
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
		Tree:           NewConversationTree(),
	}
	for _, option := range options {
		option(ret)
	}

	if ret.ConversationID == uuid.Nil {
		ret.ConversationID = uuid.New()
	}

	return ret
}

func (c *ManagerImpl) GetConversation() Conversation {
	return c.Tree.GetLeftMostThread(c.Tree.RootID)
}

func (c *ManagerImpl) GetMessage(ID NodeID) (*Message, bool) {
	return c.Tree.GetMessage(ID)
}

func (c *ManagerImpl) AppendMessages(messages ...*Message) {
	c.Tree.AppendMessages(messages)
}

func (c *ManagerImpl) AttachMessages(parentID NodeID, messages ...*Message) {
	c.Tree.AttachThread(parentID, messages)
}

func (c *ManagerImpl) PrependMessages(messages ...*Message) {
	c.Tree.PrependThread(messages)
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
