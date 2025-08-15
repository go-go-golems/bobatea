package chat

import (
	"github.com/go-go-golems/geppetto/pkg/events"
	"github.com/google/uuid"
)

type StreamMetadata struct {
	ID            uuid.UUID             `json:"id" yaml:"id"`
	EventMetadata *events.EventMetadata `json:"metadata" yaml:"metadata"`
}

type StreamStartMsg struct {
	StreamMetadata
}

type StreamStatusMsg struct {
	StreamMetadata
	Text string
}

type StreamCompletionMsg struct {
	StreamMetadata
	Delta      string
	Completion string
}

type StreamDoneMsg struct {
	StreamMetadata
	Completion string
}

type StreamCompletionError struct {
	StreamMetadata
	Err error
}
