package main

import (
	"context"
	conversation2 "github.com/go-go-golems/geppetto/pkg/conversation"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
	conversationui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
    "github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/pkg/errors"
)

type FakeBackend struct {
	p         *tea.Program
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.Mutex
}

var _ chat.Backend = &FakeBackend{}

func NewFakeBackend() *FakeBackend {
	return &FakeBackend{}
}

func (f *FakeBackend) SetProgram(p *tea.Program) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.p = p
}

func (f *FakeBackend) Start(ctx context.Context, msgs []*conversation2.Message) (tea.Cmd, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.isRunning {
		return nil, errors.New("Step is already running")
	}

	if f.p == nil {
		return nil, errors.New("Program not set")
	}

	return func() tea.Msg {
		ctx, f.cancel = context.WithCancel(ctx)
		if len(msgs) == 0 {
			return nil
		}
        lastMsg := msgs[len(msgs)-1]
        content := lastMsg.Content.String()
        words := strings.Fields(content)
		reversedWords := reverseWords(words)
		msg := strings.Join(reversedWords, " ")

		metadata := conversationui.StreamMetadata{
			ID:       conversation2.NewNodeID(),
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
            // Recognize tool commands and emit timeline tool entities, as if produced by a real agent tool call
            if strings.HasPrefix(content, "/weather ") {
                // Emit entity lifecycle messages directly (created -> updated -> completed)
                localID := conversation2.NewNodeID().String()
                f.p.Send(
                    timeline.UIEntityCreated{
                        ID:       timeline.EntityID{LocalID: localID, Kind: "tool_call"},
                        Renderer: timeline.RendererDescriptor{Key: "renderer.tool.get_weather.v1", Kind: "tool_call"},
                        Props:    map[string]any{"location": "Paris", "units": "celsius"},
                        StartedAt: time.Now(),
                    },
                )
                f.p.Send(timeline.UIEntityUpdated{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}, Patch: map[string]any{"result": "22C, Sunny"}, Version: 1, UpdatedAt: time.Now()})
                f.p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}})
                // Also send assistant confirmation text via normal stream to exercise both paths
                f.p.Send(conversationui.StreamDoneMsg{StreamMetadata: metadata, Completion: "Using weather tool..."})
                return
            }
            if strings.HasPrefix(content, "/search ") {
                localID := conversation2.NewNodeID().String()
                f.p.Send(
                    timeline.UIEntityCreated{
                        ID:       timeline.EntityID{LocalID: localID, Kind: "tool_call"},
                        Renderer: timeline.RendererDescriptor{Key: "renderer.tool.web_search.v1", Kind: "tool_call"},
                        Props:    map[string]any{"query": "golang bubbletea timeline ui"},
                        StartedAt: time.Now(),
                    },
                )
                f.p.Send(timeline.UIEntityUpdated{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}, Patch: map[string]any{"result": "Found 3 relevant links"}, Version: 1, UpdatedAt: time.Now()})
                f.p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}})
                f.p.Send(conversationui.StreamDoneMsg{StreamMetadata: metadata, Completion: "Searching the web..."})
                return
            }
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
