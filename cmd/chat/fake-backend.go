package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

func (f *FakeBackend) Start(ctx context.Context, prompt string) (tea.Cmd, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	log.Debug().Str("component", "fake_backend").Msg("Start: starting")

	if f.isRunning {
		log.Debug().Str("component", "fake_backend").Msg("Start: already running")
		return nil, errors.New("Step is already running")
	}

	if f.p == nil {
		log.Debug().Str("component", "fake_backend").Msg("Start: program not set")
		return nil, errors.New("Program not set")
	}

	f.isRunning = true
	log.Debug().Str("component", "fake_backend").Bool("is_running", f.isRunning).Msg("Start: initializing")

	return func() tea.Msg {
		log.Debug().Str("component", "fake_backend").Msg("Backend command: executing")
		ctx, f.cancel = context.WithCancel(ctx)
		content := prompt
		words := strings.Fields(content)
		reversedWords := reverseWords(words)
		localID := uuid.New().String()
		// Stream using timeline lifecycle; no legacy Stream* metadata

		go func() {
			log.Debug().Str("component", "fake_backend").Msg("Goroutine: started streaming loop")
			tick := time.Tick(100 * time.Millisecond)
			idx := 0
			defer func() {
				log.Debug().Str("component", "fake_backend").Msg("Goroutine: finishing, sending BackendFinishedMsg")
				f.p.Send(chat.BackendFinishedMsg{})
				f.cancel()
				f.cancel = nil
				f.isRunning = false
				log.Debug().Str("component", "fake_backend").Bool("is_running", f.isRunning).Msg("Goroutine: isRunning=false")
			}()
			// Recognize tool commands and emit timeline tool entities, as if produced by a real agent tool call
			if strings.HasPrefix(content, "/weather ") {
				// Emit entity lifecycle messages directly (created -> updated -> completed)
				f.p.Send(
					timeline.UIEntityCreated{
						ID:        timeline.EntityID{LocalID: localID, Kind: "tool_call"},
						Renderer:  timeline.RendererDescriptor{Key: "renderer.tool.get_weather.v1", Kind: "tool_call"},
						Props:     map[string]any{"location": "Paris", "units": "celsius"},
						StartedAt: time.Now(),
					},
				)
				f.p.Send(timeline.UIEntityUpdated{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}, Patch: map[string]any{"result": "22C, Sunny"}, Version: 1, UpdatedAt: time.Now()})
				f.p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}})
				return
			}
			if strings.HasPrefix(content, "/checkbox") {
				f.p.Send(timeline.UIEntityCreated{
					ID:        timeline.EntityID{LocalID: localID, Kind: "tool_call"},
					Renderer:  timeline.RendererDescriptor{Key: "renderer.test.checkbox.v1", Kind: "tool_call"},
					Props:     map[string]any{"label": "Enable turbo mode", "checked": false},
					StartedAt: time.Now(),
				})
				// keep it interactive; no completion yet
				return
			}
			if strings.HasPrefix(content, "/search ") {
				f.p.Send(timeline.UIEntityCreated{
					ID:        timeline.EntityID{LocalID: localID, Kind: "tool_call"},
					Renderer:  timeline.RendererDescriptor{Key: "renderer.tool.web_search.v1", Kind: "tool_call"},
					Props:     map[string]any{"query": "golang bubbletea timeline ui", "spin": 0},
					StartedAt: time.Now(),
				})
				// Stream progressive updates to showcase UIEntityUpdated
				links := []string{
					"https://golang.org",
					"https://github.com/charmbracelet/bubbletea",
					"https://github.com/go-go-golems/bobatea",
				}
				acc := ""
				for i, link := range links {
					time.Sleep(300 * time.Millisecond)
					if acc != "" {
						acc += ", "
					}
					acc += link
					f.p.Send(timeline.UIEntityUpdated{
						ID:        timeline.EntityID{LocalID: localID, Kind: "tool_call"},
						Patch:     map[string]any{"results": strings.Split(acc, ", "), "spin": i + 1},
						Version:   int64(i + 1),
						UpdatedAt: time.Now(),
					})
				}
				f.p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: localID, Kind: "tool_call"}})
				return
			}
			// Default: stream an assistant llm_text entity using timeline lifecycle
			log.Debug().Str("component", "fake_backend").Msg("Goroutine: creating assistant llm_text entity")
			assistantID := uuid.New().String()
			f.p.Send(timeline.UIEntityCreated{
				ID:        timeline.EntityID{LocalID: assistantID, Kind: "llm_text"},
				Renderer:  timeline.RendererDescriptor{Kind: "llm_text"},
				Props:     map[string]any{"role": "assistant", "text": ""},
				StartedAt: time.Now(),
			})
			for {
				select {
				case <-ctx.Done():
					log.Debug().Str("component", "fake_backend").Msg("Goroutine: ctx.Done")
					return

				case <-tick:
					if idx < len(reversedWords) {
						completion := strings.Join(reversedWords[:idx+1], " ")
						log.Debug().Int("idx", idx).Str("component", "fake_backend").Msg("Goroutine: sending UIEntityUpdated for assistant text")
						f.p.Send(timeline.UIEntityUpdated{
							ID:        timeline.EntityID{LocalID: assistantID, Kind: "llm_text"},
							Patch:     map[string]any{"text": completion},
							Version:   int64(idx + 1),
							UpdatedAt: time.Now(),
						})
						idx++
					} else {
						log.Debug().Str("component", "fake_backend").Msg("Goroutine: completing assistant llm_text entity")
						f.p.Send(timeline.UIEntityCompleted{ID: timeline.EntityID{LocalID: assistantID, Kind: "llm_text"}})
						return
					}
				}
			}
		}()

		log.Debug().Str("component", "fake_backend").Msg("Backend command: returning (no immediate UI msg)")
		return nil
	}, nil
}

func ptrFloat(v float64) *float64 { return &v }

// SubmitPrompt starts a streaming run from a single prompt string.
// SubmitPrompt removed: Start now accepts a plain prompt string

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
