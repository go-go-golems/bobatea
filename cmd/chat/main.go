package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/geppetto/pkg/conversation"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
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

var httpAddr string

func init() {
	rootCmd.AddCommand(fakeCmd)
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().StringVarP(&httpAddr, "addr", "a", ":8080", "HTTP server address")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runFakeBackend(cmd *cobra.Command, args []string) {
	runChat(func() chat.Backend {
		return NewFakeBackend()
	})
}

func runHTTPBackend(cmd *cobra.Command, args []string) {
	runChat(func() chat.Backend {
		return NewHTTPBackend("/backend", WithLogFile("/tmp/http-backend.log"))
	})
}

func runChat(backendFactory func() chat.Backend) {
	status := &chat.Status{}

	manager := conversation.NewManager(conversation.WithMessages(
		conversation.NewChatMessage(conversation.RoleSystem, "Welcome to the chat application!"),
	))

	backend := backendFactory()

	options := []tea.ProgramOption{
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	}

	p := tea.NewProgram(
		chat.InitialModel(manager, backend, chat.WithStatus(status)),
		options...,
	)

	// Set up the HTTP server
	r := mux.NewRouter()

	// Set up the user backend
	userBackend := chat.NewUserBackend(status, chat.WithLogFile("/tmp/http-backend.log"))
	userBackend.SetProgram(p)
	r.PathPrefix("/user").Handler(http.StripPrefix("/user", userBackend.Router()))

	// Set up the HTTP backend
	if httpBackend, ok := backend.(*HTTPBackend); ok {
		httpBackend.SetRouter(r.PathPrefix("/backend").Subrouter())
	}

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

	// Set the program for the backend after initialization
	if setterBackend, ok := backend.(interface{ SetProgram(*tea.Program) }); ok {
		setterBackend.SetProgram(p)
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
