package diff

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
)

// itemAdapter adapts DiffItem to bubbles/list.Item
type itemAdapter struct{ item DiffItem }

func (i itemAdapter) Title() string { return i.item.Name() }
func (i itemAdapter) Description() string {
	total := 0
	for _, c := range i.item.Categories() {
		if c == nil {
			continue
		}
		total += len(c.Changes())
	}
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("%d changes", total)
}
func (i itemAdapter) FilterValue() string { return i.item.Name() }

func newItemList(items []DiffItem, styles Styles) list.Model {
	lstItems := make([]list.Item, len(items))
	for idx := range items {
		lstItems[idx] = itemAdapter{item: items[idx]}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)

	l := list.New(lstItems, delegate, 0, 0)
	l.Title = "Items"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.DisableQuitKeybindings()
	return l
}

func setListItems(l *list.Model, items []DiffItem) {
	wrapped := make([]list.Item, len(items))
	for i := range items {
		wrapped[i] = itemAdapter{item: items[i]}
	}
	l.SetItems(wrapped)
	l.Select(0)
}
