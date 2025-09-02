package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/logutil"
	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/rs/zerolog"
)

// ThemeDemo evaluator that demonstrates different themes
type ThemeDemo struct {
	currentTheme string
}

func NewThemeDemo() *ThemeDemo {
	return &ThemeDemo{
		currentTheme: "dark",
	}
}

func (e *ThemeDemo) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
	code = strings.TrimSpace(code)

	if code == "" {
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": "Empty input"}})
		return nil
	}

	// Demonstrate colorful output
	switch code {
	case "rainbow":
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": "🌈 Rainbow colors! 🌈"}})
		return nil
	case "colors":
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": "Red, Green, Blue, Yellow, Magenta, Cyan"}})
		return nil
	case "demo":
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": "This is a demonstration of themed output!\nTry different themes to see the styling change."}})
		return nil
	}

	// Default response
	emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": fmt.Sprintf("Current theme: %s | You said: %s", e.currentTheme, code)}})
	return nil
}

func (e *ThemeDemo) GetPrompt() string {
	return "theme> "
}

func (e *ThemeDemo) GetName() string {
	return "Theme Demo"
}

func (e *ThemeDemo) SupportsMultiline() bool {
	return false
}

func (e *ThemeDemo) GetFileExtension() string {
	return ".theme"
}

// Custom theme data removed in timeline-centric example to focus on hotkeys and streaming output.

// Theme switcher application
type ThemeSwitcherApp struct {
	repl         *repl.Model
	evaluator    *ThemeDemo
	currentTheme string
}

// NewThemeSwitcherApp wires a pre-built REPL model and evaluator and adds F1-F9 hotkeys.
func NewThemeSwitcherApp(model *repl.Model, evaluator *ThemeDemo) *ThemeSwitcherApp {
	return &ThemeSwitcherApp{
		repl:         model,
		evaluator:    evaluator,
		currentTheme: "dark",
	}
}

// Note: no custom slash commands in this example; use evaluator outputs and F-keys.

func (app *ThemeSwitcherApp) Init() tea.Cmd {
	return app.repl.Init()
}

func (app *ThemeSwitcherApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f1":
			app.currentTheme = "default"
			app.evaluator.currentTheme = "default"
			return app, nil
		case "f2":
			app.currentTheme = "dark"
			app.evaluator.currentTheme = "dark"
			return app, nil
		case "f3":
			app.currentTheme = "light"
			app.evaluator.currentTheme = "light"
			return app, nil
		case "f4":
			app.currentTheme = "cyberpunk"
			app.evaluator.currentTheme = "cyberpunk"
			return app, nil
		case "f5":
			app.currentTheme = "ocean"
			app.evaluator.currentTheme = "ocean"
			return app, nil
		case "f6":
			app.currentTheme = "forest"
			app.evaluator.currentTheme = "forest"
			return app, nil
		case "f7":
			app.currentTheme = "sunset"
			app.evaluator.currentTheme = "sunset"
			return app, nil
		case "f8":
			app.currentTheme = "monochrome"
			app.evaluator.currentTheme = "monochrome"
			return app, nil
		case "f9":
			app.currentTheme = "rainbow"
			app.evaluator.currentTheme = "rainbow"
			return app, nil
		}
	}

	var cmd tea.Cmd
	updatedModel, cmd := app.repl.Update(msg)
	if replModel, ok := updatedModel.(*repl.Model); ok {
		app.repl = replModel
	}
	return app, cmd
}

func (app *ThemeSwitcherApp) View() string {
	view := app.repl.View()

	// Add theme information at the bottom
	themeInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render(fmt.Sprintf("Current theme: %s | F1-F9: Quick theme switch | Commands: demo, colors, rainbow", app.currentTheme))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		view,
		"",
		themeInfo,
	)
}

func main() {
	// Silence logs for TUI
	logutil.InitTUILoggingToDiscard(zerolog.ErrorLevel)

	// Theme demo: run REPL with ThemeDemo evaluator via Watermill bus
	evaluator := NewThemeDemo()
	config := repl.Config{
		Title:          "Theme Demo",
		Placeholder:    "Try: demo, colors, rainbow. Use F1-F9 to change theme label.",
		Width:          100,
		EnableHistory:  true,
		MaxHistorySize: 200,
	}

	bus, err := eventbus.NewInMemoryBus()
	if err != nil {
		log.Fatal(err)
	}
	repl.RegisterReplToTimelineTransformer(bus)

	model := repl.NewModel(evaluator, config, bus.Publisher)
	app := NewThemeSwitcherApp(model, evaluator)
	p := tea.NewProgram(app, tea.WithAltScreen())
	timeline.RegisterUIForwarder(bus, p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 2)
	go func() { errs <- bus.Run(ctx) }()
	go func() { _, e := p.Run(); cancel(); errs <- e }()
	if e := <-errs; e != nil {
		log.Fatal(e)
	}
}
