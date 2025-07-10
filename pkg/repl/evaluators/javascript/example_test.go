package javascript_test

import (
	"context"
	"fmt"
	"log"

	"github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
)

func ExampleEvaluator() {
	// Create a JavaScript evaluator with default configuration
	evaluator, err := javascript.NewWithDefaults()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Basic JavaScript evaluation
	result, err := evaluator.Evaluate(ctx, "2 + 3")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result: %s\n", result)

	// Variable assignment and retrieval
	_, err = evaluator.Evaluate(ctx, "let name = 'World'")
	if err != nil {
		log.Fatal(err)
	}

	result, err = evaluator.Evaluate(ctx, "name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Name: %s\n", result)

	// Function definition and call
	_, err = evaluator.Evaluate(ctx, "function greet(name) { return 'Hello, ' + name + '!'; }")
	if err != nil {
		log.Fatal(err)
	}

	result, err = evaluator.Evaluate(ctx, "greet(name)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Greeting: %s\n", result)

	// Output:
	// Result: 5
	// Name: World
	// Greeting: Hello, World!
}

func ExampleEvaluator_customConfiguration() {
	// Create a JavaScript evaluator with custom configuration
	config := javascript.Config{
		EnableModules:     true,
		EnableConsoleLog:  true,
		EnableNodeModules: true,
		CustomModules: map[string]interface{}{
			"math": map[string]interface{}{
				"pi":  3.14159,
				"add": func(a, b float64) float64 { return a + b },
			},
		},
	}

	evaluator, err := javascript.New(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use custom module
	result, err := evaluator.Evaluate(ctx, "math.pi")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Pi: %s\n", result)

	result, err = evaluator.Evaluate(ctx, "math.add(10, 5)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Addition: %s\n", result)

	// Output:
	// Pi: 3.14159
	// Addition: 15
}

func ExampleEvaluator_interfaceProperties() {
	evaluator, err := javascript.NewWithDefaults()
	if err != nil {
		log.Fatal(err)
	}

	// Interface properties
	fmt.Printf("Name: %s\n", evaluator.GetName())
	fmt.Printf("Prompt: %s\n", evaluator.GetPrompt())
	fmt.Printf("File Extension: %s\n", evaluator.GetFileExtension())
	fmt.Printf("Supports Multiline: %t\n", evaluator.SupportsMultiline())

	// Output:
	// Name: JavaScript
	// Prompt: js>
	// File Extension: .js
	// Supports Multiline: true
}
