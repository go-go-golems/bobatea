package contextpanel

import (
	"context"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Trigger describes why a context panel request was triggered.
type Trigger string

const (
	TriggerToggleOpen    Trigger = "toggle-open"
	TriggerTyping        Trigger = "typing"
	TriggerManualRefresh Trigger = "manual-refresh"
)

type Dock string

const (
	DockAboveRepl Dock = "above-repl"
	DockRight     Dock = "right"
	DockLeft      Dock = "left"
	DockBottom    Dock = "bottom"
)

// Request captures current input context for contextual panel lookup.
type Request struct {
	Input      string
	CursorByte int
	RequestID  uint64
	Trigger    Trigger
}

// Document describes what should be shown in the contextual panel.
type Document struct {
	Show        bool
	Title       string
	Subtitle    string
	Markdown    string
	Diagnostics []string
	VersionTag  string
}

type Provider interface {
	GetContextPanel(ctx context.Context, req Request) (Document, error)
}

type DebounceMsg struct {
	RequestID uint64
}

type ResultMsg struct {
	RequestID uint64
	Doc       Document
	Err       error
}

type Config struct {
	Debounce           time.Duration
	RequestTimeout     time.Duration
	PrefetchWhenHidden bool
	Dock               Dock
	WidthPercent       int
	HeightPercent      int
	Margin             int
}

type OverlayLayout struct {
	PanelX        int
	PanelY        int
	PanelWidth    int
	PanelHeight   int
	ContentWidth  int
	ContentHeight int
}

type RenderOptions struct {
	ToggleBinding  string
	RefreshBinding string
	PinBinding     string
	FooterRenderer func(s string) string
	PanelStyle     *lipgloss.Style
	TitleStyle     *lipgloss.Style
	SubtitleStyle  *lipgloss.Style
}
