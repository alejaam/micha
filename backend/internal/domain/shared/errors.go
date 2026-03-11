package shared

import "errors"

var (
	ErrInvalidMoney = errors.New("invalid money amount")
	ErrNotFound     = errors.New("resource not found")
	// ErrAlreadyDeleted is returned when an expense has already been soft-deleted.
	ErrAlreadyDeleted = errors.New("expense already deleted")
	// ErrAlreadyExists is returned when a resource already exists with the same unique identifier.
	ErrAlreadyExists = errors.New("already exists")
)
