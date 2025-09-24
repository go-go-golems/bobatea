package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatValue converts a value to its string representation
func formatValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

// censorValue returns a censored version of a value
func censorValue(value any) string {
	if value == nil {
		return ""
	}
	// Simple redaction strategy - could be enhanced
	return "[redacted]"
}

// renderChangeLine renders a single change line with prefix and styling
func renderChangeLine(prefix string, value any, redacted bool, style lipgloss.Style) string {
	if value == nil {
		return ""
	}

	// Handle empty strings with special labels
	if str, ok := value.(string); ok && str == "" {
		label := "EMPTY"
		if strings.HasPrefix(prefix, "-") {
			label = "REMOVED"
		} else if strings.HasPrefix(prefix, "+") {
			label = "ADDED"
		}
		return style.Render(fmt.Sprintf("%s %s", prefix, label))
	}

	display := formatValue(value)
	if redacted {
		display = censorValue(value)
	}
	return style.Render(fmt.Sprintf("%s %s", prefix, display))
}

// renderChangeLines renders before/after values for a change
func renderChangeLines(ch Change, redacted bool, styles Styles) (string, string) {
	var left, right string

	switch ch.Status() {
	case ChangeStatusAdded:
		right = renderChangeLine("+", ch.After(), redacted && ch.Sensitive(), styles.AddedLine)
	case ChangeStatusRemoved:
		left = renderChangeLine("-", ch.Before(), redacted && ch.Sensitive(), styles.RemovedLine)
	case ChangeStatusUpdated:
		left = renderChangeLine("-", ch.Before(), redacted && ch.Sensitive(), styles.RemovedLine)
		right = renderChangeLine("+", ch.After(), redacted && ch.Sensitive(), styles.AddedLine)
	}

	return left, right
}
