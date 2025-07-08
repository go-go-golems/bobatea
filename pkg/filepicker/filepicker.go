package filepicker

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Messages for compatibility with existing bobatea filepicker API
type SelectFileMsg struct {
	Path string
}

type CancelFilePickerMsg struct{}

// File represents a file or directory with extended metadata
type File struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	Mode     os.FileMode
	Selected bool
	Hidden   bool
}

// ViewState represents the current state of the file picker
type ViewState int

const (
	ViewStateNormal ViewState = iota
	ViewStateConfirmDelete
	ViewStateRename
	ViewStateCreateFile
	ViewStateCreateDir
	ViewStateSearch
	ViewStateGlob
)

// Operation represents file operations
type Operation int

const (
	OpNone Operation = iota
	OpCopy
	OpCut
)

// SortMode represents different sorting options
type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
	SortByDate
	SortByType
)

// AdvancedModel represents the advanced file picker model
type AdvancedModel struct {
	currentPath   string
	files         []File
	filteredFiles []File
	cursor        int
	selectedFiles []string // For final selection
	width         int
	height        int
	cancelled     bool
	showIcons     bool
	showSizes     bool
	err           error

	// Multi-selection state
	multiSelected map[string]bool

	// Operations
	clipboard   []string
	clipboardOp Operation

	// UI state
	viewState    ViewState
	confirmFiles []string
	textInput    textinput.Model
	searchInput  textinput.Model
	globInput    textinput.Model
	help         help.Model
	keys         advancedKeyMap

	// Tier 4 features
	showPreview    bool
	showHidden     bool
	detailedView   bool
	sortMode       SortMode
	searchQuery    string
	globPattern    string
	previewContent string
	previewWidth   int

	// Navigation history
	history        []string // Stack of visited directories
	historyIndex   int      // Current position in history (-1 means at the end)
	maxHistorySize int      // Maximum history entries to keep

	// Directory selection mode
	directorySelectionMode bool

	// Directory restriction (jail)
	jailDirectory string // Absolute path of the jail directory, empty means no restriction
}

// advancedKeyMap defines the key bindings for the advanced file picker
type advancedKeyMap struct {
	// Navigation
	Up   key.Binding
	Down key.Binding
	Home key.Binding
	End  key.Binding

	// Selection
	Enter          key.Binding
	Space          key.Binding
	SelectAll      key.Binding
	DeselectAll    key.Binding
	SelectAllFiles key.Binding

	// File operations
	Delete  key.Binding
	Copy    key.Binding
	Cut     key.Binding
	Paste   key.Binding
	Rename  key.Binding
	NewFile key.Binding
	NewDir  key.Binding

	// Navigation
	Escape    key.Binding
	Backspace key.Binding
	Refresh   key.Binding
	Back      key.Binding
	Forward   key.Binding

	// Tier 4 features
	TogglePreview key.Binding
	Search        key.Binding
	Glob          key.Binding
	ClearGlob     key.Binding
	ToggleHidden  key.Binding
	ToggleDetail  key.Binding
	CycleSort     key.Binding

	// Directory selection
	SelectCurrentDir   key.Binding
	ToggleDirSelection key.Binding

	// System
	Help key.Binding
	Quit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k advancedKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k advancedKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Home, k.End},
		{k.Enter, k.Space, k.SelectAll, k.DeselectAll},
		{k.Copy, k.Cut, k.Paste, k.Delete},
		{k.Rename, k.NewFile, k.NewDir, k.Refresh},
		{k.TogglePreview, k.Search, k.Glob, k.ClearGlob, k.ToggleHidden, k.ToggleDetail},
		{k.CycleSort, k.Backspace, k.Back, k.Forward},
		{k.SelectCurrentDir, k.ToggleDirSelection},
		{k.Escape, k.Help, k.Quit},
	}
}

// FullHelpForMode returns keybindings for the expanded help view based on the current mode
func (fp *AdvancedModel) FullHelpForMode() [][]key.Binding {
	if fp.directorySelectionMode {
		// Update the Space key help text for directory mode
		spaceKey := key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select directory"),
		)

		return [][]key.Binding{
			{fp.keys.Up, fp.keys.Down, fp.keys.Home, fp.keys.End},
			{fp.keys.Enter, spaceKey, fp.keys.SelectAll, fp.keys.DeselectAll},
			{fp.keys.Copy, fp.keys.Cut, fp.keys.Paste, fp.keys.Delete},
			{fp.keys.Rename, fp.keys.NewFile, fp.keys.NewDir, fp.keys.Refresh},
			{fp.keys.TogglePreview, fp.keys.Search, fp.keys.Glob, fp.keys.ClearGlob, fp.keys.ToggleHidden, fp.keys.ToggleDetail},
			{fp.keys.CycleSort, fp.keys.Backspace, fp.keys.Back, fp.keys.Forward},
			{fp.keys.SelectCurrentDir, fp.keys.ToggleDirSelection},
			{fp.keys.Escape, fp.keys.Help, fp.keys.Quit},
		}
	} else {
		// Standard file mode help
		return fp.keys.FullHelp()
	}
}

// defaultAdvancedKeyMap returns the default key bindings for the advanced file picker
func defaultAdvancedKeyMap() advancedKeyMap {
	return advancedKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "first"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "last"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "navigate/select"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select item"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		DeselectAll: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "deselect all"),
		),
		SelectAllFiles: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "select all items"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),
		Cut: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "cut"),
		),
		Paste: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "paste"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		NewFile: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new file"),
		),
		NewDir: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "new directory"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "up directory"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("f5"),
			key.WithHelp("f5", "refresh"),
		),
		Back: key.NewBinding(
			key.WithKeys("alt+left", "h"),
			key.WithHelp("alt+←/h", "back"),
		),
		Forward: key.NewBinding(
			key.WithKeys("alt+right", "l"),
			key.WithHelp("alt+→/l", "forward"),
		),
		TogglePreview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle preview"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Glob: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "glob filter"),
		),
		ClearGlob: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "clear glob"),
		),
		ToggleHidden: key.NewBinding(
			key.WithKeys("f2"),
			key.WithHelp("f2", "toggle hidden"),
		),
		ToggleDetail: key.NewBinding(
			key.WithKeys("f3"),
			key.WithHelp("f3", "toggle details"),
		),
		CycleSort: key.NewBinding(
			key.WithKeys("f4"),
			key.WithHelp("f4", "cycle sort"),
		),
		SelectCurrentDir: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "select current directory"),
		),
		ToggleDirSelection: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle directory selection mode"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// Styles for Tier 4
var (
	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230"))

	multiSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("33")).
				Foreground(lipgloss.Color("230"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	dirSelectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Background(lipgloss.Color("17")).
				Bold(true)

	hiddenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	previewTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true)

	searchStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("99"))

	confirmStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("255")).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// Option represents a configuration option for the file picker
type Option func(*AdvancedModel)

// WithStartPath sets the starting directory path
func WithStartPath(path string) Option {
	return func(fp *AdvancedModel) {
		fp.currentPath = path
	}
}

// WithShowPreview sets whether to show file preview
func WithShowPreview(show bool) Option {
	return func(fp *AdvancedModel) {
		fp.showPreview = show
	}
}

