package appshared

// IDGenerator abstracts unique ID generation for testability.
type IDGenerator interface {
	NewID() string
}
