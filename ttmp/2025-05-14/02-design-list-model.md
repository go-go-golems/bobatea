## TL;DR

Below you’ll find **(1)** a concise table of the **Bubble Tea messages** your custom list needs to recognise, how each is handled, and which ones you should bubble‐up to the parent model; **(2)** a **re-usable, pagination-free `ListModel`** that keeps a sliding window of results (`MaxVisible`) and gracefully truncates long strings to the current terminal width; and **(3)** integration hints so you can drop the component into the completion-selector you built earlier.  Everything compiles against Bubble Tea v0.25+, uses `go-runewidth` for width-aware truncation, and follows Charmbracelet’s nested-model pattern.([Go Packages][1], [YouTube][2])

---

## 1 · Incoming messages and forwarding rules

| Msg type                                             | Consumed? | Action inside `ListModel`                                                     | Forward to parent?          |
| ---------------------------------------------------- | --------- | ----------------------------------------------------------------------------- | --------------------------- |
| `setItemsMsg []Item`<br>(new data from completioner) | **Yes**   | Replace `items`, reset cursor & offset if needed                              | **No**                      |
| `tea.KeyMsg`<br>**Up / k / Ctrl-P**                  | **Yes**   | Move cursor up; adjust `offset` when cursor leaves window                     | No                          |
| `tea.KeyMsg`<br>**Down / j / Ctrl-N**                | **Yes**   | Move cursor down; adjust `offset`                                             | No                          |
| `tea.KeyMsg` other                                   | No        | —                                                                             | **Yes** — let parent decide |
| `tea.WindowSizeMsg`                                  | **Yes**   | Store `width`, recompute `truncateWidth` (`width – padding`), force re-render | No                          |
| any msg not matched                                  | No        | —                                                                             | **Yes** (return unchanged)  |

*`ListModel.Update` returns `(ListModel, tea.Cmd)`.  For messages we don’t consume we simply return the model unchanged and **the original message** so the parent can keep processing.  This is the standard nested-model pattern recommended by Charmbracelet.([YouTube][2], [Inngest][3])*

---

## 2 · Reusable list component

### 2.1 Public API

```go
// Item is the minimal interface the list needs.
type Item interface {
	Display() string  // what will be printed
	ID() string       // unique id; not shown but can be used by parent
}

// New(maxVisible int) ListModel
// .SetItems([]Item)
// .Cursor() int
// .Selected() Item
// Bubble Tea plumbing: Init(), Update(msg), View()
```

### 2.2 Implementation

```go
package listmodel

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbletea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth" // width-aware trimming :contentReference[oaicite:2]{index=2}
)

type setItemsMsg []Item

type ListModel struct {
	items       []Item
	cursor      int
	offset      int          // top of the visible window
	maxVisible  int          // rows to display
	width       int          // current terminal width
	truncateW   int          // width available for text
	pointerRune string       // e.g. "› "
}

// New creates a list that shows at most maxVisible rows at once.
func New(maxVisible int) ListModel {
	return ListModel{
		maxVisible:  maxVisible,
		pointerRune: "› ",
	}
}

/* ---------- Bubble Tea component plumbing ---------- */

func (m ListModel) Init() tea.Cmd { return nil }

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case setItemsMsg:
		m.items = msg
		m.cursor, m.offset = 0, 0
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset--
				}
			}
			return m, nil
		case "down", "j", "ctrl+n":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.maxVisible {
					m.offset++
				}
			}
			return m, nil
		default:
			// not consumed → bubble up
			return m, tea.Printf("")
		}

	case tea.WindowSizeMsg: // emitted automatically on resize :contentReference[oaicite:3]{index=3}
		m.width = msg.Width
		m.truncateW = max(0, m.width-len(m.pointerRune))
		return m, nil
	}

	return m, nil // unhandled → propagate
}

func (m ListModel) View() string {
	if len(m.items) == 0 {
		return "(no results)"
	}
	var b strings.Builder
	limit := min(m.maxVisible, len(m.items))
	for i := 0; i < limit; i++ {
		idx := m.offset + i
		if idx >= len(m.items) {
			break
		}
		prefix := "  "
		if idx == m.cursor {
			prefix = m.pointerRune
		}
		text := m.items[idx].Display()
		if m.truncateW > 0 {
			text = runewidth.Truncate(text, m.truncateW, "…") // ellipsis when wider :contentReference[oaicite:4]{index=4}
		}
		b.WriteString(prefix + text + "\n")
	}
	return b.String()
}

/* ---------- helpers ---------- */

func (m *ListModel) SetItems(it []Item) tea.Cmd { return func() tea.Msg { return setItemsMsg(it) } }

func (m ListModel) Cursor() int   { return m.cursor }
func (m ListModel) Selected() Item { if len(m.items) == 0 { return nil }; return m.items[m.cursor] }

func min(a,b int) int { if a<b { return a }; return b }
func max(a,b int) int { if a>b { return a }; return b }
```