// WithShowHidden sets whether to show hidden files
func WithShowHidden(show bool) Option {
	return func(fp *AdvancedModel) {
		fp.showHidden = show
	}
}

// WithCompatibilityMode enables compatibility mode (no additional effect currently)
func WithCompatibilityMode(compat bool) Option {
	return func(fp *AdvancedModel) {
		// This is a placeholder for future compatibility mode features
		// Currently, the Model type already provides compatibility wrapping
	}
}

// WithShowIcons sets whether to show file icons
func WithShowIcons(show bool) Option {
	return func(fp *AdvancedModel) {
		fp.showIcons = show
	}
}

// WithShowSizes sets whether to show file sizes
func WithShowSizes(show bool) Option {
	return func(fp *AdvancedModel) {
		fp.showSizes = show
	}
}

// WithDetailedView sets whether to show detailed file information
func WithDetailedView(detailed bool) Option {
	return func(fp *AdvancedModel) {
		fp.detailedView = detailed
	}
}

// WithSortMode sets the initial sort mode
func WithSortMode(mode SortMode) Option {
	return func(fp *AdvancedModel) {
		fp.sortMode = mode
	}
}

// WithPreviewWidth sets the width of the preview panel
func WithPreviewWidth(width int) Option {
	return func(fp *AdvancedModel) {
		fp.previewWidth = width
	}
}

// WithMaxHistorySize sets the maximum number of history entries
func WithMaxHistorySize(size int) Option {
	return func(fp *AdvancedModel) {
		fp.maxHistorySize = size
	}
}

// WithDirectorySelection enables or disables directory selection mode
func WithDirectorySelection(enabled bool) Option {
	return func(fp *AdvancedModel) {
		fp.directorySelectionMode = enabled
	}
}

// WithGlobPattern sets the initial glob pattern filter
func WithGlobPattern(pattern string) Option {
	return func(fp *AdvancedModel) {
		fp.globPattern = pattern
	}
}

// WithJailDirectory sets a directory restriction boundary - navigation will be limited to this directory and subdirectories
func WithJailDirectory(path string) Option {
	return func(fp *AdvancedModel) {
		if absPath, err := filepath.Abs(path); err == nil {
			fp.jailDirectory = absPath
		}
	}
}

// New creates a new file picker with the specified options
func New(options ...Option) *AdvancedModel {
	// Get current working directory as default
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	ti := textinput.New()
	ti.Placeholder = "Enter name..."
	ti.CharLimit = 255

	si := textinput.New()
	si.Placeholder = "Search files..."
	si.CharLimit = 100

	gi := textinput.New()
	gi.Placeholder = "Enter glob pattern (e.g., *.go, test_*)..."
	gi.CharLimit = 100

	fp := &AdvancedModel{
		currentPath:    wd,
		showIcons:      true,
		showSizes:      true,
		multiSelected:  make(map[string]bool),
		clipboard:      []string{},
		clipboardOp:    OpNone,
		viewState:      ViewStateNormal,
		textInput:      ti,
		searchInput:    si,
		globInput:      gi,
		help:           help.New(),
		keys:           defaultAdvancedKeyMap(),
		showPreview:    true,
		showHidden:     false,
		detailedView:   true,
		sortMode:       SortByName,
		previewWidth:   40,
		history:        make([]string, 0),
		historyIndex:   -1,
		maxHistorySize: 50,
	}

	// Apply options
	for _, option := range options {
		option(fp)
	}

	// Resolve the starting path
	if absPath, err := filepath.Abs(fp.currentPath); err == nil {
		fp.currentPath = absPath
	}

	// Add initial directory to history
	fp.addToHistory(fp.currentPath)

	// Validate starting path against jail directory if set
	if fp.jailDirectory != "" {
		if !fp.isWithinJail(fp.currentPath) {
			// If current path is outside jail, move to jail directory
			fp.currentPath = fp.jailDirectory
			fp.addToHistory(fp.currentPath)
		}
	}

	fp.loadDirectory()
	return fp
}

// NewAdvancedModel creates a new advanced file picker
// Deprecated: Use New(WithStartPath(startPath)) instead
func NewAdvancedModel(startPath string) *AdvancedModel {
	return New(WithStartPath(startPath))
}

