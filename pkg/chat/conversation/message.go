package conversation

import (
	"fmt"
)

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
