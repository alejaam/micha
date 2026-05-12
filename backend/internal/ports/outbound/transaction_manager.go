package outbound

import "context"

// TransactionManager defines the contract for running operations within a database transaction.
type TransactionManager interface {
	// Run executes the given function within a database transaction.
	// The transaction is committed if fn returns nil, rolled back otherwise.
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