// isWithinJail checks if the given path is within the jail directory
func (fp *AdvancedModel) isWithinJail(path string) bool {
	if fp.jailDirectory == "" {
		return true // No jail restriction
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Clean paths to handle .. and . components
	cleanJail := filepath.Clean(fp.jailDirectory)
	cleanPath := filepath.Clean(absPath)

	// Check if path is within jail (must be equal or a subdirectory)
	return cleanPath == cleanJail || strings.HasPrefix(cleanPath+string(filepath.Separator), cleanJail+string(filepath.Separator))
}

// isAtJailRoot checks if the current path is at the jail root
func (fp *AdvancedModel) isAtJailRoot() bool {
	if fp.jailDirectory == "" {
		return false // No jail restriction
	}

	cleanJail := filepath.Clean(fp.jailDirectory)
	cleanCurrent := filepath.Clean(fp.currentPath)
	return cleanJail == cleanCurrent
}

// validateNavigationPath checks if navigation to the given path is allowed
func (fp *AdvancedModel) validateNavigationPath(path string) bool {
	if fp.jailDirectory == "" {
		return true // No jail restriction
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	return fp.isWithinJail(absPath)
}

// CompatFilepicker provides compatibility with the old bubbles filepicker interface
type CompatFilepicker struct {
	DirAllowed       bool
	FileAllowed      bool
	CurrentDirectory string
	Height           int
	advancedModel    *AdvancedModel
}

// Model provides backward compatibility with the original bobatea filepicker API
type Model struct {
	*AdvancedModel
	sentCancelMsg bool
	sentSelectMsg bool

	// Compatibility fields
	Title        string
	Error        string
	Filepicker   CompatFilepicker
	SelectedPath string
}

// NewModel creates a new file picker with backward compatibility
// This maintains the original bobatea API while using the advanced implementation
func NewModel() Model {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	advModel := New(WithStartPath(wd))

	return Model{
		AdvancedModel: advModel,
		Filepicker: CompatFilepicker{
			DirAllowed:       true,
			FileAllowed:      true,
			CurrentDirectory: wd,
			Height:           10,
			advancedModel:    advModel,
		},
	}
}

// NewModelWithOptions creates a new file picker with backward compatibility using options
func NewModelWithOptions(options ...Option) Model {
	advModel := New(options...)

	return Model{
		AdvancedModel: advModel,
		Filepicker: CompatFilepicker{
			DirAllowed:       true,
			FileAllowed:      true,
			CurrentDirectory: advModel.currentPath,
			Height:           10,
			advancedModel:    advModel,
		},
	}
}

// Init initializes the compatibility wrapper
func (m Model) Init() tea.Cmd {
	return m.AdvancedModel.Init()
}

// Update handles updates for the compatibility wrapper
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Sync compatibility settings to advanced model if they changed
	if m.Filepicker.CurrentDirectory != m.currentPath {
		m.currentPath = m.Filepicker.CurrentDirectory
		m.loadDirectory()
	}

	// Delegate to the advanced model
	updatedAdvanced, cmd := m.AdvancedModel.Update(msg)

	// Update our wrapper
	m.AdvancedModel = updatedAdvanced.(*AdvancedModel)

	// Sync back to compatibility fields
	m.Filepicker.CurrentDirectory = m.currentPath
	m.Filepicker.advancedModel = m.AdvancedModel

	// Check if we need to send compatibility messages
	var compatCmd tea.Cmd

	if m.cancelled && !m.sentCancelMsg {
		m.sentCancelMsg = true
		compatCmd = func() tea.Msg {
			return CancelFilePickerMsg{}
		}
	} else if len(m.selectedFiles) > 0 && !m.sentSelectMsg {
		m.sentSelectMsg = true
		m.SelectedPath = m.selectedFiles[0]
		compatCmd = func() tea.Msg {
			return SelectFileMsg{Path: m.selectedFiles[0]}
		}
	}

	if compatCmd != nil {
		return m, tea.Batch(cmd, compatCmd)
	}

	return m, cmd
}

// View renders the compatibility wrapper
func (m Model) View() string {
	return m.AdvancedModel.View()
}

// Init initializes the file picker
func (fp *AdvancedModel) Init() tea.Cmd {
	return nil
}

// addToHistory adds a directory to the navigation history
func (fp *AdvancedModel) addToHistory(path string) {
	// Don't add paths outside jail to history
	if !fp.isWithinJail(path) {
		return
	}

	// If we're in the middle of history (user went back), truncate forward history
	if fp.historyIndex >= 0 && fp.historyIndex < len(fp.history)-1 {
		fp.history = fp.history[:fp.historyIndex+1]
	}

	// Don't add duplicate consecutive entries
	if len(fp.history) > 0 && fp.history[len(fp.history)-1] == path {
		return
	}

	// Add to history
	fp.history = append(fp.history, path)

	// Limit history size
	if len(fp.history) > fp.maxHistorySize {
		fp.history = fp.history[1:]
	}

	// Reset history index to end
	fp.historyIndex = -1
}

// canGoBack returns true if we can navigate back in history
func (fp *AdvancedModel) canGoBack() bool {
	if fp.historyIndex == -1 {
		return len(fp.history) > 1
	}
	return fp.historyIndex > 0
}

// canGoForward returns true if we can navigate forward in history
func (fp *AdvancedModel) canGoForward() bool {
	return fp.historyIndex >= 0 && fp.historyIndex < len(fp.history)-1
}

// goBack navigates to the previous directory in history
func (fp *AdvancedModel) goBack() {
	if !fp.canGoBack() {
		return
	}

	if fp.historyIndex == -1 {
		fp.historyIndex = len(fp.history) - 2
	} else {
		fp.historyIndex--
	}

	fp.navigateToHistoryIndex()
}

// goForward navigates to the next directory in history
func (fp *AdvancedModel) goForward() {
	if !fp.canGoForward() {
		return
	}

	fp.historyIndex++
	fp.navigateToHistoryIndex()
}

// navigateToHistoryIndex navigates to the directory at the current history index
func (fp *AdvancedModel) navigateToHistoryIndex() {
	if fp.historyIndex < 0 || fp.historyIndex >= len(fp.history) {
		return
	}

	targetPath := fp.history[fp.historyIndex]

	// Validate against jail directory
	if !fp.isWithinJail(targetPath) {
		return
	}

	fp.currentPath = targetPath
	fp.cursor = 0
	fp.multiSelected = make(map[string]bool)
	fp.searchQuery = ""
	fp.globPattern = ""
	fp.loadDirectory()
}

// Update handles messages for Tier 4
func (fp *AdvancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fp.width = msg.Width
		fp.height = msg.Height
		fp.help.Width = msg.Width

	case tea.KeyMsg:
		switch fp.viewState {
		case ViewStateNormal:
			return fp.updateNormal(msg)
		case ViewStateConfirmDelete:
			return fp.updateConfirmDelete(msg)
		case ViewStateRename, ViewStateCreateFile, ViewStateCreateDir:
			return fp.updateTextInput(msg)
		case ViewStateSearch:
			return fp.updateSearch(msg)
		case ViewStateGlob:
			return fp.updateGlob(msg)
		}
	}

	return fp, cmd
}

