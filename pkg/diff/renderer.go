package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderItemDetail renders the right-hand detail pane for a given item.
func renderItemDetail(item DiffItem, redacted bool, styles Styles, searchQuery string) string {
	if item == nil {
		return ""
	}

	header := styles.Title.Render(item.Name())
 	var sections []string

 	for _, cat := range item.Categories() {
 		if cat == nil {
 			continue
 		}
 		var lines []string
 		for _, ch := range cat.Changes() {
 			if ch == nil {
 				continue
 			}

 			path := ch.Path()
 			before := valueToString(ch.Before(), redacted && ch.Sensitive(), styles)
 			after := valueToString(ch.After(), redacted && ch.Sensitive(), styles)

 			// Search filtering in-detail: if query doesn't match path or values, we still render
 			// the category but can dim non-matching lines in a future iteration. For MVP, show all.

 			var left, right string
 			switch ch.Status() {
 			case ChangeStatusAdded:
 				left = ""
 				right = styles.AddedLine.Render(fmt.Sprintf("+ %s", after))
 			case ChangeStatusRemoved:
 				left = styles.RemovedLine.Render(fmt.Sprintf("- %s", before))
 				right = ""
 			default: // updated
 				left = styles.RemovedLine.Render(fmt.Sprintf("- %s", before))
 				right = styles.AddedLine.Render(fmt.Sprintf("+ %s", after))
 			}

 			var pathLine string
 			if path != "" {
 				pathLine = styles.Path.Render(path)
 			}

 			if left != "" {
 				if pathLine != "" {
 					lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, left, "  ", pathLine))
 				} else {
 					lines = append(lines, left)
 				}
 			}
 			if right != "" {
 				if pathLine != "" {
 					lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, right, "  ", pathLine))
 				} else {
 					lines = append(lines, right)
 				}
 			}
 		}

 		if len(lines) == 0 {
 			continue
 		}
 		section := lipgloss.JoinVertical(
 			lipgloss.Left,
 			styles.CategoryHeader.Render(cat.Name()),
 			strings.Join(lines, "\n"),
 		)
 		sections = append(sections, section)
 	}

 	return lipgloss.JoinVertical(
 		lipgloss.Left,
 		header,
 		strings.Join(sections, "\n\n"),
 	)
}

func valueToString(v any, censored bool, styles Styles) string {
	if v == nil {
		return ""
	}
 	if censored {
 		return styles.SensitiveValue.Render("[redacted]")
 	}
 	return fmt.Sprint(v)
}

// itemMatchesQuery returns true if the item matches the lowercased query.
func itemMatchesQuery(item DiffItem, lowerQuery string) bool {
	if item == nil {
 		return false
 	}
 	if lowerQuery == "" {
 		return true
 	}
 	if strings.Contains(strings.ToLower(item.Name()), lowerQuery) {
 		return true
 	}
 	for _, cat := range item.Categories() {
 		if cat == nil {
 			continue
 		}
 		for _, ch := range cat.Changes() {
 			if ch == nil {
 				continue
 			}
 			if strings.Contains(strings.ToLower(ch.Path()), lowerQuery) {
 				return true
 			}
 			before := strings.ToLower(fmt.Sprint(ch.Before()))
 			after := strings.ToLower(fmt.Sprint(ch.After()))
 			if strings.Contains(before, lowerQuery) || strings.Contains(after, lowerQuery) {
 				return true
 			}
 		}
 	}
 	return false
}


