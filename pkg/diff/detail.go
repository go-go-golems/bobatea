package diff

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// detailModel wraps a viewport for the detail panel
type detailModel struct {
	viewport viewport.Model
}

// newDetailModel creates a new detail model with a viewport
func newDetailModel() detailModel {
	vp := viewport.New(0, 0)
	return detailModel{viewport: vp}
}

// SetSize sets the size of the detail viewport
func (d *detailModel) SetSize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	d.viewport.Width = width
	d.viewport.Height = height
}

// SetContent sets the content of the detail viewport and goes to top
func (d *detailModel) SetContent(content string) {
	d.viewport.SetContent(content)
	d.viewport.GotoTop()
}

// Update handles tea messages for the detail viewport
func (d *detailModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return cmd
}

// View renders the detail viewport
func (d *detailModel) View() string {
	return d.viewport.View()
}