// updateNormal handles normal view state with Tier 4 features
func (fp *AdvancedModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, fp.keys.Quit):
		fp.cancelled = true
		return fp, tea.Quit

	case key.Matches(msg, fp.keys.Escape):
		if fp.searchQuery != "" {
			fp.searchQuery = ""
			fp.filterFiles()
		} else if fp.globPattern != "" {
			fp.globPattern = ""
			fp.filterFiles()
		} else {
			fp.cancelled = true
			return fp, tea.Quit
		}

	case key.Matches(msg, fp.keys.Help):
		fp.help.ShowAll = !fp.help.ShowAll

	case key.Matches(msg, fp.keys.TogglePreview):
		fp.showPreview = !fp.showPreview

	case key.Matches(msg, fp.keys.Search):
		fp.searchInput.SetValue("")
		fp.searchInput.Focus()
		fp.viewState = ViewStateSearch
		return fp, textinput.Blink

	case key.Matches(msg, fp.keys.Glob):
		fp.globInput.SetValue(fp.globPattern)
		fp.globInput.Focus()
		fp.viewState = ViewStateGlob
		return fp, textinput.Blink

	case key.Matches(msg, fp.keys.ClearGlob):
		fp.globPattern = ""
		fp.filterFiles()

	case key.Matches(msg, fp.keys.ToggleHidden):
		fp.showHidden = !fp.showHidden
		fp.loadDirectory()

	case key.Matches(msg, fp.keys.ToggleDetail):
		fp.detailedView = !fp.detailedView

	case key.Matches(msg, fp.keys.CycleSort):
		fp.sortMode = (fp.sortMode + 1) % 4
		fp.sortFiles()

	case key.Matches(msg, fp.keys.ToggleDirSelection):
		fp.directorySelectionMode = !fp.directorySelectionMode

	case key.Matches(msg, fp.keys.SelectCurrentDir):
		if fp.directorySelectionMode {
			fp.selectedFiles = []string{fp.currentPath}
			return fp, tea.Quit
		}

	case key.Matches(msg, fp.keys.Up):
		if fp.cursor > 0 {
			fp.cursor--
			fp.updatePreview()
		}

	case key.Matches(msg, fp.keys.Down):
		if fp.cursor < len(fp.filteredFiles)-1 {
			fp.cursor++
			fp.updatePreview()
		}

	case key.Matches(msg, fp.keys.Home):
		fp.cursor = 0
		fp.updatePreview()

	case key.Matches(msg, fp.keys.End):
		if len(fp.filteredFiles) > 0 {
			fp.cursor = len(fp.filteredFiles) - 1
			fp.updatePreview()
		}

	case key.Matches(msg, fp.keys.Space):
		if len(fp.filteredFiles) > 0 {
			file := fp.filteredFiles[fp.cursor]
			if file.Name != ".." {
				// In directory selection mode, only select directories
				// In normal mode, select both files and directories
				if fp.directorySelectionMode {
					// Only select directories in directory selection mode
					if file.IsDir {
						if fp.multiSelected[file.Path] {
							delete(fp.multiSelected, file.Path)
						} else {
							fp.multiSelected[file.Path] = true
						}
					}
				} else {
					// Select any item in normal mode
					if fp.multiSelected[file.Path] {
						delete(fp.multiSelected, file.Path)
					} else {
						fp.multiSelected[file.Path] = true
					}
				}
			}
		}

	case key.Matches(msg, fp.keys.SelectAll):
		for _, file := range fp.filteredFiles {
			if file.Name != ".." {
				fp.multiSelected[file.Path] = true
			}
		}

	case key.Matches(msg, fp.keys.DeselectAll):
		fp.multiSelected = make(map[string]bool)

	case key.Matches(msg, fp.keys.SelectAllFiles):
		for _, file := range fp.filteredFiles {
			if file.Name != ".." {
				// In directory selection mode, only select directories
				// In normal mode, select all items (both files and directories)
				if fp.directorySelectionMode {
					if file.IsDir {
						fp.multiSelected[file.Path] = true
					}
				} else {
					fp.multiSelected[file.Path] = true
				}
			}
		}

	case key.Matches(msg, fp.keys.Enter):
		if len(fp.filteredFiles) > 0 {
			selectedFile := fp.filteredFiles[fp.cursor]
			if selectedFile.IsDir {
				// Always navigate into directories, regardless of mode
				var newPath string
				if selectedFile.Name == ".." {
					newPath = filepath.Dir(fp.currentPath)
				} else {
					newPath = selectedFile.Path
				}
				// Validate navigation path against jail directory
				if fp.validateNavigationPath(newPath) {
					fp.currentPath = newPath
					fp.addToHistory(newPath)
					fp.cursor = 0
					fp.multiSelected = make(map[string]bool)
					fp.searchQuery = ""
					fp.globPattern = ""
					fp.loadDirectory()
				}
			} else {
				// For files, behavior depends on mode
				if fp.directorySelectionMode {
					// In directory selection mode, pressing Enter on a file does nothing
					// Only directories can be selected/navigated
					return fp, nil
				} else {
					// In normal mode, select the file
					if len(fp.multiSelected) > 0 {
						fp.selectedFiles = make([]string, 0, len(fp.multiSelected))
						for path := range fp.multiSelected {
							fp.selectedFiles = append(fp.selectedFiles, path)
						}
					} else {
						fp.selectedFiles = []string{selectedFile.Path}
					}
					return fp, tea.Quit
				}
			}
		}

	case key.Matches(msg, fp.keys.Backspace):
		newPath := filepath.Dir(fp.currentPath)
		// Prevent navigation outside jail directory
		if fp.validateNavigationPath(newPath) {
			fp.currentPath = newPath
			fp.addToHistory(newPath)
			fp.cursor = 0
			fp.multiSelected = make(map[string]bool)
			fp.searchQuery = ""
			fp.globPattern = ""
			fp.loadDirectory()
		}

	case key.Matches(msg, fp.keys.Back):
		fp.goBack()

	case key.Matches(msg, fp.keys.Forward):
		fp.goForward()

	case key.Matches(msg, fp.keys.Refresh):
		fp.loadDirectory()

	case key.Matches(msg, fp.keys.Delete):
		filesToDelete := fp.getSelectedFiles()
		if len(filesToDelete) > 0 {
			fp.confirmFiles = filesToDelete
			fp.viewState = ViewStateConfirmDelete
		}

	case key.Matches(msg, fp.keys.Copy):
		fp.clipboard = fp.getSelectedFiles()
		fp.clipboardOp = OpCopy

	case key.Matches(msg, fp.keys.Cut):
		fp.clipboard = fp.getSelectedFiles()
		fp.clipboardOp = OpCut

	case key.Matches(msg, fp.keys.Paste):
		if len(fp.clipboard) > 0 {
			fp.performPaste()
		}

	case key.Matches(msg, fp.keys.Rename):
		if len(fp.filteredFiles) > 0 && fp.filteredFiles[fp.cursor].Name != ".." {
			fp.textInput.SetValue(fp.filteredFiles[fp.cursor].Name)
			fp.textInput.CursorEnd()
			fp.textInput.Focus()
			fp.viewState = ViewStateRename
			return fp, textinput.Blink
		}

	case key.Matches(msg, fp.keys.NewFile):
		fp.textInput.SetValue("")
		fp.textInput.Focus()
		fp.viewState = ViewStateCreateFile
		return fp, textinput.Blink

	case key.Matches(msg, fp.keys.NewDir):
		fp.textInput.SetValue("")
		fp.textInput.Focus()
		fp.viewState = ViewStateCreateDir
		return fp, textinput.Blink
	}

	return fp, nil
}

// updateSearch handles search input state
func (fp *AdvancedModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter", "esc":
		fp.searchQuery = fp.searchInput.Value()
		fp.searchInput.Blur()
		fp.viewState = ViewStateNormal
		fp.filterFiles()
		if fp.cursor >= len(fp.filteredFiles) {
			fp.cursor = 0
		}
		fp.updatePreview()

	default:
		fp.searchInput, cmd = fp.searchInput.Update(msg)
	}

	return fp, cmd
}

// updateGlob handles glob input state
func (fp *AdvancedModel) updateGlob(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter", "esc":
		fp.globPattern = fp.globInput.Value()
		fp.globInput.Blur()
		fp.viewState = ViewStateNormal
		fp.filterFiles()
		if fp.cursor >= len(fp.filteredFiles) {
			fp.cursor = 0
		}
		fp.updatePreview()

	default:
		fp.globInput, cmd = fp.globInput.Update(msg)
	}

	return fp, cmd
}

// updateConfirmDelete handles delete confirmation
func (fp *AdvancedModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		fp.performDelete()
		fp.viewState = ViewStateNormal
	case "n", "N", "esc":
		fp.viewState = ViewStateNormal
	}
	return fp, nil
}

// updateTextInput handles text input states
func (fp *AdvancedModel) updateTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(fp.textInput.Value())
		if name != "" {
			switch fp.viewState {
			case ViewStateRename:
				fp.performRename(name)
			case ViewStateCreateFile:
				fp.performCreateFile(name)
			case ViewStateCreateDir:
				fp.performCreateDir(name)
			case ViewStateNormal, ViewStateConfirmDelete, ViewStateSearch, ViewStateGlob:
				// These states shouldn't be handled here
			}
		}
		fp.textInput.Blur()
		fp.viewState = ViewStateNormal

	case "esc":
		fp.textInput.Blur()
		fp.viewState = ViewStateNormal

	default:
		fp.textInput, cmd = fp.textInput.Update(msg)
	}

	return fp, cmd
}

// filterFiles filters files based on search query and glob pattern
func (fp *AdvancedModel) filterFiles() {
	if fp.searchQuery == "" && fp.globPattern == "" {
		fp.filteredFiles = fp.files
		return
	}

	fp.filteredFiles = []File{}
	query := strings.ToLower(fp.searchQuery)

	for _, file := range fp.files {
		// Skip parent directory ".." from glob filtering
		if file.Name == ".." {
			fp.filteredFiles = append(fp.filteredFiles, file)
			continue
		}

		// Apply search filter
		matchesSearch := fp.searchQuery == "" || strings.Contains(strings.ToLower(file.Name), query)

		// Apply glob filter
		matchesGlob := fp.globPattern == ""
		if !matchesGlob && fp.globPattern != "" {
			// Use filepath.Match for glob pattern matching
			if matched, err := filepath.Match(fp.globPattern, file.Name); err == nil && matched {
				matchesGlob = true
			}
		}

		// File must match both filters (if active)
		if matchesSearch && matchesGlob {
			fp.filteredFiles = append(fp.filteredFiles, file)
		}
	}
}

