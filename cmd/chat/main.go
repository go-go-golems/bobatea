package main

import (
    "fmt"
    "os"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/bobatea/pkg/chat"
    "github.com/go-go-golems/bobatea/pkg/timeline"
    "github.com/spf13/cobra"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

var rootCmd = &cobra.Command{
	Use:   "chat",
	Short: "A chat application with different backend options",
	Long:  `This chat application allows you to choose between a fake backend and an HTTP backend.`,
}

var fakeCmd = &cobra.Command{
	Use:   "fake",
	Short: "Run the chat application with a fake backend",
	Run:   runFakeBackend,
}

// removed http command

// chatCmd mirrors the fake backend for convenience
var chatCmd = &cobra.Command{
    Use:   "chat",
    Short: "Run the chat application (alias of fake)",
    Run:   runFakeBackend,
}

// var httpAddr string // removed

func init() {
	rootCmd.AddCommand(fakeCmd)
    rootCmd.AddCommand(chatCmd)
}

func main() {
	// Initialize zerolog to log to /tmp/fake-chat.log
	logFile, err := os.OpenFile("/tmp/fake-chat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

    zerolog.TimeFieldFormat = time.StampMilli
    // Filter out trace-level logs for readability unless overridden
    if lvl, ok := os.LookupEnv("BOBATEA_LOG_LEVEL"); ok {
        if parsed, err := zerolog.ParseLevel(lvl); err == nil {
            zerolog.SetGlobalLevel(parsed)
        } else {
            zerolog.SetGlobalLevel(zerolog.DebugLevel)
        }
    } else {
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    }
    // Configure writer without colors/non-ASCII for file logs and add caller info
    cw := zerolog.ConsoleWriter{Out: logFile, NoColor: true, TimeFormat: time.StampMilli, PartsOrder: []string{"time","level","caller","message"}}
    logger := zerolog.New(cw).With().Caller().Timestamp().Logger()
    log.Logger = logger

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runFakeBackend(cmd *cobra.Command, args []string) {
    runChatWithOptions(func() chat.Backend { return NewFakeBackend() }, func(reg *timeline.Registry) {
        log.Debug().Str("component", "main").Msg("registering tool renderers via timeline hook")
        reg.RegisterModelFactory(ToolWeatherFactory{})
        reg.RegisterModelFactory(ToolWebSearchFactory{})
        reg.RegisterModelFactory(CheckboxFactory{})
    })
}

func runChatWithOptions(backendFactory func() chat.Backend, tlHook func(*timeline.Registry)) {
	status := &chat.Status{}

	backend := backendFactory()

	options := []tea.ProgramOption{
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	}

    model := chat.InitialModel(backend, chat.WithStatus(status), chat.WithTimelineRegister(tlHook))
    p := tea.NewProgram(model, options...)

	// Set the program for the backend after initialization
	if setterBackend, ok := backend.(interface{ SetProgram(*tea.Program) }); ok {
		setterBackend.SetProgram(p)
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
