package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-go-golems/bobatea/pkg/repl"
	jsrepl "github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
)

func main() {
	ctx := context.Background()
	e, err := jsrepl.NewWithDefaults()
	if err != nil {
		log.Fatalf("new evaluator: %v", err)
	}

	_, err = e.Evaluate(ctx, "function greetUser(name) { return 'hi ' + name; }")
	if err != nil {
		log.Fatalf("define function: %v", err)
	}
	_, err = e.Evaluate(ctx, "const dataBucket = { count: 1, label: 'demo' }")
	if err != nil {
		log.Fatalf("define object: %v", err)
	}

	check := func(input string) {
		res, cErr := e.CompleteInput(ctx, repl.CompletionRequest{
			Input:      input,
			CursorByte: len(input),
			Reason:     repl.CompletionReasonShortcut,
			Shortcut:   "tab",
		})
		if cErr != nil {
			log.Fatalf("complete %q: %v", input, cErr)
		}

		values := make([]string, 0, len(res.Suggestions))
		for _, s := range res.Suggestions {
			values = append(values, s.Value)
		}

		fmt.Printf("INPUT=%q show=%v suggestions=%d\n", input, res.Show, len(values))
		fmt.Printf("  has greetUser=%v, has dataBucket=%v\n",
			has(values, "greetUser"), has(values, "dataBucket"))
		if len(values) > 0 {
			limit := min(12, len(values))
			fmt.Printf("  first %d: %s\n", limit, strings.Join(values[:limit], ", "))
		}
	}

	check("gre")
	check("dataB")
	check("dataBucket.")

	help, err := e.GetHelpBar(ctx, repl.HelpBarRequest{
		Input:      "greetUser",
		CursorByte: len("greetUser"),
		Reason:     repl.HelpBarReasonManual,
	})
	if err != nil {
		log.Fatalf("help bar: %v", err)
	}
	fmt.Printf("HELPBAR greetUser show=%v kind=%q text=%q\n", help.Show, help.Kind, help.Text)
}

func has(values []string, needle string) bool {
	for _, v := range values {
		if v == needle {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
