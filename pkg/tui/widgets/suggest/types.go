package suggest

import (
	"context"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
)

// Reason describes why a suggestion request was triggered.
type Reason string

const (
	ReasonDebounce Reason = "debounce"
	ReasonShortcut Reason = "shortcut"
	ReasonManual   Reason = "manual"
)

type Placement string

const (
	PlacementAuto   Placement = "auto"
	PlacementAbove  Placement = "above"
	PlacementBelow  Placement = "below"
	PlacementBottom Placement = "bottom"
)

type HorizontalGrow string

const (
	HorizontalGrowRight HorizontalGrow = "right"
	HorizontalGrowLeft  HorizontalGrow = "left"
)

type Request struct {
	Input      string
	CursorByte int
	Reason     Reason
	Shortcut   string
	RequestID  uint64
}

type Result struct {
	Suggestions []autocomplete.Suggestion
	ReplaceFrom int
	ReplaceTo   int
	Show        bool
}

type Provider interface {
	CompleteInput(ctx context.Context, req Request) (Result, error)
}

type Config struct {
	Debounce       time.Duration
	RequestTimeout time.Duration
	MaxVisible     int
	PageSize       int
	MaxWidth       int
	MaxHeight      int
	MinWidth       int
	Margin         int
	OffsetX        int
	OffsetY        int
	NoBorder       bool
	Placement      Placement
	HorizontalGrow HorizontalGrow
}

type DebounceMsg struct {
	RequestID uint64
}

type ResultMsg struct {
	RequestID uint64
	Result    Result
	Err       error
}

type OverlayLayout struct {
	PopupX       int
	PopupY       int
	PopupWidth   int
	VisibleRows  int
	ContentWidth int
}

type Styles struct {
	Item     lipgloss.Style
	Selected lipgloss.Style
	Popup    lipgloss.Style
}

type Action int

const (
	ActionCancel Action = iota
	ActionPrev
	ActionNext
	ActionPageUp
	ActionPageDown
	ActionAccept
)

type Buffer interface {
	Value() string
	CursorByte() int
	SetValue(value string)
	SetCursorByte(cursor int)
}
