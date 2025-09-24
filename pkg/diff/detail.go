package diff

import (
    "github.com/charmbracelet/bubbles/viewport"
)

type detailModel struct { viewport viewport.Model }

func newDetailModel() detailModel {
    return detailModel{ viewport: viewport.New(0, 0) }
}

func (m *detailModel) SetSize(width, height int) {
    if width < 0 { width = 0 }
    if height < 0 { height = 0 }
    m.viewport.Width = width
    m.viewport.Height = height
}

func (m *detailModel) SetContent(content string) {
    m.viewport.SetContent(content)
    m.viewport.GotoTop()
}

func (m *detailModel) View() string { return m.viewport.View() }


