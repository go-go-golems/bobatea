package chat

import (
	"context"
	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat/conversation"
)

// These messages are used by the backend to send new streaming data

type Backend interface {
	Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error)
	Interrupt()
	Kill()

	IsFinished() bool
}

type BackendFinishedMsg struct{}
