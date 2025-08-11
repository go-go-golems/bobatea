package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"

	conversation2 "github.com/go-go-golems/geppetto/pkg/conversation"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
	conversationui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type HTTPBackend struct {
	p               *tea.Program
	cancel          context.CancelFunc
	isRunning       bool
	messages        []*conversation2.Message
	status          string
	lastMessage     *conversationui.StreamCompletionMsg
	lastError       error
	mu              sync.Mutex
	completion      string
	logger          zerolog.Logger
	currentMetadata conversationui.StreamMetadata
}

type HTTPBackendOption func(*HTTPBackend)

func WithLogger(logger zerolog.Logger) HTTPBackendOption {
	return func(hb *HTTPBackend) {
		hb.logger = logger
	}
}

func WithLogFile(path string) HTTPBackendOption {
	return func(hb *HTTPBackend) {
		logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		hb.logger = zerolog.New(logFile).With().Timestamp().Logger()
	}
}

var _ chat.Backend = &HTTPBackend{}

func NewHTTPBackend(prefix string, options ...HTTPBackendOption) *HTTPBackend {
	backend := &HTTPBackend{
		status: "idle",
		logger: zerolog.New(io.Discard), // Default to a no-op logger
	}

	for _, option := range options {
		option(backend)
	}

	backend.logger.Info().Msg("HTTPBackend initialized")
	return backend
}

func (h *HTTPBackend) SetProgram(p *tea.Program) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.p = p
	h.logger.Debug().Msg("Program set for HTTPBackend")
}

func (h *HTTPBackend) Start(ctx context.Context, msgs []*conversation2.Message) (tea.Cmd, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Info().Msg("Starting HTTPBackend")

	if h.isRunning {
		return nil, errors.New("Backend is already running")
	}

	h.messages = msgs
	h.isRunning = true
	h.status = "waiting"

	// Update the currentMetadata
    h.currentMetadata = conversationui.StreamMetadata{
        ID:       conversation2.NewNodeID(),
    }

    h.logger.Info().
        Int("messageCount", len(msgs)).
        Str("messageID", h.currentMetadata.ID.String()).
        Msg("HTTPBackend started")

	return func() tea.Msg {
		ctx, h.cancel = context.WithCancel(ctx)
		return conversationui.StreamStartMsg{
			StreamMetadata: h.currentMetadata,
		}
	}, nil
}

func (h *HTTPBackend) Interrupt() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cancel != nil {
		h.cancel()
	}
	h.status = "interrupted"
	h.logger.Info().Msg("HTTPBackend interrupted")
}

func (h *HTTPBackend) Kill() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cancel != nil {
		h.cancel()
	}
	h.isRunning = false
	h.status = "killed"
	h.logger.Info().Msg("HTTPBackend killed (server still running)")
}

func (h *HTTPBackend) IsFinished() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return !h.isRunning
}

