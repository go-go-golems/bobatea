package diff

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focus int

const (
	focusList focus = iota
	focusDetail
	focusSearch
)

// Model implements the Bubble Tea model for the diff component.
type Model struct {
	provider DataProvider
	config   Config
	styles   Styles
	keys     keyMap

	list   list.Model
	detail detailModel
	search searchModel

	width        int
	height       int
	leftWidth    int
	rightWidth   int
	bodyHeight   int
	headerHeight int
	footerHeight int

	focus        focus
	redacted     bool
	splitRatio   float64
	statusFilter StatusFilter
	filtersOn    bool

	items        []DiffItem
	visibleItems []DiffItem
}

// NewModel creates a new diff model with the given provider and configuration.
func NewModel(provider DataProvider, config Config) Model {
	return NewModelWith(provider, config)
}

// NewModelWith creates a new diff model and applies optional Config options.
func NewModelWith(provider DataProvider, config Config, options ...Option) Model {
	for _, opt := range options {
		opt(&config)
	}
	styles := defaultStyles()
	keys := newKeyMap()

	items := provider.Items()
	l := newItemList(items, styles)
	search := newSearchModel()
	detail := newDetailModel()

	m := Model{
		provider:     provider,
		config:       config,
		styles:       styles,
		keys:         keys,
		list:         l,
		detail:       detail,
		search:       search,
		focus:        focusList,
		redacted:     config.RedactSensitive,
		splitRatio:   nonZeroOr(config.SplitPaneRatio, 0.35),
		statusFilter: config.InitialFilter,
		filtersOn:    config.EnableStatusFilters,
		items:        items,
		visibleItems: filterItems(items, ""),
	}

	// Initialize list content to visible items
	m.resetListItems()
	m.updateDetailContent()

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.computeLayout()
		m.applyContentSizes()
		m.updateDetailContent()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			if m.focus == focusList {
				m.focus = focusDetail
			} else {
				m.focus = focusList
			}
		case key.Matches(msg, m.keys.Search):
			if !m.config.EnableSearch {
				break
			}
			m.search.Show()
			m.focus = focusSearch
			// Recompute layout to account for search widget height
			m.computeLayout()
			m.applyContentSizes()
			return m, tea.Batch(cmds...)
		case key.Matches(msg, m.keys.Escape):
			if m.focus == focusSearch {
				m.focus = focusList
				m.search.Hide()
				m.visibleItems = filterItems(m.items, m.search.Query())
				m.resetListItems()
				// Recompute layout after hiding search
				m.computeLayout()
				m.applyContentSizes()
			}
		case key.Matches(msg, m.keys.ToggleRedact):
			m.redacted = !m.redacted
			m.updateDetailContent()
		case key.Matches(msg, m.keys.FilterAdded):
			if m.filtersOn {
				m.statusFilter.ShowAdded = !m.statusFilter.ShowAdded
				m.updateDetailContent()
			}
		case key.Matches(msg, m.keys.FilterRemoved):
			if m.filtersOn {
				m.statusFilter.ShowRemoved = !m.statusFilter.ShowRemoved
				m.updateDetailContent()
			}
		case key.Matches(msg, m.keys.FilterUpdated):
			if m.filtersOn {
				m.statusFilter.ShowUpdated = !m.statusFilter.ShowUpdated
				m.updateDetailContent()
			}
		}

		if m.focus == focusSearch {
			cmd := m.search.Update(msg)
			cmds = append(cmds, cmd)

			q := m.search.Query()
			m.visibleItems = filterItems(m.items, q)
			m.resetListItems()
			m.updateDetailContent()
		} else {
			switch m.focus {
			case focusList:
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
				m.updateDetailContent()
			case focusDetail:
				cmd := m.detail.Update(msg)
				cmds = append(cmds, cmd)
			case focusSearch:
				// no-op: handled above
			}
		}

	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	left := m.renderList()
	right := m.renderDetail()
	head := m.renderHeader()
	foot := m.renderFooter()
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return lipgloss.JoinVertical(lipgloss.Left, head, row, foot)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.computeLayout()
	m.applyContentSizes()
	m.updateDetailContent()
}

func (m *Model) SetRedactSensitive(enabled bool) {
	m.redacted = enabled
	m.updateDetailContent()
}

