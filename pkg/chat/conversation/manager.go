package conversation

import (
	"github.com/go-go-golems/glazed/pkg/helpers/maps"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"strings"
)

type Manager interface {
	GetConversation() Conversation
	AddMessages(msgs ...*Message)
	SaveToFile(filename string) error
}

// CreateManager creates a new Context ManagerImpl. It is used by the code generator
// to initialize a conversation by passing a custom glazed struct for params.
//
// The systemPrompt and prompt templates are rendered using the params.
// Messages are also rendered using the params before being added to the manager.
//
// ManagerOptions can be passed to further customize the manager on creation.
func CreateManager(
	systemPrompt string,
	prompt string,
	messages []*Message,
	params interface{},
	options ...ManagerOption,
) (*ManagerImpl, error) {
	// convert the params to map[string]interface{}
	var ps map[string]interface{}
	if _, ok := params.(map[string]interface{}); !ok {
		var err error
		ps, err = maps.GlazedStructToMap(params)
		if err != nil {
			return nil, err
		}
	} else {
		ps = params.(map[string]interface{})
	}

	manager := NewManager()

	if systemPrompt != "" {
		systemPromptTemplate, err := templating.CreateTemplate("system-prompt").Parse(systemPrompt)
		if err != nil {
			return nil, err
		}

		var systemPromptBuffer strings.Builder
		err = systemPromptTemplate.Execute(&systemPromptBuffer, ps)
		if err != nil {
			return nil, err
		}

		// TODO(manuel, 2023-12-07) Only do this conditionally, or maybe if the system prompt hasn't been set yet, if you use an agent.
		manager.AddMessages(NewMessage(systemPromptBuffer.String(), RoleSystem))
	}

	for _, message := range messages {
		messageTemplate, err := templating.CreateTemplate("message").Parse(message.Text)
		if err != nil {
			return nil, err
		}

		var messageBuffer strings.Builder
		err = messageTemplate.Execute(&messageBuffer, ps)
		if err != nil {
			return nil, err
		}
		s_ := messageBuffer.String()

		manager.AddMessages(NewMessage(s_, message.Role, WithTime(message.Time)))
	}

	// render the prompt
	if prompt != "" {
		// TODO(manuel, 2023-02-04) All this could be handle by some prompt renderer kind of thing
		promptTemplate, err := templating.CreateTemplate("prompt").Parse(prompt)
		if err != nil {
			return nil, err
		}

		// TODO(manuel, 2023-02-04) This is where multisteps would work differently, since
		// the prompt would be rendered at execution time
		var promptBuffer strings.Builder
		err = promptTemplate.Execute(&promptBuffer, ps)
		if err != nil {
			return nil, err
		}

		manager.AddMessages(NewMessage(promptBuffer.String(), RoleUser))
	}

	for _, option := range options {
		option(manager)
	}

	return manager, nil
}
