package chat

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type UserBackend struct {
	p      *tea.Program
	mu     sync.Mutex
	logger zerolog.Logger
}

type UserBackendOption func(*UserBackend)

func WithLogger(logger zerolog.Logger) UserBackendOption {
	return func(ub *UserBackend) {
		ub.logger = logger
	}
}

func WithLogFile(path string) UserBackendOption {
	return func(ub *UserBackend) {
		logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		ub.logger = zerolog.New(logFile).With().Timestamp().Logger()
	}
}

func NewUserBackend(options ...UserBackendOption) *UserBackend {
	ub := &UserBackend{
		logger: zerolog.New(io.Discard), // Default to a no-op logger
	}

	for _, option := range options {
		option(ub)
	}

	ub.logger.Debug().Msg("UserBackend initialized")
	return ub
}

func (u *UserBackend) SetProgram(p *tea.Program) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.p = p
	u.logger.Debug().Msg("Program set for UserBackend")
}

func (u *UserBackend) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/toggle-help", u.handleToggleHelp).Methods("POST")
	r.HandleFunc("/unfocus-message", u.handleUnfocusMessage).Methods("POST")
	r.HandleFunc("/quit", u.handleQuit).Methods("POST")
	r.HandleFunc("/focus-message", u.handleFocusMessage).Methods("POST")
	r.HandleFunc("/select-next-message", u.handleSelectNextMessage).Methods("POST")
	r.HandleFunc("/select-prev-message", u.handleSelectPrevMessage).Methods("POST")
	r.HandleFunc("/submit-message", u.handleSubmitMessage).Methods("POST")
	r.HandleFunc("/copy-to-clipboard", u.handleCopyToClipboard).Methods("POST")
	r.HandleFunc("/copy-last-response-to-clipboard", u.handleCopyLastResponseToClipboard).Methods("POST")
	r.HandleFunc("/copy-last-source-blocks-to-clipboard", u.handleCopyLastSourceBlocksToClipboard).Methods("POST")
	r.HandleFunc("/copy-source-blocks-to-clipboard", u.handleCopySourceBlocksToClipboard).Methods("POST")
	r.HandleFunc("/save-to-file", u.handleSaveToFile).Methods("POST")
	r.HandleFunc("/cancel-completion", u.handleCancelCompletion).Methods("POST")
	r.HandleFunc("/dismiss-error", u.handleDismissError).Methods("POST")
	r.HandleFunc("/input-text", u.handleInputText).Methods("POST")
	u.logger.Debug().Msg("Router set up for UserBackend")
	return r
}

func (u *UserBackend) sendUserAction(msg UserActionMsg) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.p != nil {
		u.logger.Info().Msgf("Sending user action: %T", msg)
		u.p.Send(msg)
	} else {
		u.logger.Error().Msg("Program not set for UserBackend")
	}
}

func (u *UserBackend) handleToggleHelp(w http.ResponseWriter, r *http.Request) {
	u.logger.Debug().Msg("Handling toggle help request")
	u.sendUserAction(ToggleHelpMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleUnfocusMessage(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(UnfocusMessageMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleQuit(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(QuitMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleFocusMessage(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(FocusMessageMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleSelectNextMessage(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(SelectNextMessageMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleSelectPrevMessage(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(SelectPrevMessageMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleSubmitMessage(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(SubmitMessageMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleCopyToClipboard(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(CopyToClipboardMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleCopyLastResponseToClipboard(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(CopyLastResponseToClipboardMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleCopyLastSourceBlocksToClipboard(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(CopyLastSourceBlocksToClipboardMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleCopySourceBlocksToClipboard(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(CopySourceBlocksToClipboardMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleSaveToFile(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(SaveToFileMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleCancelCompletion(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(CancelCompletionMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleDismissError(w http.ResponseWriter, r *http.Request) {
	u.sendUserAction(DismissErrorMsg{})
	w.WriteHeader(http.StatusOK)
}

func (u *UserBackend) handleInputText(w http.ResponseWriter, r *http.Request) {
	u.logger.Debug().Msg("Handling input text request")
	var input struct {
		Text string `json:"text"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		u.logger.Error().Err(err).Msg("Failed to decode input text")
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	u.logger.Debug().Str("text", input.Text).Msg("Input text received")
	u.sendUserAction(InputTextMsg{Text: input.Text})
	w.WriteHeader(http.StatusOK)
}
