package repl

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	mode_keymap "github.com/go-go-golems/bobatea/pkg/mode-keymap"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/rs/zerolog/log"
)

func (m *Model) updateInput(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Trace().Interface("k", k).Str("key", k.String()).Msg("updating input")
	prevValue := m.textInput.Value()
	prevCursor := m.textInput.Position()

	if handled, cmd := m.handleCommandPaletteInput(k); handled {
		return m, cmd
	}

	if handled, cmd := m.handleHelpDrawerShortcuts(k); handled {
		return m, cmd
	}

	if handled, cmd := m.handleCompletionNavigation(k); handled {
		return m, cmd
	}

	if cmd := m.triggerCompletionFromShortcut(k); cmd != nil {
		return m, cmd
	}

	switch {
	case key.Matches(k, m.keyMap.ToggleFocus):
		m.focus = "timeline"
		m.textInput.Blur()
		m.sh.SetSelectionVisible(true)
		m.updateKeyBindings()
		return m, nil
	case key.Matches(k, m.keyMap.Submit):
		input := m.textInput.Value()
		if strings.TrimSpace(input) == "" {
			return m, nil
		}
		m.textInput.Reset()
		m.helpBar.visible = false
		if m.config.EnableHistory {
			m.history.Add(input, "", false)
			m.history.ResetNavigation()
		}
		return m, m.submit(input)
	case key.Matches(k, m.keyMap.HistoryPrev):
		if m.config.EnableHistory {
			if entry := m.history.NavigateUp(); entry != "" {
				m.textInput.SetValue(entry)
			}
		}
		return m, tea.Batch(
			m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpBarIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpDrawerIfNeeded(prevValue, prevCursor),
		)
	case key.Matches(k, m.keyMap.HistoryNext):
		if m.config.EnableHistory {
			entry := m.history.NavigateDown()
			m.textInput.SetValue(entry)
		}
		return m, tea.Batch(
			m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpBarIfNeeded(prevValue, prevCursor),
			m.scheduleDebouncedHelpDrawerIfNeeded(prevValue, prevCursor),
		)
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(k)
	return m, tea.Batch(
		cmd,
		m.scheduleDebouncedCompletionIfNeeded(prevValue, prevCursor),
		m.scheduleDebouncedHelpBarIfNeeded(prevValue, prevCursor),
		m.scheduleDebouncedHelpDrawerIfNeeded(prevValue, prevCursor),
	)
}

func (m *Model) updateTimeline(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(k, m.keyMap.ToggleFocus):
		m.focus = "input"
		m.textInput.Focus()
		m.sh.SetSelectionVisible(false)
		m.updateKeyBindings()
		return m, nil
	case key.Matches(k, m.keyMap.TimelinePrev):
		m.sh.SelectPrev()
		return m, nil
	case key.Matches(k, m.keyMap.TimelineNext):
		m.sh.SelectNext()
		return m, nil
	case key.Matches(k, m.keyMap.TimelineEnterExit):
		if m.sh.IsEntering() {
			m.sh.ExitSelection()
		} else {
			m.sh.EnterSelection()
		}
		return m, nil
	case key.Matches(k, m.keyMap.CopyCode):
		return m, m.sh.SendToSelected(timeline.EntityCopyCodeMsg{})
	case key.Matches(k, m.keyMap.CopyText):
		return m, m.sh.SendToSelected(timeline.EntityCopyTextMsg{})
	}
	// route keys to shell/controller (e.g., Tab cycles inside entity)
	cmd := m.sh.HandleMsg(k)
	return m, cmd
}

func (m *Model) updateKeyBindings() {
	mode_keymap.EnableMode(&m.keyMap, m.focus)
}
