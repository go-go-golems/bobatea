# AGENT.md

## Build & Test Commands
- Build: `make build`
- Test all: `make test`
- Run single test: `go test ./path/to/package -run TestName`
- Security: `make gosec govulncheck`

## Binary & Build Management
- **Prefer `go run`**: Run binaries with `go run ./cmd/component` when possible
- **Build directory**: If compilation is needed, build into `build/` directory at project root
- **Avoid in-place builds**: Don't build binaries in their source directories (keeps git clean)
- **Examples**: All examples should be runnable with `go run .` from their directory

## Code Style Guidelines
- Go version: 1.24+
- Errors: Use pkg/errors for wrapping
- Testing: Use testify/assert package
- Imports: Group standard library, third party, and project imports
- Naming: CamelCase (exported), camelCase (unexported)
- Documentation: Comments for exported items start with name
- UI: Use charmbracelet libraries (bubbles, bubbletea, lipgloss)
- Prefer dependency injection over global state
- Follow Go idioms for channels and goroutines


<goGuidelines>
When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
When building web applications, use htmx, bootstrap and the templ templating language.
Always use a context argument when appropriate.
Use cobra for command-line applications.
Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
Use github.com/pkg/errors for wrapping errors.
When starting goroutines, use errgroup.
</goGuidelines>

Don't fix linting errors.
Build the binary with `make build`, but don't run it unless told to.