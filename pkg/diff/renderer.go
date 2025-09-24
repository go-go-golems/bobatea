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

	// Lowercased search for detail-side filtering
	lq := strings.ToLower(strings.TrimSpace(searchQuery))

	a, r, u := countChangesFiltered(item, lq, statusFilter, filtersOn)
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

			// Status filtering
			if filtersOn {
				switch ch.Status() {
				case ChangeStatusAdded:
					if !statusFilter.ShowAdded {
						continue
					}
				case ChangeStatusRemoved:
					if !statusFilter.ShowRemoved {
						continue
					}
				case ChangeStatusUpdated:
					if !statusFilter.ShowUpdated {
						continue
					}
				}
			}

			// Detail-side search filtering: only include matching lines
			if lq != "" {
				bp := strings.ToLower(fmt.Sprint(ch.Before()))
				ap := strings.ToLower(fmt.Sprint(ch.After()))
				if !strings.Contains(strings.ToLower(path), lq) && !strings.Contains(bp, lq) && !strings.Contains(ap, lq) {
					continue
				}
			}

			left, right := renderChangeLines(ch, redacted, styles)

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


// countChanges returns (added, removed, updated) counts
// (removed) countChanges superseded by countChangesFiltered

// countChangesFiltered applies status and search filtering to count detail lines
func countChangesFiltered(item DiffItem, lq string, status StatusFilter, filtersOn bool) (int, int, int) {
	var a, r, u int
	for _, cat := range item.Categories() {
		if cat == nil {
			continue
		}
		for _, ch := range cat.Changes() {
			if ch == nil {
				continue
			}
			if filtersOn {
				switch ch.Status() {
				case ChangeStatusAdded:
					if !status.ShowAdded {
						continue
					}
				case ChangeStatusRemoved:
					if !status.ShowRemoved {
						continue
					}
				case ChangeStatusUpdated:
					if !status.ShowUpdated {
						continue
					}
				}
			}
			if lq != "" {
				if !strings.Contains(strings.ToLower(ch.Path()), lq) &&
					!strings.Contains(strings.ToLower(fmt.Sprint(ch.Before())), lq) &&
					!strings.Contains(strings.ToLower(fmt.Sprint(ch.After())), lq) {
					continue
				}
			}
			switch ch.Status() {
			case ChangeStatusAdded:
				a++
			case ChangeStatusRemoved:
				r++
			case ChangeStatusUpdated:
				u++
			}
		}
	}
	return a, r, u
}

func renderFilterLine(f StatusFilter, styles Styles) string {
	badge := func(label string, on bool, stOn, stOff lipgloss.Style) string {
		if on {
			return stOn.Render(label + " ON")
		}
		return stOff.Render(label + " OFF")
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
			// Match path
			if strings.Contains(strings.ToLower(ch.Path()), lowerQuery) {
				return true
			}
			// Match values (stringify)
			before := strings.ToLower(fmt.Sprint(ch.Before()))
			after := strings.ToLower(fmt.Sprint(ch.After()))
			if strings.Contains(before, lowerQuery) || strings.Contains(after, lowerQuery) {
				return true
			}
		}
	}
	return false
}
