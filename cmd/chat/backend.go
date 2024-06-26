package main

import (
	"context"
	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
	conversationui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
	"github.com/go-go-golems/bobatea/pkg/conversation"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type FakeBackend struct {
	p         *tea.Program
	cancel    context.CancelFunc
	isRunning bool
}

var _ chat.Backend = &FakeBackend{}

func NewBackend(p *tea.Program) *FakeBackend {
	return &FakeBackend{
		p: p,
	}
}

func (f *FakeBackend) Start(ctx context.Context, msgs []*conversation.Message) (tea.Cmd, error) {
	if f.isRunning {
		return nil, errors.New("Step is already running")
	}
	return func() tea.Msg {
		ctx, f.cancel = context.WithCancel(ctx)
		if len(msgs) == 0 {
			return nil
		}
		lastMsg := msgs[len(msgs)-1]
		words := strings.Fields(lastMsg.Content.String())
		reversedWords := reverseWords(words)
		msg := strings.Join(reversedWords, " ")

		metadata := conversationui.StreamMetadata{
			ID:       conversation.NewNodeID(),
			ParentID: lastMsg.ID,
		}

		go func() {
			tick := time.Tick(200 * time.Millisecond)
			idx := 0
			defer func() {
				f.p.Send(chat.BackendFinishedMsg{})
				f.cancel()
				f.cancel = nil
				f.isRunning = false
			}()
			for {
				select {
				case <-ctx.Done():
					return

				case <-tick:
					if idx < len(reversedWords) {
						completion := strings.Join(reversedWords[:idx+1], " ")
						f.p.Send(
							conversationui.StreamCompletionMsg{
								StreamMetadata: metadata,
								Delta:          reversedWords[idx] + " ",
								Completion:     completion,
							},
						)
						idx++
					} else {
						f.p.Send(conversationui.StreamDoneMsg{
							StreamMetadata: metadata,
							Completion:     msg,
						})
						return
					}
				}
			}
		}()

		return conversationui.StreamStartMsg{
			StreamMetadata: metadata,
		}
	}, nil
}

func (f *FakeBackend) Interrupt() {
	if f.cancel != nil {
		f.cancel()
	}
}

func (f *FakeBackend) Kill() {
	if f.cancel != nil {
		f.cancel()
	}
}

func (f *FakeBackend) IsFinished() bool {
	return !f.isRunning
}

func reverseWords(words []string) []string {
	reversed := make([]string, len(words))
	for i, word := range words {
		reversed[len(words)-1-i] = word
	}
	return reversed
}
