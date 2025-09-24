package diff

import (
	"fmt"
 	"sort"
 	"strings"

 	"github.com/charmbracelet/bubbles/list"
 	"github.com/charmbracelet/bubbles/textinput"
 	"github.com/charmbracelet/bubbles/viewport"
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
 	provider      DataProvider
 	config        Config
 	styles        Styles

 	list          list.Model
 	detail        viewport.Model
 	searchInput   textinput.Model

 	width         int
 	height        int
 	leftWidth     int
 	rightWidth    int
	bodyHeight    int
	headerHeight  int
	footerHeight  int

 	focus         focus
 	redacted      bool
 	splitRatio    float64
	statusFilter  StatusFilter
	filtersOn     bool

 	items         []DiffItem
 	visibleItems  []DiffItem
 	searchQuery   string
 	showSearch    bool
}

// NewModel creates a new diff model with the given provider and configuration.
func NewModel(provider DataProvider, config Config) Model {
    return NewModelWith(provider, config)
}

// NewModelWith creates a new diff model and applies optional Config options.
func NewModelWith(provider DataProvider, config Config, options ...Option) Model {
    for _, opt := range options { opt(&config) }
 	styles := defaultStyles()

 	items := provider.Items()
 	wrapped := make([]list.Item, len(items))
 	for i := range items {
 		wrapped[i] = itemAdapter{item: items[i]}
 	}

 	l := list.New(wrapped, list.NewDefaultDelegate(), 0, 0)
 	l.Title = "Items"
 	l.SetShowHelp(false)
 	l.SetFilteringEnabled(false)
 	l.SetShowPagination(false)
 	l.DisableQuitKeybindings()

 	input := textinput.New()
 	input.Placeholder = "Search"
 	input.Prompt = "/ "
 	input.CharLimit = 0
 	input.Focus()

	m := Model{
 		provider:     provider,
 		config:       config,
 		styles:       styles,
 		list:         l,
 		detail:       viewport.New(0, 0),
 		searchInput:  input,
 		focus:        focusList,
 		redacted:     config.RedactSensitive,
 		splitRatio:   nonZeroOr(config.SplitPaneRatio, 0.35),
		statusFilter: config.InitialFilter,
		filtersOn:    config.EnableStatusFilters,
 		items:        items,
 		visibleItems: filterItems(items, ""),
 		searchQuery:  "",
		showSearch:   false,
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
 		//nolint:exhaustive
 		switch msg.String() {
 		case "ctrl+c", "q":
 			return m, tea.Quit
 		case "tab":
 			if m.focus == focusList {
 				m.focus = focusDetail
 			} else {
 				m.focus = focusList
 			}
		case "/":
            if !m.config.EnableSearch { break }
 			m.showSearch = true
 			m.focus = focusSearch
 			m.searchInput.Focus()
 		case "esc":
 			if m.focus == focusSearch {
 				m.focus = focusList
 				m.showSearch = false
 				m.searchQuery = ""
 				m.visibleItems = filterItems(m.items, m.searchQuery)
 				m.resetListItems()
 			}
 		case "r":
 			m.redacted = !m.redacted
 			m.updateDetailContent()
		case "1":
			if m.filtersOn { m.statusFilter.ShowAdded = !m.statusFilter.ShowAdded; m.updateDetailContent() }
		case "2":
			if m.filtersOn { m.statusFilter.ShowRemoved = !m.statusFilter.ShowRemoved; m.updateDetailContent() }
		case "3":
			if m.filtersOn { m.statusFilter.ShowUpdated = !m.statusFilter.ShowUpdated; m.updateDetailContent() }
 		}

 		if m.focus == focusSearch {
 			var cmd tea.Cmd
 			m.searchInput, cmd = m.searchInput.Update(msg)
 			cmds = append(cmds, cmd)

			q := strings.TrimSpace(m.searchInput.Value())
 			if q != m.searchQuery {
 				m.searchQuery = q
 				m.visibleItems = filterItems(m.items, q)
 				m.resetListItems()
 			}
 		} else if m.focus == focusList {
 			var cmd tea.Cmd
 			m.list, cmd = m.list.Update(msg)
 			cmds = append(cmds, cmd)
 			m.updateDetailContent()
 		} else if m.focus == focusDetail {
 			var cmd tea.Cmd
 			m.detail, cmd = m.detail.Update(msg)
 			cmds = append(cmds, cmd)
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

	search := ""
	if m.showSearch {
		search = m.searchInput.View()
	}

	meta := fmt.Sprintf("  [tab] switch  [/] search  [r] redact  [q] quit")

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.styles.Title.Render(" "+title+" "),
		" ",
		search,
		" ",
		meta,
	)
}

func (m *Model) renderList() string {
 	style := m.styles.ListBase
 	if m.focus == focusList {
 		style = m.styles.ListFocused
 	}
	content := m.list.View()
	return style.Width(m.leftWidth).Height(m.bodyHeight).Render(content)
}

func (m *Model) renderDetail() string {
 	style := m.styles.DetailBase
 	if m.focus == focusDetail {
 		style = m.styles.DetailFocused
 	}
	content := m.detail.View()
	return style.Width(m.rightWidth).Height(m.bodyHeight).Render(content)
}

func (m *Model) resetListItems() {
 	wrapped := make([]list.Item, len(m.visibleItems))
 	for i := range m.visibleItems {
 		wrapped[i] = itemAdapter{item: m.visibleItems[i]}
 	}
 	m.list.SetItems(wrapped)
 	m.list.Select(0)
}

func (m *Model) updateDetailContent() {
 	idx := m.list.Index()
 	if idx < 0 || idx >= len(m.visibleItems) {
 		m.detail.SetContent("")
 		return
 	}
 	item := m.visibleItems[idx]
    content := renderItemDetail(item, m.redacted, m.styles, m.searchQuery, m.statusFilter, m.filtersOn)
 	m.detail.SetContent(content)
 	m.detail.GotoTop()
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

func maxInt(a, b int) int {
 	if a > b {
 		return a
 	}
 	return b
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
	m.detail.Width = rightContentW
	m.detail.Height = rightContentH
}

// renderFooter returns a simple help line footer.
func (m *Model) renderFooter() string {
    help := "↑/↓ move  tab switch  / search  r redact  1/2/3 filter +/−/~  q quit"
	return lipgloss.NewStyle().Faint(true).Render(help)
}

// itemAdapter adapts DiffItem to bubbles/list.Item
type itemAdapter struct{ item DiffItem }

func (i itemAdapter) Title() string { return i.item.Name() }
func (i itemAdapter) Description() string {
 	// Optional: basic counts of changes
 	total := 0
 	for _, c := range i.item.Categories() {
 		if c == nil {
 			continue
 		}
 		total += len(c.Changes())
 	}
 	if total == 0 {
 		return ""
 	}
 	return fmt.Sprintf("%d changes", total)
}
func (i itemAdapter) FilterValue() string { return i.item.Name() }


