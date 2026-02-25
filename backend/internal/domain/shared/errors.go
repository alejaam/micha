package shared

import "errors"

var (
	ErrInvalidMoney   = errors.New("invalid money amount")
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyDeleted = errors.New("expense already deleted")
)
