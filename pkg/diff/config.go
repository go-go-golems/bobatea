package diff

// Config controls the diff model behavior.
type Config struct {
	Title           string
	RedactSensitive bool
	SplitPaneRatio  float64 // default 0.35
}

// DefaultConfig returns sensible defaults for the diff model.
func DefaultConfig() Config {
	return Config{
		Title:           "Diff",
		RedactSensitive: true,
		SplitPaneRatio:  0.35,
	}
}


