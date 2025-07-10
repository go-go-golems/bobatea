# Custom Theme Example

This example demonstrates the comprehensive theming capabilities of the bobatea REPL component. It shows how to create custom themes, switch between themes dynamically, and build a theme-aware application.

## What it does

- **Custom Theme Creation** - Shows how to create beautiful custom themes
- **Dynamic Theme Switching** - Switch themes at runtime using keyboard shortcuts
- **Theme Management** - Manage and organize multiple themes
- **Theme Commands** - Add theme-related slash commands
- **Visual Demonstration** - See how themes affect the entire REPL interface

## Running the example

```bash
go run main.go
```

## Key Features Demonstrated

- **Custom Theme Definition** - Creating themes with lipgloss styles
- **Theme Registration** - Managing theme collections
- **Dynamic Theme Switching** - Runtime theme changes
- **Theme Commands** - Slash commands for theme management
- **Visual Feedback** - Immediate visual updates when switching themes
- **Theme Persistence** - Maintaining theme state across interactions

## Available Themes

### Built-in Themes
- **Default** - Standard theme with balanced colors
- **Dark** - Dark theme optimized for low-light environments
- **Light** - Light theme for bright environments

### Custom Themes
- **Cyberpunk** - Neon colors with a futuristic feel
- **Ocean** - Blue tones inspired by the ocean
- **Forest** - Green tones inspired by nature
- **Sunset** - Warm orange and yellow tones
- **Monochrome** - Elegant grayscale theme
- **Rainbow** - Vibrant multi-color theme

## What you'll see

When you run this example, you'll see a REPL that demonstrates different themes:

### Theme Switching Interface
```
┌─────────────────────────────────────────────────────┐
│ Theme Switcher Demo                                 │
├─────────────────────────────────────────────────────┤
│ theme> /themes                                      │
│ Built-in themes: default, dark, light              │
│ Custom themes: cyberpunk, ocean, forest, sunset    │
│ Current theme: dark                                 │
│                                                     │
│ theme> demo                                         │
│ This is a demonstration of themed output!           │
│ Try different themes to see the styling change.     │
│                                                     │
│ theme> _                                            │
├─────────────────────────────────────────────────────┤
│ Current theme: dark | F1-F9: Quick theme switch    │
│ /theme <name>: Switch theme | /themes: List themes │
└─────────────────────────────────────────────────────┘
```

### Theme Examples

#### Cyberpunk Theme
- Title: Bright white on purple background
- Prompt: Bright cyan
- Result: Bright green
- Error: Bright red
- Info: Bright yellow

#### Ocean Theme
- Title: White on deep blue background
- Prompt: Ocean blue
- Result: Light blue
- Error: Coral
- Info: Aqua

#### Forest Theme
- Title: White on dark green background
- Prompt: Forest green
- Result: Light green
- Error: Red-orange
- Info: Sage green

## Code Structure

### Custom Theme Definition

```go
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
    // ... more themes
}
```

### Theme Management

```go
type ThemeSwitcherApp struct {
    repl         repl.Model
    evaluator    *ThemeDemo
    currentTheme string
    themeList    []string
}

func (app *ThemeSwitcherApp) addThemeCommands() {
    // Add theme switching command
    app.repl.AddCustomCommand("theme", func(args []string) tea.Cmd {
        return func() tea.Msg {
            if len(args) == 0 {
                return repl.EvaluationCompleteMsg{
                    Input:  "/theme",
                    Output: fmt.Sprintf("Current theme: %s\nAvailable themes: %s", 
                        app.currentTheme, strings.Join(app.themeList, ", ")),
                    Error:  nil,
                }
            }
            
            themeName := args[0]
            
            // Check built-in themes
            if theme, ok := repl.BuiltinThemes[themeName]; ok {
                app.repl.SetTheme(theme)
                app.currentTheme = themeName
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
                return repl.EvaluationCompleteMsg{
                    Input:  "/theme " + themeName,
                    Output: fmt.Sprintf("Switched to custom theme: %s", themeName),
                    Error:  nil,
                }
            }
            
            return repl.EvaluationCompleteMsg{
                Input:  "/theme " + themeName,
                Output: fmt.Sprintf("Theme '%s' not found", themeName),
                Error:  fmt.Errorf("theme not found"),
            }
        }
    })
}
```

### Keyboard Shortcuts

```go
func (app *ThemeSwitcherApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "f1":
            app.repl.SetTheme(repl.BuiltinThemes["default"])
            app.currentTheme = "default"
            return app, nil
        case "f2":
            app.repl.SetTheme(repl.BuiltinThemes["dark"])
            app.currentTheme = "dark"
            return app, nil
        case "f3":
            app.repl.SetTheme(repl.BuiltinThemes["light"])
            app.currentTheme = "light"
            return app, nil
        case "f4":
            app.repl.SetTheme(customThemes["cyberpunk"])
            app.currentTheme = "cyberpunk"
            return app, nil
        // ... more theme shortcuts
        }
    }
    
    var cmd tea.Cmd
    app.repl, cmd = app.repl.Update(msg)
    return app, cmd
}
```

## Theme Creation Guide

### Step 1: Define Your Color Palette

