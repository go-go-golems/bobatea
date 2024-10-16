# Bobatea Chat UI Tutorial

Bobatea Chat UI is a Go library for building interactive terminal-based chat interfaces that communicate with language model (LLM) stream completion endpoints. It provides a high-level abstraction for handling the chat UI and manages the conversation state, allowing developers to focus on implementing the backend interaction with the LLM API.

## Key Components

1. **Backend**: Implements the `chat.Backend` interface to handle LLM communication.
2. **Conversation Manager**: Manages the conversation state and history.
3. **Chat UI Model**: Renders the conversation and handles user input.

### Message Types

The backend communicates with the UI using these message types:

- `StreamStartMsg`: Indicates the start of streaming. The UI appends a new message to show the assistant has started processing.
- `StreamStatusMsg`: Provides status updates during streaming. Can be used to show loading indicators.
- `StreamCompletionMsg`: Contains partial or complete responses. The UI updates the message content in real-time.
- `StreamDoneMsg`: Signals the end of streaming. The UI finalizes the message content.
- `StreamCompletionError`: Communicates errors during streaming. The UI can display an error state.
- `BackendFinishedMsg`: Indicates the backend has finished processing and will not send further messages.

### UI Modes

The chat UI operates in three main modes:

1. **User Input Mode**: The default mode for entering new messages. Users can type and submit messages to the backend.
2. **Stream Completion Mode**: Displays streaming responses from the backend. Shows a loading indicator and updates the conversation in real-time.
3. **Move Around Mode**: Allows navigation through conversation history. Users can select previous messages and copy their content.

### Conversation Tree Structure

The conversation is stored in a tree-like structure, allowing for complex conversation flows:

- Each message is a node in the tree with a unique ID.
- Messages have parent-child relationships.
- The tree structure enables features like branching conversations and context-aware responses.

## Message Types and Order

The chat UI uses several message types to manage the conversation flow and update the UI. Here's an explanation of each message type and the order in which they should be sent:

1. **StreamStartMsg**: 
   - Sent when a new streaming operation begins.
   - Creates a new message in the conversation.
   - Should be the first message sent for each new response.

2. **StreamCompletionMsg**: 
   - Sent for each chunk of the streaming response.
   - Updates the content of the current message.
   - Can be sent multiple times during a single response.

3. **StreamStatusMsg**: 
   - Optional message to provide status updates.
   - Does not directly update the conversation but can be used for UI feedback.
   - Can be sent at any point during the streaming process.

4. **StreamDoneMsg**: 
   - Sent when the streaming operation is complete.
   - Finalizes the content of the current message.
   - Should be the last message sent for each response.

5. **StreamCompletionError**: 
   - Sent if an error occurs during the streaming process.
   - Can be used to display error messages in the UI.

6. **BackendFinishedMsg**: 
   - Sent when the backend has completed all processing.
   - Signals that no more messages will be sent for the current operation.

### Message Flow Example

1. Backend receives a user message and starts processing:
   - Send `StreamStartMsg`
2. As the response is generated:
   - Send multiple `StreamCompletionMsg` with partial responses
   - Optionally send `StreamStatusMsg` for progress updates
3. When the response is complete:
   - Send `StreamDoneMsg`
4. If an error occurs at any point:
   - Send `StreamCompletionError`
5. When all processing is finished:
   - Send `BackendFinishedMsg`

## Implementation Steps

### 1. Implement the Backend

The backend is a crucial component in the Bobatea Chat UI, responsible for processing user inputs, streaming responses, managing conversation state, and handling interruptions. It serves as the bridge between the user interface and the language model API. Before we dive into the code, let's understand the key aspects of a backend:

#### Key Concepts:

1. **Interface Implementation**: The backend must implement the `chat.Backend` interface, which defines methods for starting, interrupting, and managing the backend's lifecycle.

2. **Asynchronous Processing**: To keep the UI responsive, the backend processes user input and generates responses in a separate goroutine.

3. **Streaming Simulation**: The backend simulates a streaming response, sending incremental updates to the UI.

4. **Message Passing**: Communication between the backend and UI is achieved through Bubble Tea's message passing system.

5. **State Management**: The backend maintains its own state, including whether it's currently running and handles for cancellation.

6. **Context Usage**: A context is used for cancellation and timeout management.

#### Backend Interface

```go
type Backend interface {
    Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error)
    Interrupt()
    Kill()
    IsFinished() bool
}
```

This interface defines four essential methods:

1. `Start`: Initiates the backend processing with the given context and messages.
2. `Interrupt`: Interrupts the current processing.
3. `Kill`: Forcefully stops the backend.
4. `IsFinished`: Checks if the backend has finished processing.

#### Backend and Bubble Tea Integration

The backend is tightly integrated with the Bubble Tea program:


1. **Program Injection**: The Bubble Tea program is injected into the backend during initialization:

   ```go
   backend := &FakeBackend{}
   p := tea.NewProgram(model, options...)
   backend.p = p // Inject the program into the backend
   ```

2. **Message Sending**: The backend uses the injected program to send messages to the UI:

   ```go
   f.p.Send(conversationui.StreamCompletionMsg{
       StreamMetadata: metadata,
       Delta:          reversedWords[idx] + " ",
       Completion:     completion,
   })
   ```

