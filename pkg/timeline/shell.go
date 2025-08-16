package timeline

import (
    "time"

    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/rs/zerolog/log"
)

// Shell wraps a Controller with viewport management and selection helpers so it can
// be embedded as a reusable timeline UI component.
type Shell struct {
    ctrl           *Controller
    viewport       viewport.Model
    width, height  int
    scrollToBottom bool
}

// NewShell constructs a Shell with a fresh Controller backed by the provided registry.
func NewShell(reg *Registry) *Shell {
    return &Shell{
        ctrl:           NewController(reg),
        viewport:       viewport.New(0, 0),
        scrollToBottom: true,
    }
}

// Controller returns the underlying timeline controller.
func (s *Shell) Controller() *Controller { return s.ctrl }

// Init initializes the underlying viewport.
func (s *Shell) Init() tea.Cmd { return s.viewport.Init() }

// SetSize sets shell dimensions and forwards sizes to the controller.
func (s *Shell) SetSize(width, height int) {
    s.width = width
    s.height = height
    s.ctrl.SetSize(width, height)
    s.viewport.Width = width
    s.viewport.Height = height
    s.viewport.YPosition = 0
}

// SetScrollToBottom toggles auto scroll-to-bottom behavior on refreshes.
func (s *Shell) SetScrollToBottom(v bool) { s.scrollToBottom = v }

// AtBottom returns whether the viewport is currently at the bottom.
func (s *Shell) AtBottom() bool { return s.viewport.AtBottom() }

// View returns the current timeline view as rendered within the viewport.
func (s *Shell) View() string {
    s.viewport.SetContent(s.ctrl.View())
    return s.viewport.View()
}

// UpdateViewport forwards messages to the viewport (e.g., scrolling) and returns any command.
func (s *Shell) UpdateViewport(msg tea.Msg) tea.Cmd {
    var cmd tea.Cmd
    s.viewport, cmd = s.viewport.Update(msg)
    return cmd
}

// RefreshView regenerates the viewport content and optionally scrolls to bottom.
func (s *Shell) RefreshView(goToBottom bool) {
    v := s.ctrl.View()
    s.viewport.SetContent(v)
    if goToBottom || s.scrollToBottom {
        s.viewport.GotoBottom()
    }
}

// GotoBottom forces the viewport to scroll to the bottom.
func (s *Shell) GotoBottom() { s.viewport.GotoBottom() }

// ScrollDown scrolls the viewport by n lines.
func (s *Shell) ScrollDown(n int) { s.viewport.ScrollDown(n) }

// ScrollUp scrolls the viewport by n lines.
func (s *Shell) ScrollUp(n int) { s.viewport.ScrollUp(n) }

// Lifecycle wrappers
func (s *Shell) OnCreated(e UIEntityCreated)  { s.ctrl.OnCreated(e); s.RefreshView(false) }
func (s *Shell) OnUpdated(e UIEntityUpdated)  { s.ctrl.OnUpdated(e); s.RefreshView(false) }
func (s *Shell) OnCompleted(e UIEntityCompleted) { s.ctrl.OnCompleted(e); s.RefreshView(false) }
func (s *Shell) OnDeleted(e UIEntityDeleted)  { s.ctrl.OnDeleted(e); s.RefreshView(false) }

// Selection helpers and routing
func (s *Shell) SelectLast() { s.ctrl.SelectLast(); s.RefreshView(false) }
func (s *Shell) SelectNext() { s.ctrl.SelectNext(); s.ScrollToSelected() }
func (s *Shell) SelectPrev() { s.ctrl.SelectPrev(); s.ScrollToSelected() }

func (s *Shell) EnterSelection() { s.ctrl.EnterSelection(); s.RefreshView(false) }
func (s *Shell) ExitSelection()  { s.ctrl.ExitSelection(); s.RefreshView(false) }
func (s *Shell) IsEntering() bool { return s.ctrl.IsEntering() }

func (s *Shell) SetSelectionVisible(v bool) { s.ctrl.SetSelectionVisible(v); s.RefreshView(false) }
func (s *Shell) Unselect()                  { s.ctrl.Unselect(); s.RefreshView(false) }

func (s *Shell) HandleMsg(msg tea.Msg) tea.Cmd {
    cmd := s.ctrl.HandleMsg(msg)
    s.RefreshView(false)
    return cmd
}

func (s *Shell) SendToSelected(msg tea.Msg) tea.Cmd { return s.ctrl.SendToSelected(msg) }

func (s *Shell) ViewAndSelectedPosition() (string, int, int) { return s.ctrl.ViewAndSelectedPosition() }
func (s *Shell) SelectedIndex() int                        { return s.ctrl.SelectedIndex() }

// ScrollToSelected mirrors the computation used in Chat model to keep selection in view.
func (s *Shell) ScrollToSelected() {
    v, off, h := s.ctrl.ViewAndSelectedPosition()
    s.viewport.SetContent(v)

    midScreenOffset := s.viewport.YOffset + s.viewport.Height/2
    msgEndOffset := off + h
    bottomOffset := s.viewport.YOffset + s.viewport.Height

    if off > midScreenOffset && msgEndOffset > bottomOffset {
        newOffset := off - max(s.viewport.Height-h-1, s.viewport.Height/2)
        s.viewport.SetYOffset(newOffset)
        log.Trace().Int("new_y_offset", newOffset).Msg("Shell: scrolled down to show entity")
    } else if off < s.viewport.YOffset {
        s.viewport.SetYOffset(off)
        log.Trace().Int("new_y_offset", off).Msg("Shell: scrolled up to show entity")
    }
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

// Helpers for querying last assistant response
func (s *Shell) GetLastLLMByRole(role string) (EntityID, map[string]any, bool) {
    return s.ctrl.GetLastLLMByRole(role)
}

// Version helpers if callers need monotonic updates; callers can pass time.Now().UnixNano()
func VersionNow() int64 { return time.Now().UnixNano() }


