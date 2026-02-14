package repl

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderHelp() string {
	if m.help.ShowAll {
		return m.help.FullHelpView(m.computeHelpColumns(m.keyMap.fullHelpBindings()))
	}
	return m.help.ShortHelpView(m.keyMap.ShortHelp())
}

func (m *Model) computeHelpColumns(all []key.Binding) [][]key.Binding {
	enabled := enabledHelpBindings(all)
	if len(enabled) == 0 {
		return nil
	}

	width := m.help.Width
	if width <= 0 {
		return [][]key.Binding{enabled}
	}

	for cols := len(enabled); cols >= 1; cols-- {
		groups := splitHelpColumns(enabled, cols)
		if fullHelpGroupsFit(m, groups) {
			return groups
		}
	}

	return [][]key.Binding{enabled}
}

func enabledHelpBindings(bindings []key.Binding) []key.Binding {
	enabled := make([]key.Binding, 0, len(bindings))
	for _, kb := range bindings {
		if kb.Enabled() {
			enabled = append(enabled, kb)
		}
	}
	return enabled
}

func splitHelpColumns(bindings []key.Binding, cols int) [][]key.Binding {
	if len(bindings) == 0 || cols <= 0 {
		return nil
	}
	rows := (len(bindings) + cols - 1) / cols
	ret := make([][]key.Binding, 0, cols)
	for c := range cols {
		start := c * rows
		if start >= len(bindings) {
			break
		}
		end := start + rows
		if end > len(bindings) {
			end = len(bindings)
		}
		ret = append(ret, bindings[start:end])
	}
	return ret
}

func fullHelpGroupsFit(m *Model, groups [][]key.Binding) bool {
	if len(groups) == 0 {
		return true
	}
	if m.help.Width <= 0 {
		return true
	}

	totalWidth := 0
	separator := m.help.Styles.FullSeparator.Inline(true).Render(m.help.FullSeparator)
	separatorWidth := lipgloss.Width(separator)
	for _, group := range groups {
		if !hasEnabledHelpBindings(group) {
			continue
		}

		if totalWidth > 0 {
			totalWidth += separatorWidth
		}

		keys := make([]string, 0, len(group))
		descs := make([]string, 0, len(group))
		for _, kb := range group {
			if !kb.Enabled() {
				continue
			}
			h := kb.Help()
			keys = append(keys, h.Key)
			descs = append(descs, h.Desc)
		}

		col := lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.help.Styles.FullKey.Render(strings.Join(keys, "\n")),
			" ",
			m.help.Styles.FullDesc.Render(strings.Join(descs, "\n")),
		)
		totalWidth += lipgloss.Width(col)
		if totalWidth > m.help.Width {
			return false
		}
	}

	return true
}

func hasEnabledHelpBindings(group []key.Binding) bool {
	for _, kb := range group {
		if kb.Enabled() {
			return true
		}
	}
	return false
}