The UI's `Update` method receives these messages and updates the model accordingly.

3. **Asynchronous Processing**: The backend typically processes user input and generates responses in a separate goroutine to avoid blocking the UI:

   ```go
   go func() {
       // Processing logic here
       // Send messages using f.p.Send(...)
   }()
   ```


#### Backend-UI Integration:

Now that we understand the concepts, let's implement a `FakeBackend` that demonstrates these principles:

1. Define the backend struct:

````go
type FakeBackend struct {
    p         *tea.Program
    cancel    context.CancelFunc
    isRunning bool
}
````

2. Implement the `chat.Backend` interface methods:

````go
func (f *FakeBackend) Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error) {
    // Implement streaming logic here (we'll detail this below)
}

func (f *FakeBackend) Interrupt() {
    if f.cancel != nil {
        f.cancel()
    }
}

func (f *FakeBackend) Kill() {
    f.Interrupt()
    f.isRunning = false
}

func (f *FakeBackend) IsFinished() bool {
    return !f.isRunning
}
````

3. Implement the `Start` method, which is the core of the backend:

````go
func (f *FakeBackend) Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error) {
    return func() tea.Msg {
        ctx, f.cancel = context.WithCancel(ctx)
        lastMsg := msgs[len(msgs)-1]
        words := strings.Fields(lastMsg.Content.String())
        reversedWords := reverseWords(words)
        msg := strings.Join(reversedWords, " ")

        metadata := conversationui.StreamMetadata{
            ID:       conversation.NewNodeID(),
            ParentID: lastMsg.ID,
        }

        // Start asynchronous processing
        go func() {
            tick := time.Tick(200 * time.Millisecond)
            idx := 0
            defer func() {
                f.p.Send(chat.BackendFinishedMsg{})
                f.cancel()
                f.cancel = nil
                f.isRunning = false
            }()
            
            // Simulate streaming response
            for {
                select {
                case <-ctx.Done():
                    return
                case <-tick:
                    if idx < len(reversedWords) {
                        completion := strings.Join(reversedWords[:idx+1], " ")
                        f.p.Send(
                            conversationui.StreamCompletionMsg{
                                StreamMetadata: metadata,
                                Delta:          reversedWords[idx] + " ",
                                Completion:     completion,
                            },
                        )
                        idx++
                    } else {
                        f.p.Send(conversationui.StreamDoneMsg{
                            StreamMetadata: metadata,
                            Completion:     msg,
                        })
                        return
                    }
                }
            }
        }()

        return conversationui.StreamStartMsg{
            StreamMetadata: metadata,
        }
    }, nil
}
````

This implementation demonstrates:
- Processing the last message in the conversation
- Simulating a streaming response by reversing words
- Sending incremental updates using `StreamCompletionMsg`
- Handling completion with `StreamDoneMsg`
- Using a goroutine for asynchronous processing

#### Exercises:

1. **Real API Integration**: Modify the backend to connect to a real language model API instead of reversing words.
2. **Conversation Context**: Implement a method to maintain conversation context across multiple messages.
3. **Error Handling**: Add comprehensive error handling, including network errors and API limits.
5. **Backend Configuration**: Add configuration options to the backend (e.g., response length, temperature) and expose them to the user.

### 2. Set Up the Conversation Manager

The Conversation Manager handles the storage and retrieval of conversation messages. Initialize it with any initial messages:

````go
manager := conversation.NewManager(conversation.WithMessages(
    conversation.NewChatMessage(conversation.RoleSystem, "Initial prompt"),
))
````

The manager provides methods to:
- Append new messages
- Retrieve the current conversation
- Save and load conversations from files

#### Exercises:

1. **Initial Prompts**: Set up the conversation manager with a system prompt and an initial user message.
2. **Conversation Branching**: Implement a method to create a new conversation branch from a specific message.
3. **Conversation Export**: Add functionality to export the conversation history to a markdown file.

### 3. Create the Chat UI Model

The Chat UI Model renders the conversation and handles user input. Initialize it with the manager and backend:

````go
backend := &FakeBackend{}
model := chat.InitialModel(manager, backend)
````

The model manages:
- Rendering the conversation tree
- Handling user input
- Updating the UI based on backend responses

#### Exercises:

1. **Custom Styling**: Modify the chat UI model to use a different color scheme for user and assistant messages.
2. **Message Timestamps**: Add timestamps to each message in the UI.
3. **Typing Indicator**: Implement a typing indicator that appears when the backend is processing a response.

### 4. Run the Bubble Tea Program

Set up and run the Bubble Tea program, which orchestrates the UI updates and user interactions:

````go
options := []tea.ProgramOption{
    tea.WithMouseCellMotion(),
    tea.WithAltScreen(),
}
p := tea.NewProgram(model, options...)
backend.p = p // Inject the program into the backend

if _, err := p.Run(); err != nil {
    panic(err)
}
````

#### Exercises:

1. **Command-line Arguments**: Modify the main program to accept command-line arguments for initial prompts or configuration.
2. **Multiple Backends**: Implement the ability to switch between different backend implementations at runtime.
3. **Conversation Saving**: Add functionality to save and load the entire conversation state, including the backend's internal state.