// sortFiles sorts files according to current sort mode
func (fp *AdvancedModel) sortFiles() {
	sort.Slice(fp.files, func(i, j int) bool {
		// Always keep parent directory at top
		if fp.files[i].Name == ".." {
			return true
		}
		if fp.files[j].Name == ".." {
			return false
		}

		// Directories first (except when sorting by type)
		if fp.sortMode != SortByType && fp.files[i].IsDir != fp.files[j].IsDir {
			return fp.files[i].IsDir
		}

		switch fp.sortMode {
		case SortBySize:
			return fp.files[i].Size < fp.files[j].Size
		case SortByDate:
			return fp.files[i].ModTime.After(fp.files[j].ModTime)
		case SortByType:
			extI := strings.ToLower(filepath.Ext(fp.files[i].Name))
			extJ := strings.ToLower(filepath.Ext(fp.files[j].Name))
			if extI != extJ {
				return extI < extJ
			}
			return strings.ToLower(fp.files[i].Name) < strings.ToLower(fp.files[j].Name)
		case SortByName:
			return strings.ToLower(fp.files[i].Name) < strings.ToLower(fp.files[j].Name)
		default:
			return strings.ToLower(fp.files[i].Name) < strings.ToLower(fp.files[j].Name)
		}
	})

	fp.filterFiles()
}

// updatePreview updates the preview content for current file
func (fp *AdvancedModel) updatePreview() {
	if !fp.showPreview || len(fp.filteredFiles) == 0 {
		fp.previewContent = ""
		return
	}

	file := fp.filteredFiles[fp.cursor]

	if file.IsDir {
		fp.previewContent = fp.buildDirectoryPreview(file)
	} else {
		fp.previewContent = fp.buildFilePreview(file)
	}
}

// buildDirectoryPreview builds preview content for directories
func (fp *AdvancedModel) buildDirectoryPreview(file File) string {
	var content strings.Builder

	content.WriteString(previewTitleStyle.Render(file.Name) + "\n")
	content.WriteString("Type: Directory\n")
	content.WriteString(fmt.Sprintf("Modified: %s\n", file.ModTime.Format("Jan 02, 2006 15:04")))
	content.WriteString(fmt.Sprintf("Permissions: %s\n", file.Mode.String()))

	// Try to count items in directory
	if entries, err := os.ReadDir(file.Path); err == nil {
		visibleCount := 0
		hiddenCount := 0
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				hiddenCount++
			} else {
				visibleCount++
			}
		}
		content.WriteString(fmt.Sprintf("Items: %d", visibleCount))
		if hiddenCount > 0 {
			content.WriteString(fmt.Sprintf(" (%d hidden)", hiddenCount))
		}
		content.WriteString("\n")
	}

	return content.String()
}

// buildFilePreview builds preview content for files
func (fp *AdvancedModel) buildFilePreview(file File) string {
	var content strings.Builder

	content.WriteString(previewTitleStyle.Render(file.Name) + "\n")
	content.WriteString(fmt.Sprintf("Size: %s\n", fp.formatFileSize(file.Size)))
	content.WriteString(fmt.Sprintf("Modified: %s\n", file.ModTime.Format("Jan 02, 2006 15:04")))
	content.WriteString(fmt.Sprintf("Permissions: %s\n", file.Mode.String()))
	content.WriteString(strings.Repeat("─", 20) + "\n")

	// Try to preview file content
	if fp.isTextFile(file.Name) && file.Size < 10*1024 { // Only preview small text files
		if preview := fp.readFilePreview(file.Path); preview != "" {
			content.WriteString(preview)
		} else {
			content.WriteString("[Unable to read file]")
		}
	} else if fp.isImageFile(file.Name) {
		content.WriteString("[Image file]\n")
		if info, err := os.Stat(file.Path); err == nil {
			content.WriteString(fmt.Sprintf("Size: %dx? pixels\n", info.Size()))
		}
	} else if fp.isArchiveFile(file.Name) {
		content.WriteString("[Archive file]\n")
		content.WriteString("Use 'file' command for details")
	} else {
		content.WriteString("[Binary file]")
	}

	return content.String()
}

// isTextFile checks if a file is likely a text file
func (fp *AdvancedModel) isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExts := []string{
		".txt", ".md", ".readme", ".go", ".py", ".js", ".html", ".css",
		".json", ".xml", ".yml", ".yaml", ".toml", ".ini", ".conf",
		".sh", ".bat", ".ps1", ".php", ".rb", ".pl", ".java", ".cpp",
		".c", ".h", ".hpp", ".rs", ".swift", ".kt", ".scala", ".clj",
	}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}
	return false
}

// isImageFile checks if a file is an image
func (fp *AdvancedModel) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".tiff", ".ico"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// isArchiveFile checks if a file is an archive
func (fp *AdvancedModel) isArchiveFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	archiveExts := []string{".zip", ".tar", ".gz", ".rar", ".7z", ".bz2", ".xz", ".lz", ".lzma"}

	for _, archExt := range archiveExts {
		if ext == archExt {
			return true
		}
	}
	return false
}

// readFilePreview reads a preview of a text file
func (fp *AdvancedModel) readFilePreview(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer func() {
		_ = file.Close() // Ignore close errors in defer
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineCount := 0
	maxLines := 15

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		if len(line) > 50 {
			line = line[:50] + "..."
		}
		lines = append(lines, line)
		lineCount++
	}

	if lineCount == maxLines {
		lines = append(lines, "...")
	}

	return strings.Join(lines, "\n")
}

// getSelectedFiles returns the list of selected files
func (fp *AdvancedModel) getSelectedFiles() []string {
	if len(fp.multiSelected) > 0 {
		files := make([]string, 0, len(fp.multiSelected))
		for path := range fp.multiSelected {
			files = append(files, path)
		}
		return files
	}

	if len(fp.filteredFiles) > 0 && fp.filteredFiles[fp.cursor].Name != ".." {
		return []string{fp.filteredFiles[fp.cursor].Path}
	}

	return []string{}
}

// File operation methods (same as Tier 3, but working with filteredFiles)
func (fp *AdvancedModel) performDelete() {
	for _, filePath := range fp.confirmFiles {
		if err := os.RemoveAll(filePath); err != nil {
			fp.err = fmt.Errorf("failed to delete %s: %v", filepath.Base(filePath), err)
			return
		}
		delete(fp.multiSelected, filePath)
	}
	fp.loadDirectory()
}

