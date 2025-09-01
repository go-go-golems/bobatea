package main

import (
    "context"
    "fmt"
    "log"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
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

func (e *ThemeDemo) Evaluate(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return "Empty input", nil
	}

	// Handle theme-related commands
	if strings.HasPrefix(code, "theme ") {
		themeName := strings.TrimPrefix(code, "theme ")
		e.currentTheme = themeName
		return fmt.Sprintf("Theme changed to: %s", themeName), nil
	}

	// Demonstrate colorful output
	if code == "rainbow" {
		return "🌈 Rainbow colors! 🌈", nil
	}

	if code == "colors" {
		return "Red, Green, Blue, Yellow, Magenta, Cyan", nil
	}

	if code == "demo" {
		return "This is a demonstration of themed output!\nTry different themes to see the styling change.", nil
	}

	// Default response
	return fmt.Sprintf("Current theme: %s | You said: %s", e.currentTheme, code), nil
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

// Custom themes
var customThemes = map[string]repl.Theme{
	"cyberpunk": {
		Name: "Cyberpunk",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("55")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("51")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("201")).
				Italic(true),
		},
	},

	"ocean": {
		Name: "Ocean",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("24")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("117")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("203")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("123")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("67")).
				Italic(true),
		},
	},

	"forest": {
		Name: "Forest",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("22")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("34")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("120")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("160")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("142")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("100")).
				Italic(true),
		},
	},

	"sunset": {
		Name: "Sunset",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("124")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("208")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("215")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("173")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("130")).
				Italic(true),
		},
	},

	"monochrome": {
		Name: "Monochrome",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("238")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("246")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true),
		},
	},

	"rainbow": {
		Name: "Rainbow",
		Styles: repl.Styles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("90")).
				Padding(0, 1),

			Prompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true),

			Result: lipgloss.NewStyle().
				Foreground(lipgloss.Color("51")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true),

			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("207")).
				Italic(true),

			HelpText: lipgloss.NewStyle().
				Foreground(lipgloss.Color("45")).
				Italic(true),
		},
	},
}

// Theme switcher application
type ThemeSwitcherApp struct {
	repl         repl.Model
	evaluator    *ThemeDemo
	currentTheme string
	themeList    []string
}

func NewThemeSwitcherApp() *ThemeSwitcherApp {
	evaluator := NewThemeDemo()

	config := repl.Config{
		Title:          "Theme Switcher Demo",
		Placeholder:    "Try: 'demo', 'colors', 'rainbow', F1-F6 to switch themes",
		Width:          100,
		EnableHistory:  true,
		MaxHistorySize: 200,
	}

	model := repl.NewModel(evaluator, config)

	// Get all theme names
	themeList := []string{"default", "dark", "light"}
	for name := range customThemes {
		themeList = append(themeList, name)
	}

	app := &ThemeSwitcherApp{
		repl:         model,
		evaluator:    evaluator,
		currentTheme: "dark",
		themeList:    themeList,
	}

	// Set initial theme
	app.repl.SetTheme(repl.BuiltinThemes["dark"])

	// Add theme switching commands
	app.addThemeCommands()

	return app
}

func (app *ThemeSwitcherApp) addThemeCommands() {
	// Add theme switching command
	app.repl.AddCustomCommand("theme", func(args []string) tea.Cmd {
		return func() tea.Msg {
			if len(args) == 0 {
				return repl.EvaluationCompleteMsg{
					Input:  "/theme",
					Output: fmt.Sprintf("Current theme: %s\nAvailable themes: %s", app.currentTheme, strings.Join(app.themeList, ", ")),
					Error:  nil,
				}
			}

			themeName := args[0]

			// Check built-in themes
			if theme, ok := repl.BuiltinThemes[themeName]; ok {
				app.repl.SetTheme(theme)
				app.currentTheme = themeName
				app.evaluator.currentTheme = themeName
				return repl.EvaluationCompleteMsg{
					Input:  "/theme " + themeName,
					Output: fmt.Sprintf("Switched to built-in theme: %s", themeName),
					Error:  nil,
				}
			}

			// Check custom themes
			if theme, ok := customThemes[themeName]; ok {
				app.repl.SetTheme(theme)
				app.currentTheme = themeName
				app.evaluator.currentTheme = themeName
				return repl.EvaluationCompleteMsg{
					Input:  "/theme " + themeName,
					Output: fmt.Sprintf("Switched to custom theme: %s", themeName),
					Error:  nil,
				}
			}

			return repl.EvaluationCompleteMsg{
				Input:  "/theme " + themeName,
				Output: fmt.Sprintf("Theme '%s' not found. Available themes: %s", themeName, strings.Join(app.themeList, ", ")),
				Error:  fmt.Errorf("theme not found"),
			}
		}
	})

	// Add theme list command
	app.repl.AddCustomCommand("themes", func(args []string) tea.Cmd {
		return func() tea.Msg {
			var builtinThemes []string
			for name := range repl.BuiltinThemes {
				builtinThemes = append(builtinThemes, name)
			}

			var customThemeNames []string
			for name := range customThemes {
				customThemeNames = append(customThemeNames, name)
			}

			output := fmt.Sprintf("Built-in themes: %s\nCustom themes: %s\nCurrent theme: %s",
				strings.Join(builtinThemes, ", "),
				strings.Join(customThemeNames, ", "),
				app.currentTheme)

			return repl.EvaluationCompleteMsg{
				Input:  "/themes",
				Output: output,
				Error:  nil,
			}
		}
	})

	// Add theme demo command
	app.repl.AddCustomCommand("demo", func(args []string) tea.Cmd {
		return func() tea.Msg {
			return repl.EvaluationCompleteMsg{
				Input:  "/demo",
				Output: "This is a demonstration of the current theme!\nTry different themes to see how the styling changes.\nUse F1-F6 keys or /theme <name> to switch themes.",
				Error:  nil,
			}
		}
	})
}

