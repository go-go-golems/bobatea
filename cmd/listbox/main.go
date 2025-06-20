package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/listbox"
)

// Define a simple item implementation
type fruit struct {
	id      string
	display string
}

func (f fruit) ID() string      { return f.id }
func (f fruit) Display() string { return f.display }

// Main application model
type model struct {
	list     listbox.ListModel
	done     bool
	selected listbox.Item
}

func initialModel() model {
	// Create a list with 10 visible items
	list := listbox.New(10)

	return model{
		list: list,
		done: false,
	}
}

func (m model) Init() tea.Cmd {
	// Set initial items
	return m.list.SetItems(getItems())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case listbox.SelectedMsg:
		// Handle item selection
		m.done = true
		m.selected = msg.Item
		return m, nil
	}

	// Delegate update to the list component
	newList, cmd := m.list.Update(msg)
	m.list = newList.(listbox.ListModel)
	return m, cmd
}

func (m model) View() string {
	if m.done {
		return fmt.Sprintf(
			"You selected: ID=%s, Display=%s\n\nPress q to quit.",
			m.selected.ID(),
			m.selected.Display(),
		)
	}

	return "Listbox Demo\n\n" +
		"Use up/down arrows to navigate\n" +
		"Press enter to select an item\n" +
		"Press q to quit\n\n" +
		m.list.View()
}

func getItems() []listbox.Item {
	return []listbox.Item{
		fruit{id: "1", display: "Apple - A delicious red fruit that keeps the doctor away"},
		fruit{id: "2", display: "Banana - Yellow, curved fruit rich in potassium"},
		fruit{id: "3", display: "Cherry - Small red fruit with a pit inside"},
		fruit{id: "4", display: "Dragon Fruit - Exotic tropical fruit with scaly outer skin"},
		fruit{id: "5", display: "Elderberry - Dark purple berry used in syrups and jams"},
		fruit{id: "6", display: "Fig - Sweet, pear-shaped fruit with a soft skin"},
		fruit{id: "7", display: "Grape - Small, sweet, juicy berries that grow in clusters"},
		fruit{id: "8", display: "Honeydew - Light green melon with a sweet, refreshing taste"},
		fruit{id: "9", display: "Kiwi - Small fuzzy brown fruit with green flesh inside"},
		fruit{id: "10", display: "Lemon - Sour yellow citrus fruit used for flavoring"},
		fruit{id: "11", display: "Mango - Sweet tropical fruit with orange flesh"},
		fruit{id: "12", display: "Nectarine - Sweet stone fruit similar to a peach but without fuzz"},
		fruit{id: "13", display: "Orange - Citrus fruit known for its vitamin C content"},
		fruit{id: "14", display: "Papaya - Tropical fruit with orange flesh and black seeds"},
		fruit{id: "15", display: "Quince - Hard, aromatic fruit used in preserves and jellies"},
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
