package commandpalette

// ExecutedMsg is sent when a command is executed
type ExecutedMsg struct {
	Command string
	Data    interface{}
}
