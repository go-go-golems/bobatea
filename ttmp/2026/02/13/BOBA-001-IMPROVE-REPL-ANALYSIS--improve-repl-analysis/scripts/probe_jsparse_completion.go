package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/go-go-goja/pkg/jsparse"
)

type probeCase struct {
	name string
	src  string // includes "|" as cursor marker
}

func main() {
	cases := []probeCase{
		{
			name: "property completion on object literal",
			src: strings.TrimSpace(`
const obj = { alpha: 1, beta: 2, gamma: 3 };
obj.|
`),
		},
		{
			name: "property completion with partial text",
			src: strings.TrimSpace(`
const obj = { alpha: 1, beta: 2, gamma: 3 };
obj.al|
`),
		},
		{
			name: "identifier completion in global scope",
			src: strings.TrimSpace(`
const localName = 42;
function greet(name) { return "hi " + name; }
loc|
`),
		},
		{
			name: "builtin completion",
			src: strings.TrimSpace(`
console.|
`),
		},
	}

	for _, tc := range cases {
		runCase(tc)
		fmt.Println(strings.Repeat("-", 72))
	}
}

func runCase(tc probeCase) {
	source, row, col, ok := locateCursor(tc.src)
	if !ok {
		fmt.Printf("[ERROR] %s: no cursor marker found\n", tc.name)
		return
	}

	analysis := jsparse.Analyze(tc.name+".js", source, nil)
	parser, err := jsparse.NewTSParser()
	if err != nil {
		fmt.Printf("[ERROR] %s: tree-sitter init failed: %v\n", tc.name, err)
		return
	}
	defer parser.Close()
	root := parser.Parse([]byte(source))

	ctx := analysis.CompletionContextAt(root, row, col)
	candidates := analysis.CompleteAt(root, row, col)
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Label < candidates[j].Label })

	fmt.Printf("case: %s\n", tc.name)
	fmt.Printf("cursor: row=%d col=%d\n", row, col)
	if analysis.ParseErr != nil {
		fmt.Printf("goja parse error: %v\n", analysis.ParseErr)
	} else {
		fmt.Println("goja parse error: <nil>")
	}
	fmt.Printf("completion kind: %s\n", completionKindName(ctx.Kind))
	fmt.Printf("base expr: %q\n", ctx.BaseExpr)
	fmt.Printf("partial: %q\n", ctx.PartialText)
	fmt.Printf("candidate count: %d\n", len(candidates))
	for i, c := range candidates {
		if i >= 10 {
			fmt.Printf("  ... (%d more)\n", len(candidates)-10)
			break
		}
		fmt.Printf("  - %-16s kind=%d detail=%s\n", c.Label, c.Kind, c.Detail)
	}
}

func locateCursor(src string) (clean string, row int, col int, ok bool) {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		idx := strings.Index(line, "|")
		if idx == -1 {
			continue
		}
		lines[i] = strings.Replace(line, "|", "", 1)
		return strings.Join(lines, "\n"), i, idx, true
	}
	return src, 0, 0, false
}

func completionKindName(k jsparse.CompletionKind) string {
	switch k {
	case jsparse.CompletionNone:
		return "none"
	case jsparse.CompletionProperty:
		return "property"
	case jsparse.CompletionIdentifier:
		return "identifier"
	case jsparse.CompletionArgument:
		return "argument"
	default:
		return fmt.Sprintf("unknown(%d)", k)
	}
}