func (app *ThemeSwitcherApp) Init() tea.Cmd {
	return app.repl.Init()
}

func (app *ThemeSwitcherApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f1":
			app.repl.SetTheme(repl.BuiltinThemes["default"])
			app.currentTheme = "default"
			app.evaluator.currentTheme = "default"
			return app, nil
		case "f2":
			app.repl.SetTheme(repl.BuiltinThemes["dark"])
			app.currentTheme = "dark"
			app.evaluator.currentTheme = "dark"
			return app, nil
		case "f3":
			app.repl.SetTheme(repl.BuiltinThemes["light"])
			app.currentTheme = "light"
			app.evaluator.currentTheme = "light"
			return app, nil
		case "f4":
			app.repl.SetTheme(customThemes["cyberpunk"])
			app.currentTheme = "cyberpunk"
			app.evaluator.currentTheme = "cyberpunk"
			return app, nil
		case "f5":
			app.repl.SetTheme(customThemes["ocean"])
			app.currentTheme = "ocean"
			app.evaluator.currentTheme = "ocean"
			return app, nil
		case "f6":
			app.repl.SetTheme(customThemes["forest"])
			app.currentTheme = "forest"
			app.evaluator.currentTheme = "forest"
			return app, nil
		case "f7":
			app.repl.SetTheme(customThemes["sunset"])
			app.currentTheme = "sunset"
			app.evaluator.currentTheme = "sunset"
			return app, nil
		case "f8":
			app.repl.SetTheme(customThemes["monochrome"])
			app.currentTheme = "monochrome"
			app.evaluator.currentTheme = "monochrome"
			return app, nil
		case "f9":
			app.repl.SetTheme(customThemes["rainbow"])
			app.currentTheme = "rainbow"
			app.evaluator.currentTheme = "rainbow"
			return app, nil
		}
	}

	var cmd tea.Cmd
	updatedModel, cmd := app.repl.Update(msg)
	if replModel, ok := updatedModel.(repl.Model); ok {
		app.repl = replModel
	}
	return app, cmd
}

func (app *ThemeSwitcherApp) View() string {
	view := app.repl.View()

	// Add theme information at the bottom
	themeInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render(fmt.Sprintf("Current theme: %s | F1-F9: Quick theme switch | /theme <name>: Switch theme | /themes: List themes", app.currentTheme))

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

    // Simplified theme demo: run REPL with ThemeDemo evaluator via Watermill bus
    evaluator := NewThemeDemo()
    config := repl.Config{
        Title:                "Theme Demo (simplified)",
        Placeholder:          "Type 'demo' or 'colors'",
        Width:                100,
        EnableHistory:        true,
        MaxHistorySize:       200,
    }

    bus, err := eventbus.NewInMemoryBus()
    if err != nil { log.Fatal(err) }
    repl.RegisterReplToTimelineTransformer(bus)

    model := repl.NewModel(evaluator, config, bus.Publisher)
    p := tea.NewProgram(model, tea.WithAltScreen())
    timeline.RegisterUIForwarder(bus, p)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    errs := make(chan error, 2)
    go func() { errs <- bus.Run(ctx) }()
    go func() { _, e := p.Run(); cancel(); errs <- e }()
    if e := <-errs; e != nil { log.Fatal(e) }
}
