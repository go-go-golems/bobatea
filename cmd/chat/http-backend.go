package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
	conversationui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
	"github.com/go-go-golems/bobatea/pkg/conversation"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type HTTPBackend struct {
	p               *tea.Program
	cancel          context.CancelFunc
	isRunning       bool
	messages        []*conversation.Message
	status          string
	lastMessage     *conversationui.StreamCompletionMsg
	lastError       error
	mu              sync.Mutex
	server          *http.Server
	completion      string
	logger          zerolog.Logger
	currentMetadata conversationui.StreamMetadata
}

var _ chat.Backend = &HTTPBackend{}

func NewHTTPBackend(addr string) *HTTPBackend {
	// Set up zerolog
	logFile, err := os.OpenFile("/tmp/http-backend.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger := zerolog.New(logFile).With().Timestamp().Logger()

	backend := &HTTPBackend{
		status: "idle",
		logger: logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/status", backend.handleStatus).Methods("GET")
	r.HandleFunc("/start", backend.handleStart).Methods("POST")
	r.HandleFunc("/completion", backend.handleCompletion).Methods("POST")
	r.HandleFunc("/status-update", backend.handleStatusUpdate).Methods("POST")
	r.HandleFunc("/done", backend.handleDone).Methods("POST")
	r.HandleFunc("/error", backend.handleError).Methods("POST")
	r.HandleFunc("/finish", backend.handleFinish).Methods("POST")

	backend.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := backend.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	backend.logger.Info().Msg("HTTPBackend initialized")
	return backend
}

func (h *HTTPBackend) SetProgram(p *tea.Program) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.p = p
	h.logger.Debug().Msg("Program set for HTTPBackend")
}

func (h *HTTPBackend) Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error) {
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
	parentID := conversation.NullNode
	if len(msgs) > 0 {
		parentID = msgs[len(msgs)-1].ID
	}
	h.currentMetadata = conversationui.StreamMetadata{
		ID:       conversation.NewNodeID(),
		ParentID: parentID,
	}

	h.logger.Info().
		Int("messageCount", len(msgs)).
		Str("parentID", h.currentMetadata.ParentID.String()).
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
		Messages    []*conversation.Message             `json:"messages"`
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
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Handling start request")

	if !h.isRunning {
		h.logger.Error().
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	h.completion = "" // Reset completion for new session
	h.logger.Info().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Sending StreamStartMsg to program")
	h.p.Send(conversationui.StreamStartMsg{
		StreamMetadata: h.currentMetadata,
	})
	w.WriteHeader(http.StatusOK)

	h.logger.Info().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Stream started")
}

func (h *HTTPBackend) handleCompletion(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Handling completion request")

	if !h.isRunning {
		h.logger.Error().
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamCompletionMsg
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.logger.Error().
			Err(err).
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
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
	h.logger.Info().
		Str("delta", msg.Delta).
		Str("completion", h.completion).
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Sending StreamCompletionMsg to program")
	h.p.Send(msg)
	w.WriteHeader(http.StatusOK)

	h.logger.Debug().
		Str("delta", msg.Delta).
		Str("completion", h.completion).
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Completion updated")
}

func (h *HTTPBackend) handleStatusUpdate(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Handling status update request")

	if !h.isRunning {
		h.logger.Error().
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamStatusMsg
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.logger.Error().
			Err(err).
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Failed to decode StreamStatusMsg")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.StreamMetadata = h.currentMetadata // Use the current metadata

	h.logger.Info().
		Str("text", msg.Text).
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Sending StreamStatusMsg to program")
	h.p.Send(msg)
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPBackend) handleDone(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Handling done request")

	if !h.isRunning {
		h.logger.Error().
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamDoneMsg
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.logger.Error().
			Err(err).
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Failed to decode StreamDoneMsg")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.StreamMetadata = h.currentMetadata // Use the current metadata

	// Use accumulated completion if not provided in the message
	if msg.Completion == "" {
		msg.Completion = h.completion
	}

	h.logger.Info().
		Str("completion", msg.Completion).
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Sending StreamDoneMsg to program")
	h.p.Send(msg)

	h.isRunning = false
	h.status = "finished"
	h.completion = "" // Reset completion for next session
	h.logger.Info().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Sending BackendFinishedMsg to program")
	h.p.Send(chat.BackendFinishedMsg{})
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPBackend) handleError(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Handling error request")

	if !h.isRunning {
		h.logger.Error().
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Backend is not running")
		http.Error(w, "Backend is not running", http.StatusBadRequest)
		return
	}

	var msg conversationui.StreamCompletionError
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.logger.Error().
			Err(err).
			Str("parentID", h.currentMetadata.ParentID.String()).
			Str("messageID", h.currentMetadata.ID.String()).
			Msg("Failed to decode StreamCompletionError")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.StreamMetadata = h.currentMetadata // Use the current metadata

	h.lastError = msg.Err
	h.isRunning = false
	h.status = "error"
	h.logger.Error().
		Err(msg.Err).
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Sending StreamCompletionError to program")
	h.p.Send(msg)

	h.isRunning = false
	h.status = "finished"
	h.logger.Info().
		Str("parentID", h.currentMetadata.ParentID.String()).
		Str("messageID", h.currentMetadata.ID.String()).
		Msg("Sending BackendFinishedMsg to program")
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
