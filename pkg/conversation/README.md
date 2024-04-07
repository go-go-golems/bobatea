# Conversation Package

The `conversation` package provides a tree-like structure for storing and managing conversation messages in an LLM chatbot. It allows for traversing the conversation in various ways and supports different types of message content.

## Features

- Represents a conversation as a tree of messages
- Supports different types of message content (e.g., chat messages)
- Allows traversing the conversation tree in various ways (e.g., linear thread, leftmost thread)
- Provides methods for inserting, attaching, and prepending messages to the conversation tree
- Supports saving and loading conversation trees to/from JSON files
- Includes a `Manager` interface for high-level management of conversations

## Installation

```bash
go get github.com/your/package/conversation
```

## Usage

### Creating a Conversation Tree

```go
tree := conversation.NewConversationTree()
```

### Inserting Messages

```go
message1 := conversation.NewChatMessage(conversation.RoleUser, "Hello!")
message2 := conversation.NewChatMessage(conversation.RoleAssistant, "Hi there!")

tree.InsertMessages(message1, message2)
```

### Traversing the Conversation Tree

```go
thread := tree.GetConversationThread(message2.ID)
leftmostThread := tree.GetLeftMostThread(tree.RootID)
```

### Saving and Loading Conversation Trees

```go
err := tree.SaveToFile("conversation.json")
if err != nil {
    // Handle error
}

loadedTree := conversation.NewConversationTree()
err = loadedTree.LoadFromFile("conversation.json")
if err != nil {
    // Handle error
}
```

### Using the Manager

```go
manager, err := conversation.CreateManager(
    "System prompt",
    "User prompt",
    []*conversation.Message{},
    nil,
)
if err != nil {
    // Handle error
}

manager.AppendMessages(message1, message2)
conversation := manager.GetConversation()
```

## Message Content Types

The package supports different types of message content. Currently, the following content types are available:

- `ChatMessageContent`: Represents a chat message with a role (system, assistant, user) and text content.

You can define your own message content types by implementing the `MessageContent` interface.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvement, please open an issue or submit a pull request.

## License

This package is licensed under the [MIT License](LICENSE).