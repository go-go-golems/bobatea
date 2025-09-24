package diff

// Config controls the diff model behavior.
type Config struct {
	Title               string
	RedactSensitive     bool
	SplitPaneRatio      float64 // default 0.35
	EnableSearch        bool
	EnableStatusFilters bool
	InitialFilter       StatusFilter
}

// DefaultConfig returns sensible defaults for the diff model.
func DefaultConfig() Config {
	return Config{
		Title:               "Diff",
		RedactSensitive:     true,
		SplitPaneRatio:      0.35,
		EnableSearch:        true,
		EnableStatusFilters: true,
		InitialFilter:       StatusFilter{ShowAdded: true, ShowRemoved: true, ShowUpdated: true},
	}
}

// Option mutates the Config when building a Model.
type Option func(*Config)

// WithSearch enables or disables search UI and matching.
func WithSearch(enabled bool) Option { return func(c *Config) { c.EnableSearch = enabled } }

// StatusFilter controls which change statuses are visible in detail view.
type StatusFilter struct {
	ShowAdded   bool
	ShowRemoved bool
	ShowUpdated bool
}

// WithStatusFilters enables status filters and sets an initial value.
func WithStatusFilters(enabled bool, initial StatusFilter) Option {
	return func(c *Config) {
		c.EnableStatusFilters = enabled
		c.InitialFilter = initial
	}
}