func (h *HTTPBackend) handleStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().Msg("Handling status request")

	status := struct {
		Status      string                              `json:"status"`
		Messages    []*conversation2.Message            `json:"messages"`
		LastMessage *conversationui.StreamCompletionMsg `json:"last_message,omitempty"`
		LastError   string                              `json:"last_error,omitempty"`
	}{
		Status:      h.status,
		Messages:    h.messages,
		LastMessage: h.lastMessage,
	}

	if h.lastError != nil {
		status.LastError = h.lastError.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (h *HTTPBackend) handleStart(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

    h.logger.Debug().
        Str("messageID", h.currentMetadata.ID.String()).
        Msg("Handling start request")

    if !h.isRunning {
        h.logger.Error().
            Str("messageID", h.currentMetadata.ID.String()).
            Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	h.completion = "" // Reset completion for new session
    h.logger.Info().
        Str("messageID", h.currentMetadata.ID.String()).
        Msg("Sending StreamStartMsg to program")
	h.p.Send(conversationui.StreamStartMsg{
		StreamMetadata: h.currentMetadata,
	})
	w.WriteHeader(http.StatusOK)

    h.logger.Info().
        Str("messageID", h.currentMetadata.ID.String()).
        Msg("Stream started")
}

func (h *HTTPBackend) handleCompletion(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

    l := h.logger.With().
        Str("messageID", h.currentMetadata.ID.String()).
        Logger()
    l.Debug().Msg("Handling completion request")

    if !h.isRunning {
        l.Error().Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamCompletionMsg
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        l.Error().
            Err(err).
            Msg("Failed to decode StreamCompletionMsg")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.StreamMetadata = h.currentMetadata // Use the current metadata

	// Accumulate delta if completion is not provided
	if msg.Completion == "" {
		h.completion += msg.Delta
		msg.Completion = h.completion
	} else {
		// If completion is provided, update the stored completion
		h.completion = msg.Completion
	}

    h.lastMessage = &msg
    l.Info().
        Str("delta", msg.Delta).
        Str("completion", h.completion).
        Msg("Sending StreamCompletionMsg to program")
	h.p.Send(msg)
	w.WriteHeader(http.StatusOK)

    l.Debug().
        Str("delta", msg.Delta).
        Str("completion", h.completion).
        Msg("Completion updated")
}

func (h *HTTPBackend) handleStatusUpdate(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

    l := h.logger.With().
        Str("messageID", h.currentMetadata.ID.String()).
        Logger()
    l.Debug().Msg("Handling status update request")

    if !h.isRunning {
        l.Error().Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamStatusMsg
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        l.Error().
            Err(err).
            Msg("Failed to decode StreamStatusMsg")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.StreamMetadata = h.currentMetadata // Use the current metadata

    l.Info().
        Str("text", msg.Text).
        Msg("Sending StreamStatusMsg to program")
	h.p.Send(msg)
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPBackend) handleDone(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

    l := h.logger.With().
        Str("messageID", h.currentMetadata.ID.String()).
        Logger()
    l.Debug().Msg("Handling done request")

    if !h.isRunning {
        l.Error().Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamDoneMsg
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        l.Error().
            Err(err).
            Msg("Failed to decode StreamDoneMsg")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.StreamMetadata = h.currentMetadata // Use the current metadata

	// Use accumulated completion if not provided in the message
	if msg.Completion == "" {
		msg.Completion = h.completion
	}

    l.Info().
        Str("completion", msg.Completion).
        Msg("Sending StreamDoneMsg to program")
	h.p.Send(msg)

	h.isRunning = false
	h.status = "finished"
	h.completion = "" // Reset completion for next session
    l.Info().Msg("Sending BackendFinishedMsg to program")
	h.p.Send(chat.BackendFinishedMsg{})
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPBackend) handleError(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

    l := h.logger.With().
        Str("messageID", h.currentMetadata.ID.String()).
        Logger()
    l.Debug().Msg("Handling error request")

    if !h.isRunning {
        l.Error().Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamCompletionError
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        l.Error().
            Err(err).
            Msg("Failed to decode StreamCompletionError")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.StreamMetadata = h.currentMetadata // Use the current metadata

	h.lastError = msg.Err
	h.isRunning = false
	h.status = "error"
    l.Error().
        Err(msg.Err).
        Msg("Sending StreamCompletionError to program")
	h.p.Send(msg)

	h.isRunning = false
	h.status = "finished"
    l.Info().Msg("Sending BackendFinishedMsg to program")
	h.p.Send(chat.BackendFinishedMsg{})
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPBackend) handleFinish(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().Msg("Handling finish request")

	if !h.isRunning {
		h.logger.Error().Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	h.isRunning = false
	h.status = "finished"
	h.logger.Info().Msg("Sending BackendFinishedMsg to program")
	h.p.Send(chat.BackendFinishedMsg{})
	w.WriteHeader(http.StatusOK)
}

// Add this method to the HTTPBackend struct
func (h *HTTPBackend) SetRouter(r *mux.Router) {
	r.HandleFunc("/status", h.handleStatus).Methods("GET")
	r.HandleFunc("/start", h.handleStart).Methods("POST")
	r.HandleFunc("/completion", h.handleCompletion).Methods("POST")
	r.HandleFunc("/status-update", h.handleStatusUpdate).Methods("POST")
	r.HandleFunc("/done", h.handleDone).Methods("POST")
	r.HandleFunc("/error", h.handleError).Methods("POST")
	r.HandleFunc("/finish", h.handleFinish).Methods("POST")
}