func (fp *AdvancedModel) performPaste() {
	for _, src := range fp.clipboard {
		dst := filepath.Join(fp.currentPath, filepath.Base(src))

		switch fp.clipboardOp {
		case OpCopy:
			if err := fp.copyFile(src, dst); err != nil {
				fp.err = fmt.Errorf("failed to copy %s: %v", filepath.Base(src), err)
				return
			}
		case OpCut:
			if err := os.Rename(src, dst); err != nil {
				fp.err = fmt.Errorf("failed to move %s: %v", filepath.Base(src), err)
				return
			}
		case OpNone:
			// No operation to perform
		}
	}

	if fp.clipboardOp == OpCut {
		fp.clipboard = []string{}
		fp.clipboardOp = OpNone
	}

	fp.loadDirectory()
}

func (fp *AdvancedModel) copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return fp.copyDir(src, dst)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close() // Ignore close errors in defer
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close() // Ignore close errors in defer
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (fp *AdvancedModel) copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := fp.copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func (fp *AdvancedModel) performRename(newName string) {
	if len(fp.filteredFiles) > 0 && fp.filteredFiles[fp.cursor].Name != ".." {
		oldPath := fp.filteredFiles[fp.cursor].Path
		newPath := filepath.Join(fp.currentPath, newName)

		if err := os.Rename(oldPath, newPath); err != nil {
			fp.err = fmt.Errorf("failed to rename: %v", err)
			return
		}

		delete(fp.multiSelected, oldPath)
		fp.loadDirectory()
	}
}

func (fp *AdvancedModel) performCreateFile(name string) {
	filePath := filepath.Join(fp.currentPath, name)

	file, err := os.Create(filePath)
	if err != nil {
		fp.err = fmt.Errorf("failed to create file: %v", err)
		return
	}
	_ = file.Close() // Ignore close errors

	fp.loadDirectory()
}

func (fp *AdvancedModel) performCreateDir(name string) {
	dirPath := filepath.Join(fp.currentPath, name)

	if err := os.Mkdir(dirPath, 0755); err != nil {
		fp.err = fmt.Errorf("failed to create directory: %v", err)
		return
	}

	fp.loadDirectory()
}

// View renders the advanced file picker (Tier 4)
func (fp *AdvancedModel) View() string {
	if fp.width == 0 || fp.height == 0 {
		return "Loading..."
	}

	switch fp.viewState {
	case ViewStateConfirmDelete:
		return fp.viewConfirmDelete()
	case ViewStateNormal, ViewStateRename, ViewStateCreateFile, ViewStateCreateDir, ViewStateSearch, ViewStateGlob:
		return fp.viewNormal()
	default:
		return fp.viewNormal()
	}
}

// viewNormal renders the normal file picker view with preview panel
func (fp *AdvancedModel) viewNormal() string {
	// Calculate panel widths
	var fileListWidth, previewWidth int
	if fp.showPreview {
		previewWidth = (fp.width * fp.previewWidth) / 100
		if previewWidth < 20 {
			previewWidth = 2
		}
		if previewWidth > fp.width-40 {
			previewWidth = fp.width - 40
		}
		fileListWidth = fp.width - previewWidth - 6 // Account for borders and separator
	} else {
		fileListWidth = fp.width - 4
		previewWidth = 0
	}

	// Build file list panel
	filePanel := fp.buildFileListPanel(fileListWidth)

	if fp.showPreview {
		// Build preview panel
		previewPanel := fp.buildPreviewPanel(previewWidth)

		// Combine panels side by side
		panelsView := lipgloss.JoinHorizontal(
			lipgloss.Top,
			borderStyle.Width(fileListWidth).Render(filePanel),
			borderStyle.Width(previewWidth).Render(previewPanel),
		)

		// Add help below panels
		helpView := fp.buildHelpView()
		if helpView != "" {
			return panelsView + "\n" + helpView
		}
		return panelsView
	} else {
		return borderStyle.Width(fileListWidth).Render(filePanel) + "\n" + fp.buildHelpView()
	}
}

// buildHelpView creates a help view that adapts to the current mode
func (fp *AdvancedModel) buildHelpView() string {
	if fp.help.ShowAll {
		// Create a custom help model for this mode
		customKeyMap := &dynamicKeyMap{
			model: fp,
		}
		return fp.help.View(customKeyMap)
	} else {
		return fp.help.View(fp.keys)
	}
}

// dynamicKeyMap provides mode-aware help text
type dynamicKeyMap struct {
	model *AdvancedModel
}

func (d *dynamicKeyMap) ShortHelp() []key.Binding {
	return d.model.keys.ShortHelp()
}

func (d *dynamicKeyMap) FullHelp() [][]key.Binding {
	return d.model.FullHelpForMode()
}

// IsDirectorySelectionMode returns whether directory selection mode is enabled
func (fp *AdvancedModel) IsDirectorySelectionMode() bool {
	return fp.directorySelectionMode
}

// SetSize sets the width and height of the file picker
func (fp *AdvancedModel) SetSize(width, height int) {
	fp.width = width
	fp.height = height
	fp.help.Width = width
}

// buildFileListPanel builds the file list panel content
func (fp *AdvancedModel) buildFileListPanel(width int) string {
	var b strings.Builder
	contentWidth := width - 2

	// Title with status
	title := titleStyle.Render("File Explorer")
	if fp.directorySelectionMode {
		title += statusStyle.Render(" - Directory Selection Mode")
	}
	if fp.searchQuery != "" {
		title += statusStyle.Render(fmt.Sprintf(" (search: %s)", fp.searchQuery))
	}
	if fp.globPattern != "" {
		title += statusStyle.Render(fmt.Sprintf(" (glob: %s)", fp.globPattern))
	}
	if len(fp.multiSelected) > 0 {
		title += statusStyle.Render(fmt.Sprintf(" (%d selected)", len(fp.multiSelected)))
	}
	b.WriteString(title + "\n")

	// Current path (show relative path from jail if jailed)
	displayPath := fp.currentPath
	if fp.jailDirectory != "" {
		if relPath, err := filepath.Rel(fp.jailDirectory, fp.currentPath); err == nil && !strings.HasPrefix(relPath, "..") {
			if relPath == "." {
				displayPath = "[jail]"
			} else {
				displayPath = "[jail]/" + relPath
			}
		}
	}
	path := pathStyle.Render("Path: " + displayPath)
	b.WriteString(path + "\n")

	// Separator
	b.WriteString(strings.Repeat("─", contentWidth) + "\n")

	// File list
	contentHeight := fp.height - 9 // Account for title, path, status, search, help

	startIdx := 0
	endIdx := len(fp.filteredFiles)

	// Calculate visible range for scrolling
	if len(fp.filteredFiles) > contentHeight {
		if fp.cursor >= contentHeight/2 {
			startIdx = fp.cursor - contentHeight/2
			endIdx = startIdx + contentHeight
			if endIdx > len(fp.filteredFiles) {
				endIdx = len(fp.filteredFiles)
				startIdx = endIdx - contentHeight
				if startIdx < 0 {
					startIdx = 0
				}
			}
		} else {
			endIdx = contentHeight
		}
	}

	for i := startIdx; i < endIdx; i++ {
		file := fp.filteredFiles[i]
		line := fp.formatFileEntry(file, i == fp.cursor, contentWidth)
		b.WriteString(line + "\n")
	}

	// Fill remaining space
	remaining := contentHeight - (endIdx - startIdx)
	for i := 0; i < remaining; i++ {
		b.WriteString(strings.Repeat(" ", contentWidth) + "\n")
	}

	// Separator
	b.WriteString(strings.Repeat("─", contentWidth) + "\n")

	// Search input (if active)
	if fp.viewState == ViewStateSearch {
		b.WriteString(searchStyle.Render("Search: ") + fp.searchInput.View())
	} else if fp.searchQuery != "" {
		b.WriteString(searchStyle.Render(fmt.Sprintf("Search: %s (%d matches)", fp.searchQuery, len(fp.filteredFiles))))
	}

	// Glob input (if active)
	if fp.viewState == ViewStateGlob {
		b.WriteString(searchStyle.Render("Glob: ") + fp.globInput.View())
	} else if fp.globPattern != "" {
		b.WriteString(searchStyle.Render(fmt.Sprintf("Glob: %s (%d matches)", fp.globPattern, len(fp.filteredFiles))))
	}

	// Text input (if active)
	if fp.viewState == ViewStateRename || fp.viewState == ViewStateCreateFile || fp.viewState == ViewStateCreateDir {
		var prompt string
		switch fp.viewState {
		case ViewStateRename:
			prompt = "Rename: "
		case ViewStateCreateFile:
			prompt = "New file: "
		case ViewStateCreateDir:
			prompt = "New directory: "
		case ViewStateNormal, ViewStateConfirmDelete, ViewStateSearch, ViewStateGlob:
			// These states don't need prompts
		}
		if fp.viewState == ViewStateSearch {
			b.WriteString("\n")
		}
		b.WriteString(prompt + fp.textInput.View())
	}

	// Add line break only if we had search, glob, or text input
	if fp.viewState == ViewStateSearch || fp.viewState == ViewStateGlob ||
		fp.viewState == ViewStateRename || fp.viewState == ViewStateCreateFile ||
		fp.viewState == ViewStateCreateDir || fp.searchQuery != "" || fp.globPattern != "" {
		b.WriteString("\n")
	}

	// Status line
	status := fp.buildStatusLine()
	b.WriteString(status)

	// Error display
	if fp.err != nil {
		b.WriteString("\n" + errorStyle.Render("Error: "+fp.err.Error()))
		fp.err = nil
	}

	return b.String()
}

