package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/geppetto/pkg/conversation"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
    "github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/gorilla/mux"
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

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Run the chat application with an HTTP backend",
	Run:   runHTTPBackend,
}

// chatCmd mirrors the fake backend for convenience
var chatCmd = &cobra.Command{
    Use:   "chat",
    Short: "Run the chat application (alias of fake)",
    Run:   runFakeBackend,
}

var httpAddr string

func init() {
	rootCmd.AddCommand(fakeCmd)
	rootCmd.AddCommand(httpCmd)
    rootCmd.AddCommand(chatCmd)

	httpCmd.Flags().StringVarP(&httpAddr, "addr", "a", ":8080", "HTTP server address")
}

func main() {
	// Initialize zerolog to log to /tmp/fake-chat.log
	logFile, err := os.OpenFile("/tmp/fake-chat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
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
    cw := zerolog.ConsoleWriter{Out: logFile, NoColor: true, PartsOrder: []string{"time","level","caller","message"}}
    logger := zerolog.New(cw).With().Caller().Timestamp().Logger()
    log.Logger = logger

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runFakeBackend(cmd *cobra.Command, args []string) {
    runChatWithOptions(func() chat.Backend { return NewFakeBackend() }, func(reg *timeline.Registry) {
        reg.Register(&ToolWeatherRenderer{})
        reg.Register(&ToolWebSearchRenderer{})
    })
}

func runHTTPBackend(cmd *cobra.Command, args []string) {
    runChatWithOptions(func() chat.Backend { return NewHTTPBackend("/backend", WithLogFile("/tmp/http-backend.log")) }, func(reg *timeline.Registry) {
        reg.Register(&ToolWeatherRenderer{})
        reg.Register(&ToolWebSearchRenderer{})
    })
}

func runChatWithOptions(backendFactory func() chat.Backend, tlHook func(*timeline.Registry)) {
	status := &chat.Status{}

	manager := conversation.NewManager(conversation.WithMessages(
		conversation.NewChatMessage(conversation.RoleSystem, "Welcome to the chat application!"),
	))

	backend := backendFactory()

	options := []tea.ProgramOption{
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	}

    model := chat.InitialModel(manager, backend, chat.WithStatus(status), chat.WithTimelineRegister(tlHook))
    p := tea.NewProgram(model, options...)

	// Set up the user backend
	userBackend := chat.NewUserBackend(status, chat.WithLogFile("/tmp/http-backend.log"))
	userBackend.SetProgram(p)

	// Set up the HTTP backend
	if httpBackend, ok := backend.(*HTTPBackend); ok {
		// Set up the HTTP server
		r := mux.NewRouter()
		r.PathPrefix("/user").Handler(http.StripPrefix("/user", userBackend.Router()))
		httpBackend.SetRouter(r.PathPrefix("/backend").Subrouter())

		// Start the HTTP server with timeouts
		go func() {
			server := &http.Server{
				Addr:         httpAddr,
				Handler:      r,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
			}
			if err := server.ListenAndServe(); err != nil {
				fmt.Printf("Error running HTTP server: %v\n", err)
				os.Exit(1)
			}
		}()
	}

	// Set the program for the backend after initialization
	if setterBackend, ok := backend.(interface{ SetProgram(*tea.Program) }); ok {
		setterBackend.SetProgram(p)
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
