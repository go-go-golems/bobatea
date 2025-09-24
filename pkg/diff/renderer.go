package diff

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/charmbracelet/lipgloss"
)

// renderItemDetail renders the right-hand detail pane for a given item.
func renderItemDetail(item DiffItem, redacted bool, styles Styles, searchQuery string, statusFilter StatusFilter, filtersOn bool) string {
	if item == nil {
		return ""
	}

    a, r, u := countChanges(item)
    badges := lipgloss.JoinHorizontal(lipgloss.Left,
        styles.BadgeAdded.Render("+"+strconv.Itoa(a)), " ",
        styles.BadgeRemoved.Render("-"+strconv.Itoa(r)), " ",
        styles.BadgeUpdated.Render("~"+strconv.Itoa(u)),
    )
    header := lipgloss.JoinHorizontal(lipgloss.Left, styles.Title.Render(item.Name()), "  ", badges)
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

            // Status filtering
            if filtersOn {
                switch ch.Status() {
                case ChangeStatusAdded:
                    if !statusFilter.ShowAdded { continue }
                case ChangeStatusRemoved:
                    if !statusFilter.ShowRemoved { continue }
                default: // updated
                    if !statusFilter.ShowUpdated { continue }
                }
            }

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

    // Optional filter line
    filterLine := ""
    if filtersOn {
        filterLine = renderFilterLine(statusFilter, styles)
    }

    return lipgloss.JoinVertical(
 		lipgloss.Left,
        header,
        filterLine,
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

// countChanges returns (added, removed, updated) counts
func countChanges(item DiffItem) (int, int, int) {
    var a, r, u int
    for _, cat := range item.Categories() {
        if cat == nil { continue }
        for _, ch := range cat.Changes() {
            if ch == nil { continue }
            switch ch.Status() {
            case ChangeStatusAdded:
                a++
            case ChangeStatusRemoved:
                r++
            default:
                u++
            }
        }
    }
    return a, r, u
}

func renderFilterLine(f StatusFilter, styles Styles) string {
    badge := func(label string, on bool, stOn, stOff lipgloss.Style) string {
        if on { return stOn.Render(label+" ON") }
        return stOff.Render(label+" OFF")
    }
    return lipgloss.JoinHorizontal(lipgloss.Left,
        badge("+", f.ShowAdded, styles.BadgeAdded, styles.FilterOff), "   ",
        badge("-", f.ShowRemoved, styles.BadgeRemoved, styles.FilterOff), "   ",
        badge("~", f.ShowUpdated, styles.BadgeUpdated, styles.FilterOff),
    )
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


