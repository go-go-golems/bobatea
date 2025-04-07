# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands
- Build: `make build`
- Test all: `make test`
- Run single test: `go test ./path/to/package -run TestName`
- Lint: `make lint` or `make lintmax` (for more verbose output)
- Security checks: `make gosec govulncheck`

## Code Style Guidelines
- Go version: 1.24+
- Error handling: Use pkg/errors for wrapping errors
- Testing: Use testify/assert package for assertions
- Imports: Group standard library, third party, and project imports
- Naming: Use CamelCase for exported functions/types, camelCase for unexported
- Documentation: Comments for exported functions/types should start with the name
- Prefer dependency injection over global state
- Use structs with proper field tags when appropriate
- Leverage the charmbracelet libraries (bubbles, bubbletea, lipgloss) for UI
- Follow Go standard idioms for channel usage and goroutine management