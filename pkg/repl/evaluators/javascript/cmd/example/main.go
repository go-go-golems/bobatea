package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
)

func main() {
	fmt.Println("JavaScript Evaluator Example")
	fmt.Println("============================")

	// Create evaluator with default configuration
	evaluator, err := javascript.NewWithDefaults()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Display evaluator information
	fmt.Printf("Evaluator Name: %s\n", evaluator.GetName())
	fmt.Printf("Prompt: %s\n", evaluator.GetPrompt())
	fmt.Printf("Supports Multiline: %t\n", evaluator.SupportsMultiline())
	fmt.Printf("File Extension: %s\n", evaluator.GetFileExtension())
	fmt.Println()

	// Example 1: Basic arithmetic
	fmt.Println("Example 1: Basic Arithmetic")
	result, err := evaluator.Evaluate(ctx, "2 + 3 * 4")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("2 + 3 * 4 = %s\n", result)
	fmt.Println()

	// Example 2: Variable assignment
	fmt.Println("Example 2: Variable Assignment")
	_, err = evaluator.Evaluate(ctx, "let name = 'World'")
	if err != nil {
		log.Fatal(err)
	}

	result, err = evaluator.Evaluate(ctx, "name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("name = %s\n", result)
	fmt.Println()

	// Example 3: Function definition and call
	fmt.Println("Example 3: Function Definition and Call")
	_, err = evaluator.Evaluate(ctx, "function greet(name) { return 'Hello, ' + name + '!'; }")
	if err != nil {
		log.Fatal(err)
	}

	result, err = evaluator.Evaluate(ctx, "greet(name)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("greet(name) = %s\n", result)
	fmt.Println()

	// Example 4: Console logging
	fmt.Println("Example 4: Console Logging")
	_, err = evaluator.Evaluate(ctx, "console.log('Hello from JavaScript!')")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()

	// Example 5: Error handling
	fmt.Println("Example 5: Error Handling")
	_, err = evaluator.Evaluate(ctx, "invalid syntax")
	if err != nil {
		fmt.Printf("Error (expected): %v\n", err)
	}
	fmt.Println()

	// Example 6: Using set/get variables from Go
	fmt.Println("Example 6: Go-JavaScript Variable Exchange")
	err = evaluator.SetVariable("goVar", "Hello from Go!")
	if err != nil {
		log.Fatal(err)
	}

	result, err = evaluator.Evaluate(ctx, "goVar")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("goVar = %s\n", result)

	_, err = evaluator.Evaluate(ctx, "let jsVar = 'Hello from JavaScript!'")
	if err != nil {
		log.Fatal(err)
	}

	goValue, err := evaluator.GetVariable("jsVar")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("jsVar (from Go) = %s\n", goValue)
	fmt.Println()

	// Example 7: Available modules
	fmt.Println("Example 7: Available Modules")
	modules := evaluator.GetAvailableModules()
	if len(modules) > 0 {
		fmt.Printf("Available modules: %v\n", modules)
	} else {
		fmt.Println("No modules available")
	}
	fmt.Println()

	// Example 8: Code validation
	fmt.Println("Example 8: Code Validation")
	codes := []string{
		"2 + 3",
		"function test() { return 42; }",
		"invalid syntax {",
	}

	for _, code := range codes {
		isValid := evaluator.IsValidCode(code)
		fmt.Printf("'%s' is valid: %t\n", code, isValid)
	}
	fmt.Println()

	// Example 9: Custom configuration
	fmt.Println("Example 9: Custom Configuration")
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

	customEvaluator, err := javascript.New(config)
	if err != nil {
		log.Fatal(err)
	}

	result, err = customEvaluator.Evaluate(ctx, "math.pi")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("math.pi = %s\n", result)

	result, err = customEvaluator.Evaluate(ctx, "math.add(10, 5)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("math.add(10, 5) = %s\n", result)
	fmt.Println()

	fmt.Println("All examples completed successfully!")
}
