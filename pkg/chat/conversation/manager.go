package conversation

type ConversationManager interface {
	GetMessages() []*Message
	AddMessages(msgs ...*Message)
	SaveToFile(filename string) error
}
