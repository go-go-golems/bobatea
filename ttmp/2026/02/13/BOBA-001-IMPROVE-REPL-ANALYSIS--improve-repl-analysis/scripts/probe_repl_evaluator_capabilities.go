package main

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
)

type hasHelpText interface {
	GetHelpText() string
}

type hasCodeValidation interface {
	IsValidCode(code string) bool
}

type hasModules interface {
	GetAvailableModules() []string
}

func main() {
	jsEval, err := javascript.NewWithDefaults()
	if err != nil {
		panic(err)
	}

	var asInterface repl.Evaluator = jsEval

	fmt.Println("repl.Evaluator method surface:")
	printInterfaceMethods((*repl.Evaluator)(nil))
	fmt.Println()

	fmt.Println("javascript.Evaluator method surface:")
	printConcreteMethods(jsEval)
	fmt.Println()

	fmt.Println("capability checks via optional assertions:")
	if v, ok := asInterface.(hasHelpText); ok {
		fmt.Printf("  - has help text: yes (len=%d)\n", len(v.GetHelpText()))
	} else {
		fmt.Println("  - has help text: no")
	}
	if v, ok := asInterface.(hasCodeValidation); ok {
		fmt.Printf("  - has code validation: yes (\"let x = 1\" => %v)\n", v.IsValidCode("let x = 1"))
	} else {
		fmt.Println("  - has code validation: no")
	}
	if v, ok := asInterface.(hasModules); ok {
		fmt.Printf("  - has module inventory: yes (%d modules)\n", len(v.GetAvailableModules()))
	} else {
		fmt.Println("  - has module inventory: no")
	}
}

func printInterfaceMethods(ifacePtr any) {
	t := reflect.TypeOf(ifacePtr).Elem()
	methods := make([]string, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		methods = append(methods, m.Name)
	}
	sort.Strings(methods)
	for _, name := range methods {
		fmt.Printf("  - %s\n", name)
	}
}

func printConcreteMethods(v any) {
	t := reflect.TypeOf(v)
	methods := make([]string, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		methods = append(methods, m.Name)
	}
	sort.Strings(methods)
	for _, name := range methods {
		fmt.Printf("  - %s\n", name)
	}
}
