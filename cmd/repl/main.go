package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/go-go-golems/bobatea/pkg/repl/evaluators/example"
	"github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
	"github.com/go-go-golems/bobatea/pkg/repl/evaluators/math"
	"github.com/spf13/cobra"
)



func main() {
	rootCmd := &cobra.Command{
		Use:   "repl-demo",
		Short: "Demonstration of the REPL functionality",
		Long: `A comprehensive demonstration of the extracted REPL functionality.
This demo shows multiple evaluators, theming, and configuration options.`,
	}

	// JavaScript REPL command
	jsCmd := &cobra.Command{
		Use:   "js",
		Short: "Run JavaScript REPL",
		Long:  "Start a JavaScript REPL with full ES5/ES6 support",
		Run: func(cmd *cobra.Command, args []string) {
			theme, _ := cmd.Flags().GetString("theme")
			width, _ := cmd.Flags().GetInt("width")
			title, _ := cmd.Flags().GetString("title")
			disableHistory, _ := cmd.Flags().GetBool("no-history")
			disableEditor, _ := cmd.Flags().GetBool("no-editor")

			jsEval, err := javascript.NewWithDefaults()
			if err != nil {
				log.Fatal(err)
			}

			config := repl.DefaultConfig()
			config.Width = width
			config.EnableHistory = !disableHistory
			config.EnableExternalEditor = !disableEditor
			if title != "" {
				config.Title = title
			}

			model := repl.NewModel(jsEval, config)

			if theme != "" {
				if t, ok := repl.BuiltinThemes[theme]; ok {
					model.SetTheme(t)
				}
			}

			p := tea.NewProgram(model, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	// Math REPL command
	mathCmd := &cobra.Command{
		Use:   "math",
		Short: "Run Math REPL",
		Long:  "Start a simple math calculator REPL",
		Run: func(cmd *cobra.Command, args []string) {
			theme, _ := cmd.Flags().GetString("theme")
			width, _ := cmd.Flags().GetInt("width")
			title, _ := cmd.Flags().GetString("title")
			disableHistory, _ := cmd.Flags().GetBool("no-history")
			disableEditor, _ := cmd.Flags().GetBool("no-editor")

			mathEval := math.NewEvaluator()

			config := repl.DefaultConfig()
			config.Width = width
			config.EnableHistory = !disableHistory
			config.EnableExternalEditor = !disableEditor
			if title != "" {
				config.Title = title
			}

			model := repl.NewModel(mathEval, config)

			if theme != "" {
				if t, ok := repl.BuiltinThemes[theme]; ok {
					model.SetTheme(t)
				}
			}

			p := tea.NewProgram(model, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	// Example REPL command
	exampleCmd := &cobra.Command{
		Use:   "example",
		Short: "Run Example REPL",
		Long:  "Start the example REPL with echo and basic math",
		Run: func(cmd *cobra.Command, args []string) {
			theme, _ := cmd.Flags().GetString("theme")
			width, _ := cmd.Flags().GetInt("width")
			title, _ := cmd.Flags().GetString("title")
			disableHistory, _ := cmd.Flags().GetBool("no-history")
			disableEditor, _ := cmd.Flags().GetBool("no-editor")

			exampleEval := repl.NewExampleEvaluator()

			config := repl.DefaultConfig()
			config.Width = width
			config.EnableHistory = !disableHistory
			config.EnableExternalEditor = !disableEditor
			if title != "" {
				config.Title = title
			}

			model := repl.NewModel(exampleEval, config)

			if theme != "" {
				if t, ok := repl.BuiltinThemes[theme]; ok {
					model.SetTheme(t)
				}
			}

			p := tea.NewProgram(model, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	// Multi-evaluator switcher command
	multiCmd := &cobra.Command{
		Use:   "multi",
		Short: "Run Multi-evaluator REPL",
		Long:  "Start a REPL that allows switching between different evaluators",
		Run: func(cmd *cobra.Command, args []string) {
			runMultiEvaluator(cmd)
		},
	}

	// Common flags
	for _, cmd := range []*cobra.Command{jsCmd, mathCmd, exampleCmd, multiCmd} {
		cmd.Flags().String("theme", "default", "Theme to use (default, dark, light)")
		cmd.Flags().Int("width", 80, "Terminal width")
		cmd.Flags().String("title", "", "Custom title for the REPL")
		cmd.Flags().Bool("no-history", false, "Disable command history")
		cmd.Flags().Bool("no-editor", false, "Disable external editor support")
	}

	// List themes command
	themesCmd := &cobra.Command{
		Use:   "themes",
		Short: "List available themes",
		Long:  "Display all available themes with their names",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available themes:")
			for name, theme := range repl.BuiltinThemes {
				fmt.Printf("  - %s: %s\n", name, theme.Name)
			}
		},
	}

	rootCmd.AddCommand(jsCmd, mathCmd, exampleCmd, multiCmd, themesCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runMultiEvaluator(cmd *cobra.Command) {
	theme, _ := cmd.Flags().GetString("theme")
	width, _ := cmd.Flags().GetInt("width")
	title, _ := cmd.Flags().GetString("title")
	disableHistory, _ := cmd.Flags().GetBool("no-history")
	disableEditor, _ := cmd.Flags().GetBool("no-editor")

	// Create evaluators
	jsEval, err := javascript.NewWithDefaults()
	if err != nil {
		log.Fatal(err)
	}

	mathEval := NewMathEvaluator()
	exampleEval := repl.NewExampleEvaluator()

	evaluators := map[string]repl.Evaluator{
		"js":      jsEval,
		"math":    mathEval,
		"example": exampleEval,
	}

	currentEval := "js"

	config := repl.DefaultConfig()
	config.Width = width
	config.EnableHistory = !disableHistory
	config.EnableExternalEditor = !disableEditor
	if title != "" {
		config.Title = title
	} else {
		config.Title = "Multi-Evaluator REPL"
	}

	model := repl.NewModel(evaluators[currentEval], config)

	if theme != "" {
		if t, ok := repl.BuiltinThemes[theme]; ok {
			model.SetTheme(t)
		}
	}

	// Add custom commands for switching evaluators
	model.AddCustomCommand("switch", func(args []string) tea.Cmd {
		if len(args) == 0 {
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/switch",
					Output: "Available evaluators: js, math, example",
					Error:  nil,
				}
			}
		}

		newEval := args[0]
		if eval, ok := evaluators[newEval]; ok {
			currentEval = newEval
			// Note: In a real implementation, you'd need to rebuild the model
			// This is a simplified example
			return func() tea.Msg {
				return repl.EvaluationCompleteMsg{
					Input:  "/switch " + newEval,
					Output: fmt.Sprintf("Switched to %s evaluator", eval.GetName()),
					Error:  nil,
				}
			}
		}

		return func() tea.Msg {
			return repl.EvaluationCompleteMsg{
				Input:  "/switch " + newEval,
				Output: fmt.Sprintf("Unknown evaluator: %s", newEval),
				Error:  fmt.Errorf("unknown evaluator: %s", newEval),
			}
		}
	})

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
