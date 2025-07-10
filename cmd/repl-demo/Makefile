# REPL Demo Makefile

# Build the main demo application
build:
	go build -o repl-demo .

# Run the JavaScript REPL
js:
	go run . js

# Run the Math REPL
math:
	go run . math

# Run the Example REPL
example:
	go run . example

# Run the Multi-evaluator REPL
multi:
	go run . multi

# List available themes
themes:
	go run . themes

# Run examples
run-simple:
	go run ./examples/simple-embedding.go

run-custom:
	go run ./examples/custom-evaluator.go

run-multi:
	go run ./examples/multi-evaluator-switcher.go

run-theme:
	go run ./examples/custom-theme.go

run-config:
	go run ./examples/configuration-demo.go

# Run all examples (for testing)
test-examples:
	@echo "Building all examples..."
	go build ./examples/simple-embedding.go
	go build ./examples/custom-evaluator.go
	go build ./examples/multi-evaluator-switcher.go
	go build ./examples/custom-theme.go
	go build ./examples/configuration-demo.go
	@echo "All examples build successfully!"
	rm -f simple-embedding custom-evaluator multi-evaluator-switcher custom-theme configuration-demo

# Run the main demo with different themes
demo-dark:
	go run . js --theme dark --title "Dark Theme Demo"

demo-light:
	go run . js --theme light --title "Light Theme Demo"

# Clean build artifacts
clean:
	rm -f repl-demo
	rm -f simple-embedding custom-evaluator multi-evaluator-switcher custom-theme configuration-demo

# Format the code
fmt:
	go fmt ./...

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy

# Help
help:
	@echo "Available commands:"
	@echo "  build       - Build the main demo application"
	@echo "  js          - Run JavaScript REPL"
	@echo "  math        - Run Math REPL"
	@echo "  example     - Run Example REPL"
	@echo "  multi       - Run Multi-evaluator REPL"
	@echo "  themes      - List available themes"
	@echo ""
	@echo "Examples:"
	@echo "  run-simple  - Run simple embedding example"
	@echo "  run-custom  - Run custom evaluator example"
	@echo "  run-multi   - Run multi-evaluator switcher example"
	@echo "  run-theme   - Run custom theme example"
	@echo "  run-config  - Run configuration demo"
	@echo ""
	@echo "Demo variations:"
	@echo "  demo-dark   - JavaScript REPL with dark theme"
	@echo "  demo-light  - JavaScript REPL with light theme"
	@echo ""
	@echo "Development:"
	@echo "  test-examples - Build all examples to test compilation"
	@echo "  clean       - Clean build artifacts"
	@echo "  fmt         - Format code"
	@echo "  test        - Run tests"
	@echo "  deps        - Install/update dependencies"

.PHONY: build js math example multi themes run-simple run-custom run-multi run-theme run-config test-examples demo-dark demo-light clean fmt test deps help
