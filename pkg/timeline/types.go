package timeline

import "time"

// EntityID identifies a UI entity in the timeline.
type EntityID struct {
	RunID   string `json:"run_id,omitempty"`
	TurnID  string `json:"turn_id,omitempty"`
	BlockID string `json:"block_id,omitempty"`
	LocalID string `json:"local_id,omitempty"`
	Kind    string `json:"kind"`
}

// RendererDescriptor selects a renderer implementation on the UI side.
type RendererDescriptor struct {
	Key  string `json:"key"`
	Kind string `json:"kind"`
}

// UIEntityCreated is emitted when an entity is created/started.
type UIEntityCreated struct {
	ID        EntityID           `json:"id"`
	Renderer  RendererDescriptor `json:"renderer"`
	Props     map[string]any     `json:"props"`
	StartedAt time.Time          `json:"started_at"`
	Labels    map[string]string  `json:"labels,omitempty"`
}

// UIEntityUpdated streams updates to an existing entity.
type UIEntityUpdated struct {
	ID        EntityID       `json:"id"`
	Patch     map[string]any `json:"patch"`
	Version   int64          `json:"version"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// UIEntityCompleted finalizes the entity state.
type UIEntityCompleted struct {
	ID     EntityID       `json:"id"`
	Result map[string]any `json:"result,omitempty"`
}

// UIEntityDeleted removes an entity from the timeline.
type UIEntityDeleted struct {
	ID EntityID `json:"id"`
}

// Messages sent to interactive EntityModels via their Update method
type EntitySelectedMsg struct{ ID EntityID }
type EntityUnselectedMsg struct{ ID EntityID }
type EntityPropsUpdatedMsg struct {
	ID    EntityID
	Patch map[string]any
}

// Size and focus messages
type EntitySetSizeMsg struct{ Width, Height int }
type EntityFocusMsg struct{ ID EntityID }
type EntityBlurMsg struct{ ID EntityID }

// Actions that parent can request from models
type EntityCopyTextMsg struct{ ID EntityID }
type EntityCopyCodeMsg struct{ ID EntityID }

// Messages emitted by models for side-effects handled by parent
type CopyTextRequestedMsg struct{ Text string }
type CopyCodeRequestedMsg struct{ Code string }

// Entity lifecycle messages as Bubble Tea messages (optional alternative to direct calls)
type EntityCreatedMsg struct {
	ID        EntityID
	Renderer  RendererDescriptor
	Props     map[string]any
	StartedAt time.Time
}
type EntityCompletedMsg struct {
	ID     EntityID
	Result map[string]any
}

// Removed legacy LLMMeta and local LLMInferenceData.
