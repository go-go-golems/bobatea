package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	renderers "github.com/go-go-golems/bobatea/pkg/timeline/renderers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type demoModel struct {
	vp      viewport.Model
	ctrl    *timeline.Controller
	counter int
}

func (m demoModel) Init() tea.Cmd {
	log.Info().Msg("Init called")
	return tea.EnterAltScreen
}

func (m demoModel) View() string {
	header := "Timeline demo: press t = stream text, o = tool calls, q = quit\n\n"
	body := m.vp.View()
	// Ensure we always print header even if viewport height is 0
	out := header + body
	log.Info().Int("view_len", len(out)).Int("body_len", len(body)).Int("vp_w", m.vp.Width).Int("vp_h", m.vp.Height).Msg("View called")
	return out
}

// message types for scheduled steps
type startText struct{ id timeline.EntityID }
type updateText struct {
	id      timeline.EntityID
	text    string
	version int64
}
type finishText struct {
	id   timeline.EntityID
	text string
}

type startTools struct{ id timeline.EntityID }
type addTool struct {
	id      timeline.EntityID
	call    map[string]any
	version int64
}

func (m demoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		// Reserve space for the header so it stays visible
		headerLines := 2
		m.vp.Width = v.Width
		if v.Height > headerLines {
			m.vp.Height = v.Height - headerLines
		} else {
			m.vp.Height = 0
		}
		m.vp.YPosition = headerLines
		m.ctrl.SetSize(v.Width, v.Height)
		log.Info().Int("w", v.Width).Int("h", v.Height).Int("vp_w", m.vp.Width).Int("vp_h", m.vp.Height).Msg("WindowSizeMsg")
		m.vp.SetContent(m.ctrl.View())
	case tea.KeyMsg:
		log.Info().Str("key", v.String()).Msg("KeyMsg")
		switch v.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			// Ask selected entity to copy either code block or text
			// First try code; if none, model should fall back to plain text
			return m, m.ctrl.SendToSelected(timeline.EntityCopyCodeMsg{})
		case "t":
			// start a streaming text entity
			id := timeline.EntityID{LocalID: fmt.Sprintf("text-%d", m.counter), Kind: "llm_text"}
			m.counter++
			now := time.Now()
			m.ctrl.OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "llm_text"}, Props: map[string]any{"text": ""}, StartedAt: now})
			// schedule partials
			return m, tea.Batch(
				tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg { return updateText{id: id, text: "Hello", version: 1} }),
				tea.Tick(700*time.Millisecond, func(time.Time) tea.Msg { return updateText{id: id, text: "Hello, world", version: 2} }),
				tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg { return finishText{id: id, text: "Hello, world!"} }),
			)
		case "o":
			// start a tools panel and stream calls/results
			id := timeline.EntityID{LocalID: fmt.Sprintf("tools-%d", m.counter), Kind: "tool_calls_panel"}
			m.counter++
			now := time.Now()
			m.ctrl.OnCreated(timeline.UIEntityCreated{ID: id, Renderer: timeline.RendererDescriptor{Kind: "tool_calls_panel"}, Props: map[string]any{"calls": []any{}}, StartedAt: now})
			return m, tea.Batch(
				tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
					return addTool{id: id, version: 1, call: map[string]any{"id": "a", "name": "search", "status": "pending"}}
				}),
				tea.Tick(900*time.Millisecond, func(time.Time) tea.Msg {
					return addTool{id: id, version: 2, call: map[string]any{"id": "a", "name": "search", "status": "done", "result": "ok"}}
				}),
				tea.Tick(1300*time.Millisecond, func(time.Time) tea.Msg {
					return addTool{id: id, version: 3, call: map[string]any{"id": "b", "name": "write", "status": "pending"}}
				}),
				tea.Tick(1800*time.Millisecond, func(time.Time) tea.Msg {
					return addTool{id: id, version: 4, call: map[string]any{"id": "b", "name": "write", "status": "done", "result": "saved"}}
				}),
			)
		}
	case updateText:
		log.Info().Str("entity", v.id.LocalID).Int64("version", v.version).Str("text", v.text).Msg("updateText")
		m.ctrl.OnUpdated(timeline.UIEntityUpdated{ID: v.id, Patch: map[string]any{"text": v.text}, Version: v.version, UpdatedAt: time.Now()})
	case finishText:
		log.Info().Str("entity", v.id.LocalID).Str("text", v.text).Msg("finishText")
		m.ctrl.OnCompleted(timeline.UIEntityCompleted{ID: v.id, Result: map[string]any{"text": v.text}})
	case addTool:
		log.Info().Str("entity", v.id.LocalID).Int64("version", v.version).Msg("addTool")
		// Replace calls entirely for simplicity in demo
		m.ctrl.OnUpdated(timeline.UIEntityUpdated{ID: v.id, Patch: map[string]any{"calls": []any{v.call}}, Version: v.version, UpdatedAt: time.Now()})
	case timeline.CopyTextRequestedMsg:
		log.Info().Str("copied_text_len", fmt.Sprintf("%d", len(v.Text))).Msg("CopyTextRequested")
	case timeline.CopyCodeRequestedMsg:
		log.Info().Str("copied_code_len", fmt.Sprintf("%d", len(v.Code))).Msg("CopyCodeRequested")
	}
	m.vp.SetContent(m.ctrl.View())
	return m, nil
}

func main() {
	// initialize logging to tmp file
	f, err := os.OpenFile("/tmp/timeline-demo.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err == nil {
		log.Logger = zerolog.New(f).With().Timestamp().Logger()
		log.Info().Msg("logger initialized")
	}
	reg := timeline.NewRegistry()
	reg.RegisterModelFactory(renderers.NewLLMTextFactory())
	reg.RegisterModelFactory(renderers.ToolCallsPanelFactory{})
	reg.RegisterModelFactory(renderers.PlainFactory{})
	ctrl := timeline.NewController(reg)
	m := demoModel{vp: viewport.New(0, 0), ctrl: ctrl}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
