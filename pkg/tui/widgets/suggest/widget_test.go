package suggest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/autocomplete"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	result          Result
	err             error
	requests        []Request
	panicOnComplete bool
}

func (f *fakeProvider) CompleteInput(_ context.Context, req Request) (Result, error) {
	if f.panicOnComplete {
		panic("fake suggest panic")
	}
	f.requests = append(f.requests, req)
	return f.result, f.err
}

type fakeBuffer struct {
	value  string
	cursor int
}

func (b *fakeBuffer) Value() string         { return b.value }
func (b *fakeBuffer) CursorByte() int       { return b.cursor }
func (b *fakeBuffer) SetValue(value string) { b.value = value }
func (b *fakeBuffer) SetCursorByte(cursor int) {
	b.cursor = cursor
}

func newTestWidget(provider *fakeProvider) *Widget {
	return New(provider, Config{
		Debounce:       time.Nanosecond,
		RequestTimeout: 50 * time.Millisecond,
		MaxVisible:     8,
		MaxWidth:       56,
		MaxHeight:      12,
		MinWidth:       24,
		Margin:         1,
		Placement:      PlacementAuto,
		HorizontalGrow: HorizontalGrowRight,
	})
}

func TestDebounceCoalescesToLatestRequest(t *testing.T) {
	provider := &fakeProvider{}
	w := newTestWidget(provider)

	cmd1 := w.OnBufferChanged("", 0, "a", 1)
	require.NotNil(t, cmd1)
	msg1, ok := cmd1().(DebounceMsg)
	require.True(t, ok)

	cmd2 := w.OnBufferChanged("a", 1, "ab", 2)
	require.NotNil(t, cmd2)
	msg2, ok := cmd2().(DebounceMsg)
	require.True(t, ok)

	assert.Nil(t, w.HandleDebounce(context.Background(), msg1, "ab", 2))

	active := w.HandleDebounce(context.Background(), msg2, "ab", 2)
	require.NotNil(t, active)
	_, ok = active().(ResultMsg)
	require.True(t, ok)

	require.Len(t, provider.requests, 1)
	assert.Equal(t, ReasonDebounce, provider.requests[0].Reason)
	assert.Equal(t, msg2.RequestID, provider.requests[0].RequestID)
}

func TestResultDropsStaleResponse(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetRequestSeq(2)

	w.HandleResult(ResultMsg{
		RequestID: 1,
		Result: Result{
			Show: true,
			Suggestions: []autocomplete.Suggestion{
				{Id: "1", Value: "log", DisplayText: "log"},
			},
		},
	})
	assert.False(t, w.Visible())
	assert.Equal(t, uint64(0), w.LastRequestID())

	w.HandleResult(ResultMsg{
		RequestID: 2,
		Result: Result{
			Show: true,
			Suggestions: []autocomplete.Suggestion{
				{Id: "1", Value: "log", DisplayText: "log"},
			},
			ReplaceFrom: 0,
			ReplaceTo:   2,
		},
	})
	assert.True(t, w.Visible())
	assert.Equal(t, 0, w.Selection())
	assert.Equal(t, uint64(2), w.LastRequestID())
}

func TestNavigationAndApply(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetLastResult(Result{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "const", DisplayText: "const"},
			{Id: "2", Value: "console", DisplayText: "console"},
		},
		ReplaceFrom: 0,
		ReplaceTo:   4,
	})
	w.SetReplaceFrom(0)
	w.SetReplaceTo(4)
	w.SetSelection(0)

	buffer := &fakeBuffer{value: "cons", cursor: 4}
	handled := w.HandleNavigation(ActionNext, buffer)
	require.True(t, handled)
	assert.Equal(t, 1, w.Selection())

	handled = w.HandleNavigation(ActionAccept, buffer)
	require.True(t, handled)
	assert.Equal(t, "console", buffer.Value())
	assert.False(t, w.Visible())
}

func TestPageNavigation(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetVisibleRows(3)
	w.SetLastResult(Result{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "a", DisplayText: "a"},
			{Id: "2", Value: "b", DisplayText: "b"},
			{Id: "3", Value: "c", DisplayText: "c"},
			{Id: "4", Value: "d", DisplayText: "d"},
			{Id: "5", Value: "e", DisplayText: "e"},
			{Id: "6", Value: "f", DisplayText: "f"},
			{Id: "7", Value: "g", DisplayText: "g"},
		},
	})
	buffer := &fakeBuffer{}

	handled := w.HandleNavigation(ActionPageDown, buffer)
	require.True(t, handled)
	assert.Equal(t, 3, w.Selection())
	assert.GreaterOrEqual(t, w.ScrollTop(), 1)

	handled = w.HandleNavigation(ActionPageUp, buffer)
	require.True(t, handled)
	assert.Equal(t, 0, w.Selection())
	assert.Equal(t, 0, w.ScrollTop())
}

func TestDebounceDoesNotHideVisiblePopup(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetLastResult(Result{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
		},
		ReplaceFrom: 0,
		ReplaceTo:   2,
	})
	cmd := w.OnBufferChanged("co", 2, "con", 3)
	require.NotNil(t, cmd)
	assert.True(t, w.Visible())
}

