package repl

import (
	lipglossv2 "charm.land/lipgloss/v2"
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/commandpalette"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	renderers "github.com/go-go-golems/bobatea/pkg/timeline/renderers"
	"github.com/rs/zerolog/log"
)

// Model is a timeline-first REPL shell: timeline transcript + input line.
type Model struct {
	evaluator Evaluator
	config    Config
	styles    Styles

	// input & history
	history   *History
	textInput textinput.Model

	// layout
	width, height int

	// timeline shell (viewport + controller)
	reg    *timeline.Registry
	sh     *timeline.Shell
	focus  string // "input" or "timeline"
	help   help.Model
	keyMap KeyMap

	// bus publisher
	pub     message.Publisher
	turnSeq int
	appCtx  context.Context
	appStop context.CancelFunc

	// refresh scheduling
	refreshPending   bool
	refreshScheduled bool

	completion completionModel
	helpBar    helpBarModel
	helpDrawer helpDrawerModel
	palette    commandPaletteModel
}

// NewModel constructs a new REPL shell with timeline transcript.
func NewModel(evaluator Evaluator, config Config, pub message.Publisher) *Model {
	if config.Prompt == "" {
		config.Prompt = evaluator.GetPrompt()
	}
	ti := textinput.New()
	ti.Prompt = config.Prompt
	ti.Placeholder = config.Placeholder
	ti.Focus()
	ti.Width = max(10, config.Width-10)

	reg := timeline.NewRegistry()
	// Register base widgets
	reg.RegisterModelFactory(renderers.TextFactory{})
	reg.RegisterModelFactory(renderers.NewMarkdownFactory())
	reg.RegisterModelFactory(renderers.StructuredDataFactory{})
	reg.RegisterModelFactory(renderers.LogEventFactory{})
	reg.RegisterModelFactory(renderers.StructuredLogEventFactory{})

	sh := timeline.NewShell(reg)

	var completer InputCompleter
	if c, ok := evaluator.(InputCompleter); ok {
		completer = c
	}
	var helpBarProvider HelpBarProvider
	if p, ok := evaluator.(HelpBarProvider); ok {
		helpBarProvider = p
	}
	var helpDrawerProvider HelpDrawerProvider
	if p, ok := evaluator.(HelpDrawerProvider); ok {
		helpDrawerProvider = p
	}
	autocompleteCfg := normalizeAutocompleteConfig(config.Autocomplete)
	if !autocompleteCfg.Enabled {
		completer = nil
	}
	helpBarCfg := normalizeHelpBarConfig(config.HelpBar)
	if !helpBarCfg.Enabled {
		helpBarProvider = nil
	}
	helpDrawerCfg := normalizeHelpDrawerConfig(config.HelpDrawer)
	if !helpDrawerCfg.Enabled {
		helpDrawerProvider = nil
	}
	commandPaletteCfg := normalizeCommandPaletteConfig(config.CommandPalette)

	focusToggleKey := autocompleteCfg.FocusToggleKey
	if focusToggleKey == "" {
		if completer != nil {
			focusToggleKey = "ctrl+t"
		} else {
			focusToggleKey = "tab"
		}
	}

	ret := &Model{
		evaluator: evaluator,
		config:    config,
		styles:    DefaultStyles(),
		history:   NewHistory(config.MaxHistorySize),
		textInput: ti,
		width:     config.Width,
		reg:       reg,
		sh:        sh,
		focus:     "input",
		help:      help.New(),
		keyMap:    NewKeyMap(autocompleteCfg, helpDrawerCfg, commandPaletteCfg, focusToggleKey),
		pub:       pub,
		completion: completionModel{
			provider:   completer,
			debounce:   autocompleteCfg.Debounce,
			reqTimeout: autocompleteCfg.RequestTimeout,
			maxVisible: autocompleteCfg.MaxSuggestions,
			pageSize:   autocompleteCfg.OverlayPageSize,
			maxWidth:   autocompleteCfg.OverlayMaxWidth,
			maxHeight:  autocompleteCfg.OverlayMaxHeight,
			minWidth:   autocompleteCfg.OverlayMinWidth,
			margin:     autocompleteCfg.OverlayMargin,
			offsetX:    autocompleteCfg.OverlayOffsetX,
			offsetY:    autocompleteCfg.OverlayOffsetY,
			noBorder:   autocompleteCfg.OverlayNoBorder,
			placement:  autocompleteCfg.OverlayPlacement,
			horizontal: autocompleteCfg.OverlayHorizontalGrow,
		},
		helpBar: helpBarModel{
			provider:   helpBarProvider,
			debounce:   helpBarCfg.Debounce,
			reqTimeout: helpBarCfg.RequestTimeout,
		},
		helpDrawer: helpDrawerModel{
			provider:      helpDrawerProvider,
			debounce:      helpDrawerCfg.Debounce,
			reqTimeout:    helpDrawerCfg.RequestTimeout,
			prefetch:      helpDrawerCfg.PrefetchWhenHidden,
			dock:          helpDrawerCfg.Dock,
			widthPercent:  helpDrawerCfg.WidthPercent,
			heightPercent: helpDrawerCfg.HeightPercent,
			margin:        helpDrawerCfg.Margin,
		},
		palette: commandPaletteModel{
			ui:               commandpalette.New(),
			enabled:          commandPaletteCfg.Enabled,
			openKeys:         commandPaletteCfg.OpenKeys,
			closeKeys:        commandPaletteCfg.CloseKeys,
			slashEnabled:     commandPaletteCfg.SlashOpenEnabled,
			slashPolicy:      commandPaletteCfg.SlashPolicy,
			maxVisible:       commandPaletteCfg.MaxVisibleItems,
			overlayPlacement: commandPaletteCfg.OverlayPlacement,
			overlayMargin:    commandPaletteCfg.OverlayMargin,
			overlayOffsetX:   commandPaletteCfg.OverlayOffsetX,
			overlayOffsetY:   commandPaletteCfg.OverlayOffsetY,
		},
	}
	ret.appCtx, ret.appStop = context.WithCancel(context.Background())
	if ret.helpDrawer.provider == nil {
		ret.keyMap.HelpDrawerToggle.SetEnabled(false)
		ret.keyMap.HelpDrawerClose.SetEnabled(false)
		ret.keyMap.HelpDrawerRefresh.SetEnabled(false)
	}
	ret.updateKeyBindings()
	return ret
}

