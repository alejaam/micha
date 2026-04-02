package shared

import "errors"

var (
	ErrInvalidMoney = errors.New("invalid money amount")
	ErrNotFound     = errors.New("resource not found")
	// ErrForbidden is returned when a caller is authenticated but not allowed to perform the action.
	ErrForbidden = errors.New("forbidden")
	// ErrAlreadyDeleted is returned when an expense has already been soft-deleted.
	ErrAlreadyDeleted = errors.New("expense already deleted")
	// ErrAlreadyExists is returned when a resource already exists with the same unique identifier.
	ErrAlreadyExists = errors.New("already exists")
	// ErrInvalidCredentials is returned when the email or password is incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidID is returned when an entity ID is empty or invalid.
	ErrInvalidID = errors.New("invalid id")
	// ErrInvalidName is returned when a name field is empty or invalid.
	ErrInvalidName = errors.New("invalid name")
	// ErrInvalidPercentage is returned when a percentage value is outside the valid range (0-100).
	ErrInvalidPercentage = errors.New("invalid percentage")
	// ErrInvalidDateRange is returned when start date is after end date.
	ErrInvalidDateRange = errors.New("invalid date range")
	// ErrInvalidStatus is returned when a status value is not recognized.
	ErrInvalidStatus = errors.New("invalid status")
)