```go
// Choose your colors (use lipgloss.Color)
primaryColor := lipgloss.Color("39")    // Blue
secondaryColor := lipgloss.Color("46")  // Green
accentColor := lipgloss.Color("226")    // Yellow
errorColor := lipgloss.Color("196")     // Red
backgroundColor := lipgloss.Color("24") // Dark blue
```

### Step 2: Create the Theme Structure

```go
myTheme := repl.Theme{
    Name: "My Custom Theme",
    Styles: repl.Styles{
        Title: lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("15")).
            Background(backgroundColor).
            Padding(0, 1),
        
        Prompt: lipgloss.NewStyle().
            Foreground(primaryColor).
            Bold(true),
        
        Result: lipgloss.NewStyle().
            Foreground(secondaryColor),
        
        Error: lipgloss.NewStyle().
            Foreground(errorColor).
            Bold(true),
        
        Info: lipgloss.NewStyle().
            Foreground(accentColor).
            Italic(true),
        
        HelpText: lipgloss.NewStyle().
            Foreground(lipgloss.Color("243")).
            Italic(true),
    },
}
```

### Step 3: Register and Use the Theme

```go
// Add to your theme collection
customThemes["my-theme"] = myTheme

// Apply the theme
model.SetTheme(myTheme)
```

## Try These Commands

### Theme Commands
- `/theme dark` - Switch to dark theme
- `/theme cyberpunk` - Switch to cyberpunk theme
- `/theme ocean` - Switch to ocean theme
- `/themes` - List all available themes
- `/demo` - Show theme demonstration

### Test Commands
- `demo` - Show themed output
- `colors` - Display color demonstration
- `rainbow` - Show rainbow text
- `error` - Test error styling

### Keyboard Shortcuts
- **F1** - Default theme
- **F2** - Dark theme
- **F3** - Light theme
- **F4** - Cyberpunk theme
- **F5** - Ocean theme
- **F6** - Forest theme
- **F7** - Sunset theme
- **F8** - Monochrome theme
- **F9** - Rainbow theme

## Advanced Theme Features

### Conditional Styling

```go
// Create themes that adapt to content
func (e *ThemeDemo) Evaluate(ctx context.Context, code string) (string, error) {
    // Different styling based on content
    if strings.Contains(code, "error") {
        return "Error styling applied", fmt.Errorf("demonstration error")
    }
    
    if strings.Contains(code, "success") {
        return "Success! This uses success styling", nil
    }
    
    return "Regular output with standard styling", nil
}
```

### Theme Persistence

```go
// Save theme preferences
func saveThemePreference(themeName string) error {
    // Save to config file or user preferences
    return nil
}

// Load theme preferences
func loadThemePreference() (string, error) {
    // Load from config file or user preferences
    return "dark", nil
}
```

### Theme Validation

```go
// Validate theme completeness
func validateTheme(theme repl.Theme) error {
    if theme.Name == "" {
        return fmt.Errorf("theme must have a name")
    }
    
    // Check that all required styles are defined
    if theme.Styles.Title.GetForeground() == "" {
        return fmt.Errorf("theme must define title color")
    }
    
    return nil
}
```

## Color Reference

### Common Colors
- **Red**: 196, 9, 1
- **Green**: 46, 34, 2
- **Blue**: 39, 33, 4
- **Yellow**: 226, 11, 3
- **Magenta**: 201, 13, 5
- **Cyan**: 51, 14, 6
- **White**: 15, 7
- **Black**: 0, 16
- **Gray**: 243, 244, 245, 246, 247, 248

### Theme-Specific Palettes

#### Cyberpunk
- Background: 55 (purple)
- Primary: 51 (cyan)
- Secondary: 46 (green)
- Accent: 226 (yellow)

#### Ocean
- Background: 24 (deep blue)
- Primary: 39 (blue)
- Secondary: 117 (light blue)
- Accent: 123 (aqua)

#### Forest
- Background: 22 (dark green)
- Primary: 34 (green)
- Secondary: 120 (light green)
- Accent: 142 (sage)

## Best Practices

1. **Consistency** - Use a consistent color palette
2. **Contrast** - Ensure good contrast for readability
3. **Accessibility** - Consider color-blind users
4. **Context** - Use appropriate colors for different message types
5. **Testing** - Test themes in different terminal environments

## Extending the Example

You can extend this example by:

1. **Adding more themes**:
```go
customThemes["my-theme"] = createMyTheme()
```

2. **Creating theme categories**:
```go
var themeCategories = map[string][]string{
    "Dark": {"dark", "cyberpunk", "ocean"},
    "Light": {"light", "sunset"},
    "Colorful": {"rainbow", "forest"},
}
```

3. **Adding theme animations**:
```go
// Animate theme transitions
func (app *ThemeSwitcherApp) animateThemeChange(newTheme repl.Theme) tea.Cmd {
    // Implement smooth theme transitions
}
```

4. **Creating theme editor**:
```go
// Interactive theme creation
func (app *ThemeSwitcherApp) enterThemeEditor() tea.Cmd {
    // Allow users to create custom themes
}
```

## Next Steps

After trying this example, check out:

- [Embedded in App Example](../embedded-in-app/) - Integration patterns
- [Custom Evaluator Example](../custom-evaluator/) - Advanced evaluator logic
- [Basic Usage Example](../basic-usage/) - Simple implementation patterns

This example shows the power of the REPL theming system and how to create beautiful, customizable interfaces that adapt to user preferences and different environments.