// buildPreviewPanel builds the preview panel content
func (fp *AdvancedModel) buildPreviewPanel(width int) string {
	var b strings.Builder
	contentWidth := width - 2

	// Preview title
	b.WriteString(previewTitleStyle.Render("Preview") + "\n")
	b.WriteString(strings.Repeat("─", contentWidth) + "\n")

	// Preview content with better formatting
	if fp.previewContent != "" {
		lines := strings.Split(fp.previewContent, "\n")
		maxLines := fp.height - 8

		for i, line := range lines {
			if i >= maxLines {
				b.WriteString("...\n")
				break
			}
			if len(line) > contentWidth-1 {
				line = line[:contentWidth-3] + "..."
			}
			b.WriteString(" " + line + "\n") // Add leading space for readability
		}
	} else {
		b.WriteString(" No preview available\n")
	}

	return b.String()
}

// viewConfirmDelete renders the delete confirmation dialog
func (fp *AdvancedModel) viewConfirmDelete() string {
	var b strings.Builder

	b.WriteString("Delete the following files?\n\n")

	for _, filePath := range fp.confirmFiles {
		b.WriteString("• " + filepath.Base(filePath) + "\n")
	}

	b.WriteString("\n[Y] Yes    [N] No")

	dialog := confirmStyle.Render(b.String())

	return lipgloss.Place(fp.width, fp.height, lipgloss.Center, lipgloss.Center, dialog)
}

// buildStatusLine builds the status line with Tier 4 info
func (fp *AdvancedModel) buildStatusLine() string {
	var parts []string

	// File count and filtering info
	if fp.searchQuery != "" || fp.globPattern != "" {
		parts = append(parts, fmt.Sprintf("%d of %d items", len(fp.filteredFiles), len(fp.files)))
	} else {
		parts = append(parts, fmt.Sprintf("%d items", len(fp.files)))
	}

	if len(fp.multiSelected) > 0 {
		parts = append(parts, fmt.Sprintf("%d selected", len(fp.multiSelected)))
	}

	if len(fp.clipboard) > 0 {
		op := "copied"
		if fp.clipboardOp == OpCut {
			op = "cut"
		}
		parts = append(parts, fmt.Sprintf("%d %s", len(fp.clipboard), op))
	}

	// Sort mode
	sortModes := []string{"Name", "Size", "Date", "Type"}
	parts = append(parts, fmt.Sprintf("Sort: %s", sortModes[fp.sortMode]))

	// View options
	var options []string
	if fp.directorySelectionMode {
		options = append(options, "Directory Selection")
	}
	if fp.showHidden {
		options = append(options, "Hidden")
	}
	if fp.detailedView {
		options = append(options, "Details")
	}
	if fp.showPreview {
		options = append(options, "Preview")
	}
	if fp.jailDirectory != "" {
		options = append(options, "Jailed")
	}
	if len(options) > 0 {
		parts = append(parts, strings.Join(options, ","))
	}

	return statusStyle.Render(strings.Join(parts, " | "))
}