func TestTriggerShortcutUsesShortcutReason(t *testing.T) {
	provider := &fakeProvider{
		result: Result{
			Show: true,
			Suggestions: []autocomplete.Suggestion{
				{Id: "1", Value: "log", DisplayText: "log"},
			},
		},
	}
	w := newTestWidget(provider)
	cmd := w.TriggerShortcut(context.Background(), "co", 2, "tab")
	require.NotNil(t, cmd)
	msg, ok := cmd().(ResultMsg)
	require.True(t, ok)
	assert.Equal(t, uint64(1), msg.RequestID)
	require.Len(t, provider.requests, 1)
	assert.Equal(t, ReasonShortcut, provider.requests[0].Reason)
	assert.Equal(t, "tab", provider.requests[0].Shortcut)
}

func TestLayoutClampsToBounds(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetMaxVisible(10)
	w.SetMaxHeight(6)
	w.SetMaxWidth(28)
	w.SetMinWidth(16)
	w.SetLastResult(Result{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
			{Id: "2", Value: "constructor", DisplayText: strings.Repeat("X", 90)},
			{Id: "3", Value: "copyWithin", DisplayText: "copyWithin"},
			{Id: "4", Value: "count", DisplayText: "count"},
			{Id: "5", Value: "countReset", DisplayText: "countReset"},
		},
	})

	layout, ok := w.ComputeOverlayLayout(40, 10, 1, 2, "> ", strings.Repeat("x", 140), 140, w.PopupStyle(lipgloss.NewStyle().Border(lipgloss.RoundedBorder())))
	require.True(t, ok)
	assert.GreaterOrEqual(t, layout.PopupX, 0)
	assert.GreaterOrEqual(t, layout.PopupY, 0)
	assert.LessOrEqual(t, layout.PopupWidth, 40)
	assert.LessOrEqual(t, layout.PopupWidth, w.MaxWidth())
	assert.GreaterOrEqual(t, layout.PopupWidth, w.MinWidth())
	assert.GreaterOrEqual(t, layout.VisibleRows, 1)
	assert.LessOrEqual(t, layout.VisibleRows, 4)
}

func TestLayoutBottomAndGrowLeft(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetPlacement(PlacementBottom)
	w.SetMargin(1)
	w.SetLastResult(Result{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "console", DisplayText: "console"},
			{Id: "2", Value: "const", DisplayText: "const"},
			{Id: "3", Value: "continue", DisplayText: "continue"},
		},
	})

	style := w.PopupStyle(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()))
	layout, ok := w.ComputeOverlayLayout(80, 24, 1, 1, "> ", "cons", 4, style)
	require.True(t, ok)
	frameHeight := style.GetVerticalFrameSize()
	expectedY := 24 - w.Margin() - (layout.VisibleRows + frameHeight)
	assert.Equal(t, expectedY, layout.PopupY)

	w.SetHorizontalGrow(HorizontalGrowRight)
	rightLayout, ok := w.ComputeOverlayLayout(120, 24, 1, 1, "> ", "console.log", len("console.log"), style)
	require.True(t, ok)
	w.SetHorizontalGrow(HorizontalGrowLeft)
	leftLayout, ok := w.ComputeOverlayLayout(120, 24, 1, 1, "> ", "console.log", len("console.log"), style)
	require.True(t, ok)
	expectedLeftX := maxInt(0, rightLayout.PopupX-leftLayout.PopupWidth)
	assert.Equal(t, expectedLeftX, leftLayout.PopupX)
}

func TestRenderPopupUsesScrollWindow(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	w.SetVisible(true)
	w.SetSelection(4)
	w.SetScrollTop(3)
	w.SetLastResult(Result{
		Show: true,
		Suggestions: []autocomplete.Suggestion{
			{Id: "1", Value: "alpha", DisplayText: "alpha"},
			{Id: "2", Value: "beta", DisplayText: "beta"},
			{Id: "3", Value: "gamma", DisplayText: "gamma"},
			{Id: "4", Value: "delta", DisplayText: "delta"},
			{Id: "5", Value: "epsilon", DisplayText: "epsilon"},
		},
	})
	popup := w.RenderPopup(Styles{
		Item:     lipgloss.NewStyle(),
		Selected: lipgloss.NewStyle().Bold(true),
		Popup:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
	}, OverlayLayout{PopupWidth: 24, VisibleRows: 2, ContentWidth: 18})
	assert.Contains(t, popup, "delta")
	assert.Contains(t, popup, "epsilon")
	assert.NotContains(t, popup, "alpha")
}

func TestNoBorderPopupStyleHasNoFrame(t *testing.T) {
	w := newTestWidget(&fakeProvider{})
	defaultStyle := w.PopupStyle(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()))
	assert.Greater(t, defaultStyle.GetHorizontalFrameSize(), 0)
	assert.Greater(t, defaultStyle.GetVerticalFrameSize(), 0)

	w.SetNoBorder(true)
	noBorder := w.PopupStyle(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()))
	assert.Equal(t, 0, noBorder.GetHorizontalFrameSize())
	assert.Equal(t, 0, noBorder.GetVerticalFrameSize())
}
