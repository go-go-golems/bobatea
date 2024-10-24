# Add Cobra CLI with backend selection

Added Cobra CLI to allow users to choose between fake and HTTP backends.

- Converted main.go to use Cobra for command-line interface management
- Added 'fake' and 'http' subcommands for selecting backends
- Implemented flag for HTTP server address in 'http' subcommand
- Created a common runChat function for shared logic
- Improved error handling in main and runChat functions

# Add HTTP Chat Client

Added a new command-line tool to interact with the Bobatea chat HTTP backend.

- Created a new program in bobatea/cmd/chat-client
- Implemented commands for all message types (start, completion, status, done, error, finish)
- Added flags for server address and parent message ID
- Created a README.md with usage instructions and example workflow
- Integrated with the existing Bobatea chat backend

# Update HTTP Backend to Accumulate Completion

Modified the HTTP backend to keep track of the full completion status.

- Added a new `completion` field to store accumulated completion
- Updated `handleCompletion` to accumulate deltas and use full completion when provided
- Modified `handleDone` to use accumulated completion if not provided in the message
- Added completion reset in `handleStart` and `handleDone` for new sessions
- Ensured backward compatibility with clients that provide full completion

# Refactor Backend Initialization

Modified the backend initialization process to avoid circular dependencies.

- Removed tea.Program parameter from NewHTTPBackend and NewFakeBackend
- Added SetProgram method to both HTTPBackend and FakeBackend
- Updated main.go to create the backend first, then the program, and finally set the program on the backend
- Added a type assertion in runChat to call SetProgram only on backends that support it
- Improved thread safety in FakeBackend by adding a mutex

# Add get-status command to chat client

Added a new command to retrieve and display the current status of the chat backend.

- Implemented `get-status` command in the chat client
- The command sends a GET request to the `/status` endpoint of the HTTP backend
- Displays status information including current status, number of messages, last message, and last error
- Updated main.go to include the new command in the CLI options