// formatFileEntry formats a single file entry with proper table columns
func (fp *AdvancedModel) formatFileEntry(file File, isCursor bool, width int) string {
	if file.Hidden && !fp.showHidden {
		return "" // Skip hidden files if not showing them
	}

	// Column widths - responsive design with column hiding (no permissions)
	const (
		indicatorWidth = 4  // "✓▶  "
		iconWidth      = 4  // "📁  "
		sizeWidth      = 12 // "  1.23 GB  " (can get quite wide)
		dateWidth      = 10 // " Jan 02   "
		spacerWidth    = 2  // Extra spacing between sections
		sizeDateSpacer = 4  // Extra spacing between size and date (wider files)
		minNameWidth   = 25 // Minimum name column width
	)

	// Calculate which columns to show based on available width
	baseWidth := indicatorWidth + iconWidth + spacerWidth + minNameWidth + 8 // 8 for padding/safety
	showSize := fp.detailedView && fp.showSizes
	showDate := fp.detailedView

	// Progressive column hiding based on available width (no permissions column)
	fullWidth := baseWidth + sizeWidth + sizeDateSpacer + dateWidth
	if width < fullWidth {
		// Not enough space for all columns, start hiding
		if width < baseWidth+sizeWidth+sizeDateSpacer {
			showDate = false // Hide date first
		}
		if width < baseWidth+sizeWidth {
			showSize = false // Hide size last
		}
	}

	// Calculate actual fixed width based on what we're showing
	fixedWidth := indicatorWidth + iconWidth + spacerWidth
	if showSize {
		fixedWidth += sizeWidth + sizeDateSpacer
	}
	if showDate {
		fixedWidth += dateWidth
	}

	nameWidth := width - fixedWidth - 8 // -8 for border padding and extra safety margin
	if nameWidth < minNameWidth {
		nameWidth = minNameWidth
	}

	var line strings.Builder

	// Selection indicators (fixed width)
	indicator := "   "
	if isCursor && fp.multiSelected[file.Path] {
		indicator = "✓▶ "
	} else if isCursor {
		indicator = "▶  "
	} else if fp.multiSelected[file.Path] {
		indicator = "✓  "
	}
	line.WriteString(fmt.Sprintf("%-*s", indicatorWidth, indicator))

	// Icon (fixed width)
	if fp.showIcons {
		icon := fp.getFileIcon(file)
		line.WriteString(fmt.Sprintf("%-*s", iconWidth, icon))
	}

	// Spacer after icon
	line.WriteString(strings.Repeat(" ", spacerWidth))

	// Name (variable width, left-aligned)
	name := file.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}
	line.WriteString(fmt.Sprintf("%-*s", nameWidth, name))

	// Add detail columns based on available space
	if showSize {
		// Size column (right-aligned with extra spacing after)
		if !file.IsDir && file.Name != ".." {
			size := fp.formatFileSize(file.Size)
			line.WriteString(fmt.Sprintf("%*s", sizeWidth, size))
		} else {
			line.WriteString(fmt.Sprintf("%*s", sizeWidth, ""))
		}
		// Extra spacer after size column (since sizes can be wide)
		line.WriteString(strings.Repeat(" ", sizeDateSpacer))
	}

	if showDate {
		// Date column (left-aligned)
		if file.Name != ".." {
			modTime := file.ModTime.Format("Jan 02")
			line.WriteString(fmt.Sprintf("%-*s", dateWidth, modTime))
		} else {
			line.WriteString(fmt.Sprintf("%-*s", dateWidth, ""))
		}
	}

	result := line.String()

	// Ensure we don't exceed width
	if len(result) > width-2 {
		result = result[:width-2]
	}

	// Apply styling
	if file.Hidden {
		result = hiddenStyle.Render(result)
	} else if isCursor && fp.multiSelected[file.Path] {
		result = multiSelectedStyle.Render(result)
	} else if isCursor {
		result = selectedStyle.Render(result)
	} else if fp.multiSelected[file.Path] {
		result = multiSelectedStyle.Render(result)
	} else if file.IsDir {
		// Use special styling for directories in directory selection mode
		if fp.directorySelectionMode {
			result = dirSelectionStyle.Render(result)
		} else {
			result = dirStyle.Render(result)
		}
	} else {
		result = normalStyle.Render(result)
	}

	return result
}

// getFileIcon returns an appropriate icon for the file (extended for Tier 4)
func (fp *AdvancedModel) getFileIcon(file File) string {
	if file.Name == ".." {
		return "📁"
	}
	if file.IsDir {
		if file.Hidden {
			return "👻"
		}
		return "📁"
	}

	if file.Hidden {
		return "👻"
	}

	// Detailed file type detection for Tier 4
	ext := strings.ToLower(filepath.Ext(file.Name))
	filename := strings.ToLower(file.Name)

	// Archives
	switch ext {
	case ".zip", ".tar", ".gz", ".rar", ".7z", ".bz2", ".xz":
		return "📦"
	}

	// Documents
	switch ext {
	case ".pdf", ".doc", ".docx", ".odt":
		return "📋"
	case ".xls", ".xlsx", ".ods", ".csv":
		return "📊"
	case ".ppt", ".pptx", ".odp":
		return "▶️"
	}

	// Media
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".tiff":
		return "🖼️"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm":
		return "🎬"
	case ".mp3", ".wav", ".flac", ".ogg", ".aac", ".m4a":
		return "🎵"
	}

	// Code files
	switch ext {
	case ".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs", ".swift":
		return "💻"
	case ".html", ".css", ".scss", ".less":
		return "💻"
	case ".json", ".xml", ".yml", ".yaml", ".toml":
		return "💻"
	}

	// Scripts and executables
	switch ext {
	case ".sh", ".bat", ".ps1", ".cmd":
		return "⚙️"
	case ".exe", ".msi", ".deb", ".rpm", ".dmg", ".app":
		return "⚙️"
	}

	// Text files
	switch ext {
	case ".txt", ".md", ".readme", ".log":
		return "📄"
	}

	// Symlinks (would need additional detection)
	if file.Mode&os.ModeSymlink != 0 {
		return "🔗"
	}

	// Read-only files
	if file.Mode&0200 == 0 {
		return "🔒"
	}

	// Special files
	if strings.Contains(filename, "readme") {
		return "📄"
	}
	if strings.Contains(filename, "license") {
		return "📄"
	}
	if strings.Contains(filename, "makefile") || strings.Contains(filename, "dockerfile") {
		return "⚙️"
	}

	return "❓" // Unknown file type
}

// formatFileSize formats file size in human-readable format
func (fp *AdvancedModel) formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// loadDirectory loads the contents of the current directory (enhanced for Tier 4)
func (fp *AdvancedModel) loadDirectory() {
	fp.files = []File{}
	fp.err = nil

	entries, err := os.ReadDir(fp.currentPath)
	if err != nil {
		fp.err = err
		return
	}

	// Add parent directory entry if not at root and not at jail root
	if fp.currentPath != "/" && fp.currentPath != "\\" && !fp.isAtJailRoot() {
		parentPath := filepath.Dir(fp.currentPath)
		// Only add .. if parent directory is within jail (or no jail is set)
		if fp.validateNavigationPath(parentPath) {
			if info, err := os.Stat(parentPath); err == nil {
				fp.files = append(fp.files, File{
					Name:    "..",
					Path:    parentPath,
					IsDir:   true,
					Size:    0,
					ModTime: info.ModTime(),
					Mode:    info.Mode(),
					Hidden:  false,
				})
			}
		}
	}

	// Add directory entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		isHidden := strings.HasPrefix(entry.Name(), ".")

		// Skip hidden files if not showing them
		if isHidden && !fp.showHidden {
			continue
		}

		file := File{
			Name:    entry.Name(),
			Path:    filepath.Join(fp.currentPath, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
			Hidden:  isHidden,
		}
		fp.files = append(fp.files, file)
	}

	// Sort files
	fp.sortFiles()

	// Reset cursor if out of bounds
	if fp.cursor >= len(fp.filteredFiles) {
		fp.cursor = len(fp.filteredFiles) - 1
	}
	if fp.cursor < 0 {
		fp.cursor = 0
	}

	// Update preview for current file
	fp.updatePreview()
}

// GetSelected returns the selected file paths
func (fp *AdvancedModel) GetSelected() ([]string, bool) {
	if fp.cancelled {
		return nil, false
	}
	return fp.selectedFiles, len(fp.selectedFiles) > 0
}

// GetError returns any error that occurred
func (fp *AdvancedModel) GetError() error {
	return fp.err
}

// SetShowPreview sets whether to show file preview
func (fp *AdvancedModel) SetShowPreview(show bool) {
	fp.showPreview = show
}

// SetShowHidden sets whether to show hidden files
func (fp *AdvancedModel) SetShowHidden(show bool) {
	fp.showHidden = show
	fp.loadDirectory() // Reload to apply the hidden file setting
}

// GetDirectorySelectionMode returns whether directory selection mode is enabled
func (fp *AdvancedModel) GetDirectorySelectionMode() bool {
	return fp.directorySelectionMode
}

// SetDirectorySelectionMode enables or disables directory selection mode
func (fp *AdvancedModel) SetDirectorySelectionMode(enabled bool) {
	fp.directorySelectionMode = enabled
}
