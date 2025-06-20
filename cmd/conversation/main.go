package main

import (
	"github.com/charmbracelet/bubbletea"
	ui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
	conversation2 "github.com/go-go-golems/geppetto/pkg/conversation"
	"time"
)

// This example creates a new conversation manager, adds some initial messages, initializes the UI model with the manager,
// and starts the Bubble Tea program. It also demonstrates how to simulate streaming messages using a goroutine,
// sending `StreamStartMsg`, `StreamCompletionMsg`, and `StreamDoneMsg` to update the UI. The messages are sent through the
// Bubble Tea scheduler using the `p.Send` method after a delay to allow the UI to initialize first.

type model struct {
	ui ui.Model
}

func (m model) Init() tea.Cmd {
	return m.ui.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.ui, cmd = m.ui.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.ui.View()
}

var _ tea.Model = model{}

func main() {
	manager := conversation2.NewManager()

	// Add some initial messages to the manager
	msg1 := conversation2.NewChatMessage(conversation2.RoleUser, "Hello!")
	msg2 := conversation2.NewChatMessage(conversation2.RoleAssistant, "Hi there! Streaming will start soon...")
	if err := manager.AppendMessages(msg1, msg2); err != nil {
		panic(err)
	}

	m := model{
		ui: ui.NewModel(manager),
	}

	p := tea.NewProgram(m)

	// Start a goroutine to simulate streaming messages after a delay
	go func() {
		time.Sleep(1 * time.Second)

		startMsg := ui.StreamStartMsg{
			StreamMetadata: ui.StreamMetadata{
				ID:       conversation2.NewNodeID(),
				ParentID: msg2.ID,
			},
		}
		p.Send(startMsg)

		time.Sleep(500 * time.Millisecond)

		completionMsg1 := ui.StreamCompletionMsg{
			StreamMetadata: startMsg.StreamMetadata,
			Delta:          "This is the first part of the completion...",
			Completion:     "This is the first part of the completion...",
		}
		p.Send(completionMsg1)

		time.Sleep(500 * time.Millisecond)

		completionMsg2 := ui.StreamCompletionMsg{
			StreamMetadata: startMsg.StreamMetadata,
			Delta:          " and this is the second part.",
			Completion:     "This is the first part of the completion... and this is the second part.",
		}
		p.Send(completionMsg2)

		time.Sleep(500 * time.Millisecond)

		doneMsg := ui.StreamDoneMsg{
			StreamMetadata: startMsg.StreamMetadata,
			Completion:     completionMsg2.Completion,
		}
		p.Send(doneMsg)

		// send bubble tea quit message
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
