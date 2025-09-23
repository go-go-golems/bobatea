package diff

// ChangeStatus represents the type of change in a diff line.
type ChangeStatus string

const (
	// ChangeStatusAdded indicates a value was added.
	ChangeStatusAdded ChangeStatus = "added"
	// ChangeStatusUpdated indicates a value was updated.
	ChangeStatusUpdated ChangeStatus = "updated"
	// ChangeStatusRemoved indicates a value was removed.
	ChangeStatusRemoved ChangeStatus = "removed"
)

// DataProvider supplies items to the diff model.
// Implementations can adapt any domain-specific structures into DiffItem.
type DataProvider interface {
	Title() string
	Items() []DiffItem
}

// DiffItem represents a logical entity with grouped changes.
type DiffItem interface {
	ID() string
	Name() string
	Categories() []Category
}

// Category groups changes under a named section.
type Category interface {
	Name() string
	Changes() []Change
}

// Change represents a single change, including before/after values.
type Change interface {
	Path() string
	Status() ChangeStatus
	Before() any
	After() any
	Sensitive() bool
}


