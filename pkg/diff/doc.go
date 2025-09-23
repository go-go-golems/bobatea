// Package diff provides a minimal, reusable structured-data diff viewer for
// Bubble Tea applications.
//
// The MVP focuses on a two-pane layout (list/detail), substring search, and
// value redaction. It intentionally defers exporters, plugin registries,
// multiple providers, and advanced renderers to future iterations.
//
// Quick start:
//
//
//	provider := &StaticProvider{ /* implement DataProvider */ }
//	config := DefaultConfig()
//	config.Title = "My Diff"
//	model := NewModel(provider, config)
//	_ = tea.NewProgram(model, tea.WithAltScreen()).Run()
//
package diff