func (m *Model) SetSplitPaneRatio(ratio float64) {
	if ratio <= 0 {
		ratio = 0.35
	}
	m.splitRatio = ratio
	m.computeLayout()
}

// Helpers

func (m *Model) computeLayout() {
	if m.width < 0 {
		m.width = 0
	}
	if m.height < 0 {
		m.height = 0
	}
	m.leftWidth = int(float64(m.width) * m.splitRatio)
	if m.leftWidth < 20 {
		m.leftWidth = 20
	}
	m.rightWidth = m.width - m.leftWidth
	if m.rightWidth < 20 {
		m.rightWidth = 20
	}

	// Compute header/footer heights dynamically to avoid clipping content
	head := m.renderHeader()
	m.headerHeight = lipgloss.Height(head)
	if m.headerHeight < 0 {
		m.headerHeight = 0
	}
	foot := m.renderFooter()
	m.footerHeight = lipgloss.Height(foot)
	if m.footerHeight < 0 {
		m.footerHeight = 0
	}
	m.bodyHeight = m.height - m.headerHeight - m.footerHeight
	// Reserve an extra safety line to avoid top border clipping in tight layouts
	if m.bodyHeight > 0 {
		m.bodyHeight--
	}
	if m.bodyHeight < 0 {
		m.bodyHeight = 0
	}
}

func (m *Model) renderHeader() string {
	title := m.config.Title
	if title == "" && m.provider != nil {
		title = m.provider.Title()
	}
	if title == "" {
		title = "Diff"
	}

	lines := []string{m.styles.Title.Render(" " + title + " ")}
	if m.search.Visible() {
		// Show search input on its own line to ensure visibility
		lines = append(lines, m.search.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *Model) renderList() string {
	style := m.styles.ListBase
	if m.focus == focusList {
		style = m.styles.ListFocused
	}
	content := m.list.View()
	return style.Render(content)
}

func (m *Model) renderDetail() string {
	style := m.styles.DetailBase
	if m.focus == focusDetail {
		style = m.styles.DetailFocused
	}
	content := m.detail.View()
	return style.Render(content)
}

func (m *Model) resetListItems() {
	setListItems(&m.list, m.visibleItems)
}

func (m *Model) updateDetailContent() {
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.visibleItems) {
		m.detail.SetContent("")
		return
	}
	item := m.visibleItems[idx]
	content := renderItemDetail(item, m.redacted, m.styles, m.search.Query(), m.statusFilter, m.filtersOn)
	m.detail.SetContent(content)
}

func filterItems(items []DiffItem, query string) []DiffItem {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower == "" {
		out := append([]DiffItem(nil), items...)
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Name() < out[j].Name()
		})
		return out
	}
	var filtered []DiffItem
	for _, it := range items {
		if itemMatchesQuery(it, lower) {
			filtered = append(filtered, it)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Name() < filtered[j].Name()
	})
	return filtered
}

func nonZeroOr(v, fallback float64) float64 {
	if v <= 0 {
		return fallback
	}
	return v
}

// applyContentSizes computes content sizes for inner widgets accounting for borders/margins.
func (m *Model) applyContentSizes() {
	// Use base styles for frame sizes (focused/unfocused share same frame metrics)
	leftFrameW, leftFrameH := m.styles.ListBase.GetFrameSize()
	rightFrameW, rightFrameH := m.styles.DetailBase.GetFrameSize()

	leftContentW := m.leftWidth - leftFrameW
	if leftContentW < 0 {
		leftContentW = 0
	}
	leftContentH := m.bodyHeight - leftFrameH
	if leftContentH < 0 {
		leftContentH = 0
	}

	rightContentW := m.rightWidth - rightFrameW
	if rightContentW < 0 {
		rightContentW = 0
	}
	rightContentH := m.bodyHeight - rightFrameH
	if rightContentH < 0 {
		rightContentH = 0
	}

	m.list.SetSize(leftContentW, leftContentH)
	m.detail.SetSize(rightContentW, rightContentH)
	// Ensure search input has reasonable width when visible
	m.search.SetWidth(m.width)
}

// renderFooter returns a simple help line footer.
func (m *Model) renderFooter() string {
	help := "↑/↓ move  tab switch  / search  r redact  1/2/3 filter +/−/~  q quit"
	return lipgloss.NewStyle().Faint(true).Render(help)
}

// (moved to list.go)