// NewModelWithContext constructs a REPL model whose internal app context derives from ctx.
// Passing nil uses context.Background().
func NewModelWithContext(ctx context.Context, evaluator Evaluator, config Config, pub message.Publisher) *Model {
	ret := NewModel(evaluator, config, pub)
	ret.cancelAppContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ret.appCtx, ret.appStop = context.WithCancel(ctx)
	return ret
}

// Init subscribes to evaluator events.
func (m *Model) Init() tea.Cmd {
	// no blinking on text input, because it makes copy paste impossible
	return tea.Batch(m.sh.Init())
}

// Update handles TUI events.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Trace().Interface("msg", msg).Interface("type", fmt.Sprintf("%T", msg)).Msg("updating repl model")
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
		m.textInput.Width = max(10, v.Width-10)
		m.palette.ui.SetSize(v.Width, v.Height)
		helpHeight := lipgloss.Height(m.help.View(m.keyMap))
		// reserve room for title, input, and help rows
		tlHeight := max(0, v.Height-helpHeight-4)
		m.sh.SetSize(v.Width, tlHeight)
		// initial refresh to fit new size
		m.sh.RefreshView(false)
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(v, m.keyMap.Quit):
			m.cancelAppContext()
			return m, tea.Quit
		case key.Matches(v, m.keyMap.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		switch m.focus {
		case "input":
			return m.updateInput(v)
		case "timeline":
			return m.updateTimeline(v)
		}

	case timeline.UIEntityCreated:
		m.ctrl().OnCreated(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timeline.UIEntityUpdated:
		m.ctrl().OnUpdated(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timeline.UIEntityCompleted:
		m.ctrl().OnCompleted(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timeline.UIEntityDeleted:
		m.ctrl().OnDeleted(v)
		m.refreshPending = true
		return m, m.scheduleRefresh()
	case timelineRefreshMsg:
		m.refreshScheduled = false
		if m.refreshPending {
			m.sh.RefreshView(true)
			m.refreshPending = false
		}
		return m, nil

	case completionDebounceMsg:
		return m, m.handleDebouncedCompletion(v)
	case completionResultMsg:
		return m, m.handleCompletionResult(v)
	case helpBarDebounceMsg:
		return m, m.handleDebouncedHelpBar(v)
	case helpBarResultMsg:
		return m, m.handleHelpBarResult(v)
	case helpDrawerDebounceMsg:
		return m, m.handleDebouncedHelpDrawer(v)
	case helpDrawerResultMsg:
		return m, m.handleHelpDrawerResult(v)

	case cursor.BlinkMsg:
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	log.Trace().Interface("msg", msg).Msg("updating repl model default case")

	return m, nil
}

func (m *Model) View() string {
	title := m.config.Title
	if title == "" {
		title = fmt.Sprintf("%s REPL", m.evaluator.GetName())
	}

	header := m.styles.Title.Render(" " + title + " ")
	timelineView := m.sh.View()

	inputView := m.textInput.View()
	if m.focus == "timeline" {
		inputView = m.styles.HelpText.Render(inputView)
	}

	helpView := m.help.View(m.keyMap)
	baseSections := []string{
		header,
		"",
		timelineView,
		inputView,
	}
	if helpBarView := m.renderHelpBar(); helpBarView != "" {
		baseSections = append(baseSections, helpBarView)
	}
	baseSections = append(baseSections, helpView)
	base := lipgloss.JoinVertical(lipgloss.Left, baseSections...)

	if m.width <= 0 || m.height <= 0 {
		return base
	}

	completionLayout, completionOK := m.computeCompletionOverlayLayout(header, timelineView)
	completionPopup := ""
	if completionOK {
		completionPopup = m.renderCompletionPopup(completionLayout)
		if completionPopup == "" {
			m.completion.visibleRows = 0
			completionOK = false
		} else {
			m.completion.visibleRows = completionLayout.VisibleRows
			m.ensureCompletionSelectionVisible()
		}
	} else {
		m.completion.visibleRows = 0
	}

	drawerLayout, drawerOK := m.computeHelpDrawerOverlayLayout(header, timelineView)
	drawerPanel := ""
	if drawerOK {
		drawerPanel = m.renderHelpDrawerPanel(drawerLayout)
		if drawerPanel == "" {
			drawerOK = false
		}
	}

	paletteLayout, paletteOK := m.computeCommandPaletteOverlayLayout()

	if !completionOK && !drawerOK && !paletteOK {
		return base
	}

	layers := []*lipglossv2.Layer{
		lipglossv2.NewLayer(base).X(0).Y(0).Z(0).ID("repl-base"),
	}
	if drawerOK {
		layers = append(layers,
			lipglossv2.NewLayer(drawerPanel).X(drawerLayout.PanelX).Y(drawerLayout.PanelY).Z(15).ID("help-drawer-overlay"),
		)
	}
	if completionOK {
		layers = append(layers,
			lipglossv2.NewLayer(completionPopup).X(completionLayout.PopupX).Y(completionLayout.PopupY).Z(20).ID("completion-overlay"),
		)
	}
	if paletteOK {
		layers = append(layers,
			lipglossv2.NewLayer(paletteLayout.View).X(paletteLayout.PanelX).Y(paletteLayout.PanelY).Z(30).ID("command-palette-overlay"),
		)
	}

	comp := lipglossv2.NewCompositor(layers...)
	canvas := lipglossv2.NewCanvas(max(1, m.width), max(1, m.height))
	canvas.Compose(comp)
	return canvas.Render()
}
