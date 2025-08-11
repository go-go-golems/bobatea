package chatstyle

import (
    "fmt"
    "strings"
)

// RenderBox renders a single message box with the given role and text using the provided style.
// width is the outer width (including borders). The text is not markdown-rendered here.
func RenderBox(s *Style, role string, text string, width int, selected bool, focused bool) string {
    line := fmt.Sprintf("(%s): %s", role, text)
    sty := s.UnselectedMessage
    if selected {
        sty = s.SelectedMessage
    } else if focused {
        sty = s.FocusedMessage
    }
    // Compute inner width with padding
    frameW, _ := sty.GetFrameSize()
    inner := width - frameW
    if inner < 0 { inner = len(line) }
    if len(line) > inner { line = line[:inner] }
    if len(line) < inner { line = line + strings.Repeat(" ", inner-len(line)) }
    return sty.Width(width - sty.GetHorizontalPadding()).Render(line)
}