### 2.3 Why no pagination?

Instead of tracking pages we keep a **sliding window** using `offset`. Up/down merely shift the window when the cursor leaves it, which is simpler than maintaining page numbers and works well when the completioner returns only hundreds of rows.([Go Packages][4], [GitHub][5])

---

## 3 · Integration hints

1. **Wrap as a nested model**
   In your completion selector’s `model` add a field `list listmodel.ListModel`.  In `Init` call `listmodel.New(maxVisible)`.  In your `Update`, delegate:

   ```go
   m.list, cmdList := m.list.Update(msg).(listmodel.ListModel)
   ```

   Combine `cmdList` with any other commands you’re returning. Nested-model composition is idiomatic Bubble Tea and keeps each component focused.([YouTube][2])

2. **Feeding results**
   When completions arrive:

   ```go
   cmds = append(cmds, m.list.SetItems(convertToItems(matches)))
   ```

   where `convertToItems` wraps your suggestion struct into something that satisfies `listmodel.Item`.

3. **Reading the selection**
   After the user presses **Enter/Tab** in the outer model, call `sel := m.list.Selected()` to retrieve the chosen suggestion (full metadata intact).

4. **Styling**
   The component prints plain text; decorate the selection line externally if you want colours (`lipgloss` styles inside `prefix` or the line string). You can also inject a custom `pointerRune` in the constructor.([Omri Bornstein][6], [GitHub][7])

5. **Handling variable width**
   The `WindowSizeMsg` fires once at program start and again on every resize, so the list re-computes its truncation width automatically.([GitHub][8], [Inngest][3])

---

## References

1. Bubble Tea package docs — message types & nested models ([Go Packages][1])
2. Bubbles/list source for scrolling & pointer conventions ([Go Packages][4])
3. Bubble Tea GitHub example “pager” — manual window handling ([GitHub][9])
4. Bubble Tea issue on `WindowSizeMsg` ordering ([GitHub][8])
5. Inngest blog: interactive CLIs with Bubble Tea, resize handling ([Inngest][3])
6. `go-runewidth` package for width-aware truncation ([Go Packages][10])
7. `runewidth.Truncate` implementation details ([GitHub][11])
8. StackOverflow discussion on ellipsis & truncation in Go ([Stack Overflow][12])
9. Charmbracelet Bubbles README — component design tips ([GitHub][5])
10. TUI Components in Go (blog post) — handling resize & state ([Omri Bornstein][6])
11. YouTube: Bubble Tea nested models walkthrough ([YouTube][2])

[1]: https://pkg.go.dev/github.com/charmbracelet/bubbletea?utm_source=chatgpt.com "tea package - github.com/charmbracelet/bubbletea - Go Packages"
[2]: https://www.youtube.com/watch?v=uJ2egAkSkjg&utm_source=chatgpt.com "Bubble Tea Nested Models - YouTube"
[3]: https://www.inngest.com/blog/interactive-clis-with-bubbletea?utm_source=chatgpt.com "Rapidly building interactive CLIs in Go with Bubbletea - Inngest Blog"
[4]: https://pkg.go.dev/github.com/charmbracelet/bubbles/list?utm_source=chatgpt.com "list package - github.com/charmbracelet/bubbles/list - Go Packages"
[5]: https://github.com/charmbracelet/bubbles?utm_source=chatgpt.com "charmbracelet/bubbles: TUI components for Bubble Tea - GitHub"
[6]: https://applegamer22.github.io/posts/go/bubbletea/?utm_source=chatgpt.com "TUI Components in Go with Bubble Tea - Omri Bornstein"
[7]: https://github.com/charmbracelet/bubbletea?utm_source=chatgpt.com "charmbracelet/bubbletea: A powerful little TUI framework - GitHub"
[8]: https://github.com/charmbracelet/bubbletea/issues/282?utm_source=chatgpt.com "View() called before WindowSizeMsg · Issue #282 - GitHub"
[9]: https://github.com/charmbracelet/bubbletea/blob/master/examples/pager/main.go?utm_source=chatgpt.com "bubbletea/examples/pager/main.go at main - GitHub"
[10]: https://pkg.go.dev/github.com/mattn/go-runewidth?utm_source=chatgpt.com "runewidth package - github.com/mattn/go-runewidth - Go Packages"
[11]: https://github.com/mattn/go-runewidth/blob/master/runewidth.go?utm_source=chatgpt.com "go-runewidth/runewidth.go at master · mattn/go-runewidth - GitHub"
[12]: https://stackoverflow.com/questions/59955085/how-can-i-elliptically-truncate-text-in-golang?utm_source=chatgpt.com "string - How can I elliptically truncate text in golang? - Stack Overflow"
