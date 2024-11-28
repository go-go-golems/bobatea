# Bobatea Chat UI

Bobatea Chat UI is a Go library for building interactive terminal-based chat interfaces that communicate with language model (LLM) stream completion endpoints.

## Features

- Real-time streaming responses from AI assistants
- Rich conversation navigation and history
- Extensive keyboard shortcuts
- Code block and message copying
- Conversation import/export
- Customizable styling

## Installation

```bash
go get github.com/go-go-golems/bobatea/pkg/chat
```

## Using the Chat Interface

### UI Modes

The chat interface has three main modes:

1. **User Input Mode** (Default)
   - Type your message and submit it to the AI
   - See real-time streaming responses
   - Copy the last response or code blocks

2. **Navigation Mode**
   - Browse through conversation history
   - Select and copy previous messages
   - Return to input mode to continue the conversation

3. **Stream Mode**
   - Shows real-time AI responses
   - Displays progress indicators
   - Can be cancelled if needed

### Keyboard Controls

#### General
- `ctrl-?` - Show help dialog
- `alt+q` - Quit application
- `ctrl+s`, `alt+s` - Save conversation to file

#### Message Input
- `tab` - Submit your message
- `esc` or `ctrl+g` - Cancel current operation, unfocus message, or dismiss error
- `alt+l` - Copy last AI response
- `alt+k` - Copy code blocks from last response

#### Navigation
- `shift+pgup` - Scroll conversation up
- `shift+pgdown` - Scroll conversation down
- `shift+up` - Select previous message
- `shift+down` - Select next message
- `enter` - Focus selected message
- `alt+c` - Copy selected message
- `alt+d` - Copy code blocks from selected message
- `left` - Previous conversation thread
- `right` - Next conversation thread

#### Coming Soon
- Message regeneration
- Conversation branching
- Message editing
- Thread navigation

## For Developers

See [TUTORIAL.md](./TUTORIAL.md) for implementation details and backend integration guide.

## Example

Check out `cmd/chat/main.go` for a complete working example.

