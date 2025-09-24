package xdiff

import (
    "fmt"
    "sort"
)

type PathChange struct {
	Path      string
	BeforeAny any
	AfterAny  any
}

// MapDiff computes added/removed/updated by comparing two flattened maps
// (path → value). Returns three sorted slices.
func MapDiff(before, after map[string]any) (added, removed, updated []PathChange) {
	seen := make(map[string]struct{})

	for k, av := range after {
		seen[k] = struct{}{}
		bv, ok := before[k]
		if !ok {
			added = append(added, PathChange{Path: k, AfterAny: av})
			continue
		}
		if !equalAny(bv, av) {
			updated = append(updated, PathChange{Path: k, BeforeAny: bv, AfterAny: av})
		}
	}

	for k, bv := range before {
		if _, ok := seen[k]; ok {
			continue
		}
		removed = append(removed, PathChange{Path: k, BeforeAny: bv})
	}

	sort.Slice(added, func(i, j int) bool { return added[i].Path < added[j].Path })
	sort.Slice(removed, func(i, j int) bool { return removed[i].Path < removed[j].Path })
	sort.Slice(updated, func(i, j int) bool { return updated[i].Path < updated[j].Path })
	return
}

func equalAny(a, b any) bool {
    // basic equality; fine for examples
    return toString(a) == toString(b)
}

func toString(v any) string {
    switch t := v.(type) {
    case string:
        return t
    case []byte:
        return string(t)
    default:
        return fmt.Sprint(v)
    }
}


