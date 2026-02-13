package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/logutil"
	"github.com/go-go-golems/bobatea/pkg/repl"
	jsrepl "github.com/go-go-golems/bobatea/pkg/repl/evaluators/javascript"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/rs/zerolog"
)

func parseLevel(s string) zerolog.Level {
	switch strings.ToLower(s) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error", "err":
		return zerolog.ErrorLevel
	default:
		return zerolog.ErrorLevel
	}
}

func main() {
	ll := flag.String("log-level", "error", "log level: trace, debug, info, warn, error")
	lf := flag.String("log-file", "", "log file path (optional)")
	flag.Parse()

	level := parseLevel(*ll)
	if *lf != "" {
		logutil.InitTUILoggingToFile(level, *lf)
	} else {
		logutil.InitTUILoggingToDiscard(level)
	}

	evaluator, err := jsrepl.NewWithDefaults()
	if err != nil {
		log.Fatal(err)
	}

	config := repl.DefaultConfig()
	config.Title = "JavaScript REPL (jsparse autocomplete)"
	config.Placeholder = "Type console.lo or const fs = require('fs'); fs.re"
	config.Autocomplete.Enabled = true
	config.Autocomplete.FocusToggleKey = "ctrl+t"
	config.Autocomplete.TriggerKeys = []string{"tab"}
	config.Autocomplete.AcceptKeys = []string{"enter", "tab"}

	bus, err := eventbus.NewInMemoryBus()
	if err != nil {
		log.Fatal(err)
	}
	repl.RegisterReplToTimelineTransformer(bus)

	model := repl.NewModel(evaluator, config, bus.Publisher)
	programOptions := make([]tea.ProgramOption, 0, 1)
	if os.Getenv("BOBATEA_NO_ALT_SCREEN") != "1" {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	p := tea.NewProgram(model, programOptions...)
	timeline.RegisterUIForwarder(bus, p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 2)
	go func() { errs <- bus.Run(ctx) }()
	go func() {
		_, runErr := p.Run()
		cancel()
		errs <- runErr
	}()
	if runErr := <-errs; runErr != nil {
		log.Fatal(runErr)
	}
}
