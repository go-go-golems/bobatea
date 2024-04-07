package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/chat"
	conversation2 "github.com/go-go-golems/bobatea/pkg/conversation"
)

func main() {
	manager := conversation2.NewManager(conversation2.WithMessages(
		conversation2.NewChatMessage(conversation2.RoleSystem, "hahahahaha"),
	))

	backend := &FakeBackend{}
	m := chat.InitialModel(manager, backend)

	options := []tea.ProgramOption{
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		tea.WithAltScreen(),
	}
	p := tea.NewProgram(m, options...)
	backend.p = p

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
